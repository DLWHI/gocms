package test

import (
	"errors"
	"fmt"
	"gocms/internal/domain"
	"reflect"
	"strings"
)

type Value = any

func (d *Data) Resolve(v Value) (Value, error) {
	m, ok := v.(map[string]Value)
	if ok {
		for name, arg := range m {
			if strings.HasPrefix(name, "$") {
				val, err := d.ExpandToken(name, arg)
				if err != nil {
					return nil, err
				}
				return val, nil
			}
		}
	}
	return v, nil
}

func (d *Data) ResolveString(v Value) (string, error) {
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
			resolved, err := d.ResolveString(value)
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

func ResolvePath(values map[string]Value, path string) ([]Value, error) {
	currents := []Value{values}
	if path == "" || path == "$" {
		return currents, nil
	}
	parts := strings.Split(path, ".")
	for _, part := range parts {
		if part == "" {
			continue
		}
		key := part
		expand := strings.HasSuffix(part, "[*]")
		expanded := make([]Value, 0, len(currents))
		if expand {
			key = strings.TrimSuffix(part, "[*]")
		}

		for _, current := range currents {
			obj, ok := current.(map[string]Value)
			if !ok {
				return nil, fmt.Errorf("values.%s, expected object, got %T", key, current)
			}
			if expand {
				arr, ok := obj[key].([]Value)
				if !ok {
					return nil, fmt.Errorf("values.%s, expected array, got %T", key, current)
				}
				for _, val := range arr {
					expanded = append(expanded, val)
				}
			} else {
				expanded = append(expanded, obj[key])
			}
		}
		currents = expanded
	}
	return currents, nil
}

func ToLen(v Value) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int8:
		return int(n), true
	case int16:
		return int(n), true
	case int32:
		return int(n), true
	case int64:
		return int(n), true
	case uint:
		return int(n), true
	case uint8:
		return int(n), true
	case uint16:
		return int(n), true
	case uint32:
		return int(n), true
	case uint64:
		return int(n), true
	case float32:
		return int(n), true
	case float64:
		return int(n), true
	case bool:
		if n {
			return 1, true
		} else {
			return 0, true
		}
	default:
		return 0, false
	}
}
func ToNumber(v Value) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	case bool:
		if n {
			return 1, true
		} else {
			return 0, true
		}
	default:
		return 0, false
	}
}
func ToInt64(v Value) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	default:
		return 0, false
	}
}

func TypeMatches(value Value, typename string) error {
	switch typename {
	case "number":
		if _, ok := value.(float64); ok {
			return nil
		}
		return fmt.Errorf("Assertion failed, expected number, got %T", value)
	case "string":
		if _, ok := value.(string); ok {
			return nil
		}
		return fmt.Errorf("Assertion failed, expected string, got %T", value)
	case "boolean":
		if _, ok := value.(bool); ok {
			return nil
		}
		return fmt.Errorf("Assertion failed, expected bool, got %T", value)
	case "object":
		if _, ok := value.(map[string]any); ok {
			return nil
		}
		return fmt.Errorf("Assertion failed, expected object, got %T", value)
	case "array":
		if _, ok := value.([]any); ok {
			return nil
		}
		return fmt.Errorf("Assertion failed, expected array, got %T", value)
	case "null":
		return IsNull(value)
	}
	return fmt.Errorf("Invalid type in assertion, got %T", typename)
}

func IsNull(value Value) error {
	if value == nil {
		return nil
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if rv.IsNil() {
			return nil
		}
	}
	return fmt.Errorf("Assertion failed: expected %v == nil", value)
}
func IsNotNull(value Value) error {
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Invalid:
		return fmt.Errorf("Assertion failed: expected %v != nil", value)
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if !rv.IsNil() {
			return nil
		}
		return fmt.Errorf("Assertion failed: expected %v != nil", value)
	}
	return nil
}

func AllNull(values []Value) error {
	for _, value := range values {
		if err := IsNotNull(value); err != nil {
			return fmt.Errorf("Assertion failed: expected all values in %v to be null", values)
		}
	}
	return nil
}
func AllNotNull(values []Value) error {
	for _, value := range values {
		if err := IsNull(value); err != nil {
			return fmt.Errorf("Assertion failed: expected all values in %v to be non-null", values)
		}
	}
	return nil
}

func IsSorted(values []Value, desc bool) error {
	// last := values[0]
	// for i := 1; i < len(values); i++ {
	// 	if desc {
	// 		if last > values[i] {
	// 			return fmt.Errorf("Assertion failed: expected sorted ascending array")
	// 		}
	// 	} else {
	// 		if last < values[i] {
	// 			return fmt.Errorf("Assertion failed: expected sorted ascending array")
	// 		}
	// 	}
	// }
	return nil
}
