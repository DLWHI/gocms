package ads

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

func (r *RedisRepository) Create(ctx context.Context, ad *domain.Advertisement) error {
	if err := r.Repository.Create(ctx, ad); err != nil {
		return err
	}
	if ad.Group != "" {
		r.invalidate(ctx, r.keys.List(groupsList), r.keys.Key(ad.Group))
	}
	return nil
}

func (r *RedisRepository) Update(ctx context.Context, id uint, ad *domain.AdvertisementDiff) error {
	if err := r.Repository.Update(ctx, id, ad); err != nil {
		return err
	}
	r.invalidate(ctx, r.keys.Key(strconv.FormatUint(uint64(id), 10)))
	r.invalidate(ctx, r.keys.List(groupsList), r.keys.KeyWithSuffix("*", "group"))
	return nil
}

func (r *RedisRepository) Get(ctx context.Context, id uint) (*domain.Advertisement, error) {
	var ad domain.Advertisement
	var token string
	key := r.keys.Key(strconv.FormatUint(uint64(id), 10))
	err := r.cache.Get(ctx, key, &ad)
	switch {
	case err == nil:
		return &ad, nil
	case errors.Is(err, redis.Nil):
		token, err = r.cache.Lease(ctx, key)
		if err != nil {
			logger.Warn("User.Redis", "Redis error while trying to lease key", err, "key", key)
		}
	default:
		logger.Warn("User.Redis", "Redis error while trying to retrieve key", err, "key", key)
	}

	a, err := r.Repository.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	ad = *a

	if token != "" {
		if err := r.cache.Return(ctx, key, token, a, r.ttl); err != nil {
			if !errors.Is(err, redis.Nil) {
				logger.Warn("User.Redis", "Redis error while trying to store key", err, "key", key)
			}
		}
	}

	return &ad, nil
}

func (r *RedisRepository) GetGroup(ctx context.Context, group string) ([]domain.Advertisement, error) {
	var ads []domain.Advertisement
	var token string
	key := r.keys.KeyWithSuffix(group, "group")
	err := r.cache.Get(ctx, key, &ads)
	switch {
	case err == nil:
		return ads, nil
	case errors.Is(err, redis.Nil):
		token, err = r.cache.Lease(ctx, key)
		if err != nil {
			logger.Warn("User.Redis", "Redis error while trying to lease key", err, "key", key)
		}
	default:
		logger.Warn("User.Redis", "Redis error while trying to retrieve key", err, "key", key)
	}

	a, err := r.Repository.GetGroup(ctx, group)
	if err != nil {
		return nil, err
	}
	ads = a

	if token != "" {
		if err := r.cache.Return(ctx, key, token, a, r.ttl); err != nil {
			if !errors.Is(err, redis.Nil) {
				logger.Warn("User.Redis", "Redis error while trying to store key", err, "key", key)
			}
		}
	}

	return ads, nil
}

func (r *RedisRepository) GetGroups(ctx context.Context) ([]string, error) {
	var groups []string
	var token string
	key := r.keys.List(groupsList)
	err := r.cache.Get(ctx, key, &groups)
	switch {
	case err == nil:
		return groups, nil
	case errors.Is(err, redis.Nil):
		token, err = r.cache.Lease(ctx, key)
		if err != nil {
			logger.Warn("User.Redis", "Redis error while trying to lease key", err, "key", key)
		}
	default:
		logger.Warn("User.Redis", "Redis error while trying to retrieve key", err, "key", key)
	}

	g, err := r.Repository.GetGroups(ctx)
	if err != nil {
		return nil, err
	}
	groups = g

	if token != "" {
		if err := r.cache.Return(ctx, key, token, g, r.ttl); err != nil {
			if !errors.Is(err, redis.Nil) {
				logger.Warn("User.Redis", "Redis error while trying to store key", err, "key", key)
			}
		}
	}

	return groups, nil
}

func (r *RedisRepository) Delete(ctx context.Context, id uint) error {
	if err := r.Repository.Delete(ctx, id); err != nil {
		return err
	}
	r.invalidate(ctx, r.keys.Key(strconv.FormatUint(uint64(id), 10)))
	r.invalidate(ctx, r.keys.List(groupsList), r.keys.KeyWithSuffix("*", "group"))
	return nil
}

func (r *RedisRepository) invalidate(ctx context.Context, keys ...string) {
	if err := r.cache.Invalidate(ctx, keys...); err != nil {
		logger.Warn("User.Redis", "Redis error while trying to delete keys", err, "keys", keys)
	}
}
