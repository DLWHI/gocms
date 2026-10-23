package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gocms/internal/cache"
	"gocms/internal/domain"
	"gocms/internal/logger"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	Repository
	cache     cache.RedisClient
	ttl       time.Duration
	keyFormat string
	allKey    string
}

type RedisConfig struct {
	Client    *redis.Client
	TTL       time.Duration
	KeyFormat string
}

func NewRedisRepository(db Repository, config *RedisConfig) *RedisRepository {
	return &RedisRepository{
		Repository: db,
		ttl:        config.TTL,
		keyFormat:  config.KeyFormat,
		allKey:     fmt.Sprintf(config.KeyFormat, ":all"),
		cache: cache.RedisClient{
			Client:   config.Client,
			LeaseTTL: 5 * time.Second,
		},
	}
}

func (r *RedisRepository) Create(ctx context.Context, user *domain.User, pwdhash []byte) error {
	if err := r.Repository.Create(ctx, user, pwdhash); err != nil {
		return err
	}
	r.invalidate(ctx, r.allKey)
	return nil
}

func (r *RedisRepository) UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*domain.User, error) {
	user, err := r.Repository.UpdateLogin(ctx, username, timestamp)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf(r.keyFormat, username)
	r.invalidate(ctx, r.allKey, key)
	return user, nil
}

func (r *RedisRepository) Update(ctx context.Context, username string, diff map[string]interface{}) error {
	if err := r.Repository.Update(ctx, username, diff); err != nil {
		return err
	}
	key := fmt.Sprintf(r.keyFormat, username)
	r.invalidate(ctx, r.allKey, key)
	return nil
}

func (r *RedisRepository) Get(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	var token string
	key := fmt.Sprintf(r.keyFormat, username)

	err := r.cache.Get(ctx, key, &user)
	switch {
	case err == nil:
		return &user, nil
	case errors.Is(err, redis.Nil):
		token, err = r.cache.Lease(ctx, key)
		if err != nil {
			logger.Warn("User.Redis", "Redis error while trying to lease key", err, "key", key)
		}
	default:
		logger.Warn("User.Redis", "Redis error while trying to retrieve key", err, "key", key)
	}

	u, err := r.Repository.Get(ctx, username)
	if err != nil {
		return nil, err
	}
	user = *u

	if token != "" {
		if err := r.cache.Return(ctx, key, token, u, r.ttl); err != nil {
			if !errors.Is(err, redis.Nil) {
				logger.Warn("User.Redis", "Redis error while trying to store key", err, "key", key)
			}
		}
	}

	return &user, nil
}

func (r *RedisRepository) GetActive(ctx context.Context) ([]domain.User, error) {
	var users []domain.User
	var token string

	err := r.cache.Get(ctx, r.allKey, &users)
	switch {
	case err == nil:
		return users, nil
	case errors.Is(err, redis.Nil):
		token, err = r.cache.Lease(ctx, r.allKey)
		if err != nil {
			logger.Warn("User.Redis", "Redis error while trying to lease key", err, "key", r.allKey)
		}
	default:
		logger.Warn("User.Redis", "Redis error while trying to retrieve key", err, "key", r.allKey)
	}

	u, err := r.Repository.GetActive(ctx)
	if err != nil {
		return nil, err
	}
	users = u

	if token != "" {
		if err := r.cache.Return(ctx, r.allKey, token, u, r.ttl); err != nil {
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
	key := fmt.Sprintf(r.keyFormat, username)
	r.invalidate(ctx, r.allKey, key)
	return nil
}

func (r *RedisRepository) invalidate(ctx context.Context, keys ...string) {
	ictx := context.WithoutCancel(ctx)
	if _, err := r.cache.Client.Unlink(ictx, keys...).Result(); err != nil {
		logger.Warn("User.Redis", "Redis error while trying to delete keys", err, "keys", keys)
	}
}
