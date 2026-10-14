package ads

import (
	"context"
	"gocms/internal/search"
	"time"
)

type Model struct {
	ID          uint       `json:"id" gorm:"size:255;primaryKey"`
	Name        string     `json:"name" gorm:"not null"`
	Description string     `json:"description"`
	URL         string     `json:"url" gorm:"not null"`
	Priority    uint       `json:"priority" gorm:"default:0"`
	Group       string     `json:"group"`
	ExpiresAt   *time.Time `json:"expires_at"`
	Banner      string     `json:"banner"`
	Version     uint       `json:"-" gorm:"default:1"`
}

type Doc struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Priority    uint   `json:"priority"`
	Group       string `json:"group"`
	ExpiresAt   *int64 `json:"expires_at"`
	Banner      string `json:"banner"`
}

type SearchParams struct {
	search.Params
	ID            uint       `json:"id"`
	Expired       *bool      `json:"expired"`
	ExpiresAfter  *time.Time `json:"expires_after"`
	ExpiresBefore *time.Time `json:"expires_before"`
	Groups        []string   `json:"groups"`
}

type SearchResult = search.Result[Doc]

type Repository interface {
	Create(ctx context.Context, cat *Model) error
	Update(ctx context.Context, cat *Model) error
	Get(ctx context.Context, id uint) (*Model, error)
	GetGroup(ctx context.Context, group string) ([]Model, error)
	GetGroups(ctx context.Context) ([]string, error)
	Delete(ctx context.Context, id uint) error
}

type SearchEngine interface {
	Index(ctx context.Context, user *Doc) error
	Get(ctx context.Context, id uint) (*Doc, error)
	Search(ctx context.Context, filters *SearchParams) (*SearchResult, error)
	Delete(ctx context.Context, id uint) error
}

func (Model) TableName() string {
	return "advertisements"
}

func (am *Model) ToDoc() Doc {
	var ad Doc

	if am.ExpiresAt != nil {
		ts := am.ExpiresAt.Unix()
		ad.ExpiresAt = &ts
	}

	ad.ID = am.ID
	ad.Name = am.Name
	ad.Description = am.Description
	ad.URL = am.URL
	ad.Priority = am.Priority
	ad.Group = am.Group
	ad.Banner = am.Banner
	return ad
}

func (ad *Doc) ToModel() Model {
	var am Model
	if ad.ExpiresAt != nil {
		ts := time.Unix(*ad.ExpiresAt, 0).UTC()
		am.ExpiresAt = &ts
	}

	am.ID = ad.ID
	am.Name = ad.Name
	am.Description = ad.Description
	am.URL = ad.URL
	am.Priority = ad.Priority
	am.Group = ad.Group
	am.Banner = ad.Banner
	return am
}
