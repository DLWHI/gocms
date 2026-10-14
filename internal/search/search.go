package search

import "context"

type Params struct {
	SortBy string `json:"sort_by,omitempty"`
	Query  string `json:"query,omitempty"`
	Page   int64  `json:"page,omitempty"`
	Count  int64  `json:"count,omitempty"`
}

type Result[T any] struct {
	Results  []T   `json:"results"`
	LastPage int64 `json:"last_page"`
	Total    int64 `json:"total"`
}

type SearchEngine[T any] interface {
	Index(ctx context.Context, data *T) error
	Get(ctx context.Context, id string) (*T, error)
	Search(ctx context.Context, params *Params) (*Result[T], error)
	Delete(ctx context.Context, id string) error
	DeletePattern(ctx context.Context, pattern string) error
}

func ClampParams(params *Params) {
	if params.Count < 1 {
		params.Count = 1
	} else if params.Count > 100 {
		params.Count = 100
	}

	if params.Page < 1 {
		params.Page = 1
	}
}
