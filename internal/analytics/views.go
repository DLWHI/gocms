package analytics

import (
	"context"
	"errors"
	"fmt"
	"gocms/internal/errs"
	"gocms/internal/logger"
	"gocms/internal/post"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type PostViewCounterService interface {
	Read(ctx context.Context, id uint) error
	Get(ctx context.Context, id uint) (uint, error)
}

type RedisCounter struct {
	service *post.Service
	rdb     *redis.Client
	restore *redis.Script
}

func NewRedisCounter(client *redis.Client, service *post.Service) *RedisCounter {
	return &RedisCounter{
		service: service,
		rdb:     client,
		restore: redis.NewScript(`
		local liveKey = KEYS[1]
		local stagingKey = KEYS[2]
		local amount = tonumber(ARGV[1])
		
		redis.call("INCRBY", liveKey, amount)
		redis.call("DEL", stagingKey)
		
		return 1
		`),
	}
}

func (c *RedisCounter) ParseStartArgs(params ...any) (time.Duration, error) {
	if len(params) != 1 {
		return 0, fmt.Errorf("%v: invalid argument count", errs.ErrInval)
	}

	interval, ok := params[0].(time.Duration)
	if !ok {
		return 0, fmt.Errorf("%v: invalid duration", errs.ErrInval)
	}
	return interval, nil
}

func (c *RedisCounter) Start(ctx context.Context, group *sync.WaitGroup, params ...any) {
	defer group.Done()

	interval, err := c.ParseStartArgs(params...)
	if err != nil {
		logger.Error("PostViewCount.Service", "Error parsing params", err)
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	recovery := time.NewTicker(5 * interval)
	defer recovery.Stop()
	logger.Info("PostViewCount.Service", "Started analytics background worker")

	for {
		select {
		case <-ctx.Done():
			logger.Info("PostViewCount.Service", "Analytics worker received shutdown signal")
			return
		case <-ticker.C:
			if err := c.Commit(ctx); err != nil {
				logger.Error("PostViewCount.Service", "Failed to flush pending post views", err)
			}
		case <-recovery.C:
			if err := c.RecoverOrphans(ctx); err != nil {
				logger.Error("PostViewCount.Service", "Failed to recover stale view records", err)
			}
		}
	}
}

func (c *RedisCounter) Stop(ctx context.Context) error {
	logger.Info("PostViewCount.Service", "Flushing pending data...")
	if err := c.Commit(ctx); err != nil {
		logger.Error("PostViewCount.Service", "Failed to flush pending post views during shutdown", err)
		return err
	}

	logger.Info("PostViewCount.Service", "Recovering stale data...")
	if err := c.RecoverOrphans(ctx); err != nil {
		logger.Error("PostViewCount.Service", "Failed to recover stale view records", err)
		return err
	}
	return nil
}

func (c *RedisCounter) Read(ctx context.Context, id uint) error {
	key := fmt.Sprintf("cms:analytics:views:%d", id)
	if _, err := c.rdb.Incr(ctx, key).Result(); err != nil {
		logger.Error("PostViewCounter.Service", "Failed to increment cached post views", err)
		return err
	}
	return nil
}

func (c *RedisCounter) Get(ctx context.Context, id uint) (uint, error) {
	views, err := c.service.GetViews(ctx, id)
	if err != nil {
		return 0, err
	}

	key := fmt.Sprintf("cms:analytics:views:%d", id)
	cached, err := c.rdb.Get(ctx, key).Uint64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return views, nil
		}
		logger.Error("PostViewCounter.Service", "Failed to read cached post views", err)
		return 0, err
	}
	return uint(cached) + views, nil
}

func (c *RedisCounter) Commit(ctx context.Context) error {
	pattern := "cms:analytics:views*"

	iter := c.rdb.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		if strings.HasSuffix(key, ":staging") {
			continue
		}

		staging := key + ":staging"
		var id uint
		n, err := fmt.Sscanf(key, "cms:analytics:views:%d", &id)
		if err != nil || n != 1 {
			logger.Error("PostViewCounter.Service", "Skipped an analytics cache entry with an invalid post identifier", fmt.Errorf("expected one post identifier in cache key"))
			continue
		}

		if err := c.rdb.RenameNX(ctx, key, staging).Err(); err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			logger.Error("PostViewCounter.Service", "Failed to stage cached post views for flushing", err)
			continue
		}

		value, err := c.rdb.Get(ctx, staging).Uint64()
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				logger.Error("PostViewCounter.Service", "Failed to read staged post views", err)
			}
			continue
		}
		if err := c.rdb.Del(ctx, staging).Err(); err != nil {
			logger.Error("PostViewCounter.Service", "Failed to remove staged post views after reading", err)
		}

		if err := c.service.IncrementViews(ctx, id, uint(value)); err != nil {
			logger.Error("PostViewCounter.Service", "Failed to persist pending post views; restoring them to cache", err)
			if _, err := c.rdb.IncrBy(ctx, key, int64(value)).Result(); err != nil {
				return fmt.Errorf("failed to restore pending post views to cache: %w", err)
			}
			continue
		}
	}

	if err := iter.Err(); err != nil {
		return err
	}
	return nil
}

func (c *RedisCounter) RecoverOrphans(ctx context.Context) error {
	pattern := "cms:analytics:views:*:staging"
	var recovered int
	var restored int

	iter := c.rdb.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		staging := iter.Val()

		key := strings.TrimSuffix(staging, ":staging")

		value, err := c.rdb.Get(ctx, staging).Uint64()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			logger.Error("PostViewCounter.Service", "Failed to read staged post views during recovery", err)
			continue
		}
		recovered++
		if err := c.restore.Run(ctx, c.rdb, []string{key, staging}, value).Err(); err != nil {
			logger.Error("PostViewCounter.Service", "Failed to restore staged post views during recovery", err)
			continue
		}
		restored++
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if recovered > 0 || restored > 0 {
		logger.Info(
			"PostViewCounter.Service",
			"Completed analytics cache recovery",
			"stagedEntries", recovered,
			"restoredEntries", restored,
		)
	}

	return nil
}
