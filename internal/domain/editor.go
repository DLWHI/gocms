package domain

const (
	EditorActionCreate = 1 << iota
	EditorActionUpdate
	EditorActionDelete

	EditorEventName  = "editor"
	EditorEntityName = "editors"
)

type Editor struct {
	ID        string `json:"id,omitempty"`
	Slug      string `json:"slug"`
	Bio       string `json:"bio,omitempty"`
	Name      string `json:"name"`
	Locale    string `json:"locale"`
	DeletedAt *int64 `json:"suspended_at,omitempty"`
	Owner     string `json:"owner"`
}

type EditorDiff struct {
	Slug   *string `json:"slug"`
	Bio    *string `json:"bio"`
	Name   *string `json:"name"`
	Owner  *string `json:"owner"`
	Locale string  `json:"locale"`
}

type EditorEvent struct {
	Slug   string `json:"slug"`
	Action int    `json:"action"`
}
