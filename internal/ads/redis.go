package ads

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

func (r *RedisRepository) Create(ctx context.Context, ad *Model) error {
	if err := r.origin.Create(ctx, ad); err != nil {
		return err
	}
	key := fmt.Sprintf(r.keyFormat, 0)
	pattern := fmt.Sprintf("%s:%s", key, "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, pattern); err != nil {
		logger.Warn("Ads.Redis", "Failed to invalidate cached entries.", err)
	}
	return nil
}

func (r *RedisRepository) Update(ctx context.Context, ad *Model) error {
	if err := r.origin.Update(ctx, ad); err != nil {
		return err
	}
	key := fmt.Sprintf(r.keyFormat, 0)
	pattern := fmt.Sprintf("%s:%s", key, "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, pattern); err != nil {
		logger.Warn("Ads.Redis", "Failed to invalidate cached entries.", err)
	}
	return nil
}

func (r *RedisRepository) Get(ctx context.Context, id uint) (*Model, error) {
	var ad Model
	key := fmt.Sprintf(r.keyFormat, id)

	if err := cache.UnmarshalRedisKey(r.rdb, ctx, key, &ad); err != nil {
		if !errors.Is(err, redis.Nil) {
			logger.Warn("Ads.Redis", "Failed to read from cache.", err)
		}
	} else {
		return &ad, nil
	}

	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		ad, err := r.origin.Get(bg, id)
		if err != nil {
			if !errors.Is(err, errs.ErrNotFound) {
				r.group.Forget(key)
			}
			return nil, err
		}

		if err := cache.MarshalRedisKey(r.rdb, ctx, key, r.ttl, &ad); err != nil {
			logger.Warn("Ads.Redis", "Failed to write to cache.", err)
		}
		return ad, nil
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		ad = *res.Val.(*Model)
	}

	return &ad, nil
}

func (r *RedisRepository) GetGroup(ctx context.Context, group string) ([]Model, error) {
	var ads []Model

	key := fmt.Sprintf(r.keyFormat, 0)
	key = fmt.Sprintf("%s:%s", key, group)

	if err := cache.UnmarshalRedisKey(r.rdb, ctx, key, &ads); err != nil {
		if !errors.Is(err, redis.Nil) {
			logger.Warn("Ads.Redis", "Failed to read from cache.", err)
		}
	} else {
		return ads, nil
	}

	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		ads, err := r.origin.GetGroup(bg, group)
		if err != nil {
			if !errors.Is(err, errs.ErrNotFound) {
				r.group.Forget(key)
			}
			return nil, err
		}

		if err := cache.MarshalRedisKey(r.rdb, ctx, key, r.ttl, &ads); err != nil {
			logger.Warn("Ads.Redis", "Failed to write to cache.", err)
		}

		return ads, err
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		ads = res.Val.([]Model)
	}

	return ads, nil
}

func (r *RedisRepository) GetGroups(ctx context.Context) ([]string, error) {
	var groups []string

	key := fmt.Sprintf(r.keyFormat, 0)
	key += ":groups"
	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		groups, err := r.origin.GetGroups(bg)
		if err != nil {
			if !errors.Is(err, errs.ErrNotFound) {
				r.group.Forget(key)
			}
			return nil, err
		}

		return groups, err
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		groups = res.Val.([]string)
	}

	return groups, nil
}

func (r *RedisRepository) Delete(ctx context.Context, id uint) error {
	if err := r.origin.Delete(ctx, id); err != nil {
		return err
	}

	key := fmt.Sprintf(r.keyFormat, id)
	if _, err := r.rdb.Del(ctx, key).Result(); err != nil {
		logger.Warn("Ads.Redis", "Failed to delete from cache.", err)
	}

	key = fmt.Sprintf(r.keyFormat, 0)
	pattern := fmt.Sprintf("%s:%s", key, "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, pattern); err != nil {
		logger.Warn("Ads.Redis", "Failed to invalidate cached entries.", err)
	}
	return nil
}
