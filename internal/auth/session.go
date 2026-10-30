package auth

import (
	"context"
	"gocms/internal/domain"
	"time"
)

const (
	MaxTokenLength = 200
)

type SessionRepository interface {
	Create(ctx context.Context, username string) (string, error)
	Revoke(ctx context.Context, username string, suid string) error
	RevokeUser(ctx context.Context, username string) error
	Validate(ctx context.Context, token string) (string, string, error)
}

type UserRepository interface {
	UpdateLogin(ctx context.Context, username string, timestamp time.Time) error
	UpdatePassword(ctx context.Context, username string, hash []byte) error
	Get(ctx context.Context, username string) (*domain.User, error)
	GetPasswordHash(ctx context.Context, username string) ([]byte, error)
}
