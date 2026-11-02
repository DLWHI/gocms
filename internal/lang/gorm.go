package lang

import (
	"context"
	"gocms/internal/database"
	"gocms/internal/domain"

	"gorm.io/gorm"
)

type Record struct {
	Name string ``
	Code string `gorm:"primaryKey;size:2;"`
}

type GORMRepository struct {
	db *gorm.DB
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func (r *GORMRepository) Create(ctx context.Context, lang *domain.Language) error {
	err := r.db.WithContext(ctx).Create(ToRecord(lang)).Error
	return database.MapError(err)
}

func (r *GORMRepository) GetAll(ctx context.Context) ([]domain.Language, error) {
	langs := make([]Record, 0)

	err := r.db.WithContext(ctx).
		Find(&langs).Error
	if err != nil {
		return nil, err
	}

	vals := make([]domain.Language, 0, len(langs))

	for _, lang := range langs {
		vals = append(vals, *lang.ToDomain())
	}

	return vals, nil
}

func (r *GORMRepository) Delete(ctx context.Context, code string) error {
	result := r.db.WithContext(ctx).
		Where("code = ?", code).
		Delete(&Record{})
	if result.Error != nil {
		return database.MapError(result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (Record) TableName() string {
	return "languages"
}

func ToRecord(lang *domain.Language) *Record {
	return &Record{
		Name: lang.Name,
		Code: lang.Code,
	}
}

func (l *Record) ToDomain() *domain.Language {
	return &domain.Language{
		Name: l.Name,
		Code: l.Code,
	}
}
