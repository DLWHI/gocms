package lang

import (
	"context"
	"fmt"
	"gocms/internal/domain"
	"strings"
)

const (
	EntityName = "languages"
	Event      = "lang.update"
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

type Language struct {
	Name string `json:"name"`
	Code string `json:"code" gorm:"primaryKey;size:2"`
}

type Repository interface {
	Create(ctx context.Context, lang *Language) error
	GetAll(ctx context.Context) ([]Language, error)
	Delete(ctx context.Context, code string) error
}

type Service interface {
	GetSupported() []Language
	IsSupported(code string) bool
	Add(ctx context.Context, lang *Language) error
	Remove(ctx context.Context, code string) error
}

func NormalizeLangCode(value string) (string, error) {
	normalized := strings.TrimSpace(value)
	if !IsISO639_1Code(normalized) {
		return "", fmt.Errorf("%w: invalid language code", domain.ErrInval)
	}
	return normalized, nil
}

func IsISO639_1Code(value string) bool {
	_, exists := iso6391[value]
	return exists
}

func ResolveLang(ctx context.Context) (string, error) {
	code, ok := ctx.Value(ContextKey).(string)
	if !ok {
		return "", fmt.Errorf("%w: missing language code", domain.ErrInval)
	}
	return code, nil
}

func (*Language) TableName() string {
	return "languages"
}
