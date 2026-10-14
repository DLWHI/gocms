package category

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/errs"
	"net/http"
	"strings"
	"time"

	"github.com/meilisearch/meilisearch-go"
)

type MeilisearchEngine struct {
	se  meilisearch.ServiceManager
	idx string
}

func NewMeilisearchEngine(client meilisearch.ServiceManager, index string) *MeilisearchEngine {
	return &MeilisearchEngine{
		se:  client,
		idx: index,
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

func (r *MeilisearchEngine) Config() meilisearch.IndexConfig {
	return meilisearch.IndexConfig{
		Uid:        r.idx,
		PrimaryKey: "id",
	}
}

func (r *MeilisearchEngine) Settings() meilisearch.Settings {
	return meilisearch.Settings{
		SearchableAttributes: []string{"name", "slug"},
		FilterableAttributes: []string{"hidden", "navbar", "locale", "slug"},
		SortableAttributes:   []string{"name"},
	}
}

func (r *MeilisearchEngine) Index(ctx context.Context, category *Doc) error {
	index := r.se.Index(r.idx)
	category.ID = fmt.Sprintf("%s-%s", category.Slug, category.Locale)
	req := &meilisearch.DocumentsQuery{
		Filter: fmt.Sprintf("slug = %q AND locale != %q", category.Slug, category.Locale),
	}

	var res meilisearch.DocumentsResult
	err := r.se.Index(r.idx).GetDocumentsWithContext(ctx, req, &res)
	if err != nil {
		var meiliErr *meilisearch.Error
		if errors.As(err, &meiliErr) {
			if meiliErr.StatusCode == http.StatusNotFound {
				return errs.ErrNotFound
			}
		}
		return err
	}

	var categories []Doc
	if err := res.Results.DecodeInto(&categories); err != nil {
		return err
	}

	for i := range categories {
		categories[i].Slug = category.Slug
		categories[i].Navbar = category.Navbar
		categories[i].Hidden = category.Hidden
	}
	categories = append(categories, *category)

	taskInfo, err := index.AddDocumentsWithContext(ctx, categories, nil)
	task, err := r.se.TaskManager().WaitForTask(taskInfo.TaskUID, 100*time.Millisecond)
	if err != nil {
		return err
	}
	if task.Status == meilisearch.TaskStatusFailed && task.Error.Code != "index_not_found" {
		return fmt.Errorf("Error indexing: %s - %s", task.Error.Code, task.Error.Message)
	}
	return err
}

func (r *MeilisearchEngine) Get(ctx context.Context, slug string, locale string) (*Doc, error) {
	var category Doc
	docID := fmt.Sprintf("%s-%s", slug, locale)
	err := r.se.Index(r.idx).GetDocumentWithContext(ctx, docID, nil, &category)
	if err != nil {
		var meiliErr *meilisearch.Error
		if errors.As(err, &meiliErr) {
			if meiliErr.StatusCode == http.StatusNotFound {
				return nil, errs.ErrNotFound
			}
		}
		return nil, err
	}
	return &category, nil
}

func (r *MeilisearchEngine) Delete(ctx context.Context, slug string) error {
	filter := fmt.Sprintf("slug = %s", slug)
	index := r.se.Index(r.idx)

	_, err := index.DeleteDocumentsByFilterWithContext(ctx, filter, nil)
	return err
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
	index := r.se.Index(r.idx)

	params.Query = strings.TrimSpace(params.Query)

	res, err := index.SearchWithContext(ctx, params.Query, req)
	if err != nil {
		if errors.Is(err, meilisearch.ErrInvalidRequestMethod) {
			return nil, errs.ErrInval
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
