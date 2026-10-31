package domain

import (
	"context"
	"fmt"
	"strings"
)

const (
	LocaleEntityName = "locales"

	LocaleKey = "locale"

	LocaleMaxNameLength = 64
)

type Locale struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

func ResolveLocale(ctx context.Context, value string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if normalized == "" {
		locale, ok := ctx.Value(LocaleKey).(string)
		if !ok {
			return "", fmt.Errorf("%w: empty locale", ErrInval)
		}
		return locale, nil
	}
	return normalized, nil
}

func NormalizeCode(value string) (string, error) {
	normalized := strings.TrimSpace(value)

	return normalized, nil
}
