package public

import (
	"gocms/internal/domain"
	"gocms/internal/post"
	"gocms/internal/search"
	"time"
)

type PostSearchParams struct {
	search.Params
	Categories      []string `json:"categories,omitempty"`
	Authors         []string `json:"authors,omitempty"`
	PublishedAfter  *int64   `json:"published_after,omitempty"`
	PublishedBefore *int64   `json:"published_before,omitempty"`
}

type PostSearchResult = search.Result[domain.PostPreview]

func (p *PostSearchParams) ToPrivateParams() *post.SearchParams {
	var res post.SearchParams
	trueVal := true
	res.Published = &trueVal
	res.Select = []string{
		"post_id",
		"title",
		"description",
		"slug",
		"lang",
		"published_at",
		"views",
		"preview",
		"author",
		"categories",
	}

	now := time.Now().Unix()
	if p.PublishedBefore == nil && p.PublishedAfter == nil {
		res.PublishedBefore = &now
	}
	if p.PublishedBefore != nil && *p.PublishedBefore > now {
		res.PublishedBefore = &now
	}
	if p.PublishedAfter != nil && *p.PublishedAfter > now {
		res.PublishedAfter = &now
	}

	res.Count = p.Count
	res.SortBy = p.SortBy
	res.Query = p.Query
	res.Page = p.Page
	res.Categories = p.Categories
	res.Authors = p.Authors
	res.PublishedAfter = p.PublishedAfter
	res.PublishedBefore = p.PublishedBefore
	return &res
}

func ToPublicResults(posts *post.SearchResult) *PostSearchResult {
	res := PostSearchResult{
		Results:  make([]domain.PostPreview, 0, len(posts.Results)),
		LastPage: posts.LastPage,
		Total:    posts.Total,
	}

	for _, post := range posts.Results {
		res.Results = append(res.Results, *post.ToPreview())
	}

	return &res
}
