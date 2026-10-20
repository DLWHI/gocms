package locale

import (
	"context"
	"errors"
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

func (r *RedisRepository) Create(ctx context.Context, locale *Locale) error {
	if err := r.origin.Create(ctx, locale); err != nil {
		return err
	}
	if _, err := r.rdb.Unlink(ctx, r.keyFormat).Result(); err != nil {
		logger.Warn("Locale.Redis", "Failed to delete locale cache entries.", err)
	}
	return nil
}

func (r *RedisRepository) GetAll(ctx context.Context) ([]Locale, error) {
	var locales []Locale

	if err := cache.UnmarshalRedisKey(r.rdb, ctx, r.keyFormat, &locales); err != nil {
		if !errors.Is(err, redis.Nil) {
			logger.Warn("Locale.Redis", "Failed to read from cache.", err)
		}
	} else {
		return locales, nil
	}

	ch := r.group.DoChan(r.keyFormat, func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		locales, err := r.origin.GetAll(bg)
		if err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				r.group.Forget(r.keyFormat)
			}
			return nil, err
		}

		if err := cache.MarshalRedisKey(r.rdb, ctx, r.keyFormat, r.ttl, &locales); err != nil {
			logger.Warn("Locale.Redis", "Failed to write to cache.", err)
		}

		return locales, err
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		locales = res.Val.([]Locale)
	}

	return locales, nil
}

func (r *RedisRepository) Delete(ctx context.Context, locale string) error {
	if err := r.origin.Delete(ctx, locale); err != nil {
		return err
	}
	if _, err := r.rdb.Unlink(ctx, r.keyFormat).Result(); err != nil {
		logger.Warn("Locale.Redis", "Failed to delete locale cache entries.", err)
	}
	return nil
}
