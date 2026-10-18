package category

import (
	"context"
	"gocms/internal/search"
)

type Model struct {
	Slug    string `json:"slug" gorm:"size:255;primaryKey"`
	Hidden  bool   `json:"hidden"`
	Navbar  bool   `json:"navbar"`
	Version uint   `json:"-" gorm:"default:1"`
}

type Translation struct {
	Locale       string `json:"locale" gorm:"size:8;primaryKey;foreignKey:locales(code);constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Name         string `json:"name"`
	CategorySlug string `json:"slug" gorm:"size:255;primaryKey"`
	Category     Model  `json:"category_meta" gorm:"foreignKey:CategorySlug;references:Slug;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Version      uint   `json:"-" gorm:"default:1"`
}

type Document struct {
	ID     string `json:"id,omitempty"`
	Slug   string `json:"slug"`
	Hidden bool   `json:"hidden"`
	Navbar bool   `json:"navbar"`
	Locale string `json:"locale"`
	Name   string `json:"name"`
}

type MessagePayload struct {
	Slug   string `json:"slug"`
	Locale string `json:"locale"`
	Action string `json:"action"`
}

type SearchParams struct {
	search.Params
	Locale string `json:"locale,omitempty"`
	Hidden *bool  `json:"hidden,omitempty"`
	Navbar *bool  `json:"navbar,omitempty"`
}

type SearchResult = search.Result[Document]

type Repository interface {
	Create(ctx context.Context, category *Translation) error
	Update(ctx context.Context, category *Translation) error
	Get(ctx context.Context, slug string, locale string) (*Translation, error)
	GetAll(ctx context.Context, locale string, hidden *bool) ([]Translation, error)
	Delete(ctx context.Context, slug string) error
}

type SearchEngine interface {
	Index(ctx context.Context, category *Document) error
	Get(ctx context.Context, slug string, locale string) (*Document, error)
	Search(ctx context.Context, params *SearchParams) (*SearchResult, error)
	Delete(ctx context.Context, slug string) error
}

func (Model) TableName() string {
	return "categories"
}

func (Translation) TableName() string {
	return "category_translations"
}

func (cm *Translation) ToDoc() Document {
	var cd Document

	cd.Slug = cm.Category.Slug
	cd.Hidden = cm.Category.Hidden
	cd.Navbar = cm.Category.Navbar
	cd.Name = cm.Name
	cd.Locale = cm.Locale
	return cd
}

func (cd *Document) ToModel() Translation {
	var cm Translation

	cm.CategorySlug = cd.Slug
	cm.Name = cd.Name
	cm.Locale = cd.Locale
	cm.Category = Model{
		Slug:   cd.Slug,
		Hidden: cd.Hidden,
		Navbar: cd.Navbar,
	}
	return cm
}
