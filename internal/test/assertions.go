package test

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
	"reflect"
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

func (a *Assertion) Assert(d *Data, body map[string]Value) error {
	values, err := ResolvePath(body, a.Path)
	if err != nil {
		return err
	}
	if len(values) < 1 {
		return fmt.Errorf("Expected at least one value, got nothing")
	}
	expected, err := d.Resolve(a.Value)
	if err != nil {
		return fmt.Errorf("Invalid expect value submitted to assertion: %w", err)
	}
	a.Value = expected
	switch a.Op {
	case "is_null":
		return IsNull(values[0])
	case "not_null":
		return IsNotNull(values[0])
	case "eq":
		return a.opEq(values[0])
	case "ne":
		return a.opNe(values[0])
	case "lt":
		return a.opLt(values[0])
	case "gt":
		return a.opGt(values[0])
	case "lte":
		return a.opLte(values[0])
	case "gte":
		return a.opGte(values[0])
	case "type":
		return a.opType(values)
	case "len_eq":
		return a.opLenEq(values)
	case "len_lt":
		return a.opLenLt(values)
	case "len_gt":
		return a.opLenGt(values)
	case "len_lte":
		return a.opLenLte(values)
	case "len_gte":
		return a.opLenGte(values)
	case "includes":
		return a.opIncludes(values)
	case "excludes":
		return a.opExcludes(values)
	case "all_null":
		return AllNull(values)
	case "all_not_null":
		return AllNotNull(values)
	case "unix_or_null":
		return opUnixTimestamp(values)
	case "sorted_asc":
		return IsSorted(values, false)
	case "sorted_desc":
		return IsSorted(values, true)
	default:
		return fmt.Errorf("unknown operation: %s", "op")
	}
}

func (a *Assertion) opEq(value Value) error {
	if reflect.TypeOf(value) != reflect.TypeOf(a.Value) {
		return fmt.Errorf("Assertion failed, expected types to be same, got (%T, %T)", value, a.Value)
	}
	switch a.Value.(type) {
	case float64, string, bool:
		if value == a.Value {
			return nil
		}
		return fmt.Errorf("Assertion failed, expected %v == %v", value, a.Value)
	}
	return fmt.Errorf("Assertion failed, expected types to be comparable, got (%T, %T)", value, a.Value)
}
func (a *Assertion) opNe(value Value) error {
	if reflect.TypeOf(value) != reflect.TypeOf(a.Value) {
		return fmt.Errorf("Assertion failed, expected types to be same, got (%T, %T)", value, a.Value)
	}
	switch a.Value.(type) {
	case float64, string, bool:
		if value != a.Value {
			return nil
		}
		return fmt.Errorf("Assertion failed, expected %v != %v", value, a.Value)
	}
	return fmt.Errorf("Assertion failed, expected types to be comparable, got (%T, %T)", value, a.Value)
}
func (a *Assertion) opGt(value Value) error {
	if reflect.TypeOf(value) != reflect.TypeOf(a.Value) {
		return fmt.Errorf("Assertion failed, expected types to be same, got (%T, %T)", value, a.Value)
	}
	switch a.Value.(type) {
	case float64:
		lval, _ := ToNumber(value)
		rval, _ := ToNumber(a.Value)
		if lval > rval {
			return nil
		}
		return fmt.Errorf("Assertion failed, expected %v > %v", value, a.Value)
	case string:
		lval := value.(string)
		rval := a.Value.(string)
		if lval > rval {
			return nil
		}
		return fmt.Errorf("Assertion failed, expected %v > %v", value, a.Value)
	}
	return fmt.Errorf("Assertion failed, expected types to be ordered, got (%T, %T)", value, a.Value)
}
func (a *Assertion) opLt(value Value) error {
	if reflect.TypeOf(value) != reflect.TypeOf(a.Value) {
		return fmt.Errorf("Assertion failed, expected types to be same, got (%T, %T)", value, a.Value)
	}
	switch a.Value.(type) {
	case float64:
		lval, _ := ToNumber(value)
		rval, _ := ToNumber(a.Value)
		if lval < rval {
			return nil
		}
		return fmt.Errorf("Assertion failed, expected %v < %v", value, a.Value)
	case string:
		lval := value.(string)
		rval := a.Value.(string)
		if lval < rval {
			return nil
		}
		return fmt.Errorf("Assertion failed, expected %v < %v", value, a.Value)
	}
	return fmt.Errorf("Assertion failed, expected types to be ordered, got (%T, %T)", value, a.Value)
}
func (a *Assertion) opGte(value Value) error {
	if reflect.TypeOf(value) != reflect.TypeOf(a.Value) {
		return fmt.Errorf("Assertion failed, expected types to be same, got (%T, %T)", value, a.Value)
	}
	switch a.Value.(type) {
	case float64:
		lval, _ := ToNumber(value)
		rval, _ := ToNumber(a.Value)
		if lval >= rval {
			return nil
		}
		return fmt.Errorf("Assertion failed, expected %v >= %v", value, a.Value)
	case string:
		lval := value.(string)
		rval := a.Value.(string)
		if lval >= rval {
			return nil
		}
		return fmt.Errorf("Assertion failed, expected %v >= %v", value, a.Value)
	}
	return fmt.Errorf("Assertion failed, expected types to be ordered, got (%T, %T)", value, a.Value)
}
func (a *Assertion) opLte(value Value) error {
	if reflect.TypeOf(value) != reflect.TypeOf(a.Value) {
		return fmt.Errorf("Assertion failed, expected types to be same, got (%T, %T)", value, a.Value)
	}
	switch a.Value.(type) {
	case float64:
		lval, _ := ToNumber(value)
		rval, _ := ToNumber(a.Value)
		if lval <= rval {
			return nil
		}
		return fmt.Errorf("Assertion failed, expected %v <= %v", value, a.Value)
	case string:
		lval := value.(string)
		rval := a.Value.(string)
		if lval <= rval {
			return nil
		}
		return fmt.Errorf("Assertion failed, expected %v <= %v", value, a.Value)
	}
	return fmt.Errorf("Assertion failed, expected types to be ordered, got (%T, %T)", value, a.Value)
}

