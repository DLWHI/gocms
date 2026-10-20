package search

import (
	"context"
	"fmt"
	"gocms/internal/logger"
	"time"

	"github.com/meilisearch/meilisearch-go"
)

type MeilisearchIndex interface {
	Config() meilisearch.IndexConfig
	Settings() meilisearch.Settings
}

func NewMeilisearchConnection(host string, port string, apiKey string) meilisearch.ServiceManager {
	client := meilisearch.New(fmt.Sprintf("%s:%s", host, port), meilisearch.WithAPIKey(apiKey))

	_, err := client.Health()
	if err != nil {
		logger.Error("Search.Meilisearch", "Failed to connect to the search service", err)
		return nil
	}
	_, err = client.Version()
	if err != nil {
		logger.Error("Search.Meilisearch", "Failed to verify the search service version", err)
		return nil
	}
	return client
}

func CreateMeilisearchIndex(ctx context.Context, se meilisearch.ServiceManager, idx MeilisearchIndex) error {
	config := idx.Config()
	task, err := se.CreateIndexWithContext(ctx, &config)
	if err != nil {
		return err
	}
	logger.Info("Search.Meilisearch", "Search index creation enqueued")
	if err := WaitForMeilisearchTask(ctx, se, task); err != nil {
		return err
	}

	err = ConfigureMeilisearchIndex(ctx, se, idx)
	if err != nil {
		return err
	}
	return nil
}

func WipeMeilisearchIndex(ctx context.Context, se meilisearch.ServiceManager, idx MeilisearchIndex) error {
	config := idx.Config()
	task, err := se.DeleteIndexWithContext(ctx, config.Uid)
	if err != nil {
		return err
	}
	logger.Info("Search.Meilisearch", "Search index deletion enqueued")
	if err := WaitForMeilisearchTask(ctx, se, task); err != nil {
		return err
	}
	return nil
}

func ConfigureMeilisearchIndex(ctx context.Context, se meilisearch.ServiceManager, idx MeilisearchIndex) error {
	config := idx.Config()
	settings := idx.Settings()
	task, err := se.Index(config.Uid).UpdateSettingsWithContext(ctx, &settings)
	if err != nil {
		return err
	}
	logger.Info("Search.Meilisearch", "Search index settings update enqueued")
	if err := WaitForMeilisearchTask(ctx, se, task); err != nil {
		return err
	}
	return nil
}

func WaitForMeilisearchTask(ctx context.Context, se meilisearch.ServiceManager, info *meilisearch.TaskInfo) error {
	task, err := se.TaskManager().WaitForTaskWithContext(ctx, info.TaskUID, 100*time.Millisecond)
	if err != nil {
		return err
	}
	if task.Status == meilisearch.TaskStatusFailed && task.Error.Code != "index_not_found" {
		return fmt.Errorf("Error indexing: %s - %s", task.Error.Code, task.Error.Message)
	}
	return nil
}
