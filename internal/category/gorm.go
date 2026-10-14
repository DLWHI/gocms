package category

import (
	"context"
	"errors"
	"gocms/internal/errs"

	"gorm.io/gorm"
)

type GORMRepository struct {
	db *gorm.DB
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func categoryDiff(category *Model) map[string]interface{} {
	toUpdate := make(map[string]interface{})

	toUpdate["hidden"] = category.Hidden
	toUpdate["navbar"] = category.Navbar
	toUpdate["version"] = gorm.Expr("version + 1")

	return toUpdate
}

func translationDiff(category *Translation) map[string]interface{} {
	toUpdate := make(map[string]interface{})

	if category.Name != "" {
		toUpdate["name"] = category.Name
		toUpdate["version"] = gorm.Expr("version + 1")
	}

	return toUpdate
}

func (r *GORMRepository) Create(ctx context.Context, category *Translation) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *GORMRepository) Update(ctx context.Context, diff *Translation) error {
	var version struct {
		Category    uint
		Translation uint
	}
	for i := 0; i < 5; i++ {
		err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			cdiff := categoryDiff(&diff.Category)
			ctdiff := translationDiff(diff)
			err := tx.Table("categories AS c").
				Select("c.version AS category, COALESCE(ct.version, 0) AS translation").
				Joins("LEFT JOIN category_translations ct ON ct.category_slug = c.slug AND ct.locale = ?", diff.Locale).
				Where("c.slug = ?", diff.CategorySlug).
				Take(&version).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errs.ErrNotFound
				}
				return err
			}

			cdiff["version"] = gorm.Expr("version + 1")
			result := tx.Model(&Model{}).
				Where("slug = ? AND version = ?", diff.CategorySlug, version.Category).
				Updates(cdiff)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errs.ErrLock
			}

			if len(ctdiff) > 0 {
				if version.Translation == 0 {
					diff.Version = 1
					if err := tx.Create(diff).Error; err != nil {
						return err
					}
				} else {
					ctdiff["version"] = gorm.Expr("version + 1")
					result := tx.Model(&Translation{}).
						Where("category_slug = ? AND locale = ? AND version = ?", diff.CategorySlug, diff.Locale, version.Translation).
						Updates(ctdiff)
					if result.Error != nil {
						return result.Error
					}
					if result.RowsAffected == 0 {
						return errs.ErrLock
					}
				}
			}
			return nil
		})
		if err == nil {
			return nil
		} else if !errors.Is(err, errs.ErrLock) {
			return err
		}
	}
	return errs.ErrLock
}

func (r *GORMRepository) Get(ctx context.Context, slug string, locale string) (*Translation, error) {
	var res struct {
		Model
		Name   string
		Locale string
	}

	err := r.db.WithContext(ctx).
		Table("categories AS c").
		Select("ct.name, c.slug, ct.locale, c.hidden, c.navbar").
		Joins("LEFT JOIN category_translations AS ct ON ct.category_slug = c.slug AND ct.locale = ?", locale).
		Where("c.slug = ?", slug).
		Take(&res).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	return &Translation{
		Name:         res.Name,
		Locale:       locale,
		CategorySlug: res.Slug,
		Category:     res.Model,
	}, nil
}

func (r *GORMRepository) GetAll(ctx context.Context, locale string) ([]Translation, error) {
	categories := make([]Translation, 0)

	err := r.db.WithContext(ctx).Model(&Translation{}).
		Preload("Category").
		Where("locale = ?", locale).
		Find(&categories).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	return categories, nil
}

func (r *GORMRepository) GetNotHidden(ctx context.Context, locale string) ([]Translation, error) {
	categories := make([]Translation, 0)

	err := r.db.WithContext(ctx).Model(&Translation{}).
		Preload("Category").
		Where("locale = ? AND hidden != FALSE", locale).
		Find(&categories).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	return categories, nil
}

func (r *GORMRepository) Delete(ctx context.Context, slug string) error {
	result := r.db.WithContext(ctx).
		Where("slug = ?", slug).
		Delete(&Model{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errs.ErrNotFound
	}
	return nil
}
