package test

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
)

type Operation = string

type Assertion struct {
	Path  string    `json:"path"`
	Op    Operation `json:"op"`
	Value JSONValue `json:"value"`
}

type ExpectedResult struct {
	Status         int                 `json:"status"`
	BodyType       string              `json:"body"`
	Assertions     []Assertion         `json:"json_assertions"`
	ExternalChecks []ExternalAssertion `json:"external_checks"`
	Eventually     uint                `json:"eventually_ms"`
}

func (r *ExpectedResult) Check(d *Data, res *http.Response) error {
	if r.Status != res.StatusCode {
		return fmt.Errorf("Status code did not match: expected %d, got %d", r.Status, res.StatusCode)
	}
	contentType := res.Header.Get("Content-Type")
	if !IsMachingContentType(contentType, r.BodyType) {
		return fmt.Errorf("Response content type did not match: expected %q, got %q", r.BodyType, contentType)
	}

	if r.BodyType == "json" {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		var values map[string]JSONValue
		if err := json.Unmarshal(body, &values); err != nil {
			return err
		}

		for _, asertion := range r.Assertions {
			if err := asertion.Assert(d, values); err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *Assertion) Assert(d *Data, body map[string]JSONValue) error {
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
		return Eq(values[0], a.Value)
	case "ne":
		return Ne(values[0], a.Value)
	case "lt":
		return Lt(values[0], a.Value)
	case "gt":
		return Gt(values[0], a.Value)
	case "lte":
		return Lte(values[0], a.Value)
	case "gte":
		return Gte(values[0], a.Value)
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
		return Includes(values, a.Value)
	case "excludes":
		return Excludes(values, a.Value)
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

func (a *Assertion) opType(values []JSONValue) error {
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

func (a *Assertion) opLenEq(values []JSONValue) error {
	expected, ok := ToLen(a.Value)
	if !ok {
		return fmt.Errorf("Invalid length value in assertions, got %v, type %T", a.Value, a.Value)
	}
	if expected == len(values) {
		return nil
	}
	return fmt.Errorf("Assertion failed: expected len(%d) == %d", len(values), expected)
}
func (a *Assertion) opLenLt(values []JSONValue) error {
	expected, ok := ToLen(a.Value)
	if !ok {
		return fmt.Errorf("Invalid length value in assertions, got %v, type %T", a.Value, a.Value)
	}
	if expected > len(values) {
		return nil
	}
	return fmt.Errorf("Assertion failed: expected len(%d) > %d", len(values), expected)
}
func (a *Assertion) opLenGt(values []JSONValue) error {
	expected, ok := ToLen(a.Value)
	if !ok {
		return fmt.Errorf("Invalid length value in assertions, got %v, type %T", a.Value, a.Value)
	}
	if expected < len(values) {
		return nil
	}
	return fmt.Errorf("Assertion failed: expected len(%d) < %d", len(values), expected)
}
func (a *Assertion) opLenLte(values []JSONValue) error {
	expected, ok := ToLen(a.Value)
	if !ok {
		return fmt.Errorf("Invalid length value in assertions, got %v, type %T", a.Value, a.Value)
	}
	if expected >= len(values) {
		return nil
	}
	return fmt.Errorf("Assertion failed: expected len(%d) >= %d", len(values), expected)
}
func (a *Assertion) opLenGte(values []JSONValue) error {
	expected, ok := ToLen(a.Value)
	if !ok {
		return fmt.Errorf("Invalid length value in assertions, got %v, type %T", a.Value, a.Value)
	}
	if expected <= len(values) {
		return nil
	}
	return fmt.Errorf("Assertion failed: expected len(%d) <= %d", len(values), expected)
}

func opUnixTimestamp(values []JSONValue) error {
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
