package domain

const (
	CategoryActionCreate = 1 << iota
	CategoryActionUpdate
	CategoryActionDelete

	CategoryEventName  = "category"
	CategoryEntityName = "categories"
)

type Category struct {
	ID     string `json:"id,omitempty"`
	Slug   string `json:"slug"`
	Hidden bool   `json:"hidden"`
	Navbar bool   `json:"navbar"`
	Lang   string `json:"lang"`
	Name   string `json:"name"`
}

type CategoryDiff struct {
	Slug   *string `json:"slug"`
	Hidden *bool   `json:"hidden"`
	Navbar *bool   `json:"navbar"`
	Name   *string `json:"name"`
	Lang   string  `json:"lang"`
}

type CategoryEvent struct {
	Action int    `json:"action"`
	Slug   string `json:"slug"`
}
