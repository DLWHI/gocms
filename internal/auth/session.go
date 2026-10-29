package auth

import (
	"context"
	"gocms/internal/domain"
	"gocms/internal/password"
)

type SessionRepository interface {
	Create(ctx context.Context, username string) (string, error)
	Revoke(ctx context.Context, username string, suid string) error
	RevokeUser(ctx context.Context, username string) error
	Validate(ctx context.Context, token string) (string, error)
}

type UserService interface {
	PasswordMatches(ctx context.Context, username string, passwd password.Password) error
	UpdateLogin(ctx context.Context, username string) (*domain.User, error)
	ChangePasswordSelf(ctx context.Context, old password.Password, passwd password.Password) error
	Get(ctx context.Context, username string) (*domain.User, error)
}
