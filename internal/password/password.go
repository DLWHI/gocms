package password

import "encoding/json"

type Password []byte

type PasswordService interface {
	Hash(password Password) ([]byte, error)
	PasswordMatches(password Password, hash []byte) bool
	GeneratePassword(length int) (Password, error)
	IsValidPassword(password Password) error
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
