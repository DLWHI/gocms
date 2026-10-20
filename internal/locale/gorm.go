package locale

import (
	"context"
	"gocms/internal/errs"

	"gorm.io/gorm"
)

type GORMRepository struct {
	db *gorm.DB
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func (r *GORMRepository) Create(ctx context.Context, locale *Locale) error {
	return r.db.WithContext(ctx).Create(locale).Error
}

func (r *GORMRepository) GetAll(ctx context.Context) ([]Locale, error) {
	locales := make([]Locale, 0)

	err := r.db.WithContext(ctx).
		Find(&locales).Error
	if err != nil {
		return nil, err
	}

	return locales, nil
}

func (r *GORMRepository) Delete(ctx context.Context, locale string) error {
	result := r.db.WithContext(ctx).
		Where("code = ?", locale).
		Delete(&Locale{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errs.ErrNotFound
	}
	return nil
}
