package domain

const (
	AdvertisemenActionCreate = 1 << iota
	AdvertisementActionUpdate
	AdvertisementActionDelete

	AdvertisementEvent      = "ad"
	AdvertisementEntityName = "ads"
)

type Advertisement struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Priority    uint   `json:"priority"`
	Group       string `json:"group"`
	ExpiresAt   *int64 `json:"expires_at"`
	Banner      string `json:"banner"`
}

type AdvertisementDiff struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	URL         *string `json:"url"`
	Priority    *uint   `json:"priority"`
	Group       *string `json:"group"`
	ExpiresAt   *int64  `json:"expires_at"`
	Banner      *string `json:"banner"`
}

type AdvertisementEventInfo struct {
	ID     uint `json:"id"`
	Action int  `json:"action"`
}

func (u *AdvertisementDiff) Apply(target *Advertisement) {
	if u.Name != nil {
		target.Name = *u.Name
	}
	if u.Description != nil {
		target.Description = *u.Description
	}
	if u.URL != nil {
		target.URL = *u.URL
	}
	if u.Priority != nil {
		target.Priority = *u.Priority
	}
	if u.Group != nil {
		target.Group = *u.Group
	}
	if u.Banner != nil {
		target.Banner = *u.Banner
	}
	target.ExpiresAt = u.ExpiresAt
}
