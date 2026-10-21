package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gocms/internal/cache"
	"gocms/internal/logger"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	cache.RedisDecorator
	origin    Repository
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
		origin:         db,
		ttl:            config.Ttl,
		keyFormat:      config.KeyFormat,
		allKey:         fmt.Sprintf(config.KeyFormat, ":all"),
		RedisDecorator: cache.RedisDecorator{Client: config.Client},
	}
}

func (r *RedisRepository) Create(ctx context.Context, user *Record) error {
	if err := r.origin.Create(ctx, user); err != nil {
		return err
	}
	ictx := context.WithoutCancel(ctx)
	if _, err := r.Client.Unlink(ictx, r.allKey).Result(); err != nil {
		r.logError("Failed to delete key %s from cache.", r.allKey, err)
	}
	return nil
}

func (r *RedisRepository) UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*Record, error) {
	user, err := r.origin.UpdateLogin(ctx, username, timestamp)
	if err != nil {
		return nil, err
	}
	ictx := context.WithoutCancel(ctx)
	key := fmt.Sprintf(r.keyFormat, username)
	if _, err := r.Client.Unlink(ictx, key, key, r.allKey).Result(); err != nil {
		r.logError("Failed to delete key %s from cache.", key, err)
	}
	return user, nil
}

func (r *RedisRepository) UpdatePasswd(ctx context.Context, username string, pwdhash Password) error {
	return r.origin.UpdatePasswd(ctx, username, pwdhash)
}

func (r *RedisRepository) Update(ctx context.Context, username string, diff map[string]interface{}) error {
	if err := r.origin.Update(ctx, username, diff); err != nil {
		return err
	}
	ictx := context.WithoutCancel(ctx)
	key := fmt.Sprintf(r.keyFormat, username)
	if _, err := r.Client.Unlink(ictx, key, key, r.allKey).Result(); err != nil {
		r.logError("Failed to delete key %s from cache.", key, err)
	}
	return nil
}

func (r *RedisRepository) Get(ctx context.Context, username string) (*Record, error) {
	var user Record
	key := fmt.Sprintf(r.keyFormat, username)

	if err := r.RedisDecorator.Get(ctx, key, &user); err == nil {
		return &user, nil
	} else {
		if !errors.Is(err, redis.Nil) {
			r.logError("Failed to retreive key %s from cache.", key, err)
		}
	}

	token, err := r.RedisDecorator.Lease(ctx, key, r.ttl)
	if err != nil {
		r.logError("Failed to lease key %s from cache.", key, err)
	}

	u, err := r.origin.Get(ctx, username)
	if err != nil {
		return nil, err
	}
	user = *u

	if token != "" {
		if err := r.RedisDecorator.Return(ctx, key, token, u, r.ttl); err != nil {
			r.logError("Failed to store key %s to cache.", key, err)
		}
	}

	return &user, nil
}

func (r *RedisRepository) GetAll(ctx context.Context) ([]Record, error) {
	var users []Record
	if err := r.RedisDecorator.Get(ctx, r.allKey, &users); err == nil {
		return users, nil
	} else {
		if !errors.Is(err, redis.Nil) {
			r.logError("Failed to retreive key %s from cache.", r.allKey, err)
		}
	}

	token, err := r.RedisDecorator.Lease(ctx, r.allKey, r.ttl)
	if err != nil {
		r.logError("Failed to lease key %s from cache.", r.allKey, err)
	}

	users, err = r.origin.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	if token != "" {
		if err := r.RedisDecorator.Return(ctx, r.allKey, token, users, r.ttl); err != nil {
			if !errors.Is(err, redis.Nil) {
				r.logError("Failed to store key %s to cache.", r.allKey, err)
			}
		}
	}

	return users, nil
}

func (r *RedisRepository) GetPasswdHash(ctx context.Context, username string) (Password, error) {
	return r.origin.GetPasswdHash(ctx, username)
}

func (r *RedisRepository) Delete(ctx context.Context, username string) error {
	if err := r.origin.Delete(ctx, username); err != nil {
		return err
	}
	ictx := context.WithoutCancel(ctx)
	key := fmt.Sprintf(r.keyFormat, username)
	if _, err := r.Client.Unlink(ictx, key, key, r.allKey).Result(); err != nil {
		r.logError("Failed to delete key %s from cache.", key, err)
	}
	return nil
}

func (r *RedisRepository) logError(format string, key string, err error) {
	logger.Warn("User.Redis", fmt.Sprintf(format, key), err)
}
