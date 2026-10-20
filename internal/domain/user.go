package domain

import (
	"encoding/json"
	"fmt"
	"gocms/internal/errs"
	"strings"
)

type Password []byte

func (p *Password) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*p = Password(s)
	return nil
}

type User struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Permissions uint32 `json:"permissions"`
	SuspendedAt *int64 `json:"suspended_at"`
	LastLogin   *int64 `json:"last_login"`
}

func IsInvalidUsernameRune(r rune) bool {
	isAlpha := 'a' <= r && r <= 'z'
	isDigit := '0' <= r && r <= '9'
	return !(isAlpha || isDigit || r == '-' || r == '_')
}

func NormalizeUsername(value string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if normalized == "" {
		return "", fmt.Errorf("%w: username cannot be empty", errs.ErrInval)
	} else if len(normalized) > 64 {
		return "", fmt.Errorf("%w: username is longer than 64 chars", errs.ErrInval)
	} else if strings.IndexFunc(normalized, IsInvalidUsernameRune) >= 0 {
		return "", fmt.Errorf("%w: username must not contain spaces or periods", errs.ErrInval)
	}

	return normalized, nil
}
