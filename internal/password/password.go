package password

import "encoding/json"

type Password []byte

type PasswordService interface {
	Hash(pwd Password) ([]byte, error)
	PasswordMatches(pwd Password, hash []byte) bool
	GeneratePassword(length int) (Password, error)
	IsValidPassword(password Password) bool
}

func (p *Password) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*p = nil
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*p = []byte(s)
	return nil
}

func (Password) MarshalJSON() ([]byte, error) {
	return []byte(`"[********]"`), nil
}

func (Password) String() string {
	return "[********]"
}
