package domain

type Editor struct {
	Locale   string `json:"locale" gorm:"size:8;primaryKey;foreignKey:locales(code);constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Name     string `json:"name"`
	Bio      string `json:"bio"`
	User     string `json:"username" gorm:"primaryKey;size:64;"`
	UserData User   `json:"user_data" gorm:"foreignKey:User;references:Username;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Version  uint   `json:"-" gorm:"default:1"`
}

func (Editor) TableName() string {
	return "editors"
}
