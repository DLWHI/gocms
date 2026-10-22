package ads

import (
	"context"
	"errors"
	"gocms/internal/domain"
	"gocms/internal/outbox"

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
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(ad).Error; err != nil {
			return err
		}
		if err := writeMessage(tx, "create", ad.ID, ""); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *GORMRepository) Update(ctx context.Context, diff *Model) error {
	var version uint
	for i := 0; i < 5; i++ {
		err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			adiff := adDiff(diff)
			if err := tx.Table("advertisements").
				Select("version").
				Where("id = ?", diff.ID).
				Take(&version).Error; err != nil {
				return err
			}

			result := tx.Model(&Model{}).
				Where("id = ? AND version = ?", diff.ID, version).
				Updates(adiff)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return domain.ErrLock
			}
			if err := writeMessage(tx, "update", diff.ID, ""); err != nil {
				return err
			}
			return nil
		})
		if err == nil {
			return nil
		} else if !errors.Is(err, domain.ErrLock) {
			return err
		}
	}
	return domain.ErrLock
}

func (r *GORMRepository) Get(ctx context.Context, id uint) (*Model, error) {
	var ad Model

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Take(&ad).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
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
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("id = ?", id).
			Delete(&Model{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		if err := writeMessage(tx, "delete", id, ""); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *GORMRepository) DeleteGroup(ctx context.Context, group string) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("group = ?", group).
			Delete(&Model{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		if err := writeMessage(tx, "delete.group", 0, group); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func writeMessage(tx *gorm.DB, action string, id uint, group string) error {
	payload := MessagePayload{
		ID:     id,
		Group:  group,
		Action: action,
	}
	if err := outbox.WriteMessage(tx, "ads", "Ads.Repository", &payload); err != nil {
		return err
	}
	return nil
}
