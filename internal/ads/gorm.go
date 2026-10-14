package ads

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

func adDiff(ad *Model) map[string]interface{} {
	toUpdate := make(map[string]interface{})

	if ad.Name != "" {
		toUpdate["name"] = ad.Name
	}
	if ad.Description != "" {
		toUpdate["description"] = ad.Description
	}
	if ad.URL != "" {
		toUpdate["url"] = ad.URL
	}
	if ad.Priority != 0 {
		toUpdate["priority"] = ad.Priority
	}
	if ad.Group != "" {
		toUpdate["group"] = ad.Group
	}
	if ad.Banner != "" {
		toUpdate["banner"] = ad.Banner
	}

	toUpdate["expires_at"] = ad.ExpiresAt
	toUpdate["version"] = gorm.Expr("version + 1")

	return toUpdate
}

func (r *GORMRepository) Create(ctx context.Context, ad *Model) error {
	return r.db.WithContext(ctx).Create(ad).Error
}

func (r *GORMRepository) Update(ctx context.Context, diff *Model) error {
	var version uint
	for i := 0; i < 5; i++ {
		adiff := adDiff(diff)
		if err := r.db.WithContext(ctx).
			Table("advertisements").
			Select("version").
			Where("id = ?", diff.ID).
			Take(&version).Error; err != nil {
			return err
		}

		result := r.db.WithContext(ctx).
			Model(&Model{}).
			Where("id = ? AND version = ?", diff.ID, version).
			Updates(adiff)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected > 0 {
			return nil
		}
	}
	return errs.ErrLock
}

func (r *GORMRepository) Get(ctx context.Context, id uint) (*Model, error) {
	var ad Model

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Take(&ad).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	return &ad, nil
}

func (r *GORMRepository) GetGroup(ctx context.Context, group string) ([]Model, error) {
	ads := make([]Model, 0)

	res := r.db.WithContext(ctx).
		Where("advertisements.group = ?", group).
		Find(&ads)
	if res.Error != nil {
		return nil, res.Error
	}

	return ads, nil
}

func (r *GORMRepository) GetGroups(ctx context.Context) ([]string, error) {
	groups := make([]string, 0)

	res := r.db.WithContext(ctx).
		Model(&Model{}).
		Select("group").
		Group("group").
		Find(&groups)
	if res.Error != nil {
		return nil, res.Error
	}

	return groups, nil
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

func (r *GORMRepository) DeleteGroup(ctx context.Context, group string) error {
	result := r.db.WithContext(ctx).
		Where("group = ?", group).
		Delete(&Model{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errs.ErrNotFound
	}
	return nil
}
