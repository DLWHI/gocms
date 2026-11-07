package advertisement

import (
	"context"
	"gocms/internal/domain"
	"gocms/internal/search"
)

type SearchParams struct {
	search.Params
	Expired       *bool    `json:"expired,omitempty"`
	ExpiresAfter  *int64   `json:"expires_after,omitempty"`
	ExpiresBefore *int64   `json:"expires_before,omitempty"`
	Groups        []string `json:"groups,omitempty"`
	Select        []string `json:"select,omitempty"`
}

type SearchResult = search.Result[domain.Advertisement]

type SearchEngine interface {
	Index(ctx context.Context, user *domain.Advertisement) error
	Get(ctx context.Context, id uint) (*domain.Advertisement, error)
	Search(ctx context.Context, filters *SearchParams) (*SearchResult, error)
	Delete(ctx context.Context, id uint) error
}
