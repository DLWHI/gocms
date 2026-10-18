package auth

import (
	"context"
	"gocms/internal/domain"
)

type UserManager interface {
	PasswordMatches(ctx context.Context, username string, passwd domain.Password) error
	UpdateLogin(ctx context.Context, username string) (*domain.User, error)
	ChangePasswd(ctx context.Context, username string, passwd domain.Password) error
	Get(ctx context.Context, username string) (*domain.User, error)
}
