package database

import (
	"gocms/internal/ads"
	"gocms/internal/category"
	"gocms/internal/editor"
	"gocms/internal/locale"
	"gocms/internal/outbox"
	"gocms/internal/post"
	"gocms/internal/user"

	"gorm.io/gorm"
)

func SoftMigrate(db *gorm.DB) error {

	err := db.AutoMigrate(
		&ads.Model{},
		&user.Record{},
		&editor.Model{},
		&post.Meta{},
		&post.Translation{},
		&category.Record{},
		&category.RecordTranslation{},
		&locale.Locale{},
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

	if err := db.Migrator().DropTable(
		&ads.Model{},
		&locale.Locale{},
		&editor.Model{},
		&user.Record{},
		&post.Meta{},
		&post.Translation{},
		&category.Record{},
		&category.RecordTranslation{},
		&outbox.Message{},
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
