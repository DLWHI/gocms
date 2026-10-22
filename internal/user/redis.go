package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gocms/internal/cache"
	"gocms/internal/domain"
	"gocms/internal/logger"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	Repository
	cache.RedisDecorator
	ttl       time.Duration
	keyFormat string
	allKey    string
}

type RedisConfig struct {
	Client    *redis.Client
	Ttl       time.Duration
	KeyFormat string
}

func NewRedisRepository(db Repository, config *RedisConfig) *RedisRepository {
	return &RedisRepository{
		Repository: db,
		ttl:        config.Ttl,
		keyFormat:  config.KeyFormat,
		allKey:     fmt.Sprintf(config.KeyFormat, ":all"),
		RedisDecorator: cache.RedisDecorator{
			Client:   config.Client,
			LeaseTTL: 5 * time.Second,
		},
	}
}

func (r *RedisRepository) Create(ctx context.Context, user *domain.User, pwdhash domain.Password) error {
	if err := r.Repository.Create(ctx, user, pwdhash); err != nil {
		return err
	}
	r.invalidate(ctx, r.allKey)
	return nil
}

func (r *RedisRepository) UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*domain.User, error) {
	user, err := r.Repository.UpdateLogin(ctx, username, timestamp)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf(r.keyFormat, username)
	r.invalidate(ctx, r.allKey, key)
	return user, nil
}

func (r *RedisRepository) Update(ctx context.Context, username string, user *domain.User) error {
	if err := r.Repository.Update(ctx, username, user); err != nil {
		return err
	}
	key := fmt.Sprintf(r.keyFormat, username)
	r.invalidate(ctx, r.allKey, key)
	return nil
}

func (r *RedisRepository) Get(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	key := fmt.Sprintf(r.keyFormat, username)

	if err := r.RedisDecorator.Get(ctx, key, &user); err == nil {
		return &user, nil
	} else {
		if !errors.Is(err, redis.Nil) {
			logger.Warn("User.Redis", "Redis error while trying to retrieve key", err, "key", key)
		}
	}

	token, err := r.RedisDecorator.Lease(ctx, key)
	if err != nil {
		logger.Warn("User.Redis", "Redis error while trying to lease key", err, "key", key)
	}

	u, err := r.Repository.Get(ctx, username)
	if err != nil {
		return nil, err
	}
	user = *u

	if token != "" {
		if err := r.RedisDecorator.Return(ctx, key, token, u, r.ttl); err != nil {
			logger.Warn("User.Redis", "Redis error while trying to store key", err, "key", key)
		}
	}

	return &user, nil
}

func (r *RedisRepository) GetActive(ctx context.Context) ([]domain.User, error) {
	var users []domain.User
	if err := r.RedisDecorator.Get(ctx, r.allKey, &users); err == nil {
		return users, nil
	} else {
		if !errors.Is(err, redis.Nil) {
			logger.Warn("User.Redis", "Redis error while trying to retrieve key", err, "key", r.allKey)
		}
	}

	token, err := r.RedisDecorator.Lease(ctx, r.allKey)
	if err != nil {
		logger.Warn("User.Redis", "Redis error while trying to lease key", err, "key", r.allKey)
	}

	users, err = r.Repository.GetActive(ctx)
	if err != nil {
		return nil, err
	}

	if token != "" {
		if err := r.RedisDecorator.Return(ctx, r.allKey, token, users, r.ttl); err != nil {
			if !errors.Is(err, redis.Nil) {
				logger.Warn("User.Redis", "Redis error while trying to store key", err, "key", r.allKey)
			}
		}
	}

	return users, nil
}

func (r *RedisRepository) Suspend(ctx context.Context, username string) error {
	if err := r.Repository.Suspend(ctx, username); err != nil {
		return err
	}
	key := fmt.Sprintf(r.keyFormat, username)
	r.invalidate(ctx, r.allKey, key)
	return nil
}

func (r *RedisRepository) invalidate(ctx context.Context, keys ...string) {
	ictx := context.WithoutCancel(ctx)
	if _, err := r.Client.Unlink(ictx, keys...).Result(); err != nil {
		logger.Warn("User.Redis", "Redis error while trying to delete keys", err, "keys", keys)
	}
}
