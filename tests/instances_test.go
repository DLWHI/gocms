package tests

import (
	"context"
	"fmt"
	"gocms/internal/cache"
	"gocms/internal/database"
	"gocms/internal/search"
	"gocms/internal/storage"

	"github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func MySQLDatabase(app *Container) (*gorm.DB, error) {
	dbCfg := database.DBConfig{
		User:     "test",
		Password: "",
		Host:     "localhost",
		Port:     "3306",
		Name:     "test",
	}

	DB, err := database.NewMySQLConnection(&dbCfg, logger.Warn)
	if err != nil {
		return nil, err
	}

	return DB, nil
}

func RedisClient(ctx context.Context, app *Container) (*redis.Client, error) {
	rdb, err := cache.NewRedisConnection("localhost", "6379")
	if err != nil {
		return nil, err
	}

	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return rdb, nil
}

func MeilisearchClient(app *Container) (meilisearch.ServiceManager, error) {
	client := search.NewMeilisearchConnection("http://localhost", "7700", "")
	if client == nil {
		return nil, fmt.Errorf("Failed to connect to meilisearch")
	}
	return client, nil
}

func DefaultStorageHandler() *storage.Handler {
	service := storage.NewManager("./storage")
	return storage.NewHandler(service)
}
