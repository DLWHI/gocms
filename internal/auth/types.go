package auth

import (
	"context"
	"gocms/internal/domain"
)

const (
	UserPermissionKey = "userperm"
	UserNameKey       = "userperm"
	SessionIDKey      = "userperm"
)

type UserManager interface {
	PasswordMatches(ctx context.Context, username string, passwd domain.Password) error
	UpdateLogin(ctx context.Context, username string) (*domain.User, error)
	ChangePasswd(ctx context.Context, username string, passwd domain.Password) error
	Get(ctx context.Context, username string) (*domain.User, error)
}
