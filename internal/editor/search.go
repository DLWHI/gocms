package editor

import (
	"context"
	"gocms/internal/domain"
	"gocms/internal/search"
)

type SearchParams struct {
	search.Params
	Locale        string `json:"locale"`
	Owner         string `json:"owner"`
	Deleted       *bool  `json:"deleted"`
	DeletedAfter  *int64 `json:"deleted_after"`
	DeletedBefore *int64 `json:"deleted_before"`
}

type SearchResult = search.Result[domain.Editor]

type SearchEngine interface {
	Index(ctx context.Context, slug string, group ...domain.Editor) error
	Get(ctx context.Context, username string, locale string) (*domain.Editor, error)
	Search(ctx context.Context, filters *SearchParams) (*SearchResult, error)
	Delete(ctx context.Context, username string) error
}
