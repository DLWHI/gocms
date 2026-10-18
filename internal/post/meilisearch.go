package post

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/errs"
	"gocms/internal/search"
	"net/http"
	"strings"

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

	if params.ID > 0 {
		clauses = append(clauses, fmt.Sprintf("post_id = %d", params.ID))
	}

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

	if params.Locale != "" {
		clauses = append(clauses, fmt.Sprintf("locale = %q", params.Locale))
	}

	if params.CreatedAfter != nil {
		clauses = append(clauses, fmt.Sprintf("created_at >= %d", params.CreatedAfter.Unix()))
	}
	if params.CreatedBefore != nil {
		clauses = append(clauses, fmt.Sprintf("created_at <= %d", params.CreatedBefore.Unix()))
	}

	if params.PublishedAfter != nil {
		clauses = append(clauses, fmt.Sprintf("published_at >= %d", params.PublishedAfter.Unix()))
	}
	if params.PublishedBefore != nil {
		clauses = append(clauses, fmt.Sprintf("published_at <= %d", params.PublishedBefore.Unix()))
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

func buildSort(params *SearchParams) []string {
	var sortParams []string

	if params.SortBy != "" {
		sortParams = append(sortParams, params.SortBy)
	}

	if !strings.HasPrefix(params.SortBy, "post_id") {
		sortParams = append(sortParams, "post_id:desc")
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
		SearchableAttributes: []string{"title", "description", "keywords"},
		FilterableAttributes: []string{"post_id", "locale", "author", "created_at", "published_at", "categories"},
		SortableAttributes:   []string{"post_id", "title", "author", "created_at", "published_at", "views"},
	}
}

func (r *MeilisearchEngine) Index(ctx context.Context, post *Document) error {
	index := r.se.Index(r.idx)
	post.ID = fmt.Sprintf("%d-%s", post.PostID, post.Locale)
	req := &meilisearch.DocumentsQuery{
		Filter: fmt.Sprintf("post_id = %d AND locale != %q", post.PostID, post.Locale),
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

	var posts []Document
	if err := res.Results.DecodeInto(&posts); err != nil {
		return err
	}

	for i := range posts {
		posts[i].Categories = post.Categories
		posts[i].Preview = post.Preview
		posts[i].PublishedAt = post.PublishedAt
		posts[i].Views = post.Views
		posts[i].Slug = post.Slug
	}
	if post.Locale != "" {
		posts = append(posts, *post)
	}

	task, err := index.AddDocumentsWithContext(ctx, posts, nil)
	if err != nil {
		return err
	}
	if err := search.WaitForMeilisearchTask(r.se, task); err != nil {
		return err
	}
	return err
}

func (r *MeilisearchEngine) Get(ctx context.Context, id uint, locale string) (*Document, error) {
	var post Document
	docID := fmt.Sprintf("%d-%s", id, locale)
	err := r.se.Index(r.idx).GetDocumentWithContext(ctx, docID, nil, &post)
	if err != nil {
		if errors.Is(err, meilisearch.ErrInvalidRequestMethod) {
			return nil, errs.ErrNotFound
		}
		return nil, err
	}
	return &post, nil
}

func (r *MeilisearchEngine) Delete(ctx context.Context, id uint) error {
	filter := fmt.Sprintf("post_id = %d", id)
	index := r.se.Index(r.idx)

	task, err := index.DeleteDocumentsByFilterWithContext(ctx, filter, nil)
	if err != nil {
		return err
	}
	if err := search.WaitForMeilisearchTask(r.se, task); err != nil {
		return err
	}
	return nil
}

func (r *MeilisearchEngine) DeleteTranslation(ctx context.Context, id uint, locale string) error {
	docID := fmt.Sprintf("%d-%s", id, locale)
	index := r.se.Index(r.idx)
	task, err := index.DeleteDocumentWithContext(ctx, docID, nil)
	if err != nil {
		return err
	}
	if err := search.WaitForMeilisearchTask(r.se, task); err != nil {
		return err
	}
	return nil
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
