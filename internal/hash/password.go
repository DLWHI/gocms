package hash

import (
	"bytes"
	"fmt"
	"gocms/internal/domain"
)

const ()

type Password []byte

type PasswordService interface {
	Hash(pwd Password) ([]byte, error)
	PasswordMatches(pwd Password, hash []byte) bool
	GeneratePassword(length int) (Password, error)
	IsValidPassword(passwd Password) bool
}

func (p *Password) UnmarshalJSON(data []byte) error {
	defer clear(data)
	if bytes.Equal(data, []byte("null")) {
		*p = nil
		return nil
	}

	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		content := data[1 : len(data)-1]
		*p = append([]byte(nil), content...)
		return nil
	}
	return fmt.Errorf("%w: expected string for password, got %s", domain.ErrInval, string(data))
}

func (Password) MarshalJSON() ([]byte, error) {
	return []byte(`"[********]"`), nil
}

func (Password) String() string {
	return "[********]"
}
