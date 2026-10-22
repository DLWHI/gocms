package user

import (
	"context"
	"gocms/internal/domain"
	"gocms/internal/search"
)

type SearchParams struct {
	search.Params
	Permissions     *uint32 `json:"permissions"`
	Suspended       *bool   `json:"suspended"`
	SuspendedAfter  *int64  `json:"suspended_after"`
	SuspendedBefore *int64  `json:"suspended_before"`
	LastLoginAfter  *int64  `json:"last_login_after"`
	LastLoginBefore *int64  `json:"last_login_before"`
}

type SearchResult = search.Result[domain.User]

type SearchEngine interface {
	Index(ctx context.Context, user *domain.User) error
	Get(ctx context.Context, username string) (*domain.User, error)
	Search(ctx context.Context, filters *SearchParams) (*SearchResult, error)
	Delete(ctx context.Context, username string) error
}
