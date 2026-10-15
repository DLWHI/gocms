package user

import (
	"context"
	"errors"
	"gocms/internal/errs"
	"gocms/internal/outbox"
	"time"

	"gorm.io/gorm"
)

type GORMRepository struct {
	db *gorm.DB
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func (r *GORMRepository) Create(ctx context.Context, user *Model) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		if err := writeMessage(tx, "create", user.Username); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *GORMRepository) UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*Model, error) {
	var user Model
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Passwd").Where("username = ?", username).Take(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errs.ErrNotFound
			}
			return err
		}

		query := tx.Model(&Model{}).Where("username = ?", user.Username)

		if user.LastLogin == nil {
			query.Where("last_login is NULL")
		} else {
			query.Where("last_login = ?", user.LastLogin)
		}

		result := query.Update("LastLogin", timestamp)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errs.ErrLock
		}
		user.LastLogin = &timestamp
		if err := writeMessage(tx, "update", username); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *GORMRepository) UpdatePermissions(ctx context.Context, username string, permissions uint32) (*Model, error) {
	var user Model
	for i := 0; i < 5; i++ {
		err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Omit("Passwd").Where("username = ?", username).Take(&user).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errs.ErrNotFound
				}
				return err
			}

			if permissions == user.Permissions {
				return nil
			}

			result := tx.Model(&Model{}).
				Where("username = ? AND version = ?", user.Username, user.Version).
				Updates(map[string]any{
					"permissions": permissions,
					"version":     gorm.Expr("version + 1"),
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errs.ErrLock
			}
			if err := writeMessage(tx, "update", username); err != nil {
				return err
			}
			return nil
		})
		if err == nil {
			return &user, nil
		} else if !errors.Is(err, errs.ErrLock) {
			return nil, err
		}
	}
	return nil, errs.ErrLock
}

func (r *GORMRepository) UpdateEmail(ctx context.Context, username string, email string) error {
	var user Model
	for i := 0; i < 5; i++ {
		err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Omit("Passwd").Where("username = ?", username).Take(&user).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errs.ErrNotFound
				}
				return err
			}

			if email == user.Email {
				return nil
			}

			result := tx.Model(&user).Model(&Model{}).
				Where("username = ? AND version = ?", user.Username, user.Version).
				Updates(map[string]any{
					"email":   email,
					"version": gorm.Expr("version + 1"),
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errs.ErrLock
			}
			if err := writeMessage(tx, "update", username); err != nil {
				return err
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

func (r *GORMRepository) UpdatePasswd(ctx context.Context, username string, pwdhash Bytes) error {
	var user Model
	for i := 0; i < 5; i++ {
		err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Omit("Passwd").Where("username = ?", username).Take(&user).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errs.ErrNotFound
				}
				return err
			}

			result := tx.Model(&user).Model(&Model{}).
				Where("username = ? AND version = ?", user.Username, user.Version).
				Updates(map[string]any{
					"passwd":  pwdhash,
					"version": gorm.Expr("version + 1"),
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return errs.ErrLock
			}
			if err := writeMessage(tx, "update", username); err != nil {
				return err
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

func (r *GORMRepository) Get(ctx context.Context, username string) (*Model, error) {
	var user Model

	err := r.db.WithContext(ctx).
		Unscoped().
		Omit("Passwd").
		Where("username = ?", username).
		Take(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *GORMRepository) GetAll(ctx context.Context) ([]Model, error) {
	users := make([]Model, 0)

	err := r.db.WithContext(ctx).Omit("Passwd").Find(&users).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	return users, nil
}

func (r *GORMRepository) GetPasswdHash(ctx context.Context, username string) (Bytes, error) {
	var hash Bytes

	err := r.db.WithContext(ctx).Model(&Model{}).
		Select("Passwd").
		Where("username = ?", username).
		Take(&hash).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	return hash, nil
}

func (r *GORMRepository) Delete(ctx context.Context, username string) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("username = ?", username).
			Delete(&Model{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errs.ErrNotFound
		}
		if err := writeMessage(tx, "delete", username); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func writeMessage(tx *gorm.DB, action string, username string) error {
	payload := MessagePayload{
		Username: username,
		Action:   action,
	}
	if err := outbox.WriteMessage(tx, "user", "User.Repository", &payload); err != nil {
		return err
	}
	return nil
}
