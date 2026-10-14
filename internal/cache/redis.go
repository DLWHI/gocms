package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

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
