package advertisement

import (
	"context"
	"gocms/internal/domain"
)

const groupsList = "groups"

type Repository interface {
	Create(ctx context.Context, adv *domain.Advertisement) error
	Update(ctx context.Context, id uint, adv *domain.AdvertisementDiff) error
	Get(ctx context.Context, id uint) (*domain.Advertisement, error)
	GetGroup(ctx context.Context, group string) ([]domain.Advertisement, error)
	GetGroups(ctx context.Context) ([]string, error)
	Delete(ctx context.Context, id uint) error
}
