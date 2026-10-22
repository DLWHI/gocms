package main

import (
	"context"
	"gocms/internal/database"
	"gocms/internal/logger"
)

func Migrate(ctx context.Context, app *Container) error {
	if err := database.SoftMigrate(app.DB); err != nil {
		logger.Error("Wipe", "Failed to migrate DB", err)
		return err
	}
	if err := app.Posts.MeilisearchEngine.ConfigureMeilisearchIndex(ctx); err != nil {
		logger.Error("Wipe", "Failed to reconfigure post search index", err)
		return err
	}
	if err := app.Users.MeilisearchEngine.ConfigureMeilisearchIndex(ctx); err != nil {
		logger.Error("Wipe", "Failed to reconfigure user search index", err)
		return err
	}
	if err := app.Categories.MeilisearchEngine.ConfigureMeilisearchIndex(ctx); err != nil {
		logger.Error("Wipe", "Failed to reconfigure category search index", err)
		return err
	}
	if err := app.Editors.MeilisearchEngine.ConfigureMeilisearchIndex(ctx); err != nil {
		logger.Error("Wipe", "Failed to reconfigure editor search index", err)
		return err
	}
	if err := app.Ads.MeilisearchEngine.ConfigureMeilisearchIndex(ctx); err != nil {
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

	if err := app.Posts.MeilisearchEngine.WipeMeilisearchIndex(ctx); err != nil {
		logger.Error("Wipe", "Failed to wipe post search index", err)
		return err
	} else {
		if err := app.Posts.MeilisearchEngine.CreateMeilisearchIndex(ctx); err != nil {
			logger.Error("Wipe", "Failed to create post search index", err)
			return err
		}
	}

	if err := app.Users.MeilisearchEngine.WipeMeilisearchIndex(ctx); err != nil {
		logger.Error("Wipe", "Failed to wipe user search index", err)
		return err
	} else {
		if err := app.Users.MeilisearchEngine.CreateMeilisearchIndex(ctx); err != nil {
			logger.Error("Wipe", "Failed to create user search index", err)
			return err
		}
	}

	if err := app.Categories.MeilisearchEngine.WipeMeilisearchIndex(ctx); err != nil {
		logger.Error("Wipe", "Failed to wipe category search index", err)
		return err
	} else {
		if err := app.Categories.MeilisearchEngine.CreateMeilisearchIndex(ctx); err != nil {
			logger.Error("Wipe", "Failed to create category search index", err)
			return err
		}
	}

	if err := app.Editors.MeilisearchEngine.WipeMeilisearchIndex(ctx); err != nil {
		logger.Error("Wipe", "Failed to wipe editor search index", err)
		return err
	} else {
		if err := app.Editors.MeilisearchEngine.CreateMeilisearchIndex(ctx); err != nil {
			logger.Error("Wipe", "Failed to create editor search index", err)
			return err
		}
	}

	if err := app.Ads.MeilisearchEngine.WipeMeilisearchIndex(ctx); err != nil {
		logger.Error("Wipe", "Failed to wipe ads search index", err)
		return err
	} else {
		if err := app.Ads.MeilisearchEngine.CreateMeilisearchIndex(ctx); err != nil {
			logger.Error("Wipe", "Failed to create ads search index", err)
			return err
		}
	}

	logger.Info("Wipe", "Finished wiping and re-creating DB and search index schema")
	return nil
}
