package user

import (
	"context"
	"errors"
	"gocms/internal/domain"
	"gocms/internal/flow"
	"gocms/internal/outbox"
	"time"

	"gorm.io/gorm"
)

type GORMRepository struct {
	db      *gorm.DB
	retries uint
	backoff time.Duration
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{
		db:      db,
		retries: uint(5),
		backoff: 600 * time.Millisecond,
	}
}

func (r *GORMRepository) sanitize(diff map[string]interface{}) map[string]interface{} {
	toUpdate := make(map[string]interface{})

	if val, ok := diff["permissions"]; ok {
		toUpdate["permissions"] = val
	}
	if val, ok := diff["email"]; ok {
		toUpdate["email"] = val
	}

	if len(toUpdate) > 0 {
		toUpdate["version"] = gorm.Expr("version + 1")
	}

	return toUpdate
}

func (r *GORMRepository) Create(ctx context.Context, user *Record) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		if err := r.writeMessage(tx, "user", user.Username); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *GORMRepository) UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*Record, error) {
	var user Record
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Record{}).Omit("Passwd").Where("username = ?", username).Take(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrNotFound
			}
			return err
		}

		result := tx.Model(&Record{}).Where("username = ?", user.Username).Update("LastLogin", timestamp)
		if result.Error != nil {
			return result.Error
		}
		user.LastLogin = &timestamp
		if err := r.writeMessage(tx, "user", username); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GORMRepository) Update(ctx context.Context, username string, diff map[string]interface{}) error {
	err := flow.DoWithRetry(ctx, r.retries, r.backoff, func() error {
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var user Record
			sanitized := r.sanitize(diff)
			if err := tx.Omit("Passwd").Where("username = ?", username).Take(&user).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return domain.ErrNotFound
				}
				return err
			}

			if len(sanitized) > 0 {
				result := tx.Model(&Record{}).
					Where("username = ? AND version = ?", user.Username, user.Version).
					Updates(sanitized)
				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected == 0 {
					return domain.ErrLock
				}
				if err := r.writeMessage(tx, "user", username); err != nil {
					return err
				}
			}
			return nil
		})
	})
	if err != nil {
		if errors.Is(err, domain.ErrRetry) {
			return domain.ErrLock
		}
		return err
	}
	return nil
}

func (r *GORMRepository) UpdatePasswd(ctx context.Context, username string, pwdhash Password) error {
	err := flow.DoWithRetry(ctx, r.retries, r.backoff, func() error {
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var version uint
			if err := tx.Model(&Record{}).Select("Version").Where("username = ?", username).Take(&version).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return domain.ErrNotFound
				}
				return err
			}

			result := tx.Model(&Record{}).
				Where("username = ? AND version = ?", username, version).
				Updates(map[string]any{
					"passwd":  pwdhash,
					"version": gorm.Expr("version + 1"),
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return domain.ErrRetry
			}
			if err := r.writeMessage(tx, "passwd", username); err != nil {
				return err
			}
			return nil
		})
	})
	if err != nil {
		if errors.Is(err, domain.ErrRetry) {
			return domain.ErrLock
		}
		return err
	}
	return nil
}

func (r *GORMRepository) Get(ctx context.Context, username string) (*Record, error) {
	var user Record

	err := r.db.WithContext(ctx).
		Unscoped().
		Omit("Passwd").
		Where("username = ?", username).
		Take(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *GORMRepository) GetActive(ctx context.Context) ([]Record, error) {
	users := make([]Record, 0)

	err := r.db.WithContext(ctx).Omit("Passwd").Find(&users).Error
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *GORMRepository) GetPasswdHash(ctx context.Context, username string) (Password, error) {
	var hash []byte

	err := r.db.WithContext(ctx).Model(&Record{}).
		Select("Passwd").
		Where("username = ?", username).
		Limit(1).
		Scan(&hash).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return hash, nil
}

func (r *GORMRepository) Suspend(ctx context.Context, username string) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("username = ?", username).
			Delete(&Record{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		if err := r.writeMessage(tx, "user", username); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *GORMRepository) writeMessage(tx *gorm.DB, event outbox.Event, username string) error {
	if err := outbox.WriteMessage(tx, event, "User.Repository", &username); err != nil {
		return err
	}
	return nil
}
