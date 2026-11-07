package post

import (
	"context"
	"errors"
	"gocms/internal/category"
	"gocms/internal/database"
	"gocms/internal/domain"
	"gocms/internal/editor"
	"gocms/internal/flow"
	"gocms/internal/outbox"
	"gocms/internal/user"
	"time"

	"gorm.io/gorm"
)

type Record struct {
	ID          uint              `gorm:"primaryKey;autoIncrement"`
	Slug        string            `gorm:"size:128"`
	CreatedAt   time.Time         `gorm:"default:current_timestamp"`
	PublishedAt *time.Time        ``
	Views       uint              `gorm:"not null;default:0"`
	Preview     string            ``
	Author      *string           `gorm:"size:128;"`
	Editor      editor.Record     `gorm:"foreignKey:Author;references:Slug;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	CreatedBy   string            `gorm:"size:64;"`
	User        user.Record       `gorm:"foreignKey:CreatedBy;references:Username;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Categories  []category.Record `gorm:"many2many:post_categories;foreignKey:ID;joinForeignKey:PostID;references:Slug;joinReferences:CategorySlug;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;"`
	Version     uint              `gorm:"default:1"`
}

type RecordTranslation struct {
	Title       string `gorm:"not null"`
	Description string `gorm:"not null"`
	Keywords    string `gorm:"not null"`
	Markdown    string ``
	Media       string ``
	MediaSource string ``
	MediaType   string ``
	Language    string `gorm:"size:2;primaryKey"`
	PostID      uint   `gorm:"primaryKey"`
	Post        Record `gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Version     uint   `gorm:"default:1"`
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

func (r *GORMRepository) Create(ctx context.Context, post *domain.Post) error {
	record := ToRecord(post)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		if err := r.writeMessage(tx, domain.ActionCreate, post.PostID); err != nil {
			return err
		}
		post.PostID = record.Post.ID
		return nil
	})
}

func (r *GORMRepository) Update(ctx context.Context, id uint, language string, diff *domain.PostDiff) error {
	err := flow.DoWithRetry(ctx, r.retries, r.backoff, func() error {
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			recordUpdate, translationUpdate := r.patch(diff)
			var version struct {
				Post        uint
				Translation uint
			}
			err := tx.Table("posts AS p").
				Select("p.version as post, pt.version as translation").
				Joins("INNER JOIN post_translations AS pt ON pt.post_id = p.id").
				Where("p.id = ? AND pt.language = ?", id, language).
				Take(&version).Error
			if err != nil {
				return err
			}

			result := tx.Model(&Record{}).
				Where("id = ? AND version = ?", id, version.Post).
				Updates(recordUpdate)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return domain.ErrRetry
			}

			if err := tx.Model(diff).Association("Categories").Replace(&diff.Categories); err != nil {
				return err
			}

			if version.Translation == 0 {
				translationUpdate["post_id"] = id
				translationUpdate["lang"] = language
				translationUpdate["version"] = 1
				result := tx.Model(&RecordTranslation{}).
					Create(translationUpdate)
				if result.Error != nil {
					if errors.Is(err, gorm.ErrDuplicatedKey) {
						return domain.ErrRetry
					}
					return result.Error
				}
			} else if len(translationUpdate) > 0 {
				result := tx.Model(&RecordTranslation{}).
					Where("post_id = ? AND lang = ? AND version = ?", id, language, version.Translation).
					Updates(translationUpdate)
				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected == 0 {
					return domain.ErrRetry
				}
			}

			if err := r.writeMessage(tx, domain.ActionEdit, id); err != nil {
				return err
			}

			return nil
		})
	})
	return database.MapError(err)
}

func (r *GORMRepository) SetViews(ctx context.Context, id uint, value uint) error {
	err := flow.DoWithRetry(ctx, r.retries, r.backoff, func() error {
		var views uint
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Table("posts").Select("views").Where("id = ?", id).Take(&views).Error; err != nil {
				return err
			}

			result := tx.Model(&Record{}).
				Where("id = ? AND views = ?", id, views).
				Update("views", value)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return domain.ErrRetry
			}

			if err := r.writeMessage(tx, domain.ActionEdit, id); err != nil {
				return err
			}

			return nil
		})
	})
	return database.MapError(err)
}

func (r *GORMRepository) IncrementViews(ctx context.Context, id uint, value uint) (uint, error) {
	var views uint
	err := database.MapError(flow.DoWithRetry(ctx, r.retries, r.backoff, func() error {
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Table("posts").Select("views").Where("id = ?", id).Take(&views).Error; err != nil {
				return err
			}

			result := tx.Model(&Record{}).
				Where("id = ? AND views = ?", id, views).
				Update("views", gorm.Expr("views + ?", value))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return domain.ErrRetry
			}

			if err := r.writeMessage(tx, domain.ActionEdit, id); err != nil {
				return err
			}

			return nil
		})
	}))
	if err != nil {
		return 0, err
	}
	return views + value, nil
}

func (r *GORMRepository) Get(ctx context.Context, id uint, language string) (*domain.Post, error) {
	var post RecordTranslation

	result := r.db.WithContext(ctx).Model(&RecordTranslation{}).
		Preload("Post").
		Where("post_id = ? AND lang = ?", id, language).
		Take(&post)
	if result.Error != nil {
		return nil, result.Error
	}

	return post.ToDomain(), nil
}

func (r *GORMRepository) GetTranslations(ctx context.Context, id uint) ([]domain.Post, error) {
	var posts []RecordTranslation

	result := r.db.WithContext(ctx).Model(&RecordTranslation{}).
		Preload("Categories").
		Preload("Post").
		Where("post_id = ?", id).
		Find(&posts)
	if result.Error != nil {
		return nil, result.Error
	}

	vals := make([]domain.Post, 0, len(posts))

	for _, post := range posts {
		vals = append(vals, *post.ToDomain())
	}

	return vals, nil
}

func (r *GORMRepository) GetViews(ctx context.Context, id uint) (uint, error) {
	var views uint

	result := r.db.WithContext(ctx).Model(&Record{}).
		Select("views").
		Where("id = ?", id).
		Take(&views)
	if result.Error != nil {
		return 0, result.Error
	}

	return views, nil
}

func (r *GORMRepository) GetCreator(ctx context.Context, id uint) (string, error) {
	var creator string

	result := r.db.WithContext(ctx).Model(&Record{}).
		Select("created_by").
		Where("id = ?", id).
		Take(&creator)
	if result.Error != nil {
		return "", result.Error
	}

	return creator, nil
}

func (r *GORMRepository) Delete(ctx context.Context, id uint) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("id = ?", id).
			Delete(&Record{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		if err := r.writeMessage(tx, domain.ActionDelete, id); err != nil {
			return err
		}
		return nil
	})
	return database.MapError(err)
}

func (*GORMRepository) writeMessage(tx *gorm.DB, action string, id uint) error {
	payload := domain.PostEventInfo{
		PostID: id,
		Action: action,
	}
	if err := outbox.WriteMessage(tx, domain.PostEvent, "Post.Repository", &payload); err != nil {
		return err
	}
	return nil
}

func (*GORMRepository) patch(diff *domain.PostDiff) (RecordPatch, TranslationPatch) {
	record := make(map[string]interface{})
	translation := make(map[string]interface{})

	if diff.Slug != nil {
		record["slug"] = *diff.Slug
	}
	if diff.Preview != nil {
		record["preview"] = *diff.Preview
	}

	if diff.Title != nil {
		translation["title"] = diff.Title
	}
	if diff.Description != nil {
		translation["description"] = diff.Description
	}
	if diff.Keywords != nil {
		translation["keywords"] = diff.Keywords
	}
	if diff.Media != nil {
		translation["media"] = diff.Media
	}
	if diff.MediaSource != nil {
		translation["media_source"] = diff.MediaSource
	}
	if diff.MediaType != nil {
		translation["media_type"] = diff.MediaType
	}

	if len(translation) > 0 {
		translation["version"] = gorm.Expr("version + 1")
	}

	record["author"] = diff.Author
	record["published_at"] = diff.PublishedAt
	record["version"] = gorm.Expr("version + 1")

	return record, translation
}

func (*Record) TableName() string {
	return "posts"
}

func (*RecordTranslation) TableName() string {
	return "post_translations"
}

func (r *RecordTranslation) ToDomain() *domain.Post {
	var res domain.Post

	if r.Post.PublishedAt != nil {
		ts := r.Post.PublishedAt.Unix()
		res.PublishedAt = &ts
	}

	// To make json.Marshall convert no categories to empty array
	categories := make([]string, 0, len(r.Post.Categories))
	for _, category := range r.Post.Categories {
		categories = append(categories, category.Slug)
	}
	res.Categories = categories

	res.PostID = r.PostID
	res.Title = r.Title
	res.Description = r.Description
	res.Slug = r.Post.Slug
	res.Keywords = r.Keywords
	res.Language = r.Language
	res.CreatedAt = r.Post.CreatedAt.Unix()
	res.Views = r.Post.Views
	res.Preview = r.Post.Preview
	res.CreatedBy = r.Post.CreatedBy
	res.Media = r.Media
	res.MediaType = r.MediaType
	res.MediaSource = r.MediaSource
	if r.Post.Author != nil {
		res.Author = *r.Post.Author
	}

	return &res
}

func ToRecord(data *domain.Post) *RecordTranslation {
	var res RecordTranslation
	if data.PublishedAt != nil {
		ts := time.Unix(*data.PublishedAt, 0).UTC()
		res.Post.PublishedAt = &ts
	}

	if data.Author != "" {
		author := data.Author
		res.Post.Author = &author
	}

	// To make json.Marshall convert no categories to empty array
	categories := make([]category.Record, 0, len(data.Categories))
	for _, categ := range data.Categories {
		categories = append(categories, category.Record{Slug: categ})
	}
	res.Post.Categories = categories

	res.PostID = data.PostID
	res.Title = data.Title
	res.Description = data.Description
	res.Keywords = data.Keywords
	res.Language = data.Language
	res.Post.ID = data.PostID
	res.Post.Slug = data.Slug
	res.Post.Views = data.Views
	res.Post.Preview = data.Preview
	res.Media = data.Media
	res.MediaType = data.MediaType
	res.MediaSource = data.MediaSource
	res.Post.CreatedBy = data.CreatedBy

	return &res
}
