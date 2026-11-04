package domain

import (
	"fmt"
	"html"
	"net/url"
	"strings"
)

const (
	PostActionCreate = 1 << iota
	PostActionUpdate
	PostActionDelete

	PostEvent      = "post"
	PostEntityName = "posts"
)

type Post struct {
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
	Language    string   `json:"language"`
	PublishedAt *int64   `json:"published_at"`
	Preview     *string  `json:"preview"`
	Media       *string  `json:"media"`
	MediaSource *string  `json:"media_source"`
	MediaType   *string  `json:"media_type"`
	Author      *string  `json:"author"`
	Categories  []string `json:"categories"`
}

type PostEventInfo struct {
	PostID uint `json:"post_id"`
	Action int  `json:"action"`
}

func NormalizeURL(value string) (string, error) {
	u, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInval, err)
	}
	return u.String(), nil
}

func NormalizeKeywords(value string) (string, error) {
	parts := strings.Split(value, ",")
	for i := range parts {
		parts[i] = html.EscapeString(strings.TrimSpace(parts[i]))
	}
	return strings.Join(parts, ", "), nil
}
