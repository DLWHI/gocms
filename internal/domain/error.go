package domain

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrNotFound  = Error("Entity not found")
	ErrInval     = Error("Invalid argument")
	ErrCred      = Error("Invalid username or password")
	ErrExpired   = Error("Session expired")
	ErrPerm      = Error("Insufficient permissions")
	ErrLock      = Error("Record is locked")
	ErrRetry     = Error("Retry operation")
	ErrDuplicate = Error("Entity with same key exists")
	ErrAbnormal  = Error("Operation interrupted unexpectedly")
)
