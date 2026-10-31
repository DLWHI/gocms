package editor

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

func (r *RedisRepository) Create(ctx context.Context, editor *domain.Editor) error {
	if err := r.Repository.Create(ctx, editor); err != nil {
		return err
	}
	key := r.keys.List(allListBase + "-" + editor.Lang)
	if err := r.cache.Invalidate(ctx, key); err != nil {
		logger.Warn("User.Redis", "Redis error while trying to delete key", err, "key", key)
	}
	return nil
}

func (r *RedisRepository) Update(ctx context.Context, slug string, lang string, editor *domain.EditorDiff) error {
	if err := r.Repository.Update(ctx, slug, lang, editor); err != nil {
		return err
	}
	r.invalidateGroups(ctx, r.keys.Key(slug+"-"+"*"))
	r.invalidateGroups(ctx, r.keys.List(allListBase+"-"+"*"))
	return nil
}

func (r *RedisRepository) GetAny(ctx context.Context, slug string, lang string) (*domain.Editor, error) {
	var editor domain.Editor
	var token string
	key := r.keys.Key(slug + "-" + lang)
	err := r.cache.Get(ctx, key, &editor)
	switch {
	case err == nil:
		return &editor, nil
	case errors.Is(err, redis.Nil):
		token, err = r.cache.Lease(ctx, key)
		if err != nil {
			logger.Warn("User.Redis", "Redis error while trying to lease key", err, "key", key)
		}
	default:
		logger.Warn("User.Redis", "Redis error while trying to retrieve key", err, "key", key)
	}

	e, err := r.Repository.GetAny(ctx, slug, lang)
	if err != nil {
		return nil, err
	}
	editor = *e

	if token != "" {
		if err := r.cache.Return(ctx, key, token, e, r.ttl); err != nil {
			if !errors.Is(err, redis.Nil) {
				logger.Warn("User.Redis", "Redis error while trying to store key", err, "key", key)
			}
		}
	}

	return &editor, nil
}

func (r *RedisRepository) GetAll(ctx context.Context, lang string) ([]domain.Editor, error) {
	var editors []domain.Editor
	var token string
	key := r.keys.Key(allListBase + "-" + lang)
	err := r.cache.Get(ctx, key, &editors)
	switch {
	case err == nil:
		return editors, nil
	case errors.Is(err, redis.Nil):
		token, err = r.cache.Lease(ctx, key)
		if err != nil {
			logger.Warn("User.Redis", "Redis error while trying to lease key", err, "key", key)
		}
	default:
		logger.Warn("User.Redis", "Redis error while trying to retrieve key", err, "key", key)
	}

	e, err := r.Repository.GetAll(ctx, lang)
	if err != nil {
		return nil, err
	}
	editors = e

	if token != "" {
		if err := r.cache.Return(ctx, key, token, e, r.ttl); err != nil {
			if !errors.Is(err, redis.Nil) {
				logger.Warn("User.Redis", "Redis error while trying to store key", err, "key", key)
			}
		}
	}

	return editors, nil
}

func (r *RedisRepository) Delete(ctx context.Context, slug string) error {
	if err := r.Repository.Delete(ctx, slug); err != nil {
		return err
	}
	r.invalidateGroups(ctx, r.keys.Key(slug+"-"+"*"))
	r.invalidateGroups(ctx, r.keys.List(allListBase+"-"+"*"))
	return nil
}

func (r *RedisRepository) invalidateGroups(ctx context.Context, pattern string) {
	if err := r.cache.InvalidatePattern(ctx, pattern); err != nil {
		logger.Warn("User.Redis", "Redis error while trying to delete key pattern", err, "keys", pattern)
	}
}
