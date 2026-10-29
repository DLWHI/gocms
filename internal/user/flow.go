package user

import (
	"context"
	"gocms/internal/cache"
	"gocms/internal/domain"
	"gocms/internal/flow"
	"time"
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

func (r *FlowRepository) Create(ctx context.Context, user *domain.User, pwdhash []byte) error {
	if err := r.Repository.Create(ctx, user, pwdhash); err != nil {
		return err
	}
	r.group.Forget(r.keys.List(activeList))
	return nil
}

func (r *FlowRepository) UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*domain.User, error) {
	user, err := r.Repository.UpdateLogin(ctx, username, timestamp)
	if err != nil {
		return nil, err
	}
	r.group.Forget(r.keys.Key(username))
	r.group.Forget(r.keys.List(activeList))
	return user, nil
}

func (r *FlowRepository) Update(ctx context.Context, username string, diff *domain.UserDiff) error {
	if err := r.Repository.Update(ctx, username, diff); err != nil {
		return err
	}
	r.group.Forget(r.keys.Key(username))
	r.group.Forget(r.keys.List(activeList))
	return nil
}

func (r *FlowRepository) Get(ctx context.Context, username string) (*domain.User, error) {
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

func (r *FlowRepository) GetActive(ctx context.Context) ([]domain.User, error) {
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

func (r *FlowRepository) Suspend(ctx context.Context, username string) error {
	if err := r.Repository.Suspend(ctx, username); err != nil {
		return err
	}
	r.group.Forget(r.keys.Key(username))
	r.group.Forget(r.keys.List(activeList))
	return nil
}
