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
	Passwd      []byte         `gorm:"not null"`
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
	record.Passwd = pwdhash
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
			return domain.ErrExists
		}
		return err
	}
	return nil
}

func (r *GORMRepository) UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*domain.User, error) {
	var user Record
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&Record{}).Omit("Passwd").Where("username = ?", username).Take(&user).Error; err != nil {
			return err
		}

		if user.LastLogin != nil {
			if timestamp.Before(*user.LastLogin) {
				return fmt.Errorf("%w: cannot login back in time", domain.ErrInval)
			}
		}

		result := tx.Model(&Record{}).Where("username = ?", user.Username).Update("LastLogin", timestamp)
		if result.Error != nil {
			return result.Error
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

func (r *GORMRepository) Update(ctx context.Context, username string, update *domain.User) error {
	upd := ToRecord(update)
	err := flow.DoWithRetry(ctx, r.retries, r.backoff, func() error {
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var user Record
			if err := tx.Select("Version").Where("username = ?", username).Take(&user).Error; err != nil {
				return err
			}
			diff := user.Diff(&upd)
			if len(diff) > 0 {
				diff["version"] = gorm.Expr("version + 1")
				result := tx.Model(&Record{}).
					Where("username = ? AND version = ?", user.Username, user.Version).
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

func (r *GORMRepository) UpdatePasswd(ctx context.Context, username string, pwdhash []byte) error {
	if len(pwdhash) == 0 {
		return fmt.Errorf("%w: empty password", domain.ErrInval)
	}
	err := flow.DoWithRetry(ctx, r.retries, r.backoff, func() error {
		return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var version uint
			if err := tx.Model(&Record{}).Select("Version").Where("username = ?", username).Take(&version).Error; err != nil {
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
			if err := r.writeMessage(tx, domain.UserActionPasswd, username); err != nil {
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
		Omit("Passwd").
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

	err := r.db.WithContext(ctx).Omit("Passwd").Find(&users).Error
	if err != nil {
		return nil, err
	}

	vals := make([]domain.User, 0, len(users))

	for _, user := range users {
		vals = append(vals, user.ToDomain())
	}

	return vals, nil
}

func (r *GORMRepository) GetPasswdHash(ctx context.Context, username string) ([]byte, error) {
	var user Record

	err := r.db.WithContext(ctx).Model(&Record{}).
		Select("Passwd").
		Where("username = ?", username).
		Limit(1).
		Take(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return user.Passwd, nil
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
	if err != nil {
		if errors.Is(err, domain.ErrRetry) {
			return domain.ErrLock
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNotFound
		}
		return err
	}
	return nil
}

func (this *Record) Diff(other *Record) map[string]interface{} {
	diff := make(map[string]interface{})

	if this.Permissions != other.Permissions {
		diff["permissions"] = other.Permissions
	}
	if this.Email != other.Email {
		diff["email"] = other.Email
	}
	return diff
}

func (ur *Record) ToDomain() domain.User {
	var ud domain.User
	ud.Username = ur.Username

	if ur.SuspendedAt.Valid {
		ud.SuspendedAt = ur.SuspendedAt.Time.Unix()
	}

	if ur.LastLogin != nil {
		ud.LastLogin = ur.LastLogin.Unix()
	}

	ud.Email = ur.Email
	ud.Permissions = ur.Permissions
	return ud
}

func ToRecord(ud *domain.User) Record {
	var um Record

	if ud.SuspendedAt != 0 {
		um.SuspendedAt = gorm.DeletedAt{
			Time:  time.Unix(ud.SuspendedAt, 0).UTC(),
			Valid: true,
		}
	}

	if ud.LastLogin != 0 {
		t := time.Unix(ud.LastLogin, 0)
		um.LastLogin = &t
	}

	um.Username = ud.Username
	um.Email = ud.Email
	um.Permissions = ud.Permissions
	return um
}
