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
	search search.MeilisearchDecorator
}

func NewMeilisearchEngine(client meilisearch.ServiceManager, index string) *MeilisearchEngine {
	return &MeilisearchEngine{
		search: search.MeilisearchDecorator{
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

func buildFilter(params *SearchParams) string {
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

func buildSort(params *SearchParams) []string {
	var sortParams []string

	if !strings.HasPrefix(params.SortBy, "name") {
		sortParams = append(sortParams, "name:asc")
	} else {
		sortParams = append(sortParams, params.SortBy)
	}

	return sortParams
}

func (r *MeilisearchEngine) Index(ctx context.Context, category *Document) error {
	index := r.search.Client.Index(r.search.Config.Uid)
	category.ID = fmt.Sprintf("%s-%s", category.Slug, category.Locale)
	req := &meilisearch.DocumentsQuery{
		Filter: fmt.Sprintf("slug = %q AND locale != %q", category.Slug, category.Locale),
	}

	var res meilisearch.DocumentsResult
	err := r.search.Client.Index(r.search.Config.Uid).GetDocumentsWithContext(ctx, req, &res)
	if err != nil {
		var meiliErr *meilisearch.Error
		if errors.As(err, &meiliErr) {
			if meiliErr.StatusCode == http.StatusNotFound {
				return domain.ErrNotFound
			}
		}
		return err
	}

	var categories []Document
	if err := res.Results.DecodeInto(&categories); err != nil {
		return err
	}

	for i := range categories {
		categories[i].Slug = category.Slug
		categories[i].Navbar = category.Navbar
		categories[i].Hidden = category.Hidden
	}
	categories = append(categories, *category)

	task, err := index.AddDocumentsWithContext(ctx, categories, nil)
	if err != nil {
		return err
	}
	if err := r.search.WaitForMeilisearchTask(ctx, task); err != nil {
		return err
	}
	return nil
}

func (r *MeilisearchEngine) Get(ctx context.Context, slug string, locale string) (*Document, error) {
	var category Document
	docID := fmt.Sprintf("%s-%s", slug, locale)
	err := r.search.Client.Index(r.search.Config.Uid).GetDocumentWithContext(ctx, docID, nil, &category)
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

func (r *MeilisearchEngine) Delete(ctx context.Context, slug string) error {
	filter := fmt.Sprintf("slug = %q", slug)
	index := r.search.Client.Index(r.search.Config.Uid)

	task, err := index.DeleteDocumentsByFilterWithContext(ctx, filter, nil)
	if err != nil {
		return err
	}
	if err := r.search.WaitForMeilisearchTask(ctx, task); err != nil {
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
		AttributesToRetrieve: []string{
			"slug",
			"hidden",
			"navbar",
			"name",
			"locale",
		},
	}
	index := r.search.Client.Index(r.search.Config.Uid)

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

func (r *MeilisearchEngine) CreateMeilisearchIndex(ctx context.Context) error {
	return r.search.CreateMeilisearchIndex(ctx)
}

func (r *MeilisearchEngine) WipeMeilisearchIndex(ctx context.Context) error {
	return r.search.WipeMeilisearchIndex(ctx)
}

func (r *MeilisearchEngine) ConfigureMeilisearchIndex(ctx context.Context) error {
	return r.search.ConfigureMeilisearchIndex(ctx)
}
