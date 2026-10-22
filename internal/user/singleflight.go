package user

import (
	"context"
	"fmt"
	"gocms/internal/domain"
	"time"

	"golang.org/x/sync/singleflight"
)

type SingleflightRepository struct {
	Repository
	keyFormat string
	allKey    string
	group     singleflight.Group
	timeout   time.Duration
}

func NewSingleflightRepository(db Repository, keyFormat string) *SingleflightRepository {
	return &SingleflightRepository{
		Repository: db,
		keyFormat:  keyFormat,
		allKey:     fmt.Sprintf(keyFormat, ":all"),
		timeout:    20 * time.Second,
	}
}

func (r *SingleflightRepository) Create(ctx context.Context, user *domain.User, pwdhash domain.Password) error {
	if err := r.Repository.Create(ctx, user, pwdhash); err != nil {
		return err
	}
	r.group.Forget(r.allKey)
	return nil
}

func (r *SingleflightRepository) UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*domain.User, error) {
	key := fmt.Sprintf(r.keyFormat, username)
	user, err := r.Repository.UpdateLogin(ctx, username, timestamp)
	if err != nil {
		return nil, err
	}
	r.group.Forget(key)
	r.group.Forget(r.allKey)
	return user, nil
}

func (r *SingleflightRepository) Update(ctx context.Context, username string, user *domain.User) error {
	key := fmt.Sprintf(r.keyFormat, username)
	if err := r.Repository.Update(ctx, username, user); err != nil {
		return err
	}
	r.group.Forget(key)
	r.group.Forget(r.allKey)
	return nil
}

func (r *SingleflightRepository) Get(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	key := fmt.Sprintf(r.keyFormat, username)
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
		user = *res.Val.(*domain.User)
	}

	return &user, nil
}

func (r *SingleflightRepository) GetActive(ctx context.Context) ([]domain.User, error) {
	var users []domain.User
	ch := r.group.DoChan(r.allKey, func() (any, error) {
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
		users = res.Val.([]domain.User)
	}

	return users, nil
}

func (r *SingleflightRepository) Suspend(ctx context.Context, username string) error {
	key := fmt.Sprintf(r.keyFormat, username)
	if err := r.Repository.Suspend(ctx, username); err != nil {
		return err
	}
	r.group.Forget(key)
	r.group.Forget(r.allKey)
	return nil
}
