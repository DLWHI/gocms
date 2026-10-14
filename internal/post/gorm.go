package post

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

func postDiff(post *Model) map[string]interface{} {
	toUpdate := make(map[string]interface{})

	if post.Slug != "" {
		toUpdate["slug"] = post.Slug
	}
	if post.Preview != "" {
		toUpdate["preview"] = post.Preview
	}

	toUpdate["published_at"] = post.PublishedAt
	toUpdate["version"] = gorm.Expr("version + 1")

	return toUpdate
}

func translationDiff(post *Translation) map[string]interface{} {
	toUpdate := make(map[string]interface{})

	if post.Title != "" {
		toUpdate["title"] = post.Title
	}
	if post.Description != "" {
		toUpdate["description"] = post.Description
	}
	if post.Keywords != "" {
		toUpdate["keywords"] = post.Keywords
	}
	if post.Media != "" {
		toUpdate["media"] = post.Media
	}
	if post.MediaSource != "" {
		toUpdate["media_source"] = post.MediaSource
	}
	if post.MediaType != "" {
		toUpdate["media_type"] = post.MediaType
	}

	if len(toUpdate) > 0 {
		toUpdate["version"] = gorm.Expr("version + 1")
	}

	return toUpdate
}

func (r *GORMRepository) Create(ctx context.Context, post *Translation) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *GORMRepository) Update(ctx context.Context, diff *Translation) error {
	var version struct {
		Post        uint
		Translation uint
	}
	for i := 0; i < 5; i++ {
		err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			pdiff := postDiff(&diff.Post)
			ptdiff := translationDiff(diff)
			err := tx.Table("posts AS p").
				Select("p.version as post, pt.version as translation").
				Joins("LEFT JOIN post_translations AS pt ON pt.post_id = p.id AND pt.locale = ?", diff.Locale).
				Where("p.id = ?", diff.PostID).
				Take(&version).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errs.ErrNotFound
				}
				return err
			}

			result := tx.Model(&Model{}).
				Where("id = ? AND version = ?", diff.PostID, version.Post).
				Updates(pdiff)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errs.ErrLock
			}

			if err := tx.Model(&diff.Post).Association("Categories").Replace(&diff.Post.Categories); err != nil {
				return err
			}

			if len(ptdiff) > 0 {
				if version.Translation == 0 {
					diff.Version = 1
					if err := tx.Create(diff).Error; err != nil {
						return err
					}
				} else {
					result := tx.Model(&Translation{}).
						Where("post_id = ? AND locale = ? AND version = ?", diff.PostID, diff.Locale, version.Translation).
						Updates(ptdiff)
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

func (r *GORMRepository) UpdateViews(ctx context.Context, id uint, value uint) error {
	var version uint
	for i := 0; i < 5; i++ {
		err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Table("posts").Select("version").Where("id = ?", id).Take(&version).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errs.ErrNotFound
				}
				return err
			}

			result := tx.Model(&Model{}).
				Where("id = ? AND version = ?", id, version).
				Updates(map[string]any{
					"views":   value,
					"version": gorm.Expr("version + 1"),
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errs.ErrLock
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

func (r *GORMRepository) IncrementViews(ctx context.Context, id uint, value uint) error {
	var views uint
	for i := 0; i < 5; i++ {
		err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Table("posts").Select("views").Where("id = ?", id).Take(&views).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errs.ErrNotFound
				}
				return err
			}

			result := tx.Model(&Model{}).
				Where("id = ? AND views = ?", id, views).
				Update("views", gorm.Expr("views + ?", value))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errs.ErrLock
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

func (r *GORMRepository) Get(ctx context.Context, id uint, locale string) (*Translation, error) {
	var res struct {
		Model
		Title       string
		Media       string
		MediaSource string
		MediaType   string
		Description string
		Keywords    string
		Locale      string
	}

	err := r.db.WithContext(ctx).
		Table("posts AS p").
		Select("*").
		Joins("LEFT JOIN post_translations AS pt ON pt.post_id = p.id AND pt.locale = ?", locale).
		Where("p.id = ?", id).
		Take(&res).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}
	res.Model.Version = 0
	return &Translation{
		Title:       res.Title,
		Description: res.Description,
		Keywords:    res.Keywords,
		Locale:      locale,
		Media:       res.Media,
		MediaSource: res.MediaSource,
		MediaType:   res.MediaType,
		PostID:      res.ID,
		Post:        res.Model,
	}, nil
}

func (r *GORMRepository) GetTranslation(ctx context.Context, id uint, locale string) (*Translation, error) {
	var res struct {
		Model
		Title       string
		Media       string
		MediaSource string
		MediaType   string
		Description string
		Keywords    string
		Locale      string
	}

	err := r.db.WithContext(ctx).
		Table("posts AS p").
		Select("*").
		Joins("LEFT JOIN post_translations AS pt ON pt.post_id = p.id").
		Where("p.id = ? AND pt.locale = ?", id, locale).
		Take(&res).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}
	res.Model.Version = 0
	return &Translation{
		Title:       res.Title,
		Description: res.Description,
		Keywords:    res.Keywords,
		Locale:      locale,
		Media:       res.Media,
		MediaSource: res.MediaSource,
		MediaType:   res.MediaType,
		PostID:      res.ID,
		Post:        res.Model,
	}, nil
}

func (r *GORMRepository) GetCategories(ctx context.Context, id uint) ([]string, error) {
	categories := make([]string, 0)

	err := r.db.WithContext(ctx).
		Table("categories").
		Select("categories.slug").
		Joins("JOIN post_categories ON post_categories.category_slug = categories.slug").
		Where("post_categories.post_id = ?", id).
		Pluck("categories.slug", &categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *GORMRepository) GetViews(ctx context.Context, id uint) (uint, error) {
	var views uint
	result := r.db.WithContext(ctx).Model(&Model{}).
		Select("Views").
		Where("id = ?", id).
		Take(&views)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return 0, errs.ErrNotFound
		}
		return 0, result.Error
	}
	return views, nil
}

func (r *GORMRepository) AuthorOf(ctx context.Context, id uint) (string, error) {
	var author string
	result := r.db.WithContext(ctx).Model(&Model{}).
		Select("Author").
		Where("id = ?", id).
		Take(&author)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", errs.ErrNotFound
		}
		return "", result.Error
	}
	return author, nil
}

func (r *GORMRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&Model{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (r *GORMRepository) DeleteTranslation(ctx context.Context, id uint, locale string) error {
	result := r.db.WithContext(ctx).
		Where("post_id = ? AND locale = ?", id, locale).
		Delete(&Translation{})
	if result.RowsAffected == 0 {
		return errs.ErrNotFound
	}
	return result.Error
}
