package auth

import (
	"context"
	"gocms/internal/domain"
	"gocms/internal/password"
)

type UserManager interface {
	PasswordMatches(ctx context.Context, username string, passwd password.Password) error
	UpdateLogin(ctx context.Context, username string) (*domain.User, error)
	ChangePassword(ctx context.Context, username string, passwd password.Password) error
	Get(ctx context.Context, username string) (*domain.User, error)
}
