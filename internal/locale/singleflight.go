package locale

import (
	"context"
	"time"

	"gocms/internal/cache"

	"golang.org/x/sync/singleflight"
)

type SingleflightRepository struct {
	Repository
	keys    cache.Keyspace
	group   singleflight.Group
	timeout time.Duration
}

func NewSingleflightRepository(origin Repository, config cache.Keyspace) *SingleflightRepository {
	return &SingleflightRepository{
		Repository: origin,
		keys:       config,
	}
}

func (r *SingleflightRepository) Create(ctx context.Context, name string, code string) error {
	if err := r.Repository.Create(ctx, name, code); err != nil {
		return err
	}
	r.group.Forget(r.keys.List(allList))
	return nil
}

func (r *SingleflightRepository) GetAll(ctx context.Context) ([]Locale, error) {
	ch := r.group.DoChan(r.keys.List(allList), func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		locales, err := r.Repository.GetAll(bg)
		if err != nil {
			return nil, err
		}

		return locales, err
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		return res.Val.([]Locale), nil
	}
}

func (r *SingleflightRepository) Delete(ctx context.Context, code string) error {
	if err := r.Repository.Delete(ctx, code); err != nil {
		return err
	}
	r.group.Forget(r.keys.List(allList))
	return nil
}
