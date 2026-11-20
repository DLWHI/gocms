package post

import (
	"context"
	"fmt"
	"gocms/internal/domain"
	"gocms/internal/search"
)

type SearchDoc struct {
	ID          string   `json:"id"`
	PostID      uint     `json:"post_id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Slug        string   `json:"slug"`
	Lang        string   `json:"lang"`
	CreatedAt   int64    `json:"created_at,omitempty"`
	PublishedAt *int64   `json:"published_at"`
	Views       uint     `json:"views"`
	Preview     string   `json:"preview"`
	Author      *string  `json:"author"`
	CreatedBy   string   `json:"created_by,omitempty"`
	Categories  []string `json:"categories"`
}

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

type SearchResult = search.Result[SearchDoc]

type SearchEngine interface {
	Index(ctx context.Context, group ...domain.Post) error
	Get(ctx context.Context, id uint, language string) (*domain.Post, error)
	Search(ctx context.Context, params *SearchParams) (*SearchResult, error)
	SearchLang(ctx context.Context, language string, params *SearchParams) (*SearchResult, error)
	Delete(ctx context.Context, id uint) error
}

func ToSearchDoc(data *domain.Post) *SearchDoc {
	res := SearchDoc{
		ID:          fmt.Sprintf("%d-%s", data.PostID, data.Lang),
		PostID:      data.PostID,
		Title:       data.Title,
		Description: data.Description,
		Slug:        data.Slug,
		Lang:        data.Lang,
		CreatedAt:   data.CreatedAt,
		PublishedAt: new(int64),
		Views:       data.Views,
		Preview:     data.Preview,
		Author:      new(string),
		CreatedBy:   data.CreatedBy,
		Categories:  make([]string, len(data.Categories)),
	}
	if data.PublishedAt != nil {
		t := *data.PublishedAt
		res.PublishedAt = &t
	}
	if data.Author != nil {
		a := *data.Author
		res.Author = &a
	}
	copy(res.Categories, data.Categories)
	return &res
}
