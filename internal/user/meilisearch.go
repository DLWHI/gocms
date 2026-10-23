package user

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/domain"
	"gocms/internal/search"
	"net/http"
	"strings"

	"github.com/meilisearch/meilisearch-go"
)

type MeilisearchEngine struct {
	search search.MeilisearchClient
}

func NewMeilisearchEngine(client meilisearch.ServiceManager, index string) *MeilisearchEngine {
	return &MeilisearchEngine{
		search: search.MeilisearchClient{
			Client: client,
			Config: meilisearch.IndexConfig{
				Uid:        index,
				PrimaryKey: "username",
			},
			Settings: meilisearch.Settings{
				SearchableAttributes: []string{"username", "email"},
				FilterableAttributes: []string{"permissions", "suspended_at", "last_login"},
				SortableAttributes:   []string{"username", "suspended_at", "last_login"},
			},
		},
	}
}

func (ms *MeilisearchEngine) buildFilter(params *SearchParams) string {
	var clauses []string

	if params.Permissions != nil {
		clauses = append(clauses, fmt.Sprintf("permissions = %d", *params.Permissions))
	}

	if params.Suspended != nil {
		if *params.Suspended {
			clauses = append(clauses, "suspended_at IS NOT NULL")
		} else {
			clauses = append(clauses, "suspended_at IS NULL")
		}
	}

	if params.SuspendedAfter != nil {
		clauses = append(clauses, fmt.Sprintf("suspended_at >= %d", *params.SuspendedAfter))
	}
	if params.SuspendedBefore != nil {
		clauses = append(clauses, fmt.Sprintf("suspended_at <= %d", *params.SuspendedBefore))
	}

	if params.LastLoginAfter != nil {
		clauses = append(clauses, fmt.Sprintf("last_login >= %d", *params.LastLoginAfter))
	}
	if params.LastLoginBefore != nil {
		clauses = append(clauses, fmt.Sprintf("last_login <= %d", *params.LastLoginBefore))
	}

	return strings.Join(clauses, " AND ")
}

func (ms *MeilisearchEngine) buildSort(params *SearchParams) []string {
	var sortParams []string

	if params.SortBy != "" {
		sortParams = append(sortParams, params.SortBy)
	}
	if !strings.HasPrefix(params.SortBy, "username:") {
		sortParams = append(sortParams, "username:asc")
	}
	return sortParams
}

func (ms *MeilisearchEngine) Index(ctx context.Context, user *domain.User) error {
	index := ms.search.Client.Index(ms.search.Config.Uid)
	task, err := index.AddDocumentsWithContext(ctx, user, nil)
	if err != nil {
		return err
	}
	if err := ms.search.WaitForTask(ctx, task); err != nil {
		return err
	}
	return nil
}

func (ms *MeilisearchEngine) Get(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := ms.search.Client.Index(ms.search.Config.Uid).GetDocumentWithContext(ctx, username, nil, &user)
	if err != nil {
		var meiliErr *meilisearch.Error
		if errors.As(err, &meiliErr) {
			if meiliErr.StatusCode == http.StatusNotFound && meiliErr.MeilisearchApiError.Code == "document_not_found" {
				return nil, domain.ErrNotFound
			}
		}
		return nil, err
	}
	return &user, nil
}

func (ms *MeilisearchEngine) Delete(ctx context.Context, username string) error {
	index := ms.search.Client.Index(ms.search.Config.Uid)
	task, err := index.DeleteDocumentWithContext(ctx, username, nil)
	if err != nil {
		return err
	}
	if err := ms.search.WaitForTask(ctx, task); err != nil {
		return err
	}
	return nil
}

func (ms *MeilisearchEngine) Search(ctx context.Context, params *SearchParams) (*SearchResult, error) {
	var results SearchResult
	count := params.Count
	req := &meilisearch.SearchRequest{
		HitsPerPage: &count,
		Page:        params.Page,
		Filter:      ms.buildFilter(params),
		Sort:        ms.buildSort(params),
	}
	index := ms.search.Client.Index(ms.search.Config.Uid)

	query := strings.TrimSpace(params.Query)

	res, err := index.SearchWithContext(ctx, query, req)
	if err != nil {
		var meiliErr *meilisearch.Error
		if errors.As(err, &meiliErr) {
			if meiliErr.StatusCode == http.StatusBadRequest &&
				(meiliErr.MeilisearchApiError.Code == "invalid_search_filter" ||
					meiliErr.MeilisearchApiError.Code == "invalid_search_sort") {
				return nil, domain.ErrInval
			}
		}
		return nil, err
	}
	results.Results = make([]domain.User, 0, len(res.Hits))
	if err := res.Hits.DecodeInto(&results.Results); err != nil {
		return nil, err
	}
	results.Total = res.TotalHits
	results.LastPage = res.TotalPages

	return &results, nil
}

func (ms *MeilisearchEngine) CreateMeilisearchIndex(ctx context.Context) error {
	return ms.search.CreateIndex(ctx)
}

func (ms *MeilisearchEngine) WipeMeilisearchIndex(ctx context.Context) error {
	return ms.search.WipeIndex(ctx)
}

func (ms *MeilisearchEngine) ConfigureMeilisearchIndex(ctx context.Context) error {
	return ms.search.ConfigureIndex(ctx)
}
