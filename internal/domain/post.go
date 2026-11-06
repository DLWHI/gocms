package domain

import (
	"html"
	"strings"
)

const (
	PostEvent = "post"
)

type Post struct {
	ID          string   `json:"id"`
	PostID      uint     `json:"post_id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Content     string   `json:"content,omitempty"`
	Slug        string   `json:"slug"`
	Keywords    string   `json:"keywords,omitempty"`
	Language    string   `json:"language"`
	CreatedAt   int64    `json:"created_at,omitempty"`
	PublishedAt *int64   `json:"published_at"`
	Views       uint     `json:"views"`
	Preview     string   `json:"preview"`
	Media       string   `json:"media,omitempty"`
	MediaSource string   `json:"media_source,omitempty"`
	MediaType   string   `json:"media_type,omitempty"`
	Author      string   `json:"author,omitempty"`
	CreatedBy   string   `json:"created_by,omitempty"`
	Categories  []string `json:"categories"`
}

type PostDiff struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Content     *string  `json:"content"`
	Slug        *string  `json:"slug"`
	Keywords    *string  `json:"keywords"`
	PublishedAt *int64   `json:"published_at"`
	Preview     *string  `json:"preview"`
	Media       *string  `json:"media"`
	MediaSource *string  `json:"media_source"`
	MediaType   *string  `json:"media_type"`
	Author      *string  `json:"author"`
	Categories  []string `json:"categories"`
}

type PostEventInfo struct {
	PostID uint   `json:"post_id"`
	Action string `json:"action"`
}

func NormalizeKeywords(value string) (string, error) {
	parts := strings.Split(value, ",")
	for i := range parts {
		parts[i] = html.EscapeString(strings.TrimSpace(parts[i]))
	}
	return strings.Join(parts, ", "), nil
}

func (u *PostDiff) Apply(target *Post) {
	target.PublishedAt = u.PublishedAt
	if u.Title != nil {
		target.Title = *u.Title
	}
	if u.Description != nil {
		target.Description = *u.Description
	}
	if u.Content != nil {
		target.Content = *u.Content
	}
	if u.Slug != nil {
		target.Slug = *u.Slug
	}
	if u.Keywords != nil {
		target.Keywords = *u.Keywords
	}
	if u.Preview != nil {
		target.Preview = *u.Preview
	}
	if u.Media != nil {
		target.Media = *u.Media
	}
	if u.MediaSource != nil {
		target.MediaSource = *u.MediaSource
	}
	if u.MediaType != nil {
		target.MediaType = *u.MediaType
	}
	if u.Author != nil {
		target.Author = *u.Author
	}
	if u.Categories != nil {
		target.Categories = u.Categories
	}
}
