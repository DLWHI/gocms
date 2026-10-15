package editor

import (
	"context"
	"errors"
	"gocms/internal/errs"
	"gocms/internal/outbox"
	"gocms/internal/user"
	"time"

	"gorm.io/gorm"
)

type GORMRepository struct {
	db *gorm.DB
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func editorDiff(editor *Model) map[string]interface{} {
	toUpdate := make(map[string]interface{})
	if editor.Name != "" {
		toUpdate["name"] = editor.Name
	}
	if editor.Bio != "" {
		toUpdate["bio"] = editor.Bio
	}
	return toUpdate
}

func (r *GORMRepository) Create(ctx context.Context, editor *Model) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&Model{}).Omit("UserData").Create(editor)
		if result.Error != nil {
			return result.Error
		}
		if err := writeMessage(tx, "create", editor.User, editor.Locale, nil); err != nil {
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
			ediff := editorDiff(diff)
			err := tx.Table("editors AS e").
				Select("e.version AS version").
				Where("u.username = ? AND u.suspended_at IS NULL", diff.User).
				Take(&version).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errs.ErrNotFound
				}
				return err
			}

			if len(ediff) > 0 {
				ediff["version"] = gorm.Expr("version + 1")
				if version == 0 {
					diff.Version = 1
					if err := tx.Create(diff).Error; err != nil {
						return err
					}
				} else {
					result := tx.Model(&Model{}).
						Where("user = ? AND locale = ? AND version = ?", diff.User, diff.Locale, version).
						Updates(ediff)
					if result.Error != nil {
						return result.Error
					}
					if result.RowsAffected == 0 {
						return errs.ErrLock
					}
				}
			}

			if err := writeMessage(tx, "update", diff.User, diff.Locale, nil); err != nil {
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

func (r *GORMRepository) Get(ctx context.Context, username string, locale string) (*Model, error) {
	var res struct {
		Username string
		Name     string
		Locale   string
		Email    string
	}

	err := r.db.WithContext(ctx).
		Table("users AS u").
		Select("u.username, u.email, e.name, e.locale").
		Joins("INNER JOIN editors AS e ON e.user = u.username AND e.locale = ?", locale).
		Where("u.username = ? AND u.suspended_at IS NULL", username).
		Take(&res).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	return &Model{
		Locale: locale,
		Name:   res.Name,
		User:   res.Username,
		UserData: user.Model{
			Email:    res.Email,
			Username: res.Username,
		},
	}, nil
}

func (r *GORMRepository) GetAll(ctx context.Context, locale string) ([]Model, error) {
	editors := make([]Model, 0)

	err := r.db.WithContext(ctx).
		Model(&Model{}).
		Omit("version").
		Preload("UserData", func(db *gorm.DB) *gorm.DB {
			return db.Omit("passwd", "version")
		}).
		Where("locale = ?", locale).
		Find(&editors).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}

	return editors, nil
}

func (r *GORMRepository) Delete(ctx context.Context, username string) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		result := tx.Where("user = ?", username).
			Delete(&Model{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errs.ErrNotFound
		}
		if err := writeMessage(tx, "delete", username, "", &now); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func writeMessage(tx *gorm.DB, action string, username string, locale string, timestamp *time.Time) error {
	payload := MessagePayload{
		Username:    username,
		Locale:      locale,
		Action:      action,
		SuspendedAt: nil,
	}
	if timestamp != nil {
		ts := timestamp.Unix()
		payload.SuspendedAt = &ts
	}
	if err := outbox.WriteMessage(tx, "editor", "Editor.Repository", &payload); err != nil {
		return err
	}
	return nil
}
