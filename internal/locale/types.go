package locale

import "context"

type Locale struct {
	Name string `json:"name"`
	Code string `json:"code" gorm:"primaryKey;size:8;"`
}

type Repository interface {
	Create(ctx context.Context, locale *Locale) error
	GetAll(ctx context.Context) ([]Locale, error)
	Delete(ctx context.Context, locale string) error
}

func (Locale) TableName() string {
	return "locales"
}
