package storage

import "time"

type ItemType string

const (
	TypeFolder ItemType = "folder"
	TypeImage  ItemType = "image"
	TypeVideo  ItemType = "video"
	TypeFile   ItemType = "file"
)

type Item struct {
	Path string    `json:"id"`
	Type ItemType  `json:"type"`
	Size int64     `json:"size"`
	Date time.Time `json:"modified"`
}

type Info struct {
	Total uint64 `json:"total"`
	Used  uint64 `json:"used"`
	Free  uint64 `json:"free"`
}
