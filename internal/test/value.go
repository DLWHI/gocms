package test

import (
	"errors"
	"fmt"
	"gocms/internal/domain"
	"strings"
)

type Value = any

func (d *Data) Resolve(v Value) (string, error) {
	m, ok := v.(map[string]Value)
	if ok {
		for name, arg := range m {
			if strings.HasPrefix(name, "$") {
				val, err := d.ExpandToken(name, arg)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("%v", val), nil
			}
		}
	}
	return fmt.Sprintf("%v", v), nil
}

func (d *Data) ResolveMap(values map[string]Value) (map[string]Value, error) {
	res := make(map[string]Value, len(values))
	for key, value := range values {
		switch value.(type) {
		case map[string]any:
			resolved, err := d.Resolve(value)
			if err != nil {
				if !errors.Is(err, domain.ErrInval) {
					return nil, err
				}
				res[key] = value
			} else {
				res[key] = resolved
			}
		default:
			res[key] = value
		}
	}
	return res, nil
}

func (d *Data) ExpandToken(name string, arg Value) (Value, error) {
	switch name {
	case "$const":
		key, ok := arg.(string)
		if !ok {
			return nil, fmt.Errorf("$const expects a string")
		}
		val, ok := d.Constants[key]
		if !ok {
			return nil, fmt.Errorf("unknown constant %q", key)
		}
		return val, nil

	case "$repeat":
		m, ok := arg.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("$repeat expects {s, n}")
		}
		s, ok1 := m["s"].(string)
		n, ok2 := m["n"].(float64)
		if !ok1 || !ok2 || n < 0 || n != float64(int(n)) {
			return nil, fmt.Errorf("$repeat needs string s and non-negative integer n")
		}
		return strings.Repeat(s, int(n)), nil

	case "$email_of_length":
		n, ok := arg.(float64)
		const domain = "@example.com"
		if !ok || int(n) <= len(domain) {
			return nil, fmt.Errorf("$email_of_length needs an integer > %d", len(domain))
		}
		return strings.Repeat("a", int(n)-len(domain)) + domain, nil
	case "$raw_session":
		actor, ok := arg.(string)
		if !ok {
			return nil, fmt.Errorf("$raw_session needs to be a string")
		}
		return d.GetSessionOf(actor), nil
	default:
		return nil, domain.ErrInval
	}
}
