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
	UserActionPasswd

	UserEventName  = "user"
	UserEntityName = "users"

	MaxUsernameLength = 64
)

type User struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Permissions uint32 `json:"permissions"`
	SuspendedAt int64  `json:"suspended_at"`
	LastLogin   int64  `json:"last_login"`
}

type UserEvent struct {
	Action   int    `json:"action"`
	Username string `json:"username"`
}

func IsInvalidUsernameRune(r rune) bool {
	isAlpha := 'a' <= r && r <= 'z'
	isDigit := '0' <= r && r <= '9'
	return !(isAlpha || isDigit || r == '-' || r == '_')
}

func NormalizeUsername(value string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if normalized == "" {
		return "", fmt.Errorf("%w: username cannot be empty", ErrInval)
	} else if len(normalized) > MaxUsernameLength {
		return "", fmt.Errorf("%w: username is longer than %d chars", ErrInval, MaxUsernameLength)
	} else if strings.IndexFunc(normalized, IsInvalidUsernameRune) >= 0 {
		return "", fmt.Errorf("%w: username must contain only letters, numbers, dashes or underscores", ErrInval)
	}

	return normalized, nil
}

func ParseEmail(email string) (string, error) {
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Name != "" {
		return "", fmt.Errorf("%w: invalid email", ErrInval)
	}
	return parsed.Address, nil
}
