package category

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

func (r *RedisRepository) Create(ctx context.Context, category *domain.Category) error {
	if err := r.Repository.Create(ctx, category); err != nil {
		return err
	}
	key := r.keys.List(notHiddenListBase + "-" + category.Locale)
	if err := r.cache.Invalidate(ctx, key); err != nil {
		logger.Warn("User.Redis", "Redis error while trying to delete key", err, "key", key)
	}
	return nil
}

func (r *RedisRepository) Update(ctx context.Context, slug string, locale string, diff *domain.CategoryDiff) error {
	if err := r.Repository.Update(ctx, slug, locale, diff); err != nil {
		return err
	}
	r.invalidateGroup(ctx, r.keys.Key(slug+"-*"))
	r.invalidateGroup(ctx, r.keys.List(notHiddenListBase+"-*"))
	return nil
}

func (r *RedisRepository) Get(ctx context.Context, slug string, locale string) (*domain.Category, error) {
	var category domain.Category
	var token string
	key := r.keys.Key(slug + "-" + locale)
	err := r.cache.Get(ctx, key, &category)
	switch {
	case err == nil:
		return &category, nil
	case errors.Is(err, redis.Nil):
		token, err = r.cache.Lease(ctx, key)
		if err != nil {
			logger.Warn("User.Redis", "Redis error while trying to lease key", err, "key", key)
		}
	default:
		logger.Warn("User.Redis", "Redis error while trying to retrieve key", err, "key", key)
	}

	c, err := r.Repository.Get(ctx, slug, locale)
	if err != nil {
		return nil, err
	}
	category = *c

	if token != "" {
		if err := r.cache.Return(ctx, key, token, c, r.ttl); err != nil {
			if !errors.Is(err, redis.Nil) {
				logger.Warn("User.Redis", "Redis error while trying to store key", err, "key", key)
			}
		}
	}

	return &category, nil
}

func (r *RedisRepository) GetAll(ctx context.Context, locale string) ([]domain.Category, error) {
	var categories []domain.Category
	var token string
	key := r.keys.List(notHiddenListBase + "-" + locale)
	err := r.cache.Get(ctx, key, &categories)
	switch {
	case err == nil:
		return categories, nil
	case errors.Is(err, redis.Nil):
		token, err = r.cache.Lease(ctx, key)
		if err != nil {
			logger.Warn("User.Redis", "Redis error while trying to lease key", err, "key", key)
		}
	default:
		logger.Warn("User.Redis", "Redis error while trying to retrieve key", err, "key", key)
	}

	c, err := r.Repository.GetAll(ctx, locale)
	if err != nil {
		return nil, err
	}
	categories = c

	if token != "" {
		if err := r.cache.Return(ctx, key, token, c, r.ttl); err != nil {
			if !errors.Is(err, redis.Nil) {
				logger.Warn("User.Redis", "Redis error while trying to store key", err, "key", key)
			}
		}
	}
	return categories, nil
}

func (r *RedisRepository) Delete(ctx context.Context, slug string) error {
	if err := r.Repository.Delete(ctx, slug); err != nil {
		return err
	}
	r.invalidateGroup(ctx, r.keys.Key(slug+"-*"))
	r.invalidateGroup(ctx, r.keys.List(notHiddenListBase+"-*"))
	return nil
}

func (r *RedisRepository) invalidateGroup(ctx context.Context, pattern string) {
	if err := r.cache.InvalidatePattern(ctx, pattern); err != nil {
		logger.Warn("User.Redis", "Redis error while trying to delete key pattern", err, "keys", pattern)
	}
}
