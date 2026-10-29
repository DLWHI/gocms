package category

import (
	"context"
	"gocms/internal/domain"
)

const (
	notHiddenListBase = "not_hidden"
)

type Repository interface {
	Create(ctx context.Context, category *domain.Category) error
	Update(ctx context.Context, slug string, locale string, diff *domain.CategoryDiff) error
	Get(ctx context.Context, slug string, locale string) (*domain.Category, error)
	GetTranslations(ctx context.Context, slug string) ([]domain.Category, error)
	GetNotHidden(ctx context.Context, locale string) ([]domain.Category, error)
	Delete(ctx context.Context, slug string) error
}
