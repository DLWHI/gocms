package locale

import "context"

type Model struct {
	Name string `json:"name"`
	Code string `json:"code" gorm:"primaryKey;size:8;"`
}

type Repository interface {
	Create(ctx context.Context, locale *Model) error
	GetAll(ctx context.Context) ([]Model, error)
	Delete(ctx context.Context, locale string) error
}

func (Model) TableName() string {
	return "locales"
}
