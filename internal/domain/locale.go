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

var iso6391 = map[string]struct{}{
	"aa": {}, "ab": {}, "ae": {}, "af": {}, "ak": {}, "am": {}, "an": {}, "ar": {},
	"as": {}, "av": {}, "ay": {}, "az": {}, "ba": {}, "be": {}, "bg": {}, "bi": {},
	"bm": {}, "bn": {}, "bo": {}, "br": {}, "bs": {}, "ca": {}, "ce": {}, "ch": {},
	"co": {}, "cr": {}, "cs": {}, "cu": {}, "cv": {}, "cy": {}, "da": {}, "de": {},
	"dv": {}, "dz": {}, "ee": {}, "el": {}, "en": {}, "eo": {}, "es": {}, "et": {},
	"eu": {}, "fa": {}, "ff": {}, "fi": {}, "fj": {}, "fo": {}, "fr": {}, "fy": {},
	"ga": {}, "gd": {}, "gl": {}, "gn": {}, "gu": {}, "gv": {}, "ha": {}, "he": {},
	"hi": {}, "ho": {}, "hr": {}, "ht": {}, "hu": {}, "hy": {}, "hz": {}, "ia": {},
	"id": {}, "ie": {}, "ig": {}, "ii": {}, "ik": {}, "io": {}, "is": {}, "it": {},
	"iu": {}, "ja": {}, "jv": {}, "ka": {}, "kg": {}, "ki": {}, "kj": {}, "kk": {},
	"kl": {}, "km": {}, "kn": {}, "ko": {}, "kr": {}, "ks": {}, "ku": {}, "kv": {},
	"kw": {}, "ky": {}, "la": {}, "lb": {}, "lg": {}, "li": {}, "ln": {}, "lo": {},
	"lt": {}, "lu": {}, "lv": {}, "mg": {}, "mh": {}, "mi": {}, "mk": {}, "ml": {},
	"mn": {}, "mr": {}, "ms": {}, "mt": {}, "my": {}, "na": {}, "nb": {}, "nd": {},
	"ne": {}, "ng": {}, "nl": {}, "nn": {}, "no": {}, "nr": {}, "nv": {}, "ny": {},
	"oc": {}, "oj": {}, "om": {}, "or": {}, "os": {}, "pa": {}, "pi": {}, "pl": {},
	"ps": {}, "pt": {}, "qu": {}, "rm": {}, "rn": {}, "ro": {}, "ru": {}, "rw": {},
	"sa": {}, "sc": {}, "sd": {}, "se": {}, "sg": {}, "si": {}, "sk": {}, "sl": {},
	"sm": {}, "sn": {}, "so": {}, "sq": {}, "sr": {}, "ss": {}, "st": {}, "su": {},
	"sv": {}, "sw": {}, "ta": {}, "te": {}, "tg": {}, "th": {}, "ti": {}, "tk": {},
	"tl": {}, "tn": {}, "to": {}, "tr": {}, "ts": {}, "tt": {}, "tw": {}, "ty": {},
	"ug": {}, "uk": {}, "ur": {}, "uz": {}, "ve": {}, "vi": {}, "vo": {}, "wa": {},
	"wo": {}, "xh": {}, "yi": {}, "yo": {}, "za": {}, "zh": {}, "zu": {},
}

type Locale struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

func ResolveLocale(ctx context.Context, value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		locale, ok := ctx.Value(LocaleKey).(string)
		if !ok || locale == "" {
			return "", fmt.Errorf("%w: empty or invalid locale", ErrInval)
		}
		return locale, nil
	}
	_, exists := iso6391[normalized]
	if !exists {
		return "", fmt.Errorf("%w: invalid locale", ErrInval)
	}
	return normalized, nil
}

func NormalizeLocaleCode(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	_, exists := iso6391[normalized]
	if !exists {
		return "", fmt.Errorf("%w: invalid locale", ErrInval)
	}
	return "", nil
}
