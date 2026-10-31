package locale

import (
	"context"
	"errors"
	"gocms/internal/domain"

	"gorm.io/gorm"
)

type Record struct {
	Name string ``
	Code string `gorm:"primaryKey;size:8;"`
}

type GORMRepository struct {
	db *gorm.DB
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func (r *GORMRepository) Create(ctx context.Context, locale *domain.Locale) error {
	if err := r.db.WithContext(ctx).Create(ToRecord(locale)).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return domain.ErrDuplicate
		}
		return err
	}
	return nil
}

func (r *GORMRepository) GetAll(ctx context.Context) ([]domain.Locale, error) {
	locales := make([]Record, 0)

	err := r.db.WithContext(ctx).
		Find(&locales).Error
	if err != nil {
		return nil, err
	}

	vals := make([]domain.Locale, 0, len(locales))

	for _, locale := range locales {
		vals = append(vals, *locale.ToDomain())
	}

	return vals, nil
}

func (r *GORMRepository) Delete(ctx context.Context, code string) error {
	result := r.db.WithContext(ctx).
		Where("code = ?", code).
		Delete(&Record{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (Record) TableName() string {
	return "locales"
}

func ToRecord(locale *domain.Locale) *Record {
	return &Record{
		Name: locale.Name,
		Code: locale.Code,
	}
}

func (l *Record) ToDomain() *domain.Locale {
	return &domain.Locale{
		Name: l.Name,
		Code: l.Code,
	}
}
