package domain

import "context"

const (
	UsernameKey  = "username"
	SessionIDKey = "session"
)

type AuthService interface {
	Authorize(ctx context.Context, entity string, action string) bool
	IsSubsetOf(ctx context.Context, permissions uint32) bool
}
