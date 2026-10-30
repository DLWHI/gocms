package user

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/domain"
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
	Version     uint           `gorm:"default:1"`
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
		if err := tx.Create(record).Error; err != nil {
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

func (r *GORMRepository) Update(ctx context.Context, username string, diff *domain.UserDiff) error {
	return r.mapErr(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		update := r.patch(diff)
		if len(update) > 0 {
			result := tx.Model(&Record{}).
				Where("username = ?", username).
				Updates(update)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return domain.ErrNotFound
			}
			if err := r.writeMessage(tx, domain.UserActionUpdate, username); err != nil {
				return err
			}
		}
		return nil
	}))
}

func (r *GORMRepository) UpdateLogin(ctx context.Context, username string, timestamp time.Time) error {
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
		return r.mapErr(err)
	}
	return nil
}

func (r *GORMRepository) UpdatePassword(ctx context.Context, username string, hash []byte) error {
	if len(hash) == 0 {
		return fmt.Errorf("%w: empty password", domain.ErrInval)
	}
	return r.mapErr(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&Record{}).
			Where("username = ?", username).
			Updates(map[string]interface{}{
				"password": hash,
				"version":  gorm.Expr("version + 1"),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}
		if err := r.writeMessage(tx, domain.UserActionPassword, username); err != nil {
			return err
		}
		return nil
	}))
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

	return user.ToDomain(), nil
}

func (r *GORMRepository) GetActive(ctx context.Context) ([]domain.User, error) {
	var users []Record

	err := r.db.WithContext(ctx).Omit("Password").Find(&users).Error
	if err != nil {
		return nil, err
	}

	vals := make([]domain.User, 0, len(users))

	for _, user := range users {
		vals = append(vals, *user.ToDomain())
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

func (*GORMRepository) writeMessage(tx *gorm.DB, action int, username string) error {
	payload := domain.UserEvent{
		Username: username,
		Action:   action,
	}
	if err := outbox.WriteMessage(tx, domain.UserEventName, "User.Repository", &payload); err != nil {
		return err
	}
	return nil
}

func (*GORMRepository) mapErr(err error) error {
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

func (*GORMRepository) patch(diff *domain.UserDiff) map[string]interface{} {
	record := make(map[string]interface{}, 0)

	if diff.Email != nil {
		record["email"] = *diff.Email
	}
	if diff.Permissions != nil {
		record["permissions"] = *diff.Permissions
	}

	if len(record) > 0 {
		record["version"] = gorm.Expr("version + 1")
	}
	return record
}

func (*Record) TableName() string {
	return domain.UserEntityName
}

func (rd *Record) ToDomain() *domain.User {
	var res domain.User
	res.Username = rd.Username

	if rd.SuspendedAt.Valid {
		t := rd.SuspendedAt.Time.Unix()
		res.SuspendedAt = &t
	}

	if rd.LastLogin != nil {
		t := rd.LastLogin.Unix()
		res.LastLogin = &t
	}

	res.Email = rd.Email
	res.Permissions = rd.Permissions
	return &res
}

func ToRecord(data *domain.User) *Record {
	var res Record

	res.Username = data.Username
	res.Email = data.Email
	res.Permissions = data.Permissions
	return &res
}
