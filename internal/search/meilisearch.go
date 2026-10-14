package search

import (
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

func CreateMeilisearchIndex(se meilisearch.ServiceManager, idx MeilisearchIndex) error {
	config := idx.Config()
	taskInfo, err := se.CreateIndex(&config)
	if err != nil {
		return err
	}
	logger.Info("Search.Meilisearch", "Search index creation enqueued")
	task, err := se.TaskManager().WaitForTask(taskInfo.TaskUID, 100*time.Millisecond)
	if err != nil {
		return err
	}
	if task.Status == meilisearch.TaskStatusFailed {
		return fmt.Errorf("Error creating index %s: %s - %s", config.Uid, task.Error.Code, task.Error.Message)
	}

	err = ConfigureMeilisearchIndex(se, idx)
	if err != nil {
		return err
	}
	return nil
}

func WipeMeilisearchIndex(se meilisearch.ServiceManager, idx MeilisearchIndex) error {
	config := idx.Config()
	taskInfo, err := se.DeleteIndex(config.Uid)
	if err != nil {
		return err
	}
	logger.Info("Search.Meilisearch", "Search index deletion enqueued")
	task, err := se.TaskManager().WaitForTask(taskInfo.TaskUID, 100*time.Millisecond)
	if err != nil {
		return err
	}
	if task.Status == meilisearch.TaskStatusFailed && task.Error.Code != "index_not_found" {
		return fmt.Errorf("Error wiping index %s: %s - %s", config.Uid, task.Error.Code, task.Error.Message)
	}
	return nil
}

func ConfigureMeilisearchIndex(se meilisearch.ServiceManager, idx MeilisearchIndex) error {
	config := idx.Config()
	settings := idx.Settings()
	taskInfo, err := se.Index(config.Uid).UpdateSettings(&settings)
	if err != nil {
		return err
	}
	logger.Info("Search.Meilisearch", "Search index settings update enqueued")
	task, err := se.TaskManager().WaitForTask(taskInfo.TaskUID, 100*time.Millisecond)
	if err != nil {
		return err
	}
	if task.Status == meilisearch.TaskStatusFailed {
		return fmt.Errorf("Error configuring index: %s - %s", task.Error.Code, task.Error.Message)
	}
	return nil
}
