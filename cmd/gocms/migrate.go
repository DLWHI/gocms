package main

import (
	"context"
	"gocms/internal/database"
	"gocms/internal/logger"
	"gocms/internal/search"
)

func Migrate(ctx context.Context, app *Container) error {
	if err := database.SoftMigrate(app.DB); err != nil {
		logger.Error("Wipe", "Failed to migrate DB", err)
		return err
	}
	engine := app.Meilisearch
	if err := search.ConfigureMeilisearchIndex(ctx, engine, app.Posts.MeilisearchEngine); err != nil {
		logger.Error("Wipe", "Failed to reconfigure post search index", err)
		return err
	}
	if err := search.ConfigureMeilisearchIndex(ctx, engine, app.Users.MeilisearchEngine); err != nil {
		logger.Error("Wipe", "Failed to reconfigure user search index", err)
		return err
	}
	if err := search.ConfigureMeilisearchIndex(ctx, engine, app.Categories.MeilisearchEngine); err != nil {
		logger.Error("Wipe", "Failed to reconfigure category search index", err)
		return err
	}
	if err := search.ConfigureMeilisearchIndex(ctx, engine, app.Editors.MeilisearchEngine); err != nil {
		logger.Error("Wipe", "Failed to reconfigure editor search index", err)
		return err
	}
	if err := search.ConfigureMeilisearchIndex(ctx, engine, app.Ads.MeilisearchEngine); err != nil {
		logger.Error("Wipe", "Failed to reconfigure ads search index", err)
		return err
	}
	logger.Info("Migrate", "Migration successful")
	return nil
}

func Wipe(ctx context.Context, app *Container) error {
	if err := database.HardMigrate(app.DB); err != nil {
		logger.Error("Wipe", "Failed to reset DB", err)
		return err
	}

	engine := app.Meilisearch

	if err := search.WipeMeilisearchIndex(ctx, engine, app.Posts.MeilisearchEngine); err != nil {
		logger.Error("Wipe", "Failed to wipe post search index", err)
		return err
	} else {
		if err := search.CreateMeilisearchIndex(ctx, engine, app.Posts.MeilisearchEngine); err != nil {
			logger.Error("Wipe", "Failed to create post search index", err)
			return err
		}
	}

	if err := search.WipeMeilisearchIndex(ctx, engine, app.Users.MeilisearchEngine); err != nil {
		logger.Error("Wipe", "Failed to wipe user search index", err)
		return err
	} else {
		if err := search.CreateMeilisearchIndex(ctx, engine, app.Users.MeilisearchEngine); err != nil {
			logger.Error("Wipe", "Failed to create user search index", err)
			return err
		}
	}

	if err := search.WipeMeilisearchIndex(ctx, engine, app.Categories.MeilisearchEngine); err != nil {
		logger.Error("Wipe", "Failed to wipe category search index", err)
		return err
	} else {
		if err := search.CreateMeilisearchIndex(ctx, engine, app.Categories.MeilisearchEngine); err != nil {
			logger.Error("Wipe", "Failed to create category search index", err)
			return err
		}
	}

	if err := search.WipeMeilisearchIndex(ctx, engine, app.Editors.MeilisearchEngine); err != nil {
		logger.Error("Wipe", "Failed to wipe editor search index", err)
		return err
	} else {
		if err := search.CreateMeilisearchIndex(ctx, engine, app.Editors.MeilisearchEngine); err != nil {
			logger.Error("Wipe", "Failed to create editor search index", err)
			return err
		}
	}

	if err := search.WipeMeilisearchIndex(ctx, engine, app.Ads.MeilisearchEngine); err != nil {
		logger.Error("Wipe", "Failed to wipe ads search index", err)
		return err
	} else {
		if err := search.CreateMeilisearchIndex(ctx, engine, app.Ads.MeilisearchEngine); err != nil {
			logger.Error("Wipe", "Failed to create ads search index", err)
			return err
		}
	}

	logger.Info("Wipe", "Finished wiping and re-creating DB and search index schema")
	return nil
}
