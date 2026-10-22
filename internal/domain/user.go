package domain

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
)

const (
	MaxUsernameLength = 64
	charset           = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;:,.<>?"

	UserActionCreate = 1 << iota
	UserActionUpdate
	UserActionSuspend
	UserActionLogin
	UserActionPasswd

	UserEventName = "user"
)

type Password []byte

type User struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Permissions uint32 `json:"permissions"`
	SuspendedAt *int64 `json:"suspended_at"`
	LastLogin   *int64 `json:"last_login"`
}

type UserEvent struct {
	Action   int    `json:"action"`
	Username string `json:"username"`
}

func (p *Password) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*p = Password(s)
	return nil
}

func (Password) MarshalJSON() ([]byte, error) {
	return []byte("[NOT_YOUR_BUISNESS]"), nil
}

func (Password) String() string {
	return "[NOT_YOUR_BUISNESS]"
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

func GeneratePassword(length int) (Password, error) {
	bytes := make([]byte, length)

	for i := range bytes {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return nil, err
		}
		bytes[i] = charset[randomIndex.Int64()]
	}

	return bytes, nil
}

func IsValidPassword(passwd Password) bool {
	if len(passwd) == 0 || len(passwd) > 72 {
		return false
	}
	return true
}

func ParseEmail(email string) (string, error) {
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Name != "" {
		return "", fmt.Errorf("%w: invalid email", ErrInval)
	}
	return parsed.Address, nil
}
