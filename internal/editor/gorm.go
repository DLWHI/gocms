package editor

import (
	"context"
	"errors"
	"gocms/internal/errs"
	"gocms/internal/user"

	"gorm.io/gorm"
)

type GORMRepository struct {
	db *gorm.DB
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func userDiff(user *user.Model) map[string]interface{} {
	toUpdate := make(map[string]interface{})

	if user.Email != "" {
		toUpdate["email"] = user.Email
		toUpdate["version"] = gorm.Expr("version + 1")
	}

	return toUpdate
}

func editorDiff(editor *Model) map[string]interface{} {
	toUpdate := make(map[string]interface{})
	if editor.Name != "" {
		toUpdate["name"] = editor.Name
		toUpdate["version"] = gorm.Expr("version + 1")
	}
	return toUpdate
}

func (r *GORMRepository) Create(ctx context.Context, editor *Model) error {
	return r.db.WithContext(ctx).Model(&Model{}).Omit("UserData").Create(editor).Error
}

func (r *GORMRepository) Update(ctx context.Context, diff *Model) error {
	var version struct {
		User   uint
		Editor uint
	}
	for i := 0; i < 5; i++ {
		err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			udiff := userDiff(&diff.UserData)
			ediff := editorDiff(diff)
			err := tx.Table("users AS u").
				Select("u.version AS user, e.version AS editor").
				Joins("LEFT JOIN editors AS e ON e.user = u.username AND e.locale = ?", diff.Locale).
				Where("u.username = ? AND u.suspended_at IS NULL", diff.User).
				Take(&version).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errs.ErrNotFound
				}
				return err
			}

			if len(udiff) > 0 {
				result := tx.Model(&user.Model{}).
					Where("username = ? AND version = ?", diff.User, version.User).
					Updates(udiff)
				if result.Error != nil {
					return result.Error
				}

				if result.RowsAffected == 0 {
					return errs.ErrLock
				}
			}

			if len(ediff) > 0 {
				if version.Editor == 0 {
					diff.Version = 1
					if err := tx.Create(diff).Error; err != nil {
						return err
					}
				} else {
					result := tx.Model(&Model{}).
						Where("user = ? AND locale = ? AND version = ?", diff.User, diff.Locale, version.Editor).
						Updates(ediff)
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
		Joins("LEFT JOIN editors AS e ON e.user = u.username AND e.locale = ?", locale).
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

func (r *GORMRepository) GetTranslation(ctx context.Context, username string, locale string) (*Model, error) {
	var res struct {
		Username string
		Name     string
		Locale   string
		Email    string
	}

	err := r.db.WithContext(ctx).
		Table("users AS u").
		Select("u.username, u.email, e.name, e.locale").
		Joins("LEFT JOIN editors AS e ON e.user = u.username", locale).
		Where("u.username = ? AND u.suspended_at IS NULL AND e.locale = ?", username, locale).
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
	result := r.db.WithContext(ctx).
		Where("username = ?", username).
		Delete(&user.Model{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errs.ErrNotFound
	}
	return nil
}
