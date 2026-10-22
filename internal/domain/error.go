package domain

type Constant string

func (e Constant) Error() string {
	return string(e)
}

const (
	ErrNotFound = Constant("Entity not found")
	ErrInval    = Constant("Invalid argument")
	ErrCred     = Constant("Invalid username or password")
	ErrExpired  = Constant("Session expired")
	ErrPerm     = Constant("Insufficient permissions")
	ErrLock     = Constant("Record is locked")
	ErrRetry    = Constant("Retry operation")
)
