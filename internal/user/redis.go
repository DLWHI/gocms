package user

import (
	"context"
	"errors"
	"time"

	"gocms/internal/cache"
	"gocms/internal/domain"
	"gocms/internal/logger"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	Repository
	cache cache.RedisClient
	ttl   time.Duration
	keys  cache.Keyspace
}

func NewRedisRepository(origin Repository, client *redis.Client, ttl time.Duration, config cache.Keyspace) *RedisRepository {
	return &RedisRepository{
		Repository: origin,
		ttl:        ttl,
		keys:       config,
		cache: cache.RedisClient{
			Client:   client,
			LeaseTTL: 5 * time.Second,
		},
	}
}

func (r *RedisRepository) Create(ctx context.Context, user *domain.User, pwdhash []byte) error {
	if err := r.Repository.Create(ctx, user, pwdhash); err != nil {
		return err
	}
	r.invalidate(ctx, r.keys.List(activeList))
	return nil
}

func (r *RedisRepository) UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*domain.User, error) {
	user, err := r.Repository.UpdateLogin(ctx, username, timestamp)
	if err != nil {
		return nil, err
	}
	r.invalidate(ctx, r.keys.Key(username), r.keys.List(activeList))
	return user, nil
}

func (r *RedisRepository) Update(ctx context.Context, username string, diff map[string]interface{}) error {
	if err := r.Repository.Update(ctx, username, diff); err != nil {
		return err
	}
	r.invalidate(ctx, r.keys.Key(username), r.keys.List(activeList))
	return nil
}

func (r *RedisRepository) Get(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	var token string
	key := r.keys.Key(username)

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
	key := r.keys.List(activeList)
	err := r.cache.Get(ctx, key, &users)
	switch {
	case err == nil:
		return users, nil
	case errors.Is(err, redis.Nil):
		token, err = r.cache.Lease(ctx, key)
		if err != nil {
			logger.Warn("User.Redis", "Redis error while trying to lease key", err, "key", key)
		}
	default:
		logger.Warn("User.Redis", "Redis error while trying to retrieve key", err, "key", key)
	}

	u, err := r.Repository.GetActive(ctx)
	if err != nil {
		return nil, err
	}
	users = u

	if token != "" {
		if err := r.cache.Return(ctx, key, token, u, r.ttl); err != nil {
			if !errors.Is(err, redis.Nil) {
				logger.Warn("User.Redis", "Redis error while trying to store key", err, "key", key)
			}
		}
	}

	return users, nil
}

func (r *RedisRepository) Suspend(ctx context.Context, username string) error {
	if err := r.Repository.Suspend(ctx, username); err != nil {
		return err
	}
	r.invalidate(ctx, r.keys.Key(username), r.keys.List(activeList))
	return nil
}

func (r *RedisRepository) invalidate(ctx context.Context, keys ...string) {
	if err := r.cache.Invalidate(ctx, keys...); err != nil {
		logger.Warn("User.Redis", "Redis error while trying to delete keys", err, "keys", keys)
	}
}
