package post

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/domain"
	"gocms/internal/search"
	"net/http"
	"strconv"
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
				SearchableAttributes: []string{"title", "description", "slug", "keywords", "post_id"},
				FilterableAttributes: []string{"post_id", "lang", "author", "created_at", "created_by", "published_at", "categories"},
				SortableAttributes:   []string{"post_id", "title", "author", "created_at", "created_by", "published_at", "views"},
			},
		},
	}
}

func (*MeilisearchEngine) buildFilter(params *SearchParams) string {
	clauses := []string{}

	if len(params.Categories) > 0 {
		categroies := make([]string, 0, len(params.Categories))
		for _, category := range params.Categories {
			categroies = append(categroies, strconv.Quote(category))
		}
		clauses = append(clauses, fmt.Sprintf("categories IN [%s]", strings.Join(categroies, ", ")))
	}

	if len(params.ExcludeCategories) > 0 {
		categroies := make([]string, 0, len(params.ExcludeCategories))
		for _, category := range params.ExcludeCategories {
			categroies = append(categroies, strconv.Quote(category))
		}
		clauses = append(clauses, fmt.Sprintf("categories NOT IN [%s]", strings.Join(categroies, ", ")))
	}

	if len(params.Authors) > 0 {
		authors := make([]string, 0, len(params.Authors))
		for _, author := range params.Authors {
			authors = append(authors, strconv.Quote(author))
		}
		clauses = append(clauses, fmt.Sprintf("author IN [%s]", strings.Join(authors, ", ")))
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

func (*MeilisearchEngine) wrapLang(clauses string, language string) string {
	if clauses == "" {
		return fmt.Sprintf("lang = %q", language)
	} else {
		return clauses + fmt.Sprintf("AND lang = %q", language)
	}
}

func (ms *MeilisearchEngine) Index(ctx context.Context, group ...domain.Post) error {
	index := ms.search.Client.Index(ms.search.Config.Uid)
	for i := range group {
		group[i].ID = fmt.Sprintf("%d-%s", group[i].PostID, group[i].Lang)
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

func (ms *MeilisearchEngine) Get(ctx context.Context, id uint, language string) (*domain.Post, error) {
	var post domain.Post
	docID := fmt.Sprintf("%d-%s", id, language)
	err := ms.search.Client.Index(ms.search.Config.Uid).GetDocumentWithContext(ctx, docID, nil, &post)
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

func (ms *MeilisearchEngine) Delete(ctx context.Context, id uint) error {
	filter := fmt.Sprintf("post_id = %d", id)
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
	req := &meilisearch.SearchRequest{
		HitsPerPage:          &params.Count,
		Page:                 params.Page,
		Filter:               ms.buildFilter(params),
		Sort:                 ms.buildSort(params),
		AttributesToRetrieve: params.Select,
	}
	return ms.perform(ctx, strings.TrimSpace(params.Query), req)
}

func (ms *MeilisearchEngine) SearchLang(ctx context.Context, language string, params *SearchParams) (*SearchResult, error) {
	req := &meilisearch.SearchRequest{
		HitsPerPage:          &params.Count,
		Page:                 params.Page,
		Filter:               ms.wrapLang(ms.buildFilter(params), language),
		Sort:                 ms.buildSort(params),
		AttributesToRetrieve: params.Select,
	}
	return ms.perform(ctx, strings.TrimSpace(params.Query), req)
}

func (ms *MeilisearchEngine) perform(ctx context.Context, query string, req *meilisearch.SearchRequest) (*SearchResult, error) {
	var results SearchResult
	index := ms.search.Client.Index(ms.search.Config.Uid)

	res, err := index.SearchWithContext(ctx, query, req)
	if err != nil {
		var meiliErr *meilisearch.Error
		if errors.As(err, &meiliErr) {
			if meiliErr.StatusCode == http.StatusBadRequest &&
				(meiliErr.MeilisearchApiError.Code == "invalid_search_filter" ||
					meiliErr.MeilisearchApiError.Code == "invalid_search_sort") {
				return nil, fmt.Errorf("%w: %v", domain.ErrInval, meiliErr.Error())
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
