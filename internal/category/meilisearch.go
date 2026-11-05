package category

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
				SearchableAttributes: []string{"name", "slug"},
				FilterableAttributes: []string{"hidden", "navbar", "lang", "slug"},
				SortableAttributes:   []string{"name", "slug"},
			},
		},
	}
}

func (*MeilisearchEngine) buildFilter(params *SearchParams, language string) string {
	clauses := []string{fmt.Sprintf("lang = %q", language)}

	if params.Hidden != nil {
		clauses = append(clauses, fmt.Sprintf("hidden = %t", *params.Hidden))
	}
	if params.Navbar != nil {
		clauses = append(clauses, fmt.Sprintf("navbar = %t", *params.Navbar))
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

func (ms *MeilisearchEngine) Index(ctx context.Context, slug string, group ...domain.Category) error {
	index := ms.search.Client.Index(ms.search.Config.Uid)

	for i := range group {
		if group[i].Slug != slug {
			return fmt.Errorf("%w: group item not from provided category slug", domain.ErrInval)
		}
		group[i].ID = fmt.Sprintf("%s-%s", slug, group[i].Lang)
	}

	task, err := index.AddDocumentsWithContext(ctx, group, nil)
	if err != nil {
		return err
	}
	if err := ms.search.WaitForTask(ctx, task); err != nil {
		return err
	}
	return nil
}

func (ms *MeilisearchEngine) Get(ctx context.Context, slug string, lang string) (*domain.Category, error) {
	var category domain.Category
	id := fmt.Sprintf("%s-%s", slug, lang)
	err := ms.search.Client.Index(ms.search.Config.Uid).GetDocumentWithContext(ctx, id, nil, &category)
	if err != nil {
		var meiliErr *meilisearch.Error
		if errors.As(err, &meiliErr) {
			if meiliErr.StatusCode == http.StatusNotFound && meiliErr.MeilisearchApiError.Code == "document_not_found" {
				return nil, domain.ErrNotFound
			}
		}
		return nil, err
	}
	return &category, nil
}

func (ms *MeilisearchEngine) Delete(ctx context.Context, slug string) error {
	filter := fmt.Sprintf("slug = %q", slug)
	index := ms.search.Client.Index(ms.search.Config.Uid)

	task, err := index.DeleteDocumentsByFilterWithContext(ctx, filter, nil)
	if err != nil {
		return err
	}
	if err := ms.search.WaitForTask(ctx, task); err != nil {
		return err
	}
	return nil
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

	results.Results = make([]domain.Category, 0, len(res.Hits))
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
