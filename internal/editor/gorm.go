package editor

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
	Slug      string         `gorm:"primaryKey;size:254;"`
	Owner     string         ``
	DeletedAt gorm.DeletedAt ``
	Version   uint           `gorm:"default:1"`
}

type RecordTranslation struct {
	Lang       string `gorm:"size:8;primaryKey;foreignKey:languages(code);constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Name       string ``
	Bio        string ``
	EditorSlug string `gorm:"size:255;primaryKey"`
	Editor     Record `gorm:"foreignKey:EditorSlug;references:Slug;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Version    uint   `gorm:"default:1"`
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

func (r *GORMRepository) Create(ctx context.Context, editor *domain.Editor) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ToRecord(editor)).Error; err != nil {
			return err
		}
		if err := r.writeMessage(tx, domain.EditorActionCreate, editor.Slug); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrDuplicate
		}
		return err
	}
	return nil
}

func (r *GORMRepository) Update(ctx context.Context, slug string, lang string, diff *domain.EditorDiff) error {
	return flow.DoWithRetry(ctx, r.retries, r.backoff, func() error {
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			recordUpdate, translationUpdate := r.patch(diff)
			if len(recordUpdate)+len(translationUpdate) == 0 {
				return nil
			}
			var version struct {
				Record      uint
				Translation uint
			}
			err := tx.Table("editors AS e").
				Select("e.version AS record, COALESCE(et.version, 0) AS translation").
				Joins("LEFT JOIN editor_translations et ON et.editor_slug = e.slug AND et.lang = ?", lang).
				Where("e.slug = ?", slug).
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
					translationUpdate["editor_slug"] = slug
					translationUpdate["lang"] = lang
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
						Where("editor_slug = ? AND lang = ? AND version = ?", slug, lang, version).
						Updates(recordUpdate)
					if result.Error != nil {
						return result.Error
					}
					if result.RowsAffected == 0 {
						return domain.ErrRetry
					}
				}
			}
			if err := r.writeMessage(tx, domain.EditorActionUpdate, slug); err != nil {
				return err
			}
			return nil
		})
	})
}

func (r *GORMRepository) GetAny(ctx context.Context, slug string, lang string) (*domain.Editor, error) {
	var record RecordTranslation
	result := r.db.WithContext(ctx).Model(&RecordTranslation{}).
		Unscoped().
		Preload("Editor").
		Where("editor_slug = ? AND lang = ?", slug, lang).
		Take(&record)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, result.Error
	}

	return record.ToDomain(), nil
}

func (r *GORMRepository) GetTranslations(ctx context.Context, slug string) ([]domain.Editor, error) {
	var editors []RecordTranslation

	result := r.db.WithContext(ctx).
		Model(&RecordTranslation{}).
		Preload("Editor").
		Where("editor_slug = ?", slug).
		Find(&editors)
	if result.Error != nil {
		return nil, result.Error
	}

	vals := make([]domain.Editor, 0, len(editors))

	for _, editor := range editors {
		vals = append(vals, *editor.ToDomain())
	}

	return vals, nil
}

func (r *GORMRepository) GetAll(ctx context.Context, lang string) ([]domain.Editor, error) {
	var editors []RecordTranslation

	result := r.db.WithContext(ctx).
		Model(&RecordTranslation{}).
		Joins("Editor").
		Where("lang = ?", lang).
		Find(&editors)
	if result.Error != nil {
		return nil, result.Error
	}

	vals := make([]domain.Editor, 0, len(editors))

	for _, editor := range editors {
		vals = append(vals, *editor.ToDomain())
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
		if err := r.writeMessage(tx, domain.EditorActionDelete, slug); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (*GORMRepository) writeMessage(tx *gorm.DB, action int, slug string) error {
	payload := domain.EditorEvent{
		Slug:   slug,
		Action: action,
	}
	if err := outbox.WriteMessage(tx, domain.EditorEventName, "Editor.Repository", &payload); err != nil {
		return err
	}
	return nil
}

func (*GORMRepository) patch(diff *domain.EditorDiff) (RecordPatch, TranslationPatch) {
	record := make(map[string]interface{}, 0)
	translation := make(map[string]interface{}, 0)
	if diff.Slug != nil {
		record["slug"] = *diff.Slug
	}
	if diff.Owner != nil {
		record["owner"] = *diff.Owner
	}

	if diff.Name != nil {
		translation["name"] = *diff.Name
	}
	if diff.Bio != nil {
		translation["bio"] = *diff.Bio
	}

	if len(record) > 0 {
		record["version"] = gorm.Expr("version + 1")
	}
	if len(translation) > 0 {
		translation["version"] = gorm.Expr("version + 1")
	}

	return RecordPatch(record), TranslationPatch(translation)
}

func (*Record) TableName() string {
	return "editors"
}

func (*RecordTranslation) TableName() string {
	return "editor_translations"
}

func (rd *RecordTranslation) ToDomain() *domain.Editor {
	var res domain.Editor

	if rd.Editor.DeletedAt.Valid {
		ts := rd.Editor.DeletedAt.Time.Unix()
		res.DeletedAt = &ts
	}

	res.Owner = rd.Editor.Owner
	res.Slug = rd.Editor.Slug
	res.Name = rd.Name
	res.Lang = rd.Lang
	res.Bio = rd.Bio
	return &res
}

func ToRecord(data *domain.Editor) *RecordTranslation {
	var res RecordTranslation

	res.Editor.Slug = data.Slug
	res.EditorSlug = data.Slug
	res.Editor.Owner = data.Owner
	res.Lang = data.Lang
	res.Name = data.Name
	res.Bio = data.Bio
	return &res
}
