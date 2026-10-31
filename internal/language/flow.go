package language

import (
	"context"
	"time"

	"gocms/internal/cache"
	"gocms/internal/domain"
	"gocms/internal/flow"
)

type FlowRepository struct {
	Repository
	keys    cache.Keyspace
	group   flow.Group
	timeout time.Duration
}

func NewFlowRepository(origin Repository, timeout time.Duration, config cache.Keyspace) *FlowRepository {
	return &FlowRepository{
		Repository: origin,
		keys:       config,
		timeout:    timeout,
	}
}

func (r *FlowRepository) Create(ctx context.Context, lang *domain.Language) error {
	if err := r.Repository.Create(ctx, lang); err != nil {
		return err
	}
	r.group.Forget(r.keys.List(allList))
	return nil
}

func (r *FlowRepository) GetAll(ctx context.Context) ([]domain.Language, error) {
	ch := r.group.DoChan(r.keys.List(allList), func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), r.timeout)
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
		return res.Val.([]domain.Language), nil
	}
}

func (r *FlowRepository) Delete(ctx context.Context, code string) error {
	if err := r.Repository.Delete(ctx, code); err != nil {
		return err
	}
	r.group.Forget(r.keys.List(allList))
	return nil
}
