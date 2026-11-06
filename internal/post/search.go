package post

import (
	"context"
	"gocms/internal/domain"
	"gocms/internal/search"
	"time"
)

type SearchParams struct {
	search.Params
	Categories        []string   `json:"categories,omitempty"`
	ExcludeCategories []string   `json:"exclude_categories,omitempty"`
	Authors           []string   `json:"authors,omitempty"`
	CreatedAfter      *time.Time `json:"created_after,omitempty"`
	CreatedBefore     *time.Time `json:"created_before,omitempty"`
	PublishedAfter    *time.Time `json:"published_after,omitempty"`
	PublishedBefore   *time.Time `json:"published_before,omitempty"`
	Published         *bool      `json:"published,omitempty"`
}

type SearchResult = search.Result[domain.Post]

type SearchEngine interface {
	Index(ctx context.Context, id uint, group ...domain.Post) error
	Get(ctx context.Context, id uint, language string) (*domain.Post, error)
	Search(ctx context.Context, language string, post *SearchParams) (*SearchResult, error)
	Delete(ctx context.Context, id uint) error
}
