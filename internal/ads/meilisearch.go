package ads

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
				SearchableAttributes: []string{"name", "description"},
				FilterableAttributes: []string{"group", "expires_at", "id"},
				SortableAttributes:   []string{"id", "name", "priority", "expires_at", "group"},
			},
		},
	}
}

func buildFilter(params *SearchParams) string {
	var clauses []string

	if params.ID != 0 {
		clauses = append(clauses, fmt.Sprintf("id = %d", params.ID))
	}

	if params.Expired != nil {
		if *params.Expired {
			clauses = append(clauses, fmt.Sprintf("expires_at < %d", time.Now().Unix()))
		} else {
			clauses = append(clauses, fmt.Sprintf("(expires_at IS NULL OR expires_at > %d)", time.Now().Unix()))
		}
	}

	if params.ExpiresAfter != nil {
		clauses = append(clauses, fmt.Sprintf("expires_at >= %d", params.ExpiresAfter.Unix()))
	}
	if params.ExpiresBefore != nil {
		clauses = append(clauses, fmt.Sprintf("expires_at <= %d", params.ExpiresBefore.Unix()))
	}

	if len(params.Groups) > 0 {
		for i := range params.Groups {
			params.Groups[i] = fmt.Sprintf("%q", params.Groups[i])
		}
		groupClauses := fmt.Sprintf("group IN [%s]", strings.Join(params.Groups, ", "))
		clauses = append(clauses, groupClauses)
	}

	return strings.Join(clauses, " AND ")
}

func buildSort(params *SearchParams) []string {
	var sortParams []string

	if params.SortBy != "" {
		sortParams = append(sortParams, params.SortBy)
	}

	if !strings.HasPrefix(params.SortBy, "name") {
		sortParams = append(sortParams, "name:asc")
	}

	return sortParams
}

func (r *MeilisearchEngine) Index(ctx context.Context, ad *Doc) error {
	index := r.search.Client.Index(r.search.Config.PrimaryKey)
	task, err := index.AddDocumentsWithContext(ctx, ad, nil)
	if err != nil {
		return err
	}
	if err := r.search.WaitForTask(ctx, task); err != nil {
		return err
	}
	return err
}

func (r *MeilisearchEngine) Get(ctx context.Context, id uint) (*Doc, error) {
	var ad Doc
	docID := strconv.FormatUint(uint64(id), 10)
	err := r.search.Client.Index(r.search.Config.Uid).GetDocumentWithContext(ctx, docID, nil, &ad)
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

func (r *MeilisearchEngine) Delete(ctx context.Context, id uint) error {
	docID := strconv.FormatUint(uint64(id), 10)
	index := r.search.Client.Index(r.search.Config.Uid)
	task, err := index.DeleteDocumentWithContext(ctx, docID, nil)
	if err != nil {
		return err
	}
	if err := r.search.WaitForTask(ctx, task); err != nil {
		return err
	}
	return err
}

func (r *MeilisearchEngine) Search(ctx context.Context, params *SearchParams) (*SearchResult, error) {
	var results SearchResult
	req := &meilisearch.SearchRequest{
		HitsPerPage:          &params.Count,
		Page:                 params.Page,
		Filter:               buildFilter(params),
		Sort:                 buildSort(params),
		AttributesToRetrieve: params.Select,
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
	return r.search.CreateIndex(ctx)
}

func (r *MeilisearchEngine) WipeMeilisearchIndex(ctx context.Context) error {
	return r.search.WipeIndex(ctx)
}

func (r *MeilisearchEngine) ConfigureMeilisearchIndex(ctx context.Context) error {
	return r.search.ConfigureIndex(ctx)
}
