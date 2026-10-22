package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"gocms/internal/domain"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	leasePrefix = "lease:"
	fillScript  = redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("SET", KEYS[1], ARGV[2], "PX", ARGV[3])
		end
		return nil
	`)
)

type RedisDecorator struct {
	Client *redis.Client
}

func NewRedisConnection(host string, port string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port), // Redis host and port
		Password: "",                               // No password set by default
		DB:       0,                                // Use default DB index 0

		// Connection Pool Settings
		PoolSize:       50,               // Max open connections
		MinIdleConns:   10,               // Warm idle connections ready to use
		MaxActiveConns: 100,              // Hard cap under extreme bursts
		PoolTimeout:    10 * time.Second, // Max time a goroutine waits for an available conn

		// Timeout Configurations (Prevent goroutine leaks)
		DialTimeout:  20 * time.Second, // Timeout for establishing new connections
		ReadTimeout:  20 * time.Second, // Timeout for socket reads
		WriteTimeout: 20 * time.Second, // Timeout for socket writes

		// Connection Lifecycle & Health Checks
		ConnMaxIdleTime: 5 * time.Minute,  // Close idle connections after 5m
		ConnMaxLifetime: 30 * time.Minute, // Recycle connections periodically
	})

	return rdb, nil
}

func (r *RedisDecorator) Get(ctx context.Context, key string, ptr any) error {
	cached, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	if strings.HasPrefix(cached, leasePrefix) {
		return redis.Nil
	}
	if err := json.Unmarshal([]byte(cached), ptr); err != nil {
		return err
	}
	return nil
}

func (r *RedisDecorator) Lease(ctx context.Context, key string, ttl time.Duration) (string, error) {
	token := strconv.FormatUint(rand.Uint64(), 36)
	if _, err := r.Client.Set(ctx, key, token, ttl).Result(); err != nil {
		return "", err
	}
	return token, nil
}

func (r *RedisDecorator) Return(ctx context.Context, key string, token string, ptrval any, ttl time.Duration) error {
	data, err := json.Marshal(ptrval)
	if err != nil {
		return domain.ErrInval
	}

	err = fillScript.Run(ctx, r.Client, []string{key}, token, data, ttl.Milliseconds()).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisDecorator) Delete(ctx context.Context, key string) error {
	if _, err := r.Client.Del(ctx, key).Result(); err != nil {
		return err
	}
	return nil
}

func UnmarshalRedisKey(rdb *redis.Client, ctx context.Context, key string, ptr any) error {
	cached, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(cached), ptr); err != nil {
		return err
	}
	return nil
}

func MarshalRedisKey(rdb *redis.Client, ctx context.Context, key string, ttl time.Duration, ptr any) error {
	data, err := json.Marshal(ptr)
	if err != nil {
		return err
	}
	if _, err := rdb.Set(ctx, key, data, ttl).Result(); err != nil {
		return err
	}
	return nil
}

func EraseRedisKeyPattern(rdb *redis.Client, ctx context.Context, pattern string) error {
	keys := make([]string, 0, 10)
	iter := rdb.Scan(ctx, 0, pattern, 10).Iterator()

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())

		if len(keys) > 9 {
			if _, err := rdb.Unlink(ctx, keys...).Result(); err != nil {
				return err
			}
			keys = keys[:0]
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		if _, err := rdb.Unlink(ctx, keys...).Result(); err != nil {
			return err
		}
	}
	return nil
}
