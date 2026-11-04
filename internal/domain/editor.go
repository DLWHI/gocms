package domain

const (
	EditorActionCreate = 1 << iota
	EditorActionUpdate
	EditorActionDelete

	EditorEvent      = "editor"
	EditorEntityName = "editors"
)

type Editor struct {
	ID        string  `json:"id,omitempty"`
	Slug      string  `json:"slug"`
	Bio       *string `json:"bio,omitempty"`
	Name      string  `json:"name"`
	Lang      string  `json:"lang"`
	DeletedAt *int64  `json:"deleted_at,omitempty"`
	Owner     string  `json:"owner"`
}

type EditorDiff struct {
	Slug  *string `json:"slug"`
	Bio   *string `json:"bio"`
	Name  *string `json:"name"`
	Owner *string `json:"owner"`
}

type EditorEventInfo struct {
	Slug   string `json:"slug"`
	Action int    `json:"action"`
}

func (u *EditorDiff) Apply(target *Editor) {
	if u.Slug != nil {
		target.Slug = *u.Slug
	}
	if u.Bio != nil {
		target.Bio = u.Bio
	}
	if u.Name != nil {
		target.Name = *u.Name
	}
	if u.Owner != nil {
		target.Owner = *u.Owner
	}
}
