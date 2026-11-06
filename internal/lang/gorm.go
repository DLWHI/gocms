package lang

import (
	"context"
	"gocms/internal/database"
	"gocms/internal/domain"
	"gocms/internal/outbox"

	"gorm.io/gorm"
)

type GORMRepository struct {
	db *gorm.DB
}

func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

func (r *GORMRepository) Create(ctx context.Context, lang *Language) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(lang).Error; err != nil {
			return err
		}

		if err := outbox.WriteMessage(tx, Event, "Language.Repository", nil); err != nil {
			return err
		}

		return nil
	})
	return database.MapError(err)
}

func (r *GORMRepository) GetAll(ctx context.Context) ([]Language, error) {
	langs := make([]Language, 0)

	err := r.db.WithContext(ctx).
		Order("code ASC").
		Find(&langs).Error
	if err != nil {
		return nil, err
	}

	return langs, nil
}

func (r *GORMRepository) Delete(ctx context.Context, code string) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("code = ?", code).
			Delete(&Language{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNotFound
		}

		if err := outbox.WriteMessage(tx, Event, "Language.Repository", nil); err != nil {
			return err
		}

		return nil
	})
	return database.MapError(err)
}
