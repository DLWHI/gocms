package domain

import (
	"fmt"
	"strings"
)

const (
	MaxSlugLength = 256
	MaxNameLength = 254
)

func IsInvalidSlugRune(r rune) bool {
	isAlpha := 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z'
	isDigit := '0' <= r && r <= '9'
	return !(isAlpha || isDigit || r == '-' || r == '_')
}

func NormalizeSlug(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", fmt.Errorf("%w: slug cannot be empty", ErrInval)
	}
	if len(normalized) > MaxSlugLength {
		return "", fmt.Errorf("%w: slug is longer than %d chars", ErrInval, MaxSlugLength)
	}
	if strings.IndexFunc(normalized, IsInvalidSlugRune) >= 0 {
		return "", fmt.Errorf("%w: slug must contain only letters, numbers, dashes or underscores", ErrInval)
	}

	return strings.ToLower(normalized), nil
}

func NormalizeName(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if len(normalized) > MaxNameLength {
		return "", fmt.Errorf("%w: name is longer than %d chars", ErrInval, MaxNameLength)
	}

	return normalized, nil
}
