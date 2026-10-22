package editor

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

func (r *RedisRepository) Create(ctx context.Context, editor *Model) error {
	if err := r.origin.Create(ctx, editor); err != nil {
		return err
	}
	allKey := fmt.Sprintf(r.keyFormat, "__all", editor.Locale)
	if _, err := r.rdb.Del(ctx, allKey).Result(); err != nil {
		logger.Warn("Editor.Redis", "Failed to delete from cache.", err)
	}
	suspendedKey := fmt.Sprintf(r.keyFormat, "__dead", "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, suspendedKey); err != nil {
		logger.Warn("Editor.Redis", "Failed to invalidate cached entries.", err)
	}
	notSuspendedKey := fmt.Sprintf(r.keyFormat, "__!dead", "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, notSuspendedKey); err != nil {
		logger.Warn("Editor.Redis", "Failed to invalidate cached entries.", err)
	}
	return nil
}

func (r *RedisRepository) Update(ctx context.Context, editor *Model) error {
	if err := r.origin.Update(ctx, editor); err != nil {
		return err
	}
	pattern := fmt.Sprintf(r.keyFormat, editor.User, "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, pattern); err != nil {
		logger.Warn("Editor.Redis", "Failed to invalidate cached entries.", err)
	}

	allKey := fmt.Sprintf(r.keyFormat, "__all", editor.Locale)
	if _, err := r.rdb.Del(ctx, allKey).Result(); err != nil {
		logger.Warn("Editor.Redis", "Failed to delete from cache.", err)
	}
	suspendedKey := fmt.Sprintf(r.keyFormat, "__dead", "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, suspendedKey); err != nil {
		logger.Warn("Editor.Redis", "Failed to invalidate cached entries.", err)
	}
	notSuspendedKey := fmt.Sprintf(r.keyFormat, "__!dead", "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, notSuspendedKey); err != nil {
		logger.Warn("Editor.Redis", "Failed to invalidate cached entries.", err)
	}
	return nil
}

func (r *RedisRepository) Get(ctx context.Context, username string, locale string) (*Model, error) {
	var editor Model
	key := fmt.Sprintf(r.keyFormat, username, locale)

	if err := cache.UnmarshalRedisKey(r.rdb, ctx, key, &editor); err != nil {
		if !errors.Is(err, redis.Nil) {
			logger.Warn("Editor.Redis", "Failed to read from cache.", err)
		}
	} else {
		return &editor, nil
	}

	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		editor, err := r.origin.Get(bg, username, locale)
		if err != nil {
			if !errors.Is(err, domain.ErrNotFound) {
				r.group.Forget(key)
			}
			return nil, err
		}

		if err := cache.MarshalRedisKey(r.rdb, ctx, key, r.ttl, &editor); err != nil {
			logger.Warn("Editor.Redis", "Failed to write to cache.", err)
		}
		return editor, err
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		editor = *res.Val.(*Model)
	}

	return &editor, nil
}

func (r *RedisRepository) GetAll(ctx context.Context, locale string, suspended *bool) ([]Model, error) {
	var editors []Model
	key := ""
	if suspended != nil {
		if *suspended {
			key = fmt.Sprintf(r.keyFormat, "__dead", locale)
		} else {
			key = fmt.Sprintf(r.keyFormat, "__!dead", locale)
		}
	} else {
		key = fmt.Sprintf(r.keyFormat, "__all", locale)
	}

	if err := cache.UnmarshalRedisKey(r.rdb, ctx, key, &editors); err != nil {
		if !errors.Is(err, redis.Nil) {
			logger.Warn("Editor.Redis", "Failed to read from cache.", err)
		}
	} else {
		return editors, nil
	}

	ch := r.group.DoChan(key, func() (any, error) {
		bg, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		editors, err := r.origin.GetAll(bg, locale, suspended)
		if err != nil {
			if !errors.Is(err, domain.ErrNotFound) {
				r.group.Forget(key)
			}
			return nil, err
		}

		if err := cache.MarshalRedisKey(r.rdb, bg, key, r.ttl, &editors); err != nil {
			logger.Warn("Editor.Redis", "Failed to write to cache.", err)
		}

		return editors, err
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case res := <-ch:
		if res.Err != nil {
			return nil, res.Err
		}
		editors = res.Val.([]Model)
	}

	return editors, nil
}

func (r *RedisRepository) Delete(ctx context.Context, username string) error {
	if err := r.origin.Delete(ctx, username); err != nil {
		return err
	}

	pattern := fmt.Sprintf(r.keyFormat, username, "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, pattern); err != nil {
		logger.Warn("Editor.Redis", "Failed to invalidate cached entries.", err)
	}

	allKey := fmt.Sprintf(r.keyFormat, "__all", "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, allKey); err != nil {
		logger.Warn("Editor.Redis", "Failed to invalidate cached entries.", err)
	}
	suspendedKey := fmt.Sprintf(r.keyFormat, "__dead", "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, suspendedKey); err != nil {
		logger.Warn("Editor.Redis", "Failed to invalidate cached entries.", err)
	}
	notSuspendedKey := fmt.Sprintf(r.keyFormat, "__!dead", "*")
	if err := cache.EraseRedisKeyPattern(r.rdb, ctx, notSuspendedKey); err != nil {
		logger.Warn("Editor.Redis", "Failed to invalidate cached entries.", err)
	}

	return nil
}
