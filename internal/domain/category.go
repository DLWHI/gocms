package domain

import (
	"fmt"
	"strings"
)

const (
	CategoryActionCreate = 1 << iota
	CategoryActionUpdate
	CategoryActionDelete

	CategoryEventName  = "category"
	CategoryEntityName = "categories"

	MaxCategoryNameLength = 254
)

type Category struct {
	ID     string `json:"id,omitempty"`
	Slug   Slug   `json:"slug"`
	Hidden bool   `json:"hidden"`
	Navbar bool   `json:"navbar"`
	Locale string `json:"locale"`
	Name   string `json:"name"`
}

type CategoryDiff struct {
	Slug   *Slug   `json:"slug"`
	Hidden *bool   `json:"hidden"`
	Navbar *bool   `json:"navbar"`
	Name   *string `json:"name"`
}

type CategoryEvent struct {
	Action int    `json:"action"`
	Slug   string `json:"slug"`
}

func NormalizeCategoryName(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if len(normalized) > MaxCategoryNameLength {
		return "", fmt.Errorf("%w: category name is longer than %d chars", ErrInval, MaxCategoryNameLength)
	}

	return normalized, nil
}
