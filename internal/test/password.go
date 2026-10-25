package test

import "encoding/json"

type Password []byte

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

func (p *Password) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(*p))
}

func (p *Password) String() string {
	return string(*p)
}
