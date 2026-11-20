package editor

import (
	"context"
	"fmt"
	"gocms/internal/domain"
	"gocms/internal/search"
)

type SearchDoc struct {
	ID        string `json:"id,omitempty"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Lang      string `json:"lang"`
	DeletedAt *int64 `json:"deleted_at,omitempty"`
	Owner     string `json:"owner"`
}

type SearchParams struct {
	search.Params
	Owner         string `json:"owner"`
	Deleted       *bool  `json:"deleted"`
	DeletedAfter  *int64 `json:"deleted_after"`
	DeletedBefore *int64 `json:"deleted_before"`
}

type SearchResult = search.Result[SearchDoc]

type SearchEngine interface {
	Index(ctx context.Context, slug string, group ...domain.Editor) error
	Get(ctx context.Context, username string, language string) (*domain.Editor, error)
	Search(ctx context.Context, filters *SearchParams) (*SearchResult, error)
	SearchLang(ctx context.Context, language string, filters *SearchParams) (*SearchResult, error)
	Delete(ctx context.Context, username string) error
}

func (d *SearchDoc) ToDomain() *domain.Editor {
	res := domain.Editor{
		Slug:  d.Slug,
		Bio:   "",
		Name:  d.Name,
		Lang:  d.Lang,
		Owner: d.Owner,
	}
	if d.DeletedAt != nil {
		t := *d.DeletedAt
		res.DeletedAt = &t
	}
	return &res
}

func ToSearchDoc(data *domain.Editor) *SearchDoc {
	res := SearchDoc{
		ID:    fmt.Sprintf("%s-%s", data.Slug, data.Lang),
		Slug:  data.Slug,
		Name:  data.Name,
		Lang:  data.Lang,
		Owner: data.Owner,
	}
	if data.DeletedAt != nil {
		t := *data.DeletedAt
		res.DeletedAt = &t
	}
	return &res
}
