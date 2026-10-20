package domain

type Locale struct {
	Name string `json:"name"`
	Code string `json:"code" gorm:"primaryKey;size:8;"`
}
