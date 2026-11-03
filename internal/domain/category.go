package domain

const (
	CategoryActionCreate = 1 << iota
	CategoryActionUpdate
	CategoryActionDelete

	CategoryEvent      = "category"
	CategoryEntityName = "categories"
)

type Category struct {
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

type CategoryEventInfo struct {
	Action int    `json:"action"`
	Slug   string `json:"slug"`
}
