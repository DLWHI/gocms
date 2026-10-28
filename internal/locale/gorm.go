package locale

import (
	"context"
	"gocms/internal/domain"

	"gorm.io/gorm"
)

type Locale struct {
	Name string ``
	Code string `gorm:"primaryKey;size:8;"`
}

type GORMRepository struct {
	db *gorm.DB
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func (r *GORMRepository) Create(ctx context.Context, name string, code string) error {
	locale := Locale{
		Name: name,
		Code: code,
	}
	return r.db.WithContext(ctx).Create(&locale).Error
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

func (r *GORMRepository) Delete(ctx context.Context, code string) error {
	result := r.db.WithContext(ctx).
		Where("code = ?", code).
		Delete(&Locale{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (Locale) TableName() string {
	return "locales"
}
