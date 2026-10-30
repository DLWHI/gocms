package domain

import (
	"fmt"
	"net/mail"
	"strings"
)

const (
	UserActionCreate = 1 << iota
	UserActionUpdate
	UserActionSuspend
	UserActionLogin
	UserActionPassword

	UserEventName  = "user"
	UserEntityName = "users"

	MaxUsernameLength = 64
	MaxEmailLength    = 254
)

type User struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Permissions uint32 `json:"permissions"`
	SuspendedAt *int64 `json:"suspended_at"`
	LastLogin   *int64 `json:"last_login"`
}

type UserDiff struct {
	Email       *string `json:"email,omitempty"`
	Permissions *uint32 `json:"permissions,omitempty"`
}

type UserEvent struct {
	Action   int    `json:"action"`
	Username string `json:"username"`
}

func IsInvalidUsernameRune(r rune) bool {
	isAlpha := 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z'
	isDigit := '0' <= r && r <= '9'
	return !(isAlpha || isDigit || r == '-' || r == '_')
}

func NormalizeUsername(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", fmt.Errorf("%w: username cannot be empty", ErrInval)
	}
	if len(normalized) > MaxUsernameLength {
		return "", fmt.Errorf("%w: username is longer than %d chars", ErrInval, MaxUsernameLength)
	}
	if strings.IndexFunc(normalized, IsInvalidUsernameRune) >= 0 {
		return "", fmt.Errorf("%w: username must contain only letters, numbers, dashes or underscores", ErrInval)
	}

	return strings.ToLower(normalized), nil
}

func NormalizeEmail(email string) (string, error) {
	if email == "" {
		return email, nil
	}
	if len(email) > MaxEmailLength {
		return "", fmt.Errorf("%w: email too long", ErrInval)
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Name != "" {
		return "", fmt.Errorf("%w: invalid email", ErrInval)
	}
	return parsed.Address, nil
}
