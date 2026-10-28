package locale

import "context"

const allList = "all"

type Repository interface {
	Create(ctx context.Context, name string, code string) error
	GetAll(ctx context.Context) ([]Locale, error)
	Delete(ctx context.Context, code string) error
}
