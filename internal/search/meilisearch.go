package search

import (
	"context"
	"fmt"
	"gocms/internal/logger"
	"time"

	"github.com/meilisearch/meilisearch-go"
)

type MeilisearchClient struct {
	Client   meilisearch.ServiceManager
	Config   meilisearch.IndexConfig
	Settings meilisearch.Settings
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

func (d *MeilisearchClient) CreateIndex(ctx context.Context) error {
	config := d.Config
	task, err := d.Client.CreateIndexWithContext(ctx, &config)
	if err != nil {
		return err
	}
	logger.Info("Search.Meilisearch", "Search index creation enqueued")
	if err := d.WaitForTask(ctx, task); err != nil {
		return err
	}

	err = d.ConfigureIndex(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (d *MeilisearchClient) WipeIndex(ctx context.Context) error {
	config := d.Config
	task, err := d.Client.DeleteIndexWithContext(ctx, config.Uid)
	if err != nil {
		return err
	}
	logger.Info("Search.Meilisearch", "Search index deletion enqueued")
	if err := d.WaitForTask(ctx, task); err != nil {
		return err
	}
	return nil
}

func (d *MeilisearchClient) ConfigureIndex(ctx context.Context) error {
	config := d.Config
	settings := d.Settings
	task, err := d.Client.Index(config.Uid).UpdateSettingsWithContext(ctx, &settings)
	if err != nil {
		return err
	}
	logger.Info("Search.Meilisearch", "Search index settings update enqueued")
	if err := d.WaitForTask(ctx, task); err != nil {
		return err
	}
	return nil
}

func (d *MeilisearchClient) WaitForTask(ctx context.Context, info *meilisearch.TaskInfo) error {
	task, err := d.Client.TaskManager().WaitForTaskWithContext(ctx, info.TaskUID, 100*time.Millisecond)
	if err != nil {
		return err
	}
	if task.Status == meilisearch.TaskStatusFailed && task.Error.Code != "index_not_found" {
		return fmt.Errorf("Error indexing: %s - %s", task.Error.Code, task.Error.Message)
	}
	return nil
}