func (a *Assertion) opType(values []Value) error {
	typename, ok := a.Value.(string)
	if !ok {
		return fmt.Errorf("Invalid type value in assertion, got %T", a.Value)
	}
	for _, value := range values {
		if err := TypeMatches(value, typename); err != nil {
			return err
		}
	}
	return nil
}

func (a *Assertion) opLenEq(values []Value) error {
	expected, ok := ToLen(a.Value)
	if !ok {
		return fmt.Errorf("Invalid length value in assertions, got %v, type %T", a.Value, a.Value)
	}
	if expected == len(values) {
		return nil
	}
	return fmt.Errorf("Assertion failed: expected len(%d) == %d", len(values), expected)
}
func (a *Assertion) opLenLt(values []Value) error {
	expected, ok := ToLen(a.Value)
	if !ok {
		return fmt.Errorf("Invalid length value in assertions, got %v, type %T", a.Value, a.Value)
	}
	if expected > len(values) {
		return nil
	}
	return fmt.Errorf("Assertion failed: expected len(%d) > %d", len(values), expected)
}
func (a *Assertion) opLenGt(values []Value) error {
	expected, ok := ToLen(a.Value)
	if !ok {
		return fmt.Errorf("Invalid length value in assertions, got %v, type %T", a.Value, a.Value)
	}
	if expected < len(values) {
		return nil
	}
	return fmt.Errorf("Assertion failed: expected len(%d) < %d", len(values), expected)
}
func (a *Assertion) opLenLte(values []Value) error {
	expected, ok := ToLen(a.Value)
	if !ok {
		return fmt.Errorf("Invalid length value in assertions, got %v, type %T", a.Value, a.Value)
	}
	if expected >= len(values) {
		return nil
	}
	return fmt.Errorf("Assertion failed: expected len(%d) >= %d", len(values), expected)
}
func (a *Assertion) opLenGte(values []Value) error {
	expected, ok := ToLen(a.Value)
	if !ok {
		return fmt.Errorf("Invalid length value in assertions, got %v, type %T", a.Value, a.Value)
	}
	if expected <= len(values) {
		return nil
	}
	return fmt.Errorf("Assertion failed: expected len(%d) <= %d", len(values), expected)
}

func (a *Assertion) opIncludes(values []Value) error {
	for _, value := range values {
		if err := a.opEq(value); err == nil {
			return nil
		}
	}
	return fmt.Errorf("Assertion failed: expected %v to be in %v", a.Value, values)
}
func (a *Assertion) opExcludes(values []Value) error {
	for _, value := range values {
		if err := a.opEq(value); err != nil {
			return nil
		}
	}
	return fmt.Errorf("Assertion failed: expected %v not to be in %v", a.Value, values)
}

func opUnixTimestamp(values []Value) error {
	for _, value := range values {
		number, ok := ToNumber(value)
		if ok {
			if math.Mod(number, 1) != 0 {
				return fmt.Errorf("Assertion failed: expected all integer or null in %v", values)
			}
		} else {
			if err := IsNull(value); err != nil {
				return fmt.Errorf("Assertion failed: expected all integer or null in %v", values)
			}
		}
	}
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
