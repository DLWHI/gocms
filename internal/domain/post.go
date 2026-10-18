package domain

import (
	"time"
)

type Post struct {
	ID          uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	Slug        string     `json:"slug"`
	CreatedAt   time.Time  `json:"created_at" gorm:"default:current_timestamp"`
	PublishedAt *time.Time `json:"published_at"`
	Views       uint       `json:"views" gorm:"not null;default:0"`
	Preview     string     `json:"preview"`
	Author      string     `json:"author" gorm:"size:64;"`
	AuthorUser  User       `json:"-" gorm:"foreignKey:Author;references:Username;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Categories  []Category `json:"-" gorm:"many2many:post_categories;foreignKey:ID;joinForeignKey:PostID;references:Slug;joinReferences:CategorySlug;constraint:OnDelete:CASCADE,OnUpdate:CASCADE;"`
	Version     uint       `json:"-" gorm:"default:1"`
}

type PostTranslation struct {
	Title       string `json:"title" gorm:"not null"`
	Description string `json:"description" gorm:"not null"`
	Keywords    string `json:"keywords" gorm:"not null"`
	Markdown    string `json:"markdown"`
	Media       string `json:"media"`
	MediaSource string `json:"media_source"`
	MediaType   string `json:"media_type"`
	Locale      string `json:"locale" gorm:"size:8;primaryKey;foreignKey:locales(code);constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	PostID      uint   `json:"post_id" gorm:"primaryKey"`
	Post        Post   `json:"post_meta" gorm:"foreignKey:PostID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Version     uint   `json:"-" gorm:"default:1"`
}

func (Post) TableName() string {
	return "posts"
}

func (PostTranslation) TableName() string {
	return "post_translations"
}
