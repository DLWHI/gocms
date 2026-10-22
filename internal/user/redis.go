package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gocms/internal/cache"
	"gocms/internal/logger"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	Repository
	cache.RedisDecorator
	ttl       time.Duration
	keyFormat string
	allKey    string
}

type RedisConfig struct {
	Client    *redis.Client
	Ttl       time.Duration
	KeyFormat string
}

func NewRedisRepository(db Repository, config *RedisConfig) *RedisRepository {
	return &RedisRepository{
		Repository:     db,
		ttl:            config.Ttl,
		keyFormat:      config.KeyFormat,
		allKey:         fmt.Sprintf(config.KeyFormat, ":all"),
		RedisDecorator: cache.RedisDecorator{Client: config.Client},
	}
}

func (r *RedisRepository) Create(ctx context.Context, user *Record) error {
	if err := r.Repository.Create(ctx, user); err != nil {
		return err
	}
	ictx := context.WithoutCancel(ctx)
	if _, err := r.Client.Unlink(ictx, r.allKey).Result(); err != nil {
		r.logError("Failed to delete key %s from cache.", r.allKey, err)
	}
	return nil
}

func (r *RedisRepository) UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*Record, error) {
	user, err := r.Repository.UpdateLogin(ctx, username, timestamp)
	if err != nil {
		return nil, err
	}
	ictx := context.WithoutCancel(ctx)
	key := fmt.Sprintf(r.keyFormat, username)
	if _, err := r.Client.Unlink(ictx, key, key, r.allKey).Result(); err != nil {
		r.logError("Failed to delete key %s from cache.", key, err)
	}
	return user, nil
}

func (r *RedisRepository) Update(ctx context.Context, username string, diff map[string]interface{}) error {
	if err := r.Repository.Update(ctx, username, diff); err != nil {
		return err
	}
	ictx := context.WithoutCancel(ctx)
	key := fmt.Sprintf(r.keyFormat, username)
	if _, err := r.Client.Unlink(ictx, key, key, r.allKey).Result(); err != nil {
		r.logError("Failed to delete key %s from cache.", key, err)
	}
	return nil
}

func (r *RedisRepository) Get(ctx context.Context, username string) (*Record, error) {
	var user Record
	key := fmt.Sprintf(r.keyFormat, username)

	if err := r.RedisDecorator.Get(ctx, key, &user); err == nil {
		return &user, nil
	} else {
		if !errors.Is(err, redis.Nil) {
			r.logError("Failed to retrieve key %s from cache.", key, err)
		}
	}

	token, err := r.RedisDecorator.Lease(ctx, key, r.ttl)
	if err != nil {
		r.logError("Failed to lease key %s from cache.", key, err)
	}

	u, err := r.Repository.Get(ctx, username)
	if err != nil {
		return nil, err
	}
	user = *u

	if token != "" {
		if err := r.RedisDecorator.Return(ctx, key, token, u, r.ttl); err != nil {
			r.logError("Failed to store key %s to cache.", key, err)
		}
	}

	return &user, nil
}

func (r *RedisRepository) GetActive(ctx context.Context) ([]Record, error) {
	var users []Record
	if err := r.RedisDecorator.Get(ctx, r.allKey, &users); err == nil {
		return users, nil
	} else {
		if !errors.Is(err, redis.Nil) {
			r.logError("Failed to retrieve key %s from cache.", r.allKey, err)
		}
	}

	token, err := r.RedisDecorator.Lease(ctx, r.allKey, r.ttl)
	if err != nil {
		r.logError("Failed to lease key %s from cache.", r.allKey, err)
	}

	users, err = r.Repository.GetActive(ctx)
	if err != nil {
		return nil, err
	}

	if token != "" {
		if err := r.RedisDecorator.Return(ctx, r.allKey, token, users, r.ttl); err != nil {
			if !errors.Is(err, redis.Nil) {
				r.logError("Failed to store key %s to cache.", r.allKey, err)
			}
		}
	}

	return users, nil
}

func (r *RedisRepository) Suspend(ctx context.Context, username string) error {
	if err := r.Repository.Suspend(ctx, username); err != nil {
		return err
	}
	ictx := context.WithoutCancel(ctx)
	key := fmt.Sprintf(r.keyFormat, username)
	if _, err := r.Client.Unlink(ictx, key, key, r.allKey).Result(); err != nil {
		r.logError("Failed to delete key %s from cache.", key, err)
	}
	return nil
}

func (r *RedisRepository) logError(format string, key string, err error) {
	logger.Warn("User.Redis", fmt.Sprintf(format, key), err)
}
