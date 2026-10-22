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
	client meilisearch.ServiceManager
	idx    string
}

func NewMeilisearchEngine(client meilisearch.ServiceManager, index string) *MeilisearchEngine {
	return &MeilisearchEngine{
		client: client,
		idx:    index,
	}
}

func buildFilter(params *SearchParams) string {
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
		clauses = append(clauses, fmt.Sprintf("suspended_at >= %d", params.SuspendedAfter.Unix()))
	}
	if params.SuspendedBefore != nil {
		clauses = append(clauses, fmt.Sprintf("suspended_at <= %d", params.SuspendedBefore.Unix()))
	}

	if params.LastLoginAfter != nil {
		clauses = append(clauses, fmt.Sprintf("last_login >= %d", params.LastLoginAfter.Unix()))
	}
	if params.LastLoginBefore != nil {
		clauses = append(clauses, fmt.Sprintf("last_login <= %d", params.LastLoginBefore.Unix()))
	}

	return strings.Join(clauses, " AND ")
}

func buildSort(params *SearchParams) []string {
	var sortParams []string

	if params.SortBy != "" {
		sortParams = append(sortParams, params.SortBy)
	}
	if !strings.HasPrefix(params.SortBy, "username") {
		sortParams = append(sortParams, "username:asc")
	}
	return sortParams
}

func (r *MeilisearchEngine) Config() meilisearch.IndexConfig {
	return meilisearch.IndexConfig{
		Uid:        r.idx,
		PrimaryKey: "username",
	}
}

func (r *MeilisearchEngine) Settings() meilisearch.Settings {
	return meilisearch.Settings{
		SearchableAttributes: []string{"username", "email"},
		FilterableAttributes: []string{"permissions", "suspended_at", "last_login"},
		SortableAttributes:   []string{"username", "suspended_at", "last_login"},
	}
}

func (r *MeilisearchEngine) Index(ctx context.Context, user *Document) error {
	index := r.client.Index(r.idx)
	task, err := index.AddDocumentsWithContext(ctx, user, nil)
	if err != nil {
		return err
	}
	if err := search.WaitForMeilisearchTask(ctx, r.client, task); err != nil {
		return err
	}
	return nil
}

func (r *MeilisearchEngine) Get(ctx context.Context, username string) (*Document, error) {
	var user Document
	err := r.client.Index(r.idx).GetDocumentWithContext(ctx, username, nil, &user)
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

func (r *MeilisearchEngine) Delete(ctx context.Context, username string) error {
	index := r.client.Index(r.idx)
	task, err := index.DeleteDocumentWithContext(ctx, username, nil)
	if err != nil {
		return err
	}
	if err := search.WaitForMeilisearchTask(ctx, r.client, task); err != nil {
		return err
	}
	return nil
}

func (r *MeilisearchEngine) Search(ctx context.Context, params *SearchParams) (*SearchResult, error) {
	var results SearchResult
	req := &meilisearch.SearchRequest{
		HitsPerPage: &params.Count,
		Page:        params.Page,
		Filter:      buildFilter(params),
		Sort:        buildSort(params),
	}
	index := r.client.Index(r.idx)

	params.Query = strings.TrimSpace(params.Query)

	res, err := index.SearchWithContext(ctx, params.Query, req)
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

	if err := res.Hits.DecodeInto(&results.Results); err != nil {
		return nil, err
	}
	results.Total = res.TotalHits
	results.LastPage = res.TotalPages

	return &results, nil
}
