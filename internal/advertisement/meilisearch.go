package advertisement

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/domain"
	"gocms/internal/search"
	"net/http"
	"strconv"
	"strings"
	"time"

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
				SearchableAttributes: []string{"name", "description", "id"},
				FilterableAttributes: []string{"group", "expires_at"},
				SortableAttributes:   []string{"id", "name", "priority", "expires_at", "group"},
			},
		},
	}
}

func (*MeilisearchEngine) buildFilter(params *SearchParams) string {
	var clauses []string

	if params.Expired != nil {
		if *params.Expired {
			clauses = append(clauses, fmt.Sprintf("expires_at < %d", time.Now().Unix()))
		} else {
			clauses = append(clauses, fmt.Sprintf("(expires_at IS NULL OR expires_at > %d)", time.Now().Unix()))
		}
	}

	if params.ExpiresAfter != nil {
		clauses = append(clauses, fmt.Sprintf("expires_at >= %d", *params.ExpiresAfter))
	}
	if params.ExpiresBefore != nil {
		clauses = append(clauses, fmt.Sprintf("expires_at <= %d", *params.ExpiresBefore))
	}

	if len(params.Groups) > 0 {
		groups := make([]string, 0, len(params.Groups))
		for _, group := range params.Groups {
			if group == "" {
				groups = append(groups, "group IS EMPTY")
			} else {
				groups = append(groups, fmt.Sprintf("group = %q", group))
			}
		}
		clauses = append(clauses, "("+strings.Join(groups, " OR ")+")")
	}

	return strings.Join(clauses, " AND ")
}

func (*MeilisearchEngine) buildSort(params *SearchParams) []string {
	var sortParams []string

	if params.SortBy != "" {
		sortParams = append(sortParams, params.SortBy)
	}

	if !strings.HasPrefix(params.SortBy, "name") {
		sortParams = append(sortParams, "name:asc")
	}

	return sortParams
}

func (ms *MeilisearchEngine) Index(ctx context.Context, ad *domain.Advertisement) error {
	index := ms.search.Client.Index(ms.search.Config.Uid)
	task, err := index.AddDocumentsWithContext(ctx, ad, nil)
	if err != nil {
		return err
	}
	if err := ms.search.WaitForTask(ctx, task); err != nil {
		return err
	}
	return err
}

func (ms *MeilisearchEngine) Get(ctx context.Context, id uint) (*domain.Advertisement, error) {
	var ad domain.Advertisement
	docID := strconv.FormatUint(uint64(id), 10)
	err := ms.search.Client.Index(ms.search.Config.Uid).GetDocumentWithContext(ctx, docID, nil, &ad)
	if err != nil {
		var meiliErr *meilisearch.Error
		if errors.As(err, &meiliErr) {
			if meiliErr.StatusCode == http.StatusNotFound {
				return nil, domain.ErrNotFound
			}
		}
		return nil, err
	}
	return &ad, nil
}

func (ms *MeilisearchEngine) Delete(ctx context.Context, id uint) error {
	docID := strconv.FormatUint(uint64(id), 10)
	index := ms.search.Client.Index(ms.search.Config.Uid)
	task, err := index.DeleteDocumentWithContext(ctx, docID, nil)
	if err != nil {
		return err
	}
	if err := ms.search.WaitForTask(ctx, task); err != nil {
		return err
	}
	return err
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
