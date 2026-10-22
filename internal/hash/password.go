package hash

import (
	"encoding/json"
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
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*p = Password(s)
	return nil
}

func (Password) MarshalJSON() ([]byte, error) {
	return []byte(`"[********]"`), nil
}

func (Password) String() string {
	return "[********]"
}
