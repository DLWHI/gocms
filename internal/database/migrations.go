package database

import (
	"gocms/internal/ads"
	"gocms/internal/category"
	"gocms/internal/editor"
	"gocms/internal/language"
	"gocms/internal/outbox"
	"gocms/internal/post"
	"gocms/internal/user"

	"gorm.io/gorm"
)

func SoftMigrate(db *gorm.DB) error {

	err := db.AutoMigrate(
		&ads.Model{},
		&user.Record{},
		&editor.Record{},
		&editor.RecordTranslation{},
		&post.Meta{},
		&post.Translation{},
		&category.Record{},
		&category.RecordTranslation{},
		&language.Record{},
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
		&ads.Model{},
		&user.Record{},
		&editor.Record{},
		&editor.RecordTranslation{},
		&post.Meta{},
		&post.Translation{},
		&category.Record{},
		&category.RecordTranslation{},
		&language.Record{},
		&outbox.Message{},
	); err != nil {
		return err
	}
	return nil
}
