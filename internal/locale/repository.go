package locale

import (
	"context"
	"gocms/internal/domain"
)

const allList = "all"

type Repository interface {
	Create(ctx context.Context, locale *domain.Locale) error
	GetAll(ctx context.Context) ([]domain.Locale, error)
	Delete(ctx context.Context, code string) error
}
