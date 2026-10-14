package database

import (
	"gocms/internal/ads"
	"gocms/internal/category"
	"gocms/internal/editor"
	"gocms/internal/locale"
	"gocms/internal/post"
	"gocms/internal/user"

	"gorm.io/gorm"
)

func SoftMigrate(db *gorm.DB) error {

	err := db.AutoMigrate(
		&ads.Model{},
		&user.Model{},
		&editor.Model{},
		&post.Meta{},
		&post.Translation{},
		&category.Model{},
		&category.Translation{},
		&locale.Model{},
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

	if err := db.Migrator().DropTable(
		&ads.Model{},
		&locale.Model{},
		&editor.Model{},
		&user.Model{},
		&post.Meta{},
		&post.Translation{},
		&category.Model{},
		&category.Translation{},
	); err != nil {
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
