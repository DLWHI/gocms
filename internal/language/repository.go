package language

import (
	"context"
	"gocms/internal/domain"
)

const allList = "all"

type Repository interface {
	Create(ctx context.Context, lang *domain.Language) error
	GetAll(ctx context.Context) ([]domain.Language, error)
	Delete(ctx context.Context, code string) error
}
