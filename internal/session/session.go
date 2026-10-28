package session

import (
	"context"
	"gocms/internal/domain"
)

type Session struct {
	ID   string      `json:"session_id"`
	User domain.User `json:"user"`
}

type Repository interface {
	Create(ctx context.Context, user *domain.User) (*Session, error)
	RevokeID(ctx context.Context, suid string) error
	RevokeUser(ctx context.Context, username string) error
	Validate(ctx context.Context, suid string) (*Session, error)
}
