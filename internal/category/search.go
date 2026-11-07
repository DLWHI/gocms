package category

import (
	"context"
	"gocms/internal/domain"
	"gocms/internal/search"
)

type SearchParams struct {
	search.Params
	Hidden *bool `json:"hidden,omitempty"`
	Navbar *bool `json:"navbar,omitempty"`
}

type SearchResult = search.Result[domain.Category]

type SearchEngine interface {
	Index(ctx context.Context, slug string, group ...domain.Category) error
	Get(ctx context.Context, slug string, language string) (*domain.Category, error)
	Search(ctx context.Context, language string, params *SearchParams) (*SearchResult, error)
	Delete(ctx context.Context, slug string) error
}
