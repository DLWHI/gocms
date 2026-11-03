package domain

import "context"

const (
	UsernameContextKey  = "username"
	SessionContextIDKey = "session"
)

type AuthService interface {
	Authorize(ctx context.Context, entity string, action string) bool
	IsSubsetOf(ctx context.Context, permissions uint32) bool
}
