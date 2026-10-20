package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gocms/internal/logger"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	origin    Repository
	rdb       *redis.Client
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
		origin:    db,
		rdb:       config.Client,
		ttl:       config.Ttl,
		keyFormat: config.KeyFormat,
		allKey:    fmt.Sprintf(config.KeyFormat, ":all"),
	}
}

func (r *RedisRepository) Create(ctx context.Context, user *User) error {
	if err := r.origin.Create(ctx, user); err != nil {
		return err
	}
	ictx := context.WithoutCancel(ctx)
	if _, err := r.rdb.Del(ictx, r.allKey).Result(); err != nil {
		r.logError("Failed to delete key %s from cache.", r.allKey, err)
	}
	return nil
}

func (r *RedisRepository) UpdateLogin(ctx context.Context, username string, timestamp time.Time) (*User, error) {
	user, err := r.origin.UpdateLogin(ctx, username, timestamp)
	if err != nil {
		return nil, err
	}
	ictx := context.WithoutCancel(ctx)
	key := fmt.Sprintf(r.keyFormat, username)
	if _, err := r.rdb.Del(ictx, key).Result(); err != nil {
		r.logError("Failed to delete key %s from cache.", key, err)
	}
	if _, err := r.rdb.Del(ictx, r.allKey).Result(); err != nil {
		r.logError("Failed to delete key %s from cache.", r.allKey, err)
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
	if _, err := r.rdb.Del(ictx, key).Result(); err != nil {
		r.logError("Failed to delete key %s from cache.", key, err)
	}
	if _, err := r.rdb.Del(ictx, r.allKey).Result(); err != nil {
		r.logError("Failed to delete key %s from cache.", r.allKey, err)
	}
	return nil
}

func (r *RedisRepository) Get(ctx context.Context, username string) (*User, error) {
	key := fmt.Sprintf(r.keyFormat, username)

	cached, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		var user User
		if err := json.Unmarshal([]byte(cached), &user); err != nil {
			r.logError("Invalid data on key %s, key will be deleted", key, err)
			ictx := context.WithoutCancel(ctx)
			if _, err := r.rdb.Del(ictx, key).Result(); err != nil {
				r.logError("Failed to delete key %s from cache.", key, err)
			}
		} else {
			return &user, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		r.logError("Failed to read key %s from cache.", key, err)
	}

	if _, err := r.rdb.Set(ctx, key, "", r.ttl).Result(); err != nil {
		r.logError("Failed to write key %s to cache.", key, err)
	}

	user, err := r.origin.Get(ctx, username)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(user)
	if err == nil {
		if _, err := r.rdb.SetXX(ctx, key, data, r.ttl).Result(); err != nil {
			r.logError("Failed to write key %s to cache.", key, err)
		}
	} else {
		r.logError("Failed to encode key %s into json.", key, err)
	}

	return user, nil
}

func (r *RedisRepository) GetAll(ctx context.Context) ([]User, error) {
	cached, err := r.rdb.Get(ctx, r.allKey).Result()
	if err == nil {
		var users []User
		if err := json.Unmarshal([]byte(cached), &users); err != nil {
			r.logError("Invalid data on key %s, key will be deleted", r.allKey, err)
			ictx := context.WithoutCancel(ctx)
			if _, err := r.rdb.Del(ictx, r.allKey).Result(); err != nil {
				r.logError("Failed to delete key %s from cache.", r.allKey, err)
			}
		} else {
			return users, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		r.logError("Failed to read key %s from cache.", r.allKey, err)
	}

	users, err := r.origin.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(users)
	if err == nil {
		if _, err := r.rdb.Set(ctx, r.allKey, data, r.ttl).Result(); err != nil {
			r.logError("Failed to write key %s to cache.", r.allKey, err)
		}
	} else {
		r.logError("Failed to encode key %s into json.", r.allKey, err)
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
	if _, err := r.rdb.Del(ictx, key).Result(); err != nil {
		r.logError("Failed to delete key %s from cache.", key, err)
	}
	if _, err := r.rdb.Del(ictx, r.allKey).Result(); err != nil {
		r.logError("Failed to delete key %s from cache.", r.allKey, err)
	}
	return nil
}

func (r *RedisRepository) logError(format string, key string, err error) {
	logger.Warn("User.Redis", fmt.Sprintf(format, key), err)
}
