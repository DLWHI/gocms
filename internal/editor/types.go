package editor

import (
	"context"
	"gocms/internal/search"
	"gocms/internal/user"
	"time"

	"gorm.io/gorm"
)

type Model struct {
	Locale   string      `json:"locale" gorm:"size:8;primaryKey;foreignKey:locales(code);constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Name     string      `json:"name"`
	Bio      string      `json:"bio"`
	User     string      `json:"username" gorm:"primaryKey;size:64;"`
	UserData user.Record `json:"user_data" gorm:"foreignKey:User;references:Username;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Version  uint        `json:"-" gorm:"default:1"`
}

type Document struct {
	ID          string `json:"id,omitempty"`
	User        string `json:"username,omitempty"`
	Email       string `json:"email"`
	Bio         string `json:"bio,omitempty"`
	Name        string `json:"name"`
	Locale      string `json:"locale"`
	SuspendedAt *int64 `json:"suspended_at,omitempty"`
}

type MessagePayload struct {
	Username    string `json:"username"`
	Locale      string `json:"locale"`
	SuspendedAt *int64 `json:"suspended_at"`
	Action      string `json:"action"`
}

type SearchParams struct {
	search.Params
	Locale          string     `json:"locale"`
	Suspended       *bool      `json:"suspended"`
	SuspendedAfter  *time.Time `json:"suspended_after"`
	SuspendedBefore *time.Time `json:"suspended_before"`
}

type SearchResult = search.Result[Document]

type Repository interface {
	Create(ctx context.Context, editor *Model) error
	Update(ctx context.Context, editor *Model) error
	Get(ctx context.Context, username string, locale string) (*Model, error)
	GetAll(ctx context.Context, locale string, suspended *bool) ([]Model, error)
	Delete(ctx context.Context, username string) error
}

type SearchEngine interface {
	Index(ctx context.Context, editor *Document) error
	Get(ctx context.Context, username string, locale string) (*Document, error)
	Search(ctx context.Context, filters *SearchParams) (*SearchResult, error)
	Delete(ctx context.Context, username string, timestamp int64) error
}

func (Model) TableName() string {
	return "editors"
}

func (em *Model) ToDoc() Document {
	var ed Document

	if em.UserData.SuspendedAt.Valid {
		ts := em.UserData.SuspendedAt.Time.Unix()
		ed.SuspendedAt = &ts
	}

	ed.User = em.User
	ed.Email = em.UserData.Email
	ed.Name = em.Name
	ed.Locale = em.Locale
	ed.Bio = em.Bio
	return ed
}

func (ed *Document) ToModel() Model {
	var em Model

	if ed.SuspendedAt != nil {
		em.UserData.SuspendedAt = gorm.DeletedAt{
			Time:  time.Unix(*ed.SuspendedAt, 0).UTC(),
			Valid: true,
		}
	}

	em.User = ed.User
	em.Locale = ed.Locale
	em.Name = ed.Name
	em.Bio = ed.Bio
	em.UserData.Username = ed.User
	em.UserData.Email = ed.Email
	return em
}
