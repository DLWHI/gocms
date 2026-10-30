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
				SearchableAttributes: []string{"slug", "name"},
				FilterableAttributes: []string{"locale", "deleted_at", "owner"},
				SortableAttributes:   []string{"slug", "name", "deleted_at"},
			},
		},
	}
}

func (*MeilisearchEngine) buildFilter(params *SearchParams) string {
	var clauses []string

	if params.Locale != "" {
		clauses = append(clauses, fmt.Sprintf("locale = %q", params.Locale))
	}
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
	if params.DeletedAfter != nil {
		clauses = append(clauses, fmt.Sprintf("deleted_at <= %d", *params.DeletedBefore))
	}

	if params.Owner != "" {
		clauses = append(clauses, fmt.Sprintf("owner = %q", params.Owner))
	}

	return strings.Join(clauses, " AND ")
}

func (*MeilisearchEngine) buildSort(params *SearchParams) []string {
	var sortParams []string
	if !strings.HasPrefix(params.SortBy, "name") {
		sortParams = append(sortParams, "name:asc")
	} else {
		sortParams = append(sortParams, params.SortBy)
	}

	return sortParams
}

func (r *MeilisearchEngine) Index(ctx context.Context, slug string, group ...domain.Editor) error {
	index := r.search.Client.Index(r.search.Config.Uid)

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
	if err := r.search.WaitForTask(ctx, task); err != nil {
		return err
	}
	return err
}

func (r *MeilisearchEngine) Get(ctx context.Context, username string, locale string) (*domain.Editor, error) {
	var editor domain.Editor
	editor.ID = fmt.Sprintf("%s-%s", editor.Slug, editor.Locale)
	err := r.search.Client.Index(r.search.Config.Uid).GetDocumentWithContext(ctx, editor.ID, nil, &editor)
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

func (r *MeilisearchEngine) Delete(ctx context.Context, slug string) error {
	filter := fmt.Sprintf("username = %q", slug)
	index := r.search.Client.Index(r.search.Config.Uid)

	task, err := index.DeleteDocumentsByFilterWithContext(ctx, filter, nil)
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
		Filter:               r.buildFilter(params),
		Sort:                 r.buildSort(params),
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
