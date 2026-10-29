package category

import (
	"context"
	"errors"
	"gocms/internal/domain"
	"gocms/internal/flow"
	"gocms/internal/outbox"
	"time"

	"gorm.io/gorm"
)

type Record struct {
	Slug    string `gorm:"size:255;primaryKey"`
	Hidden  bool   ``
	Navbar  bool   ``
	Version uint   `gorm:"default:1"`
}

type RecordTranslation struct {
	Locale       string `gorm:"size:8;primaryKey;foreignKey:locales(code);constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Name         string ``
	CategorySlug string `gorm:"size:255;primaryKey"`
	Category     Record `gorm:"foreignKey:CategorySlug;references:Slug;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Version      uint   `gorm:"default:1"`
}

type RecordPatch = map[string]interface{}
type TranslationPatch = map[string]interface{}

type GORMRepository struct {
	db      *gorm.DB
	retries uint
	backoff time.Duration
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{
		db:      db,
		retries: uint(5),
		backoff: 100 * time.Millisecond,
	}
}

func (r *GORMRepository) Create(ctx context.Context, category *domain.Category) error {
	record := ToRecord(category)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		if err := r.writeMessage(tx, domain.CategoryActionCreate, string(category.Slug)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *GORMRepository) Update(ctx context.Context, slug string, locale string, diff *domain.CategoryDiff) error {
	return r.mapErr(flow.DoWithRetry(ctx, r.retries, r.backoff, func() error {
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			recordUpdate, translationUpdate := patch(diff)
			if len(recordUpdate)+len(translationUpdate) == 0 {
				return nil
			}
			var version struct {
				Record      uint
				Translation uint
			}
			err := tx.Table("categories AS c").
				Select("c.version AS record, COALESCE(ct.version, 0) AS translation").
				Joins("LEFT JOIN category_translations ct ON ct.category_slug = c.slug AND ct.locale = ?", locale).
				Where("c.slug = ?", slug).
				Take(&version).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return domain.ErrNotFound
				}
				return err
			}

			if len(recordUpdate) > 0 {
				result := tx.Model(&Record{}).
					Where("slug = ? AND version = ?", slug, version.Record).
					Updates(recordUpdate)
				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected == 0 {
					return domain.ErrRetry
				}
			}

			if len(translationUpdate) > 0 {
				if version.Translation == 0 {
					translationUpdate["category_slug"] = slug
					translationUpdate["locale"] = locale
					result := tx.Model(&RecordTranslation{}).
						Create(recordUpdate)
					if result.Error != nil {
						if errors.Is(err, gorm.ErrDuplicatedKey) {
							return domain.ErrRetry
						}
						return result.Error
					}
				} else {
					result := tx.Model(&RecordTranslation{}).
						Where("category_slug = ? AND locale = ? AND version = ?", slug, locale, version).
						Updates(recordUpdate)
					if result.Error != nil {
						return result.Error
					}
					if result.RowsAffected == 0 {
						return domain.ErrRetry
					}
				}
			}
			if err := r.writeMessage(tx, domain.CategoryActionUpdate, slug); err != nil {
				return err
			}
			return nil
		})
	}))
}

func (r *GORMRepository) Get(ctx context.Context, slug string, locale string) (*domain.Category, error) {
	var record RecordTranslation
	result := r.db.WithContext(ctx).Model(&RecordTranslation{}).
		Preload("Category").
		Where("category_slug = ? AND locale = ?", slug, locale).
		Take(&record)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, result.Error
	}

	val := record.ToDomain()
	return &val, nil
}

func (r *GORMRepository) GetTranslations(ctx context.Context, slug string) ([]domain.Category, error) {
	var categories []RecordTranslation

	result := r.db.WithContext(ctx).Model(&RecordTranslation{}).
		Preload("Category").
		Where("category_slug = ?", slug).Find(&categories)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, result.Error
	}

	vals := make([]domain.Category, 0, len(categories))

	for _, user := range categories {
		vals = append(vals, user.ToDomain())
	}

	return vals, nil
}

func (r *GORMRepository) GetNotHidden(ctx context.Context, locale string) ([]domain.Category, error) {
	var categories []RecordTranslation

	result := r.db.WithContext(ctx).Model(&RecordTranslation{}).
		Preload("Category").
		Where("locale = ? AND hidden = false", locale).
		Find(&categories)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, result.Error
	}

	vals := make([]domain.Category, 0, len(categories))

	for _, user := range categories {
		vals = append(vals, user.ToDomain())
	}

	return vals, nil
}

func (r *GORMRepository) Delete(ctx context.Context, slug string) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("slug = ?", slug).
			Delete(&Record{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		if err := r.writeMessage(tx, domain.CategoryActionDelete, slug); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *GORMRepository) mapErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrRetry):
		return domain.ErrLock
	case errors.Is(err, gorm.ErrRecordNotFound):
		return domain.ErrNotFound
	default:
		return err
	}
}

func (r *GORMRepository) writeMessage(tx *gorm.DB, action int, slug string) error {
	payload := domain.CategoryEvent{
		Action: action,
		Slug:   slug,
	}
	if err := outbox.WriteMessage(tx, domain.CategoryEventName, "Category.Repository", &payload); err != nil {
		return err
	}
	return nil
}

func (Record) TableName() string {
	return "categories"
}

func (RecordTranslation) TableName() string {
	return "category_translations"
}

func patch(diff *domain.CategoryDiff) (RecordPatch, TranslationPatch) {
	record := make(map[string]interface{}, 0)
	translation := make(map[string]interface{}, 0)
	if diff.Slug != nil {
		record["slug"] = *diff.Slug
	}
	if diff.Hidden != nil {
		record["hidden"] = *diff.Hidden
	}
	if diff.Navbar != nil {
		record["navbar"] = *diff.Navbar
	}

	if diff.Name != nil {
		translation["name"] = *diff.Name
	}

	if len(record) > 0 {
		record["version"] = gorm.Expr("version + 1")
	}
	if len(translation) > 0 {
		translation["version"] = gorm.Expr("version + 1")
	}

	return RecordPatch(record), TranslationPatch(translation)
}

func (cm *RecordTranslation) ToDomain() domain.Category {
	var cd domain.Category

	cd.Slug = cm.Category.Slug
	cd.Hidden = cm.Category.Hidden
	cd.Navbar = cm.Category.Navbar
	cd.Name = cm.Name
	cd.Locale = cm.Locale
	return cd
}

func ToRecord(cd *domain.Category) RecordTranslation {
	var cm RecordTranslation

	cm.CategorySlug = cd.Slug
	cm.Name = cd.Name
	cm.Locale = cd.Locale
	cm.Category = Record{
		Slug:   cd.Slug,
		Hidden: cd.Hidden,
		Navbar: cd.Navbar,
	}
	return cm
}
