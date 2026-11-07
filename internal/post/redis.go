package post

import (
	"context"
	"errors"
	"strconv"
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

func (r *RedisRepository) Update(ctx context.Context, id uint, language string, diff *domain.PostDiff) error {
	if err := r.Repository.Update(ctx, id, language, diff); err != nil {
		return err
	}
	pattern := r.keys.KeyWithSuffix(strconv.FormatUint(uint64(id), 10), "translation:*")
	if err := r.cache.InvalidatePattern(ctx, pattern); err != nil {
		logger.Warn("Post.Redis", "Failed to invalidate cached entries.", err)
	}
	return nil
}

func (r *RedisRepository) Get(ctx context.Context, id uint, language string) (*domain.Post, error) {
	var post domain.Post
	var token string
	key := r.keys.KeyWithSuffix(strconv.FormatUint(uint64(id), 10), "translation:"+language)

	err := r.cache.Get(ctx, key, &post)
	switch {
	case err == nil:
		return &post, nil
	case errors.Is(err, redis.Nil):
		token, err = r.cache.Lease(ctx, key)
		if err != nil {
			logger.Warn("User.Redis", "Redis error while trying to lease key", err, "key", key)
		}
	default:
		logger.Warn("User.Redis", "Redis error while trying to retrieve key", err, "key", key)
	}

	p, err := r.Repository.Get(ctx, id, language)
	if err != nil {
		return nil, err
	}
	post = *p

	if token != "" {
		if err := r.cache.Return(ctx, key, token, p, r.ttl); err != nil {
			if !errors.Is(err, redis.Nil) {
				logger.Warn("User.Redis", "Redis error while trying to store key", err, "key", key)
			}
		}
	}

	return &post, nil
}

func (r *RedisRepository) CreatorOf(ctx context.Context, id uint) (string, error) {
	var token string
	var creator string
	key := r.keys.KeyWithSuffix(strconv.FormatUint(uint64(id), 10), "created_by")

	err := r.cache.Get(ctx, key, &creator)
	switch {
	case err == nil:
		return creator, nil
	case errors.Is(err, redis.Nil):
		token, err = r.cache.Lease(ctx, key)
		if err != nil {
			logger.Warn("User.Redis", "Redis error while trying to lease key", err, "key", key)
		}
	default:
		logger.Warn("User.Redis", "Redis error while trying to retrieve key", err, "key", key)
	}

	c, err := r.Repository.GetCreator(ctx, id)
	if err != nil {
		return "", err
	}
	creator = c

	if token != "" {
		if err := r.cache.Return(ctx, key, token, &c, r.ttl); err != nil {
			if !errors.Is(err, redis.Nil) {
				logger.Warn("User.Redis", "Redis error while trying to store key", err, "key", key)
			}
		}
	}

	return creator, nil
}

func (r *RedisRepository) Delete(ctx context.Context, id uint) error {
	if err := r.Repository.Delete(ctx, id); err != nil {
		return err
	}

	pattern := r.keys.KeyWithSuffix(strconv.FormatUint(uint64(id), 10), "translation:*")
	if err := r.cache.InvalidatePattern(ctx, pattern); err != nil {
		logger.Warn("Post.Redis", "Failed to invalidate cached entries.", err)
	}

	return nil
}
