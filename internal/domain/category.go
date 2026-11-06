package domain

const (
	CategoryActionCreate = 1 << iota
	CategoryActionUpdate
	CategoryActionDelete

	CategoryEvent      = "category"
	CategoryEntityName = "categories"
)

type Category struct {
	ID       string `json:"id,omitempty"`
	Slug     string `json:"slug"`
	Hidden   bool   `json:"hidden"`
	Navbar   bool   `json:"navbar"`
	Language string `json:"lang"`
	Name     string `json:"name"`
}

type CategoryDiff struct {
	Slug   *string `json:"slug"`
	Hidden *bool   `json:"hidden"`
	Navbar *bool   `json:"navbar"`
	Name   *string `json:"name"`
}

type CategoryEventInfo struct {
	Action int    `json:"action"`
	Slug   string `json:"slug"`
}

func (u *CategoryDiff) Apply(target *Category) {
	if u.Slug != nil {
		target.Slug = *u.Slug
	}
	if u.Hidden != nil {
		target.Hidden = *u.Hidden
	}
	if u.Navbar != nil {
		target.Navbar = *u.Navbar
	}
	if u.Name != nil {
		target.Name = *u.Name
	}
}
