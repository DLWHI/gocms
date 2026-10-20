package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gocms/internal/errs"

	"golang.org/x/sync/singleflight"
)

type SingleflightRepository struct {
	origin    Repository
	keyFormat string
	allKey    string
	group     singleflight.Group
}

func NewSingleflightRepository(db Repository, keyFormat string) *SingleflightRepository {
	return &SingleflightRepository{
		origin:    db,
		keyFormat: keyFormat,
		allKey:    fmt.Sprintf(keyFormat, ":all"),
	}
}

func (r *SingleflightRepository) Create(ctx context.Context, user *User) error {
	if err := r.origin.Create(ctx, user); err != nil {
		return err
	}
	return nil
}

func (r *SingleflightRepository) UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*User, error) {
	key := fmt.Sprintf(r.keyFormat, username)
	r.group.Forget(key)
	user, err := r.origin.UpdateLogin(ctx, username, timestamp)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *SingleflightRepository) UpdatePasswd(ctx context.Context, username string, pwdhash Password) error {
	return r.origin.UpdatePasswd(ctx, username, pwdhash)
}

func (r *SingleflightRepository) Update(ctx context.Context, username string, diff map[string]interface{}) error {
	key := fmt.Sprintf(r.keyFormat, username)
	r.group.Forget(key)
	if err := r.origin.Update(ctx, username, diff); err != nil {
		return err
	}
	return nil
}

func (r *SingleflightRepository) Get(ctx context.Context, username string) (*User, error) {
	var user User
	key := fmt.Sprintf(r.keyFormat, username)
	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		retreived, err := r.origin.Get(bg, username)
		if err != nil {
			if !errors.Is(err, errs.ErrNotFound) {
				r.group.Forget(key)
			}
			return nil, err
		}

		return retreived, err
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		user = *res.Val.(*User)
	}

	return &user, nil
}

func (r *SingleflightRepository) GetAll(ctx context.Context) ([]User, error) {
	var users []User
	ch := r.group.DoChan(r.allKey, func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		retreived, err := r.origin.GetAll(bg)
		if err != nil {
			r.group.Forget(r.allKey)
			return nil, err
		}

		return retreived, err
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		users = res.Val.([]User)
	}

	return users, nil
}

func (r *SingleflightRepository) GetPasswdHash(ctx context.Context, username string) (Password, error) {
	return r.origin.GetPasswdHash(ctx, username)
}

func (r *SingleflightRepository) Delete(ctx context.Context, username string) error {
	key := fmt.Sprintf(r.keyFormat, username)
	r.group.Forget(key)
	if err := r.origin.Delete(ctx, username); err != nil {
		return err
	}
	return nil
}
