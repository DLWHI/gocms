package tests

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
		User:     "test",
		Password: "",
		Host:     "localhost",
		Port:     "3306",
		Name:     "",
	}

	DB, err := database.NewMySQLConnection(&dbCfg)
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
	host := os.Getenv("MEILISEARCH_HOST")
	port := os.Getenv("MEILISEARCH_PORT")
	key := os.Getenv("MEILISEARCH_API")

	client := search.NewMeilisearchConnection(host, port, key)

	return client, nil
}

func DefaultStorageHandler() *storage.Handler {
	service := storage.NewManager("./test_storage")
	return storage.NewHandler(service)
}
