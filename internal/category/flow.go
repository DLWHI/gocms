package category

import (
	"context"
	"gocms/internal/cache"
	"gocms/internal/domain"
	"gocms/internal/flow"
	"time"
)

type SingleflightRepository struct {
	Repository
	keys    cache.Keyspace
	group   flow.Group
	timeout time.Duration
}

func NewSingleflightRepository(db Repository, timeout time.Duration, config cache.Keyspace) *SingleflightRepository {
	return &SingleflightRepository{
		Repository: db,
		keys:       config,
		timeout:    timeout,
	}
}

func (r *SingleflightRepository) Create(ctx context.Context, category *domain.Category) error {
	if err := r.Repository.Create(ctx, category); err != nil {
		return err
	}
	r.group.Forget(r.keys.List(notHiddenListBase + "-" + category.Locale))
	return nil
}

func (r *SingleflightRepository) Update(ctx context.Context, slug string, locale string, diff *domain.CategoryDiff) error {
	if err := r.Repository.Update(ctx, slug, locale, diff); err != nil {
		return err
	}
	r.group.ForgetPattern(r.keys.List(notHiddenListBase + "-*"))
	r.group.ForgetPattern(r.keys.Key(slug + "-*"))
	return nil
}

func (r *SingleflightRepository) Get(ctx context.Context, slug string, locale string) (*domain.Category, error) {
	key := r.keys.Key(notHiddenListBase + "-" + locale)
	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.timeout)
		defer cancel()

		retrieved, err := r.Repository.Get(bg, slug, locale)
		if err != nil {
			return nil, err
		}
		return retrieved, err
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		return res.Val.(*domain.Category), nil
	}
}

func (r *SingleflightRepository) GetNotHidden(ctx context.Context, locale string) ([]domain.Category, error) {
	key := r.keys.Key(notHiddenListBase + "-" + locale)
	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.timeout)
		defer cancel()

		retrieved, err := r.Repository.GetNotHidden(bg, locale)
		if err != nil {
			return nil, err
		}
		return retrieved, err
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		return res.Val.([]domain.Category), nil
	}
}

func (r *SingleflightRepository) Delete(ctx context.Context, slug string) error {
	if err := r.Repository.Delete(ctx, slug); err != nil {
		return err
	}
	r.group.ForgetPattern(r.keys.List(notHiddenListBase + "-*"))
	r.group.ForgetPattern(r.keys.Key(slug + "-*"))
	return nil
}
