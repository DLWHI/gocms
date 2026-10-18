package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gocms/internal/cache"
	"gocms/internal/errs"
	"gocms/internal/logger"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type RedisRepository struct {
	origin    Repository
	rdb       *redis.Client
	ttl       time.Duration
	keyFormat string
	group     singleflight.Group
}

type RedisConfig struct {
	Client    *redis.Client
	Ttl       time.Duration
	KeyFormat string
}

func NewRedisRepository(db Repository, config *RedisConfig) *RedisRepository {
	return &RedisRepository{
		origin:    db,
		rdb:       config.Client,
		ttl:       config.Ttl,
		keyFormat: config.KeyFormat,
	}
}

func (r *RedisRepository) Create(ctx context.Context, user *User) error {
	if err := r.origin.Create(ctx, user); err != nil {
		return err
	}
	allKey := fmt.Sprintf(r.keyFormat, "__all")
	if _, err := r.rdb.Del(ctx, allKey).Result(); err != nil {
		logger.Warn("User.Redis", "Failed to delete from cache.", err)
	}
	return nil
}

func (r *RedisRepository) UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*User, error) {
	user, err := r.origin.UpdateLogin(ctx, username, timestamp)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf(r.keyFormat, username)
	if _, err := r.rdb.Del(ctx, key).Result(); err != nil {
		logger.Warn("User.Redis", "Failed to delete from cache.", err)
	}
	allKey := fmt.Sprintf(r.keyFormat, "__all")
	if _, err := r.rdb.Del(ctx, allKey).Result(); err != nil {
		logger.Warn("User.Redis", "Failed to delete from cache.", err)
	}
	return user, nil
}

func (r *RedisRepository) UpdatePasswd(ctx context.Context, username string, pwdhash Password) error {
	return r.origin.UpdatePasswd(ctx, username, pwdhash)
}

func (r *RedisRepository) Update(ctx context.Context, username string, diff map[string]interface{}) (*User, error) {
	user, err := r.origin.Update(ctx, username, diff)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf(r.keyFormat, username)
	if _, err := r.rdb.Del(ctx, key).Result(); err != nil {
		logger.Warn("User.Redis", "Failed to delete from cache.", err)
	}
	allKey := fmt.Sprintf(r.keyFormat, "__all")
	if _, err := r.rdb.Del(ctx, allKey).Result(); err != nil {
		logger.Warn("User.Redis", "Failed to delete from cache.", err)
	}
	return user, nil
}

func (r *RedisRepository) Get(ctx context.Context, username string) (*User, error) {
	var user User
	key := fmt.Sprintf(r.keyFormat, username)

	if err := cache.UnmarshalRedisKey(r.rdb, ctx, key, &user); err != nil {
		if !errors.Is(err, redis.Nil) {
			logger.Warn("User.Redis", "Failed to read from cache.", err)
		}
	} else {
		return &user, nil
	}

	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		user, err := r.origin.Get(bg, username)
		if err != nil {
			if !errors.Is(err, errs.ErrNotFound) {
				r.group.Forget(key)
			}
			return nil, err
		}

		if err := cache.MarshalRedisKey(r.rdb, ctx, key, r.ttl, &user); err != nil {
			logger.Warn("User.Redis", "Failed to write to cache.", err)
		}

		return user, err
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

func (r *RedisRepository) GetAll(ctx context.Context) ([]User, error) {
	var users []User
	key := fmt.Sprintf(r.keyFormat, "__all")
	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		user, err := r.origin.GetAll(bg)
		if err != nil {
			if !errors.Is(err, errs.ErrNotFound) {
				r.group.Forget(key)
			}
			return nil, err
		}

		return user, err
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

func (r *RedisRepository) GetPasswdHash(ctx context.Context, username string) (Password, error) {
	return r.origin.GetPasswdHash(ctx, username)
}

func (r *RedisRepository) Delete(ctx context.Context, username string) error {
	if err := r.origin.Delete(ctx, username); err != nil {
		return err
	}
	key := fmt.Sprintf(r.keyFormat, username)
	if _, err := r.rdb.Del(ctx, key).Result(); err != nil {
		logger.Warn("User.Redis", "Failed to delete from cache.", err)
	}
	return nil
}
