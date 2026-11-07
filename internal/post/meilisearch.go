package post

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
				SearchableAttributes: []string{"title", "description", "keywords", "post_id"},
				FilterableAttributes: []string{"language", "author", "created_at", "created_by", "published_at", "categories"},
				SortableAttributes:   []string{"post_id", "title", "author", "created_at", "published_at", "created_by", "views"},
			},
		},
	}
}

func (*MeilisearchEngine) buildFilter(params *SearchParams, language string) string {
	clauses := []string{fmt.Sprintf("lang = %q", language)}

	if len(params.Categories) > 0 {
		for i := range params.Categories {
			params.Categories[i] = fmt.Sprintf("%q", params.Categories[i])
		}
		catClauses := fmt.Sprintf("categories IN [%s]", strings.Join(params.Categories, ", "))
		clauses = append(clauses, catClauses)
	}

	if len(params.ExcludeCategories) > 0 {
		for i := range params.ExcludeCategories {
			params.ExcludeCategories[i] = fmt.Sprintf("%q", params.ExcludeCategories[i])
		}
		catClauses := fmt.Sprintf("categories NOT IN [%s]", strings.Join(params.ExcludeCategories, ", "))
		clauses = append(clauses, catClauses)
	}

	if len(params.Authors) > 0 {
		for i := range params.Authors {
			params.Authors[i] = fmt.Sprintf("%q", params.Authors[i])
		}
		authorClauses := fmt.Sprintf("author IN [%s]", strings.Join(params.Authors, ", "))
		clauses = append(clauses, authorClauses)
	}

	if params.CreatedAfter != nil {
		clauses = append(clauses, fmt.Sprintf("created_at >= %d", *params.CreatedAfter))
	}
	if params.CreatedBefore != nil {
		clauses = append(clauses, fmt.Sprintf("created_at <= %d", *params.CreatedBefore))
	}

	if params.PublishedAfter != nil {
		clauses = append(clauses, fmt.Sprintf("published_at >= %d", *params.PublishedAfter))
	}
	if params.PublishedBefore != nil {
		clauses = append(clauses, fmt.Sprintf("published_at <= %d", *params.PublishedBefore))
	}

	if params.Published != nil {
		if *params.Published {
			clauses = append(clauses, "published_at IS NOT NULL")
		} else {
			clauses = append(clauses, "published_at IS NULL")
		}
	}

	return strings.Join(clauses, " AND ")
}

func (*MeilisearchEngine) buildSort(params *SearchParams) []string {
	var sortParams []string

	if params.SortBy != "" {
		sortParams = append(sortParams, params.SortBy)
	}

	if !strings.HasPrefix(params.SortBy, "post_id") {
		sortParams = append(sortParams, "post_id:desc")
	}

	return sortParams
}

func (r *MeilisearchEngine) Index(ctx context.Context, id uint, group ...domain.Post) error {
	index := r.search.Client.Index(r.search.Config.Uid)
	for i := range group {
		if group[i].PostID != id {
			return fmt.Errorf("%w: group item not from provided post id", domain.ErrInval)
		}
		group[i].ID = fmt.Sprintf("%d-%s", group[i].PostID, group[i].Language)
	}

	task, err := index.AddDocumentsWithContext(ctx, group, nil)
	if err != nil {
		return err
	}
	if err := r.search.WaitForTask(ctx, task); err != nil {
		return err
	}
	return err
}

func (r *MeilisearchEngine) Get(ctx context.Context, id uint, language string) (*domain.Post, error) {
	var post domain.Post
	docID := fmt.Sprintf("%d-%s", id, language)
	err := r.search.Client.Index(r.search.Config.Uid).GetDocumentWithContext(ctx, docID, nil, &post)
	if err != nil {
		var meiliErr *meilisearch.Error
		if errors.As(err, &meiliErr) {
			if meiliErr.StatusCode == http.StatusNotFound {
				return nil, domain.ErrNotFound
			}
		}
		return nil, err
	}
	return &post, nil
}

func (r *MeilisearchEngine) Delete(ctx context.Context, id uint) error {
	filter := fmt.Sprintf("post_id = %d", id)
	index := r.search.Client.Index(r.search.Config.Uid)

	task, err := index.DeleteDocumentsByFilterWithContext(ctx, filter, nil)
	if err != nil {
		return err
	}
	if err := r.search.WaitForTask(ctx, task); err != nil {
		return err
	}
	return nil
}

func (r *MeilisearchEngine) Search(ctx context.Context, language string, params *SearchParams) (*SearchResult, error) {
	var results SearchResult
	req := &meilisearch.SearchRequest{
		HitsPerPage:          &params.Count,
		Page:                 params.Page,
		Filter:               r.buildFilter(params, language),
		Sort:                 r.buildSort(params),
		AttributesToRetrieve: params.Select,
	}
	index := r.search.Client.Index(r.search.Config.Uid)

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

func (r *MeilisearchEngine) Create(ctx context.Context) error {
	return r.search.Create(ctx)
}

func (r *MeilisearchEngine) Wipe(ctx context.Context) error {
	return r.search.Wipe(ctx)
}

func (r *MeilisearchEngine) Configure(ctx context.Context) error {
	return r.search.Configure(ctx)
}

func (ms *MeilisearchEngine) Destroy(ctx context.Context) error {
	return ms.search.Destroy(ctx)
}
