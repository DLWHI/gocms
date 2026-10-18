package domain

type Category struct {
	Slug    string `json:"slug" gorm:"size:255;primaryKey"`
	Hidden  bool   `json:"hidden"`
	Navbar  bool   `json:"navbar"`
	Version uint   `json:"-" gorm:"default:1"`
}

type CategoryTranslation struct {
	Locale       string   `json:"locale" gorm:"size:8;primaryKey;foreignKey:locales(code);constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Name         string   `json:"name"`
	CategorySlug string   `json:"slug" gorm:"size:255;primaryKey"`
	Category     Category `json:"category_meta" gorm:"foreignKey:CategorySlug;references:Slug;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Version      uint     `json:"-" gorm:"default:1"`
}

func (Category) TableName() string {
	return "categories"
}

func (CategoryTranslation) TableName() string {
	return "category_translations"
}
