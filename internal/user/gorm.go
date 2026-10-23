package user

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/domain"
	"gocms/internal/flow"
	"gocms/internal/outbox"
	"time"

	"gorm.io/gorm"
)

type Record struct {
	Username    string         `gorm:"primaryKey;size:64;"`
	Password    []byte         `gorm:"not null"`
	Email       string         ``
	Permissions uint32         `gorm:"default:0"` // PermCreate
	SuspendedAt gorm.DeletedAt ``
	LastLogin   *time.Time     ``
	Version     uint           `gorm:"default:0"`
}

func (Record) TableName() string {
	return domain.UserEntityName
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

func (r *GORMRepository) Create(ctx context.Context, user *domain.User, pwdhash []byte) error {
	record := ToRecord(user)
	record.Password = pwdhash
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		if err := r.writeMessage(tx, domain.UserActionCreate, user.Username); err != nil {
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

func (r *GORMRepository) Update(ctx context.Context, username string, update *domain.UserDiff) error {
	err := flow.DoWithRetry(ctx, r.retries, r.backoff, func() error {
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var user Record
			if err := tx.Omit("Password").Where("username = ?", username).Take(&user).Error; err != nil {
				return err
			}
			diff := Map(update)
			if len(diff) > 0 {
				diff["version"] = gorm.Expr("version + 1")
				result := tx.Model(&Record{}).
					Select("").
					Where("username = ? AND version = ?", username, user.Version).
					Updates(diff)
				if result.Error != nil {
					return result.Error
				}
				if result.RowsAffected == 0 {
					return domain.ErrRetry
				}
				if err := r.writeMessage(tx, domain.UserActionUpdate, username); err != nil {
					return err
				}
			}
			return nil
		})
	})
	return r.mapErr(err)
}

func (r *GORMRepository) UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*domain.User, error) {
	var user Record
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&Record{}).
			Omit("Password").
			Where("username = ?", username).
			Take(&user).Error
		if err != nil {
			return err
		}

		if user.LastLogin != nil {
			if timestamp.Before(*user.LastLogin) {
				return fmt.Errorf("%w: cannot login back in time", domain.ErrInval)
			}
		}

		result := tx.Model(&Record{}).
			Where("username = ? AND (last_login IS NULL OR last_login <= ?)", username, timestamp).
			Update("LastLogin", timestamp)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		user.LastLogin = &timestamp
		if err := r.writeMessage(tx, domain.UserActionLogin, username); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, r.mapErr(err)
	}
	u := user.ToDomain()
	return &u, nil
}

func (r *GORMRepository) UpdatePassword(ctx context.Context, username string, pwdhash []byte) error {
	if len(pwdhash) == 0 {
		return fmt.Errorf("%w: empty password", domain.ErrInval)
	}
	err := flow.DoWithRetry(ctx, r.retries, r.backoff, func() error {
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var version uint
			err := tx.Model(&Record{}).
				Select("Version").
				Where("username = ?", username).
				Take(&version).Error
			if err != nil {
				return err
			}

			result := tx.Model(&Record{}).
				Where("username = ? AND version = ?", username, version).
				Updates(map[string]any{
					"password": pwdhash,
					"version":  gorm.Expr("version + 1"),
				})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return domain.ErrRetry
			}
			if err := r.writeMessage(tx, domain.UserActionPassword, username); err != nil {
				return err
			}
			return nil
		})
	})
	return r.mapErr(err)
}

func (r *GORMRepository) Get(ctx context.Context, username string) (*domain.User, error) {
	var user Record

	err := r.db.WithContext(ctx).
		Unscoped().
		Omit("Password").
		Where("username = ?", username).
		Take(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	u := user.ToDomain()
	return &u, nil
}

func (r *GORMRepository) GetActive(ctx context.Context) ([]domain.User, error) {
	users := make([]Record, 0)

	err := r.db.WithContext(ctx).Omit("Password").Find(&users).Error
	if err != nil {
		return nil, err
	}

	vals := make([]domain.User, 0, len(users))

	for _, user := range users {
		vals = append(vals, user.ToDomain())
	}

	return vals, nil
}

func (r *GORMRepository) GetPasswordHash(ctx context.Context, username string) ([]byte, error) {
	var user Record

	err := r.db.WithContext(ctx).Model(&Record{}).
		Select("Password").
		Where("username = ?", username).
		Limit(1).
		Take(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return user.Password, nil
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
		if err := r.writeMessage(tx, domain.UserActionSuspend, username); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *GORMRepository) writeMessage(tx *gorm.DB, action int, username string) error {
	payload := domain.UserEvent{
		Username: username,
		Action:   action,
	}
	if err := outbox.WriteMessage(tx, domain.UserEventName, "User.Repository", &payload); err != nil {
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

func Map(update *domain.UserDiff) map[string]interface{} {
	diff := make(map[string]interface{})

	if update.Permissions != nil {
		diff["permissions"] = *update.Permissions
	}
	if update.Email != nil {
		diff["email"] = *update.Email
	}
	return diff
}

func (rd *Record) ToDomain() domain.User {
	var res domain.User
	res.Username = rd.Username

	if rd.SuspendedAt.Valid {
		res.SuspendedAt = rd.SuspendedAt.Time.Unix()
	}

	if rd.LastLogin != nil {
		res.LastLogin = rd.LastLogin.Unix()
	}

	res.Email = rd.Email
	res.Permissions = rd.Permissions
	return res
}

func ToRecord(data *domain.User) Record {
	var res Record

	if data.SuspendedAt != 0 {
		res.SuspendedAt = gorm.DeletedAt{
			Time:  time.Unix(data.SuspendedAt, 0).UTC(),
			Valid: true,
		}
	}

	if data.LastLogin != 0 {
		t := time.Unix(data.LastLogin, 0).UTC()
		res.LastLogin = &t
	}

	res.Username = data.Username
	res.Email = data.Email
	res.Permissions = data.Permissions
	return res
}
