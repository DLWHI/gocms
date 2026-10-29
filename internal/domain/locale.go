package domain

const (
	LocaleEntityName = "locale"

	LocaleKey = "locale"
)

type Locale struct {
	Name string `json:"name"`
	Code string `json:"code"`
}
