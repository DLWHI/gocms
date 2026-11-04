package domain

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

const (
	MaxSlugLength = 128
	MaxNameLength = 254
)

func IsInvalidSlugRune(r rune) bool {
	isAlpha := 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z'
	isDigit := '0' <= r && r <= '9'
	return !(isAlpha || isDigit || r == '-' || r == '_')
}

func IsNonGraphic(r rune) bool {
	return !unicode.IsPrint(r)
}

func NormalizeSlug(value string) (string, error) {
	normalized := strings.TrimFunc(value, func(r rune) bool { return unicode.IsSpace(r) || r == '-' || r == '_' })
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
	if strings.IndexFunc(normalized, IsNonGraphic) >= 0 {
		return "", fmt.Errorf("%w: name must not contain non printable characters", ErrInval)
	}

	return normalized, nil
}

func NormalizeURL(value string) (string, error) {
	u, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInval, err)
	}
	return u.String(), nil
}
