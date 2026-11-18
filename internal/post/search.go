package post

import (
	"context"
	"gocms/internal/domain"
	"gocms/internal/search"
)

type SearchParams struct {
	search.Params
	Categories        []string `json:"categories,omitempty"`
	ExcludeCategories []string `json:"exclude_categories,omitempty"`
	Authors           []string `json:"authors,omitempty"`
	CreatedBy         *string  `json:"created_by,omitempty"`
	CreatedAfter      *int64   `json:"created_after,omitempty"`
	CreatedBefore     *int64   `json:"created_before,omitempty"`
	PublishedAfter    *int64   `json:"published_after,omitempty"`
	PublishedBefore   *int64   `json:"published_before,omitempty"`
	Published         *bool    `json:"published,omitempty"`
}

type SearchResult = search.Result[domain.Post]

type SearchEngine interface {
	Index(ctx context.Context, group ...domain.Post) error
	Get(ctx context.Context, id uint, language string) (*domain.Post, error)
	Search(ctx context.Context, params *SearchParams) (*SearchResult, error)
	SearchLang(ctx context.Context, language string, params *SearchParams) (*SearchResult, error)
	Delete(ctx context.Context, id uint) error
}

func ToMetaResults(posts *SearchResult) *search.Result[domain.PostMeta] {
	res := search.Result[domain.PostMeta]{
		Results:  make([]domain.PostMeta, 0, len(posts.Results)),
		LastPage: posts.LastPage,
		Total:    posts.Total,
	}

	for _, post := range posts.Results {
		res.Results = append(res.Results, *post.ToMeta())
	}

	return &res
}
