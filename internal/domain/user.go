package domain

import "encoding/json"

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
