package domain

import "time"

type Advertisement struct {
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

func (Advertisement) TableName() string {
	return "advertisements"
}
