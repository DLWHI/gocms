package auth

import (
	"context"
	"gocms/internal/domain"
	"gocms/internal/hash"
)

type UserManager interface {
	PasswordMatches(ctx context.Context, username string, passwd hash.Password) error
	UpdateLogin(ctx context.Context, username string) (*domain.User, error)
	ChangePassword(ctx context.Context, username string, passwd hash.Password) error
	Get(ctx context.Context, username string) (*domain.User, error)
}
