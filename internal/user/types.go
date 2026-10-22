package user

import (
	"context"
	"gocms/internal/domain"
	"time"
)

type Repository interface {
	Create(ctx context.Context, user *domain.User, pwdhash domain.Password) error
	UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*domain.User, error)
	Update(ctx context.Context, username string, user *domain.User) error
	UpdatePasswd(ctx context.Context, username string, pwdhash domain.Password) error
	Get(ctx context.Context, username string) (*domain.User, error)
	GetActive(ctx context.Context) ([]domain.User, error)
	GetPasswdHash(ctx context.Context, username string) (domain.Password, error)
	Suspend(ctx context.Context, username string) error
}

type SearchEngine interface {
	Index(ctx context.Context, user *domain.User) error
	Get(ctx context.Context, username string) (*domain.User, error)
	Search(ctx context.Context, filters *SearchParams) (*SearchResult, error)
	Delete(ctx context.Context, username string) error
}
