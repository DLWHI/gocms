package advertisement

import (
	"context"
	"errors"
	"gocms/internal/database"
	"gocms/internal/domain"
	"gocms/internal/flow"
	"gocms/internal/outbox"
	"time"

	"gorm.io/gorm"
)

type Record struct {
	ID          uint       `gorm:"size:255;primaryKey"`
	Name        string     `gorm:"not null"`
	Description string     ``
	URL         string     `gorm:"not null"`
	Priority    uint       `gorm:"default:0"`
	Group       string     ``
	ExpiresAt   *time.Time ``
	Banner      string     ``
	Version     uint       `gorm:"default:1"`
}

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

func (r *GORMRepository) Create(ctx context.Context, ad *domain.Advertisement) error {
	record := ToRecord(ad)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		ad.ID = record.ID
		if err := r.writeMessage(tx, domain.ActionCreate, record.ID); err != nil {
			return err
		}
		return nil
	})
	return database.MapError(err)
}

func (r *GORMRepository) Update(ctx context.Context, id uint, diff *domain.AdvertisementDiff) error {
	err := flow.DoWithRetry(ctx, r.retries, r.backoff, func() error {
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var version uint
			if err := tx.Table("advertisements").
				Select("version").
				Where("id = ?", id).
				Take(&version).Error; err != nil {
				return err
			}
			patch := r.patch(diff)

			result := tx.Model(&Record{}).
				Where("id = ? AND version = ?", id, version).
				Updates(patch)
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

func (r *GORMRepository) Get(ctx context.Context, id uint) (*domain.Advertisement, error) {
	var ad Record

	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Take(&ad).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return ad.ToDoc(), nil
}

func (r *GORMRepository) GetGroup(ctx context.Context, group string) ([]domain.Advertisement, error) {
	var ads []Record

	res := r.db.WithContext(ctx).
		Where("group = ?", group).
		Where("(expires_at IS NULL OR expires_at > now())").
		Order("priority ASC").
		Find(&ads)
	if res.Error != nil {
		return nil, res.Error
	}

	vals := make([]domain.Advertisement, 0, len(ads))

	for _, ad := range ads {
		vals = append(vals, *ad.ToDoc())
	}

	return vals, nil
}

func (r *GORMRepository) GetGroups(ctx context.Context) ([]string, error) {
	groups := make([]string, 0)

	res := r.db.WithContext(ctx).
		Model(&Record{}).
		Group("group").
		Pluck("group", &groups)
	if res.Error != nil {
		return nil, res.Error
	}

	return groups, nil
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
	payload := domain.AdvertisementEventInfo{
		ID:     id,
		Action: action,
	}
	if err := outbox.WriteMessage(tx, domain.AdvertisementEvent, "Advertisement.Repository", &payload); err != nil {
		return err
	}
	return nil
}

func (*GORMRepository) patch(diff *domain.AdvertisementDiff) map[string]interface{} {
	record := make(map[string]interface{})

	if diff.Name != nil {
		record["name"] = *diff.Name
	}
	if diff.Description != nil {
		record["description"] = *diff.Description
	}
	if diff.URL != nil {
		record["url"] = *diff.URL
	}
	if diff.Priority != nil {
		record["priority"] = *diff.Priority
	}
	if diff.Group != nil {
		record["group"] = *diff.Group
	}
	if diff.Banner != nil {
		record["banner"] = *diff.Banner
	}

	record["expires_at"] = diff.ExpiresAt
	record["version"] = gorm.Expr("version + 1")

	return record
}

func (*Record) TableName() string {
	return "advertisements"
}

func (r *Record) ToDoc() *domain.Advertisement {
	var res domain.Advertisement

	if r.ExpiresAt != nil {
		ts := r.ExpiresAt.Unix()
		res.ExpiresAt = &ts
	}

	res.ID = r.ID
	res.Name = r.Name
	res.Description = r.Description
	res.URL = r.URL
	res.Priority = r.Priority
	res.Group = r.Group
	res.Banner = r.Banner
	return &res
}

func ToRecord(data *domain.Advertisement) *Record {
	var res Record
	if data.ExpiresAt != nil {
		ts := time.Unix(*data.ExpiresAt, 0).UTC()
		res.ExpiresAt = &ts
	}

	res.ID = data.ID
	res.Name = data.Name
	res.Description = data.Description
	res.URL = data.URL
	res.Priority = data.Priority
	res.Group = data.Group
	res.Banner = data.Banner
	return &res
}
