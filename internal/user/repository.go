package user

import (
	"context"
	"gocms/internal/domain"
	"time"
)

type Repository interface {
	Create(ctx context.Context, user *domain.User, pwdhash []byte) error
	Update(ctx context.Context, username string, diff map[string]interface{}) error
	UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*domain.User, error)
	UpdatePassword(ctx context.Context, username string, pwdhash []byte) error
	Get(ctx context.Context, username string) (*domain.User, error)
	GetActive(ctx context.Context) ([]domain.User, error)
	GetPasswordHash(ctx context.Context, username string) ([]byte, error)
	Suspend(ctx context.Context, username string) error
}
