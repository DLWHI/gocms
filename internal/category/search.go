package category

import (
	"context"
	"gocms/internal/domain"
	"gocms/internal/search"
)

type SearchParams struct {
	search.Params
	Lang   string `json:"lang,omitempty"`
	Hidden *bool  `json:"hidden,omitempty"`
	Navbar *bool  `json:"navbar,omitempty"`
}

type SearchResult = search.Result[domain.Category]

type SearchEngine interface {
	Index(ctx context.Context, slug string, group ...domain.Category) error
	Get(ctx context.Context, slug string, lang string) (*domain.Category, error)
	Search(ctx context.Context, params *SearchParams) (*SearchResult, error)
	Delete(ctx context.Context, slug string) error
}
