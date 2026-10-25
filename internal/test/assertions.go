package test

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"reflect"
	"strings"
)

type Operation = string

type Assertion struct {
	Path  string    `json:"path"`
	Op    Operation `json:"op"`
	Value Value     `json:"value"`
}

type ExpectedResult struct {
	Status         int                 `json:"status"`
	BodyType       string              `json:"body"`
	Assertions     []Assertion         `json:"json_assertions"`
	ExternalChecks []ExternalAssertion `json:"external_checks"`
	Eventually     uint                `json:"eventually_ms"`
}

type ExternalAssertion struct {
	Check    string `json:"check"`
	Username string `json:"username"`
	Password string `json:"password"`
	Expect   string `json:"expect"`
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

func (expected *ExpectedResult) Check(e *Data, res *http.Response) error {
	if expected.Status != res.StatusCode {
		return fmt.Errorf("Status code did not match: expected %d, got %d", expected.Status, res.StatusCode)
	}
	contentType := res.Header.Get("Content-Type")
	if !IsMachingContentType(contentType, expected.BodyType) {
		return fmt.Errorf("Response content type did not match: expected %q, got %q", expected.BodyType, contentType)
	}

	if expected.BodyType == "json" {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		var values map[string]Value
		if err := json.Unmarshal(body, &values); err != nil {
			return err
		}

		for _, as := range expected.Assertions {
			if err := as.Assert(e, values); err != nil {
				return err
			}
		}
	}
	// assert external checks too
	return nil
}

func (as *Assertion) Assert(e *Data, body map[string]Value) error {
	// values, err := ResolvePath(body, as.Path)
	// if err != nil {
	// 	return err
	// }
	// switch as.Op {
	// case "is_null":
	// 	return OpNull(values, true)
	// case "not_null":
	// 	return OpNull(values, false)
	// case "eq", "ne":
	// 	return OpCompare(e, values, as.Value, as.Op)
	// case "type":
	// 	for _, v := range values {
	// 		if !typeMatches(v, arg) {
	// 			return fmt.Errorf("assertion failed: expected type %s, got %T", arg, v)
	// 		}
	// 	}
	// case "len_eq":
	// 	expectedLen, err := strconv.Atoi(arg)
	// 	if err != nil {
	// 		return fmt.Errorf("invalid len_eq argument: %v", err)
	// 	}
	// 	for _, v := range values {
	// 		if getLength(v) != expectedLen {
	// 			return fmt.Errorf("assertion failed: expected length %d, got %d", expectedLen, getLength(v))
	// 		}
	// 	}
	// case "len_lte":
	// 	expectedLen, err := strconv.Atoi(arg)
	// 	if err != nil {
	// 		return fmt.Errorf("invalid len_lte argument: %v", err)
	// 	}
	// 	for _, v := range values {
	// 		if getLength(v) > expectedLen {
	// 			return fmt.Errorf("assertion failed: expected length <= %d, got %d", expectedLen, getLength(v))
	// 		}
	// 	}
	// case "gte":
	// 	expectedVal, err := strconv.ParseFloat(arg, 64)
	// 	if err != nil {
	// 		return fmt.Errorf("invalid gte argument: %v", err)
	// 	}
	// 	for _, v := range values {
	// 		f, err := toFloat(v)
	// 		if err != nil || f < expectedVal {
	// 			return fmt.Errorf("assertion failed: expected >= %v, got %v", expectedVal, v)
	// 		}
	// 	}
	// case "includes":
	// 	for _, v := range values {
	// 		if !containsValue(v, arg) {
	// 			return fmt.Errorf("assertion failed: expected %v to include %v", v, arg)
	// 		}
	// 	}
	// case "excludes":
	// 	for _, v := range values {
	// 		if containsValue(v, arg) {
	// 			return fmt.Errorf("assertion failed: expected %v to exclude %v", v, arg)
	// 		}
	// 	}
	// case "all_null":
	// 	for _, v := range values {
	// 		slice, ok := toSlice(v)
	// 		if !ok {
	// 			return fmt.Errorf("all_null requires a list, got %T", v)
	// 		}
	// 		for _, item := range slice {
	// 			if item != nil && !isNil(item) {
	// 				return fmt.Errorf("assertion failed: all_null failed, found non-null element %v", item)
	// 			}
	// 		}
	// 	}
	// case "unix_or_null":
	// 	for _, v := range values {
	// 		slice, ok := toSlice(v)
	// 		if !ok {
	// 			return fmt.Errorf("unix_or_null requires a list, got %T", v)
	// 		}
	// 		for _, item := range slice {
	// 			if item != nil && !isNil(item) && !isInteger(item) {
	// 				return fmt.Errorf("assertion failed: unix_or_null failed, found non-integer element %v (%T)", item, item)
	// 			}
	// 		}
	// 	}
	// case "sorted_asc":
	// 	for _, v := range values {
	// 		slice, ok := toSlice(v)
	// 		if !ok {
	// 			return fmt.Errorf("sorted_asc requires a list, got %T", v)
	// 		}
	// 		if !isSortedAsc(slice) {
	// 			return fmt.Errorf("assertion failed: list is not sorted ascending")
	// 		}
	// 	}
	// case "sorted_desc":
	// 	for _, v := range values {
	// 		slice, ok := toSlice(v)
	// 		if !ok {
	// 			return fmt.Errorf("sorted_desc requires a list, got %T", v)
	// 		}
	// 		if !isSortedDesc(slice) {
	// 			return fmt.Errorf("assertion failed: list is not sorted descending")
	// 		}
	// 	}
	// default:
	// 	return fmt.Errorf("unknown operation: %s", op)
	// }
	return nil
}

func OpCompare(e *Data, values []Value, rhs Value, op string) error {
	// for _, val := range values {
	// 	switch typed := rhs.(type) {
	// 	case int, float64:
	// 		switch op {
	// 		case "eq":
	// 			return lhs == typed
	// 		case "ne":
	// 			return lhs != typed
	// 		case "le":
	// 			return lhs < typed
	// 		case "gt":
	// 			return lhs > typed
	// 		case "lte":
	// 			return lhs <= typed
	// 		case "gte":
	// 			return lhs >= typed
	// 		default:
	// 			return false
	// 		}
	// 	}
	// 	switch op {
	// 	case "eq":
	// 		if fmt.Sprint(v) != value {
	// 			return fmt.Errorf("Assertion failed: expected %v == %s", v, value)
	// 		}
	// 	case "ne":
	// 		if fmt.Sprint(v) == value {
	// 			return fmt.Errorf("Assertion failed: expected %v != %s", v, value)
	// 		}
	// 	}
	// }
	return nil
}

// if positive == true, checks if values[0] == nil. Checks if values[0] != nil otherwise
func OpNull(values []Value, positive bool) error {
	if len(values) < 1 {
		return fmt.Errorf("Assertion failed: expected value, got nothing")
	}

	rv := reflect.ValueOf(values[0])
	switch rv.Kind() {
	case reflect.Invalid:
		if !positive {
			return fmt.Errorf("Assertion failed: expected %v != nil", values[0])
		}
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		if positive {
			if !rv.IsNil() {
				return fmt.Errorf("Assertion failed: expected %v == nil", values[0])
			}
		} else {
			if rv.IsNil() {
				return fmt.Errorf("Assertion failed: expected %v != nil", values[0])
			}
		}
	default:
		if positive {
			return fmt.Errorf("Assertion failed: expected %v == nil", values[0])
		}
	}
	return nil
}

func OpSorted(values []Value, desc bool) error {
	// if len(values) < 1 {
	// 	return fmt.Errorf("Assertion failed: expected non-empty array")
	// }

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

func IsMachingContentType(contentType string, expected string) bool {
	mediaType, _, _ := mime.ParseMediaType(contentType)
	switch {
	case expected == "plain_text" && mediaType == "text/plain":
		return true
	case expected == "empty" && mediaType == "":
		return true
	case expected == "json" && mediaType == "application/json":
		return true
	case expected == "any":
		return true
	default:
		return false
	}
}
