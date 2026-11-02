package search

import (
	"context"
	"fmt"
	"gocms/internal/logger"
	"time"

	"github.com/meilisearch/meilisearch-go"
)

type ConfigurableIndex interface {
	Create(ctx context.Context) error
	Configure(ctx context.Context) error
	Wipe(ctx context.Context) error
	Destroy(ctx context.Context) error
}

type MeilisearchIndex struct {
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

func (d *MeilisearchIndex) Create(ctx context.Context) error {
	task, err := d.Client.CreateIndexWithContext(ctx, &d.Config)
	if err != nil {
		return err
	}
	if err := d.WaitForTask(ctx, task); err != nil {
		return err
	}

	err = d.Configure(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (d *MeilisearchIndex) Configure(ctx context.Context) error {
	task, err := d.Client.Index(d.Config.Uid).UpdateSettingsWithContext(ctx, &d.Settings)
	if err != nil {
		return err
	}
	if err := d.WaitForTask(ctx, task); err != nil {
		return err
	}
	return nil
}

func (d *MeilisearchIndex) Wipe(ctx context.Context) error {
	task, err := d.Client.Index(d.Config.Uid).DeleteAllDocumentsWithContext(ctx, nil)
	if err != nil {
		return err
	}
	if err := d.WaitForTask(ctx, task); err != nil {
		return err
	}
	return nil
}

func (d *MeilisearchIndex) Destroy(ctx context.Context) error {
	task, err := d.Client.DeleteIndexWithContext(ctx, d.Config.Uid)
	if err != nil {
		return err
	}
	if err := d.WaitForTask(ctx, task); err != nil {
		return err
	}
	return nil
}

func (d *MeilisearchIndex) WaitForTask(ctx context.Context, info *meilisearch.TaskInfo) error {
	task, err := d.Client.TaskManager().WaitForTaskWithContext(ctx, info.TaskUID, 100*time.Millisecond)
	if err != nil {
		return err
	}
	if task.Status == meilisearch.TaskStatusFailed {
		return fmt.Errorf("Error performing %s task: %s - %s", task.Type, task.Error.Code, task.Error.Message)
	}
	return nil
}
