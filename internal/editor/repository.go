package editor

import (
	"context"
	"gocms/internal/domain"
)

const (
	allListBase = "all"
)

type Repository interface {
	Create(ctx context.Context, editor *domain.Editor) error
	Update(ctx context.Context, slug string, locale string, editor *domain.EditorDiff) error
	GetAny(ctx context.Context, slug string, locale string) (*domain.Editor, error)
	GetTranslations(ctx context.Context, slug string) ([]domain.Editor, error)
	GetAll(ctx context.Context, locale string) ([]domain.Editor, error)
	Delete(ctx context.Context, slug string) error
}
