package user

import (
	"context"
	"gocms/internal/cache"
	"gocms/internal/domain"
	"time"

	"golang.org/x/sync/singleflight"
)

type SingleflightRepository struct {
	Repository
	keys    cache.Keyspace
	group   singleflight.Group
	timeout time.Duration
}

func NewSingleflightRepository(db Repository, config cache.Keyspace) *SingleflightRepository {
	return &SingleflightRepository{
		Repository: db,
		keys:       config,
		timeout:    20 * time.Second,
	}
}

func (r *SingleflightRepository) Create(ctx context.Context, user *domain.User, pwdhash []byte) error {
	if err := r.Repository.Create(ctx, user, pwdhash); err != nil {
		return err
	}
	r.group.Forget(r.keys.List(activeList))
	return nil
}

func (r *SingleflightRepository) UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*domain.User, error) {
	key := r.keys.Key(username)
	user, err := r.Repository.UpdateLogin(ctx, username, timestamp)
	if err != nil {
		return nil, err
	}
	r.group.Forget(key)
	r.group.Forget(r.keys.List(activeList))
	return user, nil
}

func (r *SingleflightRepository) Update(ctx context.Context, username string, diff map[string]interface{}) error {
	key := r.keys.Key(username)
	if err := r.Repository.Update(ctx, username, diff); err != nil {
		return err
	}
	r.group.Forget(key)
	r.group.Forget(r.keys.List(activeList))
	return nil
}

func (r *SingleflightRepository) Get(ctx context.Context, username string) (*domain.User, error) {
	key := r.keys.Key(username)
	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.timeout)
		defer cancel()

		retrieved, err := r.Repository.Get(bg, username)
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
		return res.Val.(*domain.User), nil
	}

}

func (r *SingleflightRepository) GetActive(ctx context.Context) ([]domain.User, error) {
	ch := r.group.DoChan(r.keys.List(activeList), func() (any, error) {
		bg, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.timeout)
		defer cancel()

		retrieved, err := r.Repository.GetActive(bg)
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
		return res.Val.([]domain.User), nil
	}
}

func (r *SingleflightRepository) Suspend(ctx context.Context, username string) error {
	key := r.keys.Key(username)
	if err := r.Repository.Suspend(ctx, username); err != nil {
		return err
	}
	r.group.Forget(key)
	r.group.Forget(r.keys.List(activeList))
	return nil
}
