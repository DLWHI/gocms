package post

import (
	"context"
	"gocms/internal/domain"
)

type Repository interface {
	Create(ctx context.Context, post *domain.Post) error
	Update(ctx context.Context, id uint, language string, post *domain.PostDiff) error
	SetViews(ctx context.Context, id uint, value uint) error
	GetViews(ctx context.Context, id uint) (uint, error)
	IncrementViews(ctx context.Context, id uint, value uint) (uint, error)
	Get(ctx context.Context, id uint, language string) (*domain.Post, error)
	GetTranslations(ctx context.Context, id uint) ([]domain.Post, error)
	GetBatch(ctx context.Context, ids []uint) ([]domain.Post, error)
	GetCreator(ctx context.Context, id uint) (string, error)
	Delete(ctx context.Context, id uint) error
}
