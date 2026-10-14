package post

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

func (r *RedisRepository) Create(ctx context.Context, post *Translation) error {
	return r.origin.Create(ctx, post)
}

func (r *RedisRepository) Update(ctx context.Context, post *Meta) error {
	if err := r.origin.Update(ctx, post); err != nil {
		return err
	}
	pattern := fmt.Sprintf(r.keyFormat, post.ID, "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, pattern); err != nil {
		logger.Warn("Post.Redis", "Failed to invalidate cached entries.", err)
	}
	return nil
}

func (r *RedisRepository) UpdateTranslation(ctx context.Context, post *Translation) error {
	if err := r.origin.UpdateTranslation(ctx, post); err != nil {
		return err
	}
	pattern := fmt.Sprintf(r.keyFormat, post.PostID, "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, pattern); err != nil {
		logger.Warn("Post.Redis", "Failed to invalidate cached entries.", err)
	}
	return nil
}

func (r *RedisRepository) IncrementViews(ctx context.Context, id uint, value uint) error {
	return r.origin.IncrementViews(ctx, id, value)
}

func (r *RedisRepository) Get(ctx context.Context, id uint) (*Meta, error) {
	var post Meta
	key := fmt.Sprintf(r.keyFormat, id, "__all")

	if err := cache.UnmarshalRedisKey(r.rdb, ctx, key, &post); err != nil {
		if !errors.Is(err, redis.Nil) {
			logger.Warn("Post.Redis", "Failed to read from cache.", err)
		}
	} else {
		return &post, nil
	}

	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		post, err := r.origin.Get(bg, id)
		if err != nil {
			if !errors.Is(err, errs.ErrNotFound) {
				r.group.Forget(key)
			}
			return nil, err
		}

		if err := cache.MarshalRedisKey(r.rdb, bg, key, r.ttl, &post); err != nil {
			logger.Warn("Post.Redis", "Failed to write to cache.", err)
		}

		return post, nil
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		post = *res.Val.(*Meta)
	}

	return &post, nil
}

func (r *RedisRepository) GetTranslation(ctx context.Context, id uint, locale string) (*Translation, error) {
	var post Translation
	key := fmt.Sprintf(r.keyFormat, id, locale)

	if err := cache.UnmarshalRedisKey(r.rdb, ctx, key, &post); err != nil {
		if !errors.Is(err, redis.Nil) {
			logger.Warn("Post.Redis", "Failed to read from cache.", err)
		}
	} else {
		return &post, nil
	}

	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		post, err := r.origin.GetTranslation(bg, id, locale)
		if err != nil {
			if !errors.Is(err, errs.ErrNotFound) {
				r.group.Forget(key)
			}
			return nil, err
		}

		if err := cache.MarshalRedisKey(r.rdb, bg, key, r.ttl, &post); err != nil {
			logger.Warn("Post.Redis", "Failed to write to cache.", err)
		}

		return post, nil
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		post = *res.Val.(*Translation)
	}

	return &post, nil
}

func (r *RedisRepository) GetCategories(ctx context.Context, id uint) ([]string, error) {
	categories := make([]string, 0)

	key := fmt.Sprintf(r.keyFormat, id, "categories")

	if err := cache.UnmarshalRedisKey(r.rdb, ctx, key, &categories); err != nil {
		if !errors.Is(err, redis.Nil) {
			logger.Warn("Post.Redis", "Failed to read from cache.", err)
		}
	} else {
		return categories, nil
	}

	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		post, err := r.origin.GetCategories(bg, id)
		if err != nil {
			if !errors.Is(err, errs.ErrNotFound) {
				r.group.Forget(key)
			}
			return nil, err
		}

		if err := cache.MarshalRedisKey(r.rdb, bg, key, r.ttl, &post); err != nil {
			logger.Warn("Post.Redis", "Failed to write to cache.", err)
		}

		return post, nil
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		categories = res.Val.([]string)
	}
	return categories, nil
}

func (r *RedisRepository) Delete(ctx context.Context, id uint) error {
	if err := r.origin.Delete(ctx, id); err != nil {
		return err
	}

	pattern := fmt.Sprintf(r.keyFormat, id, "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, pattern); err != nil {
		logger.Warn("Post.Redis", "Failed to invalidate cached entries.", err)
	}

	return nil
}

func (r *RedisRepository) DeleteTranslation(ctx context.Context, id uint, locale string) error {
	if err := r.origin.DeleteTranslation(ctx, id, locale); err != nil {
		return err
	}

	key := fmt.Sprintf(r.keyFormat, id, locale)
	if _, err := r.rdb.Del(ctx, key).Result(); err != nil {
		logger.Warn("Post.Redis", "Failed to delete from cache.", err)
	}
	return nil
}
