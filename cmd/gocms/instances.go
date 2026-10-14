package main

import (
	"context"
	"gocms/internal/cache"
	"gocms/internal/database"
	"gocms/internal/search"
	"gocms/internal/storage"
	"os"

	"github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func MySQLDatabase(app *Container) (*gorm.DB, error) {
	dbCfg := database.DBConfig{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Name:     os.Getenv("DB_NAME"),
	}

	DB, err := database.NewMySQLConnection(&dbCfg)
	if err != nil {
		return nil, err
	}

	return DB, nil
}

func RedisClient(ctx context.Context, app *Container) (*redis.Client, error) {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")

	rdb, err := cache.NewRedisConnection(host, port)
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
	host := os.Getenv("MEILISEARCH_HOST")
	port := os.Getenv("MEILISEARCH_PORT")
	key := os.Getenv("MEILISEARCH_API")

	client := search.NewMeilisearchConnection(host, port, key)

	return client, nil
}

func DefaultStorageHandler(cfg Config) *storage.Handler {
	path := os.Getenv("IMAGE_STORAGE_PATH")
	if cfg.Storage != "" {
		path = cfg.Storage
	} else if path == "" {
		path = "./storage"
	}
	service := storage.NewManager(path)
	return storage.NewHandler(service)
}
