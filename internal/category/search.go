package category

import (
	"context"
	"fmt"
	"gocms/internal/domain"
	"gocms/internal/search"
)

type SearchDoc struct {
	ID     string `json:"id,omitempty"`
	Slug   string `json:"slug"`
	Hidden bool   `json:"hidden"`
	Navbar bool   `json:"navbar"`
	Lang   string `json:"lang"`
	Name   string `json:"name"`
}

type SearchParams struct {
	search.Params
	Hidden *bool `json:"hidden,omitempty"`
	Navbar *bool `json:"navbar,omitempty"`
}

type SearchResult = search.Result[SearchDoc]

type SearchEngine interface {
	Index(ctx context.Context, slug string, group ...domain.Category) error
	Get(ctx context.Context, slug string, language string) (*domain.Category, error)
	Search(ctx context.Context, params *SearchParams) (*SearchResult, error)
	SearchLang(ctx context.Context, language string, params *SearchParams) (*SearchResult, error)
	Delete(ctx context.Context, slug string) error
}

func (d *SearchDoc) ToDomain() *domain.Category {
	res := domain.Category{
		Slug:   d.Slug,
		Hidden: d.Hidden,
		Navbar: d.Navbar,
		Lang:   d.Lang,
		Name:   d.Name,
	}
	return &res
}

func ToSearchDoc(data *domain.Category) *SearchDoc {
	res := SearchDoc{
		ID:     fmt.Sprintf("%s-%s", data.Slug, data.Lang),
		Slug:   data.Slug,
		Hidden: data.Hidden,
		Navbar: data.Navbar,
		Lang:   data.Lang,
		Name:   data.Name,
	}
	return &res
}
