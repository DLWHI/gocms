package tests

import (
	"context"
	"gocms/internal/ads"
	"gocms/internal/category"
	"gocms/internal/editor"
	"gocms/internal/locale"
	"gocms/internal/post"
	"gocms/internal/user"

	"github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	categories = []category.Translation{
		{
			Locale:       "en",
			Name:         "Kekw",
			CategorySlug: "cookinja",
			Category: category.Model{
				Slug: "cookinja",
			},
		}, {
			Locale:       "en",
			Name:         "Category",
			CategorySlug: "cat",
			Category: category.Model{
				Slug: "cat",
			},
		}, {
			Locale:       "ru",
			Name:         "Категория",
			CategorySlug: "cat",
			Category: category.Model{
				Slug: "cat",
			},
		}, {
			Locale:       "en",
			Name:         "Some other category",
			CategorySlug: "other-cat",
			Category: category.Model{
				Slug: "other-cat",
			},
		}, {
			Locale:       "ru",
			Name:         "Другая категория",
			CategorySlug: "other-cat",
			Category: category.Model{
				Slug: "other-cat",
			},
		},
	}
	// ----------------------------------------------------------
	users = []user.Model{
		{Username: "test"},
		{Username: "editor", Email: "editor@mail.com"},
	}
)

type Container struct {
	Context context.Context
	Stop    context.CancelFunc

	DB          *gorm.DB
	Redis       *redis.Client
	Meilisearch meilisearch.ServiceManager

	CategoryHandler *category.Handler
	UserHandler     *user.Handler
	PostHandler     *post.Handler
	AdHandler       *ads.Handler
	LocalesHandler  *locale.Handler
	EditorHandler   *editor.Handler

	RedisSessionManager *user.RedisSessionStorage
}

func NewContainer() *Container {
	return &Container{}
}

func (app *Container) Init() error {
	// app.Context, app.Stop = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// if err := os.RemoveAll("./content"); err != nil {
	// 	return err
	// }
	// app.DB = MySQLDatabase(app)
	// if app.DB == nil {
	// 	return fmt.Errorf("[ERROR] No database instance provided")
	// }
	// app.Redis = RedisClient(app)
	// if app.Redis == nil {
	// 	return fmt.Errorf("[ERROR] No working redis client, aborting")
	// }
	// app.Meilisearch = MeilisearchClient(app)
	// if app.Meilisearch == nil {
	// 	return fmt.Errorf("[ERROR] No working search client provided, aborting")
	// }
	// app.deleteIndexes()
	// if err := app.createIndexes(); err != nil {
	// 	return err
	// }

	// app.RedisSessionManager = RedisSessionManager(app)

	// app.PostHandler = DefaultPostHandler(app)
	// app.UserHandler = DefaultUserHandler(app)
	// app.CategoryHandler = DefaultCategoryHandler(app)
	// app.AdHandler = DefaultAdvertisementsHandler(app)
	// app.LocalesHandler = DefaultLocalesHandler(app)
	// app.EditorHandler = DefaultEditorHandler(app)

	// if err := database.HardMigrate(app.DB, "cms_routes_test"); err != nil {
	// 	return fmt.Errorf("[ERROR] Failed to run schema setup: %w", err)
	// }

	// if err := app.seedCategories(); err != nil {
	// 	return err
	// }
	// if err := app.seedUsers(); err != nil {
	// 	return err
	// }

	return nil
}

func (app *Container) TearDown() {
	// app.Stop()
	// if err := os.RemoveAll("./content"); err != nil {
	// 	fmt.Printf("[ERROR] Failed to run content storage cleanup: %v", err)
	// }

	// result := app.DB.Exec("DROP DATABASE IF EXISTS cms_routes_test;")
	// if result.Error != nil {
	// 	fmt.Printf("[ERROR] Failed to run schema cleanup: %v", result.Error)
	// }

	// sqlDB, err := app.DB.DB()
	// if err != nil {
	// 	fmt.Printf("[ERROR] Unable to obtain DB pool: %v", err)
	// }
	// err = sqlDB.Close()
	// if err != nil {
	// 	fmt.Printf("[ERROR] Unable to close DB pool: %v", err)
	// }

	// err = app.Redis.Close()
	// if err != nil {
	// 	fmt.Printf("[ERROR] Unable to close Redis connection: %v", err)
	// }
	// app.deleteIndexes()
}

