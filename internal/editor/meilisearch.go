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
				SearchableAttributes: []string{"username", "email", "name"},
				FilterableAttributes: []string{"locale", "username", "suspended_at"},
				SortableAttributes:   []string{"name", "username", "suspended_at", "last_login"},
			},
		},
	}
}

func buildFilter(params *SearchParams) string {
	var clauses []string

	if params.Locale != "" {
		clauses = append(clauses, fmt.Sprintf("locale = %q", params.Locale))
	}
	if params.Suspended != nil {
		if *params.Suspended {
			clauses = append(clauses, "suspended_at IS NOT NULL")
		} else {
			clauses = append(clauses, "suspended_at IS NULL")
		}
	}

	if params.SuspendedAfter != nil {
		clauses = append(clauses, fmt.Sprintf("suspended_at >= %d", params.SuspendedAfter.Unix()))
	}
	if params.SuspendedBefore != nil {
		clauses = append(clauses, fmt.Sprintf("suspended_at <= %d", params.SuspendedBefore.Unix()))
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

func (r *MeilisearchEngine) Index(ctx context.Context, editor *Document) error {
	index := r.search.Client.Index(r.search.Config.Uid)
	editor.ID = fmt.Sprintf("%s-%s", editor.User, editor.Locale)
	req := &meilisearch.DocumentsQuery{
		Filter: fmt.Sprintf("username = %q AND locale != %q", editor.User, editor.Locale),
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

	var editors []Document
	if err := res.Results.DecodeInto(&editors); err != nil {
		return err
	}

	for i := range editors {
		editors[i].Email = editor.Email
	}
	editors = append(editors, *editor)

	task, err := index.AddDocumentsWithContext(ctx, editors, nil)
	if err != nil {
		return err
	}
	if err := r.search.WaitForTask(ctx, task); err != nil {
		return err
	}
	return err
}

func (r *MeilisearchEngine) Get(ctx context.Context, username string, locale string) (*Document, error) {
	var editor Document
	editor.ID = fmt.Sprintf("%s-%s", editor.User, editor.Locale)
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

func (r *MeilisearchEngine) Delete(ctx context.Context, username string, timestamp int64) error {
	filter := fmt.Sprintf("username = %q", username)
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
		HitsPerPage: &params.Count,
		Page:        params.Page,
		Filter:      buildFilter(params),
		Sort:        buildSort(params),
		AttributesToRetrieve: []string{
			"username",
			"email",
			"desctiption",
			"name",
			"locale",
			"suspended_at",
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
	return r.search.CreateIndex(ctx)
}

func (r *MeilisearchEngine) WipeMeilisearchIndex(ctx context.Context) error {
	return r.search.WipeIndex(ctx)
}

func (r *MeilisearchEngine) ConfigureMeilisearchIndex(ctx context.Context) error {
	return r.search.ConfigureIndex(ctx)
}
