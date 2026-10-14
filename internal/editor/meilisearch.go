package editor

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
	if params.Suspended != nil {
		if *params.Suspended == true {
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

func (r *MeilisearchEngine) Config() meilisearch.IndexConfig {
	return meilisearch.IndexConfig{
		Uid:        r.idx,
		PrimaryKey: "id",
	}
}

func (r *MeilisearchEngine) Settings() meilisearch.Settings {
	return meilisearch.Settings{
		SearchableAttributes: []string{"username", "email", "name"},
		FilterableAttributes: []string{"locale", "username", "suspended_at"},
		SortableAttributes:   []string{"name", "username", "suspended_at", "last_login"},
	}
}

func (r *MeilisearchEngine) Index(ctx context.Context, editor *Doc) error {
	index := r.se.Index(r.idx)
	editor.ID = fmt.Sprintf("%s-%s", editor.User, editor.Locale)
	req := &meilisearch.DocumentsQuery{
		Filter: fmt.Sprintf("username = %q AND locale != %q", editor.User, editor.Locale),
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

	var editors []Doc
	if err := res.Results.DecodeInto(&editors); err != nil {
		return err
	}

	for i := range editors {
		editors[i].Email = editor.Email
	}
	editors = append(editors, *editor)

	taskInfo, err := index.AddDocumentsWithContext(ctx, editors, nil)
	task, err := r.se.TaskManager().WaitForTask(taskInfo.TaskUID, 100*time.Millisecond)
	if err != nil {
		return err
	}
	if task.Status == meilisearch.TaskStatusFailed && task.Error.Code != "index_not_found" {
		return fmt.Errorf("Error indexing: %s - %s", task.Error.Code, task.Error.Message)
	}
	return err
}

func (r *MeilisearchEngine) Get(ctx context.Context, username string, locale string) (*Doc, error) {
	var editor Doc
	editor.ID = fmt.Sprintf("%s-%s", editor.User, editor.Locale)
	err := r.se.Index(r.idx).GetDocumentWithContext(ctx, editor.ID, nil, &editor)
	if err != nil {
		var meiliErr *meilisearch.Error
		if errors.As(err, &meiliErr) {
			if meiliErr.StatusCode == http.StatusNotFound {
				return nil, errs.ErrNotFound
			}
		}
		return nil, err
	}
	return &editor, nil
}

func (r *MeilisearchEngine) Delete(ctx context.Context, username string) error {
	filter := fmt.Sprintf("username = %q", username)
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
			"username",
			"email",
			"desctiption",
			"name",
			"locale",
			"suspended_at",
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
