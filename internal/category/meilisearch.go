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
	search search.MeilisearchClient
}

func NewMeilisearchEngine(client meilisearch.ServiceManager, index string) *MeilisearchEngine {
	return &MeilisearchEngine{
		search: search.MeilisearchClient{
			Client: client,
			Config: meilisearch.IndexConfig{
				Uid:        index,
				PrimaryKey: "id",
			},
			Settings: meilisearch.Settings{
				SearchableAttributes: []string{"name", "slug"},
				FilterableAttributes: []string{"hidden", "navbar", "locale", "slug"},
				SortableAttributes:   []string{"name"},
			},
		},
	}
}

func (ms *MeilisearchEngine) buildFilter(params *SearchParams) string {
	var clauses []string

	if params.Locale != "" {
		clauses = append(clauses, fmt.Sprintf("locale = %q", params.Locale))
	}

	if params.Hidden != nil {
		clauses = append(clauses, fmt.Sprintf("hidden = %t", *params.Hidden))
	}
	if params.Navbar != nil {
		clauses = append(clauses, fmt.Sprintf("navbar = %t", *params.Navbar))
	}

	return strings.Join(clauses, " AND ")
}

func (ms *MeilisearchEngine) buildSort(params *SearchParams) []string {
	var sortParams []string

	if !strings.HasPrefix(params.SortBy, "name") {
		sortParams = append(sortParams, "name:asc")
	} else {
		sortParams = append(sortParams, params.SortBy)
	}

	return sortParams
}

func (ms *MeilisearchEngine) Index(ctx context.Context, slug string, group ...domain.Category) error {
	index := ms.search.Client.Index(ms.search.Config.Uid)

	for i := range group {
		if group[i].Slug != slug {
			return fmt.Errorf("%v: group item not from provided category slug", domain.ErrInval)
		}
		group[i].ID = fmt.Sprintf("%s-%s", slug, group[i].Locale)
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

func (ms *MeilisearchEngine) Get(ctx context.Context, slug string, locale string) (*domain.Category, error) {
	var category domain.Category
	docID := fmt.Sprintf("%s-%s", slug, locale)
	err := ms.search.Client.Index(ms.search.Config.Uid).GetDocumentWithContext(ctx, docID, nil, &category)
	if err != nil {
		var meiliErr *meilisearch.Error
		if errors.As(err, &meiliErr) {
			if meiliErr.StatusCode == http.StatusNotFound {
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

func (ms *MeilisearchEngine) Search(ctx context.Context, params *SearchParams) (*SearchResult, error) {
	var results SearchResult
	req := &meilisearch.SearchRequest{
		HitsPerPage:          &params.Count,
		Page:                 params.Page,
		Filter:               ms.buildFilter(params),
		Sort:                 ms.buildSort(params),
		AttributesToRetrieve: params.Select,
	}
	index := ms.search.Client.Index(ms.search.Config.Uid)

	params.Query = strings.TrimSpace(params.Query)

	res, err := index.SearchWithContext(ctx, params.Query, req)
	if err != nil {
		if errors.Is(err, meilisearch.ErrInvalidRequestMethod) {
			return nil, domain.ErrInval
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

func (ms *MeilisearchEngine) CreateMeilisearchIndex(ctx context.Context) error {
	return ms.search.CreateIndex(ctx)
}

func (ms *MeilisearchEngine) WipeMeilisearchIndex(ctx context.Context) error {
	return ms.search.WipeIndex(ctx)
}

func (ms *MeilisearchEngine) ConfigureMeilisearchIndex(ctx context.Context) error {
	return ms.search.ConfigureIndex(ctx)
}
