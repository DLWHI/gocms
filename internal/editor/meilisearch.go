package editor

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
	search search.MeilisearchIndex
}

func NewMeilisearchEngine(client meilisearch.ServiceManager, index string) *MeilisearchEngine {
	return &MeilisearchEngine{
		search: search.MeilisearchIndex{
			Client: client,
			Config: meilisearch.IndexConfig{
				Uid:        index,
				PrimaryKey: "id",
			},
			Settings: meilisearch.Settings{
				SearchableAttributes: []string{"slug", "name"},
				FilterableAttributes: []string{"lang", "deleted_at", "owner"},
				SortableAttributes:   []string{"slug", "name", "deleted_at"},
			},
		},
	}
}

func (*MeilisearchEngine) buildFilter(params *SearchParams, language string) string {
	clauses := []string{fmt.Sprintf("lang = %q", language)}

	if params.Deleted != nil {
		if *params.Deleted {
			clauses = append(clauses, "deleted_at EXISTS")
		} else {
			clauses = append(clauses, "deleted_at NOT EXISTS")
		}
	}

	if params.DeletedAfter != nil {
		clauses = append(clauses, fmt.Sprintf("deleted_at >= %d", *params.DeletedAfter))
	}
	if params.DeletedBefore != nil {
		clauses = append(clauses, fmt.Sprintf("deleted_at <= %d", *params.DeletedBefore))
	}

	if params.Owner != "" {
		clauses = append(clauses, fmt.Sprintf("owner = %q", params.Owner))
	}

	return strings.Join(clauses, " AND ")
}

func (*MeilisearchEngine) buildSort(params *SearchParams) []string {
	var sortParams []string
	if params.SortBy != "" {
		sortParams = append(sortParams, params.SortBy)
	}
	if !strings.HasPrefix(params.SortBy, "name:") {
		sortParams = append(sortParams, "name:asc")
	}

	return sortParams
}

func (ms *MeilisearchEngine) Index(ctx context.Context, slug string, group ...domain.Editor) error {
	index := ms.search.Client.Index(ms.search.Config.Uid)

	for _, editor := range group {
		if editor.Slug != slug {
			return fmt.Errorf("%w: group item not from provided editor slug", domain.ErrInval)
		}
		editor.ID = fmt.Sprintf("%s-%s", editor.Slug, editor.Lang)
	}

	task, err := index.AddDocumentsWithContext(ctx, group, nil)
	if err != nil {
		return err
	}
	if err := ms.search.WaitForTask(ctx, task); err != nil {
		return err
	}
	return err
}

func (ms *MeilisearchEngine) Get(ctx context.Context, username string, lang string) (*domain.Editor, error) {
	var editor domain.Editor
	id := fmt.Sprintf("%s-%s", editor.Slug, editor.Lang)
	err := ms.search.Client.Index(ms.search.Config.Uid).GetDocumentWithContext(ctx, id, nil, &editor)
	if err != nil {
		var meiliErr *meilisearch.Error
		if errors.As(err, &meiliErr) {
			if meiliErr.StatusCode == http.StatusNotFound {
				return nil, domain.ErrNotFound
			}
		}
		return nil, err
	}
	return &editor, nil
}

func (ms *MeilisearchEngine) Delete(ctx context.Context, slug string) error {
	filter := fmt.Sprintf("username = %q", slug)
	index := ms.search.Client.Index(ms.search.Config.Uid)

	task, err := index.DeleteDocumentsByFilterWithContext(ctx, filter, nil)
	if err != nil {
		return err
	}
	if err := ms.search.WaitForTask(ctx, task); err != nil {
		return err
	}
	return err
}

func (ms *MeilisearchEngine) Search(ctx context.Context, language string, params *SearchParams) (*SearchResult, error) {
	var results SearchResult
	req := &meilisearch.SearchRequest{
		HitsPerPage:          &params.Count,
		Page:                 params.Page,
		Filter:               ms.buildFilter(params, language),
		Sort:                 ms.buildSort(params),
		AttributesToRetrieve: params.Select,
	}
	index := ms.search.Client.Index(ms.search.Config.Uid)

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

func (ms *MeilisearchEngine) Create(ctx context.Context) error {
	return ms.search.Create(ctx)
}

func (ms *MeilisearchEngine) Wipe(ctx context.Context) error {
	return ms.search.Wipe(ctx)
}

func (ms *MeilisearchEngine) Configure(ctx context.Context) error {
	return ms.search.Configure(ctx)
}

func (ms *MeilisearchEngine) Destroy(ctx context.Context) error {
	return ms.search.Destroy(ctx)
}
