package main

import (
	"context"
	"gocms/internal/ads"
	"gocms/internal/category"
	"gocms/internal/editor"
	"gocms/internal/lang"
	"gocms/internal/logger"
	"gocms/internal/outbox"
	"gocms/internal/post"
	"gocms/internal/search"
	"gocms/internal/user"

	"github.com/meilisearch/meilisearch-go"
	"gorm.io/gorm"
)

func Migrate(ctx context.Context, app *Container) error {
	if err := SoftMigrate(app.DB); err != nil {
		logger.Error("Wipe", "Failed to migrate DB", err)
		return err
	}

	if err := ConfigureIndexes(ctx, app.Meilisearch); err != nil {
		logger.Error("Wipe", "Failed to configure search indexes", err)
		return err
	}

	logger.Info("Migrate", "Migration successful")
	return nil
}

func HardReset(ctx context.Context, app *Container) error {
	if err := HardMigrate(app.DB); err != nil {
		logger.Error("Wipe", "Failed to reset DB", err)
		return err
	}

	if err := ReCreateIndexes(ctx, app.Meilisearch); err != nil {
		logger.Error("Wipe", "Failed to reset search indexes", err)
		return err
	}

	logger.Info("Wipe", "Finished wiping and re-creating DB and search index schema")
	return nil
}

func SoftMigrate(db *gorm.DB) error {

	err := db.AutoMigrate(
		&ads.Record{},
		&user.Record{},
		&editor.Record{},
		&editor.RecordTranslation{},
		&post.Record{},
		&post.Translation{},
		&category.Record{},
		&category.RecordTranslation{},
		&lang.Language{},
		&outbox.Message{},
	)

	if err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_post_categories_category_slug ON post_categories (category_slug, post_id);").Error; err != nil {
		return err
	}

	return nil
}

func HardMigrate(db *gorm.DB) error {
	if err := db.Migrator().DropTable("post_categories"); err != nil {
		return err
	}

	if err := DropAll(db); err != nil {
		return err
	}

	if err := SoftMigrate(db); err != nil {
		return err
	}

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_post_categories_category_slug ON post_categories (category_slug, post_id);").Error; err != nil {
		return err
	}

	return nil
}

func DropAll(db *gorm.DB) error {
	err := db.Migrator().DropTable("post_categories")
	if err != nil {
		return err
	}
	if err := db.Migrator().DropTable(
		&ads.Record{},
		&user.Record{},
		&editor.Record{},
		&editor.RecordTranslation{},
		&post.Record{},
		&post.Translation{},
		&category.Record{},
		&category.RecordTranslation{},
		&lang.Language{},
		&outbox.Message{},
	); err != nil {
		return err
	}
	return nil
}

func ReCreateIndexes(ctx context.Context, client meilisearch.ServiceManager) error {
	indexes := []search.ConfigurableIndex{
		user.NewMeilisearchEngine(client, UsersSearchIndexName),
		category.NewMeilisearchEngine(client, CategoriesSearchIndexName),
		editor.NewMeilisearchEngine(client, EditorsSearchIndexName),
		post.NewMeilisearchEngine(client, PostsSearchIndexName),
		ads.NewMeilisearchEngine(client, AdsSearchIndexName),
	}

	for _, index := range indexes {
		if err := index.Destroy(ctx); err != nil {
			return err
		}
	}

	for _, index := range indexes {
		if err := index.Create(ctx); err != nil {
			return err
		}
	}

	return nil
}

func ConfigureIndexes(ctx context.Context, client meilisearch.ServiceManager) error {
	indexes := []search.ConfigurableIndex{
		user.NewMeilisearchEngine(client, UsersSearchIndexName),
		category.NewMeilisearchEngine(client, CategoriesSearchIndexName),
		editor.NewMeilisearchEngine(client, EditorsSearchIndexName),
		post.NewMeilisearchEngine(client, PostsSearchIndexName),
		ads.NewMeilisearchEngine(client, AdsSearchIndexName),
	}

	for _, index := range indexes {
		if err := index.Configure(ctx); err != nil {
			return err
		}
	}

	return nil
}
