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
		logger.Warn("User.Redis", "Redis error while trying to delete key", err, "key", r.allKey)
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
		logger.Warn("User.Redis", "Redis error while trying to delete key", err, "key", key)
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
		logger.Warn("User.Redis", "Redis error while trying to delete keys", err, "keys", []string{key, r.allKey})
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
			logger.Warn("User.Redis", "Redis error while trying to retrieve key", err, "key", key)
		}
	}

	token, err := r.RedisDecorator.Lease(ctx, key, r.ttl)
	if err != nil {
		logger.Warn("User.Redis", "Redis error while trying to lease key", err, "key", key)
	}

	u, err := r.Repository.Get(ctx, username)
	if err != nil {
		return nil, err
	}
	user = *u

	if token != "" {
		if err := r.RedisDecorator.Return(ctx, key, token, u, r.ttl); err != nil {
			logger.Warn("User.Redis", "Redis error while trying to store key", err, "key", key)
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
			logger.Warn("User.Redis", "Redis error while trying to retrieve key", err, "key", r.allKey)
		}
	}

	token, err := r.RedisDecorator.Lease(ctx, r.allKey, r.ttl)
	if err != nil {
		logger.Warn("User.Redis", "Redis error while trying to lease key", err, "key", r.allKey)
	}

	users, err = r.Repository.GetActive(ctx)
	if err != nil {
		return nil, err
	}

	if token != "" {
		if err := r.RedisDecorator.Return(ctx, r.allKey, token, users, r.ttl); err != nil {
			if !errors.Is(err, redis.Nil) {
				logger.Warn("User.Redis", "Redis error while trying to store key", err, "key", r.allKey)
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
		logger.Warn("User.Redis", "Redis error while trying to delete keys", err, "keys", []string{key, r.allKey})
	}
	return nil
}
