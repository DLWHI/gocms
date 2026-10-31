package test

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
)

type JSONValue = any

func ResolvePath(values map[string]JSONValue, path string) ([]JSONValue, error) {
	currents := []JSONValue{values}
	if path == "" || path == "$" {
		return currents, nil
	}
	parts := strings.Split(path, ".")
	for _, part := range parts {
		if part == "" || part == "$" {
			continue
		}
		key := part
		expand := strings.HasSuffix(part, "[*]")
		expanded := make([]JSONValue, 0, len(currents))
		if expand {
			key = strings.TrimSuffix(part, "[*]")
		}

		for _, current := range currents {
			obj, ok := current.(map[string]JSONValue)
			if !ok {
				return nil, fmt.Errorf("values.%s, expected object, got %T", key, current)
			}
			if expand {
				arr, ok := obj[key].([]JSONValue)
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

func ToLen(v JSONValue) (int, bool) {
	switch n := v.(type) {
	case int:
		if n < 0 {
			return 0, false
		}
		return n, true
	case int8:
		if n < 0 {
			return 0, false
		}
		return int(n), true
	case int16:
		if n < 0 {
			return 0, false
		}
		return int(n), true
	case int32:
		if n < 0 {
			return 0, false
		}
		return int(n), true
	case int64:
		if n < 0 || n > math.MaxInt {
			return 0, false
		}
		return int(n), true
	case uint:
		return int(n), true
	case uint8:
		return int(n), true
	case uint16:
		return int(n), true
	case uint32:
		if int(n) < 0 {
			return 0, false
		}
		return int(n), true
	case uint64:
		if n > math.MaxInt {
			return 0, false
		}
		return int(n), true
	case float32:
		if n < 0 {
			return 0, false
		}
		return int(n), true
	case float64:
		if n < 0 {
			return 0, false
		}
		return int(n), true
	case json.Number:
		if val, err := n.Int64(); err != nil || val > math.MaxInt || val < 0 {
			return 0, false
		} else {
			return int(val), true
		}
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
func ToNumber(v JSONValue) (float64, bool) {
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
	case json.Number:
		if val, err := n.Float64(); err != nil {
			return 0, false
		} else {
			return val, true
		}
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

func TypeMatches(value JSONValue, typename string) error {
	switch typename {
	case "number":
		switch value.(type) {
		case float64, json.Number:
			return nil
		}
		return fmt.Errorf("Assertion failed: expected number, got %T", value)
	case "string":
		if _, ok := value.(string); ok {
			return nil
		}
		return fmt.Errorf("Assertion failed: expected string, got %T", value)
	case "boolean":
		if _, ok := value.(bool); ok {
			return nil
		}
		return fmt.Errorf("Assertion failed: expected bool, got %T", value)
	case "object":
		if _, ok := value.(map[string]any); ok {
			return nil
		}
		return fmt.Errorf("Assertion failed: expected object, got %T", value)
	case "array":
		if _, ok := value.([]any); ok {
			return nil
		}
		return fmt.Errorf("Assertion failed: expected array, got %T", value)
	case "null":
		return IsNull(value)
	}
	return fmt.Errorf("Invalid type in assertion, got %T", typename)
}

func Eq(lhs JSONValue, rhs JSONValue) error {
	if reflect.TypeOf(lhs) != reflect.TypeOf(rhs) {
		return fmt.Errorf("Assertion failed: expected types to be same, got (%T, %T)", lhs, rhs)
	}
	switch lhs.(type) {
	case float64, string, bool, json.Number:
		if lhs == rhs {
			return nil
		}
		return fmt.Errorf("Assertion failed: expected %v == %v", lhs, rhs)
	}
	return fmt.Errorf("Assertion failed: expected types to be comparable, got (%T, %T)", lhs, rhs)
}
func Ne(lhs JSONValue, rhs JSONValue) error {
	if reflect.TypeOf(lhs) != reflect.TypeOf(rhs) {
		return fmt.Errorf("Assertion failed: expected types to be same, got (%T, %T)", lhs, rhs)
	}
	switch rhs.(type) {
	case float64, string, bool, json.Number:
		if lhs != rhs {
			return nil
		}
		return fmt.Errorf("Assertion failed: expected %v != %v", lhs, rhs)
	}
	return fmt.Errorf("Assertion failed: expected types to be comparable, got (%T, %T)", lhs, rhs)
}
func Gt(value JSONValue, rhs JSONValue) error {
	if reflect.TypeOf(value) != reflect.TypeOf(rhs) {
		return fmt.Errorf("Assertion failed: expected types to be same, got (%T, %T)", value, rhs)
	}
	switch rhs.(type) {
	case float64, json.Number:
		lval, _ := ToNumber(value)
		rval, _ := ToNumber(rhs)
		if lval > rval {
			return nil
		}
		return fmt.Errorf("Assertion failed: expected %v > %v", value, rhs)
	case string:
		lval := value.(string)
		rval := rhs.(string)
		if lval > rval {
			return nil
		}
		return fmt.Errorf("Assertion failed: expected %v > %v", value, rhs)
	}
	return fmt.Errorf("Assertion failed: expected types to be ordered, got (%T, %T)", value, rhs)
}
func Lt(lhs JSONValue, rhs JSONValue) error {
	if reflect.TypeOf(lhs) != reflect.TypeOf(rhs) {
		return fmt.Errorf("Assertion failed: expected types to be same, got (%T, %T)", lhs, rhs)
	}
	switch rhs.(type) {
	case float64, json.Number:
		lval, _ := ToNumber(lhs)
		rval, _ := ToNumber(rhs)
		if lval < rval {
			return nil
		}
		return fmt.Errorf("Assertion failed: expected %v < %v", lhs, rhs)
	case string:
		lval := lhs.(string)
		rval := rhs.(string)
		if lval < rval {
			return nil
		}
		return fmt.Errorf("Assertion failed: expected %v < %v", lhs, rhs)
	}
	return fmt.Errorf("Assertion failed: expected types to be ordered, got (%T, %T)", lhs, rhs)
}
func Gte(lhs JSONValue, rhs JSONValue) error {
	if reflect.TypeOf(lhs) != reflect.TypeOf(rhs) {
		return fmt.Errorf("Assertion failed: expected types to be same, got (%T, %T)", lhs, rhs)
	}
	switch rhs.(type) {
	case float64, json.Number:
		lval, _ := ToNumber(lhs)
		rval, _ := ToNumber(rhs)
		if lval >= rval {
			return nil
		}
		return fmt.Errorf("Assertion failed: expected %v >= %v", lhs, rhs)
	case string:
		lval := lhs.(string)
		rval := rhs.(string)
		if lval >= rval {
			return nil
		}
		return fmt.Errorf("Assertion failed: expected %v >= %v", lhs, rhs)
	}
	return fmt.Errorf("Assertion failed: expected types to be ordered, got (%T, %T)", lhs, rhs)
}
func Lte(lhs JSONValue, rhs JSONValue) error {
	if reflect.TypeOf(lhs) != reflect.TypeOf(rhs) {
		return fmt.Errorf("Assertion failed: expected types to be same, got (%T, %T)", lhs, rhs)
	}
	switch rhs.(type) {
	case float64, json.Number:
		lval, _ := ToNumber(lhs)
		rval, _ := ToNumber(rhs)
		if lval <= rval {
			return nil
		}
		return fmt.Errorf("Assertion failed: expected %v <= %v", lhs, rhs)
	case string:
		lval := lhs.(string)
		rval := rhs.(string)
		if lval <= rval {
			return nil
		}
		return fmt.Errorf("Assertion failed: expected %v <= %v", lhs, rhs)
	}
	return fmt.Errorf("Assertion failed: expected types to be ordered, got (%T, %T)", lhs, rhs)
}

func IsNull(value JSONValue) error {
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
func IsNotNull(value JSONValue) error {
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

func AllNull(values []JSONValue) error {
	for _, value := range values {
		if err := IsNull(value); err != nil {
			return fmt.Errorf("Assertion failed: expected all values in %v to be null", values)
		}
	}
	return nil
}
func AllNotNull(values []JSONValue) error {
	for _, value := range values {
		if err := IsNotNull(value); err != nil {
			return fmt.Errorf("Assertion failed: expected all values in %v to be non-null", values)
		}
	}
	return nil
}

func Includes(values []JSONValue, target JSONValue) error {
	for _, value := range values {
		if err := Eq(value, target); err == nil {
			return nil
		}
	}
	return fmt.Errorf("Assertion failed: expected %v to be in %v", target, values)
}
func Excludes(values []JSONValue, target JSONValue) error {
	for _, value := range values {
		if err := Eq(value, target); err == nil {
			return fmt.Errorf("Assertion failed: expected %v not to be in %v", value, values)
		}
	}
	return nil
}

func IsSorted(values []JSONValue, desc bool) error {
	last := values[0]
	for i := 1; i < len(values); i++ {
		if desc {
			if err := Gt(last, values[i]); err != nil {
				return fmt.Errorf("Assertion failed: expected sorted ascending array")
			}
		} else {
			if err := Lt(last, values[i]); err != nil {
				return fmt.Errorf("Assertion failed: expected sorted ascending array")
			}
		}
	}
	return nil
}