func (app *Container) createIndexes() error {
	// if err := search.IndexPosts(app.Meilisearch, "posts_test"); err != nil {
	// 	return fmt.Errorf("[ERROR] Failed to run index schema setup: %w", err)
	// }
	// if err := search.IndexUsers(app.Meilisearch, "users_test"); err != nil {
	// 	return fmt.Errorf("[ERROR] Failed to run user index schema setup: %w", err)
	// }
	// if err := search.IndexEditors(app.Meilisearch, "editors_test"); err != nil {
	// 	return fmt.Errorf("[ERROR] Failed to run user index schema setup: %w", err)
	// }
	// if err := search.IndexAds(app.Meilisearch, "ads_test"); err != nil {
	// 	return fmt.Errorf("[ERROR] Failed to run user index schema setup: %w", err)
	// }
	// if err := search.IndexCategories(app.Meilisearch, "categories_test"); err != nil {
	// 	return fmt.Errorf("[ERROR] Failed to run user index schema setup: %w", err)
	// }
	return nil
}

func (app *Container) deleteIndexes() {
	// client := app.Meilisearch
	// _, err := client.DeleteIndex("posts_test")
	// if err != nil {
	// 	fmt.Printf("[ERROR] Unable to clean post test index: %v", err)
	// }
	// _, err = client.DeleteIndex("users_test")
	// if err != nil {
	// 	fmt.Printf("[ERROR] Unable to clean user test index: %v", err)
	// }
	// _, err = client.DeleteIndex("editors_test")
	// if err != nil {
	// 	fmt.Printf("[ERROR] Unable to clean editor test index: %v", err)
	// }
	// _, err = client.DeleteIndex("ads_test")
	// if err != nil {
	// 	fmt.Printf("[ERROR] Unable to clean advertisements test index: %v", err)
	// }
	// _, err = client.DeleteIndex("categories_test")
	// if err != nil {
	// 	fmt.Printf("[ERROR] Unable to clean categories test index: %v", err)
	// }
}

func (app *Container) seedCategories() error {
	// var categoriesDocs []category.Doc
	// if err := app.DB.Create(&categories).Error; err != nil {
	// 	return fmt.Errorf("[ERROR] Unable to insert categories into database: %w", err)
	// }

	// for i, category := range categories {
	// 	categoriesDocs = append(categoriesDocs, category.ToDoc())
	// 	categoriesDocs[i].ID = fmt.Sprintf("%s-%s", category.CategorySlug, category.Locale)
	// }
	// taskInfo, err := app.Meilisearch.Index("categories_test").AddDocuments(&categoriesDocs, nil)
	// if err != nil {
	// 	return fmt.Errorf("[ERROR] Unable to insert categories into index: %w", err)
	// }

	// task, err := app.Meilisearch.WaitForTask(taskInfo.TaskUID, 200*time.Millisecond)
	// if err != nil {
	// 	return err
	// }
	// if task.Status == meilisearch.TaskStatusFailed {
	// 	return fmt.Errorf("Error configuring index: %s - %s", task.Error.Code, task.Error.Message)
	// }
	return nil
}

func (app *Container) seedUsers() error {
	// var usersDoc []user.Doc
	// if err := app.DB.Create(&users).Error; err != nil {
	// 	return fmt.Errorf("[ERROR] Unable to insert users into database: %w", err)
	// }

	// for _, user := range users {
	// 	usersDoc = append(usersDoc, user.ToDoc())
	// }
	// taskInfo, err := app.Meilisearch.Index("users_test").AddDocuments(&usersDoc, nil)
	// if err != nil {
	// 	return fmt.Errorf("[ERROR] Unable to insert users into index: %w", err)
	// }

	// task, err := app.Meilisearch.WaitForTask(taskInfo.TaskUID, 200*time.Millisecond)
	// if err != nil {
	// 	return err
	// }
	// if task.Status == meilisearch.TaskStatusFailed {
	// 	return fmt.Errorf("Error configuring index: %s - %s", task.Error.Code, task.Error.Message)
	// }
	return nil
}
