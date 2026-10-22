package category

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gocms/internal/cache"
	"gocms/internal/domain"
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

func (r *RedisRepository) Create(ctx context.Context, category *Translation) error {
	if err := r.origin.Create(ctx, category); err != nil {
		return err
	}
	allKey := fmt.Sprintf(r.keyFormat, "__all", "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, allKey); err != nil {
		logger.Warn("Category.Redis", "Failed to invalidate cached entries.", err)
	}
	notHiddenKey := fmt.Sprintf(r.keyFormat, "__!hidden", "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, notHiddenKey); err != nil {
		logger.Warn("Category.Redis", "Failed to invalidate cached entries.", err)
	}
	hiddenKey := fmt.Sprintf(r.keyFormat, "__hidden", "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, hiddenKey); err != nil {
		logger.Warn("Category.Redis", "Failed to invalidate cached entries.", err)
	}
	return nil
}

func (r *RedisRepository) Update(ctx context.Context, category *Translation) error {
	oldSlug := category.Category.Slug
	if err := r.origin.Update(ctx, category); err != nil {
		return err
	}
	pattern := fmt.Sprintf(r.keyFormat, oldSlug, "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, pattern); err != nil {
		logger.Warn("Category.Redis", "Failed to invalidate cached entries.", err)
	}

	allKey := fmt.Sprintf(r.keyFormat, "__all", category.Locale)
	if _, err := r.rdb.Del(ctx, allKey).Result(); err != nil {
		logger.Warn("Category.Redis", "Failed to delete from cache.", err)
	}
	notHiddenKey := fmt.Sprintf(r.keyFormat, "__!hidden", category.Locale)
	if _, err := r.rdb.Del(ctx, notHiddenKey).Result(); err != nil {
		logger.Warn("Category.Redis", "Failed to delete from cache.", err)
	}
	hiddenKey := fmt.Sprintf(r.keyFormat, "__hidden", "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, hiddenKey); err != nil {
		logger.Warn("Category.Redis", "Failed to invalidate cached entries.", err)
	}
	return nil
}

func (r *RedisRepository) Get(ctx context.Context, slug string, locale string) (*Translation, error) {
	var category Translation
	key := fmt.Sprintf(r.keyFormat, slug, locale)

	if err := cache.UnmarshalRedisKey(r.rdb, ctx, key, &category); err != nil {
		if !errors.Is(err, redis.Nil) {
			logger.Warn("Category.Redis", "Failed to read from cache.", err)
		}
	} else {
		return &category, nil
	}

	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		category, err := r.origin.Get(bg, slug, locale)
		if err != nil {
			if !errors.Is(err, domain.ErrNotFound) {
				r.group.Forget(key)
			}
			return nil, err
		}

		if err := cache.MarshalRedisKey(r.rdb, bg, key, r.ttl, &category); err != nil {
			logger.Warn("Category.Redis", "Failed to write to cache.", err)
		}
		return category, err
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		category = *res.Val.(*Translation)
	}

	return &category, nil
}

func (r *RedisRepository) GetAll(ctx context.Context, locale string, hidden *bool) ([]Translation, error) {
	var categories []Translation

	key := ""
	if hidden != nil {
		if *hidden {
			key = fmt.Sprintf(r.keyFormat, "__hidden", locale)
		} else {
			key = fmt.Sprintf(r.keyFormat, "__!hidden", locale)
		}
	} else {
		key = fmt.Sprintf(r.keyFormat, "__all", locale)
	}

	if err := cache.UnmarshalRedisKey(r.rdb, ctx, key, &categories); err != nil {
		if !errors.Is(err, redis.Nil) {
			logger.Warn("Category.Redis", "Failed to read from cache.", err)
		}
	} else {
		return categories, nil
	}

	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		categories, err := r.origin.GetAll(bg, locale, hidden)
		if err != nil {
			if !errors.Is(err, domain.ErrNotFound) {
				r.group.Forget(key)
			}
			return nil, err
		}

		if err := cache.MarshalRedisKey(r.rdb, bg, key, r.ttl, &categories); err != nil {
			logger.Warn("Category.Redis", "Failed to write to cache.", err)
		}

		return categories, err
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		categories = res.Val.([]Translation)
	}

	return categories, nil
}

func (r *RedisRepository) Delete(ctx context.Context, slug string) error {
	if err := r.origin.Delete(ctx, slug); err != nil {
		return err
	}
	pattern := fmt.Sprintf(r.keyFormat, slug, "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, pattern); err != nil {
		logger.Warn("Category.Redis", "Failed to invalidate cached entries.", err)
	}

	allKey := fmt.Sprintf(r.keyFormat, "__all", "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, allKey); err != nil {
		logger.Warn("Category.Redis", "Failed to invalidate cached entries.", err)
	}
	notHiddenKey := fmt.Sprintf(r.keyFormat, "__!hidden", "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, notHiddenKey); err != nil {
		logger.Warn("Category.Redis", "Failed to invalidate cached entries.", err)
	}
	hiddenKey := fmt.Sprintf(r.keyFormat, "__hidden", "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, hiddenKey); err != nil {
		logger.Warn("Category.Redis", "Failed to invalidate cached entries.", err)
	}
	return nil
}
