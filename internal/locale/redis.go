package locale

import (
	"context"
	"errors"
	"time"

	"gocms/internal/cache"
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
		cache: cache.RedisClient{
			Client:   client,
			LeaseTTL: 5 * time.Second,
		},
		ttl:  ttl,
		keys: config,
	}
}

func (r *RedisRepository) Create(ctx context.Context, name string, code string) error {
	if err := r.Repository.Create(ctx, name, code); err != nil {
		return err
	}
	if err := r.cache.Invalidate(ctx, r.keys.List(allList)); err != nil {
		logger.Warn("Locale.Redis", "Failed to delete locale cache entries.", err)
	}
	return nil
}

func (r *RedisRepository) GetAll(ctx context.Context) ([]Locale, error) {
	var locales []Locale
	var token string
	key := r.keys.List(allList)
	err := r.cache.Get(ctx, key, &locales)
	switch {
	case err == nil:
		return locales, nil
	case errors.Is(err, redis.Nil):
		token, err = r.cache.Lease(ctx, key)
		if err != nil {
			logger.Warn("User.Redis", "Redis error while trying to lease key", err, "key", key)
		}
	default:
		logger.Warn("User.Redis", "Redis error while trying to retrieve key", err, "key", key)
	}

	locales, err = r.Repository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	if token != "" {
		if err := r.cache.Return(ctx, key, token, &locales, r.ttl); err != nil {
			if !errors.Is(err, redis.Nil) {
				logger.Warn("User.Redis", "Redis error while trying to store key", err, "key", key)
			}
		}
	}
	return locales, nil
}

func (r *RedisRepository) Delete(ctx context.Context, code string) error {
	if err := r.Repository.Delete(ctx, code); err != nil {
		return err
	}
	if err := r.cache.Invalidate(ctx, r.keys.List(allList)); err != nil {
		logger.Warn("Locale.Redis", "Failed to delete locale cache entries.", err)
	}
	return nil
}
