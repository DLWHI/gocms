package post

import (
	"context"
	"gocms/internal/category"
	"gocms/internal/search"
	"gocms/internal/user"
	"time"
)

type Meta struct {
	ID          uint             `json:"id" gorm:"primaryKey;autoIncrement"`
	Slug        string           `json:"slug"`
	CreatedAt   time.Time        `json:"created_at" gorm:"default:current_timestamp"`
	PublishedAt *time.Time       `json:"published_at"`
	Views       uint             `json:"views" gorm:"not null;default:0"`
	Preview     string           `json:"preview"`
	Author      string           `json:"author" gorm:"size:64;"`
	AuthorUser  user.Model       `json:"-" gorm:"foreignKey:Author;references:Username;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Categories  []category.Model `json:"-" gorm:"many2many:post_categories;foreignKey:ID;joinForeignKey:PostID;references:Slug;joinReferences:CategorySlug;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;"`
	Version     uint             `json:"-" gorm:"default:1"`
}

type Translation struct {
	Title       string `json:"title" gorm:"not null"`
	Description string `json:"description" gorm:"not null"`
	Keywords    string `json:"keywords" gorm:"not null"`
	Markdown    string `json:"markdown"`
	Media       string `json:"media"`
	MediaSource string `json:"media_source"`
	MediaType   string `json:"media_type"`
	Locale      string `json:"locale" gorm:"size:8;primaryKey;foreignKey:locales(code);constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	PostID      uint   `json:"post_id" gorm:"primaryKey"`
	Post        Meta   `json:"post_meta" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Version     uint   `json:"-" gorm:"default:1"`
}

type MessagePayload struct {
	ID     uint   `json:"id"`
	Locale string `json:"locale"`
	Action string `json:"action"`
}

type Document struct {
	ID          string   `json:"id,omitempty"`
	PostID      uint     `json:"post_id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Markdown    string   `json:"markdown,omitempty"`
	Slug        string   `json:"slug"`
	Keywords    string   `json:"keywords,omitempty"`
	Locale      string   `json:"locale"`
	CreatedAt   int64    `json:"created_at,omitempty"`
	PublishedAt *int64   `json:"published_at"`
	Views       uint     `json:"views"`
	Preview     string   `json:"preview"`
	Media       string   `json:"media,omitempty"`
	MediaSource string   `json:"media_source,omitempty"`
	MediaType   string   `json:"media_type,omitempty"`
	Author      string   `json:"author,omitempty"`
	Categories  []string `json:"categories"`
}

type SearchParams struct {
	search.Params
	ID                uint       `json:"id,omitempty"`
	Categories        []string   `json:"categories,omitempty"`
	ExcludeCategories []string   `json:"exclude_categories,omitempty"`
	Authors           []string   `json:"authors,omitempty"`
	Locale            string     `json:"locale,omitempty"`
	CreatedAfter      *time.Time `json:"created_after,omitempty"`
	CreatedBefore     *time.Time `json:"created_before,omitempty"`
	PublishedAfter    *time.Time `json:"published_after,omitempty"`
	PublishedBefore   *time.Time `json:"published_before,omitempty"`
	Published         *bool      `json:"published,omitempty"`
	Select            []string   `json:"select,omitempty"`
}

type SearchResult = search.Result[Document]

type Repository interface {
	Create(ctx context.Context, post *Translation) error
	Update(ctx context.Context, post *Meta) error
	UpdateTranslation(ctx context.Context, post *Translation) error
	IncrementViews(ctx context.Context, id uint, value uint) error
	Get(ctx context.Context, id uint) (*Meta, error)
	GetTranslation(ctx context.Context, id uint, locale string) (*Translation, error)
	GetCategories(ctx context.Context, id uint) ([]string, error)
	Delete(ctx context.Context, id uint) error
	DeleteTranslation(ctx context.Context, id uint, locale string) error
}

type SearchEngine interface {
	Index(ctx context.Context, post *Document) error
	Get(ctx context.Context, id uint, locale string) (*Document, error)
	Search(ctx context.Context, post *SearchParams) (*SearchResult, error)
	Delete(ctx context.Context, id uint) error
	DeleteTranslation(ctx context.Context, id uint, locale string) error
}

func (Meta) TableName() string {
	return "posts"
}

func (Translation) TableName() string {
	return "post_translations"
}

func (pm *Meta) ToDoc() Document {
	var pd Document

	if pm.PublishedAt != nil {
		ts := pm.PublishedAt.Unix()
		pd.PublishedAt = &ts
	}

	// To make json.Marshall convert no categories to empty array
	categories := make([]string, 0, len(pm.Categories))
	for _, category := range pm.Categories {
		categories = append(categories, category.Slug)
	}
	pd.Categories = categories

	pd.PostID = pm.ID
	pd.Slug = pm.Slug
	pd.CreatedAt = pm.CreatedAt.Unix()
	pd.Views = pm.Views
	pd.Preview = pm.Preview
	pd.Author = pm.Author

	return pd
}

func (pm *Translation) ToDoc() Document {
	var pd Document

	if pm.Post.PublishedAt != nil {
		ts := pm.Post.PublishedAt.Unix()
		pd.PublishedAt = &ts
	}

	// To make json.Marshall convert no categories to empty array
	categories := make([]string, 0, len(pm.Post.Categories))
	for _, category := range pm.Post.Categories {
		categories = append(categories, category.Slug)
	}
	pd.Categories = categories

	pd.PostID = pm.PostID
	pd.Title = pm.Title
	pd.Description = pm.Description
	pd.Slug = pm.Post.Slug
	pd.Keywords = pm.Keywords
	pd.Locale = pm.Locale
	pd.CreatedAt = pm.Post.CreatedAt.Unix()
	pd.Views = pm.Post.Views
	pd.Preview = pm.Post.Preview
	pd.Media = pm.Media
	pd.MediaType = pm.MediaType
	pd.MediaSource = pm.MediaSource
	pd.Author = pm.Post.Author

	return pd
}

func (pd *Document) ToModel() Translation {
	var pm Translation
	if pd.PublishedAt != nil {
		ts := time.Unix(*pd.PublishedAt, 0).UTC()
		pm.Post.PublishedAt = &ts
	}

	// To make json.Marshall convert no categories to empty array
	categories := make([]category.Model, 0, len(pd.Categories))
	for _, categ := range pd.Categories {
		categories = append(categories, category.Model{Slug: categ})
	}
	pm.Post.Categories = categories

	pm.PostID = pd.PostID
	pm.Title = pd.Title
	pm.Description = pd.Description
	pm.Keywords = pd.Keywords
	pm.Locale = pd.Locale
	pm.Post.ID = pd.PostID
	pm.Post.Slug = pd.Slug
	pm.Post.CreatedAt = time.Unix(pd.CreatedAt, 0).UTC()
	pm.Post.Views = pd.Views
	pm.Post.Author = pd.Author
	pm.Post.Preview = pd.Preview
	pm.Media = pd.Media
	pm.MediaType = pd.MediaType
	pm.MediaSource = pd.MediaSource
	return pm
}
