package tests

import (
	"context"
	"gocms/internal/analytics"
	"gocms/internal/category"
	"gocms/internal/logger"
	"gocms/internal/outbox"
	"gocms/internal/public"
	"gocms/internal/runner"
	"gocms/internal/user"
	"log/slog"
	"sync"
	"time"

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
	group  sync.WaitGroup
	Logger *slog.Logger

	DB          *gorm.DB
	Redis       *redis.Client
	Meilisearch meilisearch.ServiceManager

	Categories *CategoryBundle
	Users      *UserBundle
	Posts      *PostBundle
	Ads        *AdsBundle
	Editors    *EditorBundle
	Locales    *LocaleBundle

	Server       runner.Runner
	ViewsCounter runner.Runner
	MessageRelay runner.Runner
}

func NewContainer() *Container {
	return &Container{}
}

func (app *Container) Init(ctx context.Context) error {
	app.Logger = slog.New(logger.NewReasonedHandler(nil))
	slog.SetDefault(app.Logger)

	db, err := MySQLDatabase(app)
	if err != nil {
		logger.Error("Init", "Failed to connect to database", err)
		return err
	}
	redis, err := RedisClient(ctx, app)
	if err != nil {
		logger.Error("Init", "Failed to connect to redis", err)
		return err
	}

	search, err := MeilisearchClient(app)
	if err != nil {
		logger.Error("Init", "Failed to connect to meilisearch", err)
		return err
	}

	sessions := user.NewRedisSessionStorage(redis, 24*time.Hour)

	categories := NewCategoryBundle(db, search, redis)
	users := NewUserBundle(db, search, redis, sessions)
	posts := NewPostBundle(db, search, redis)
	ads := NewAdsBundle(db, search, redis)
	editors := NewEditorBundle(db, search, redis)
	locales := NewLocaleBundle(db, redis)

	views := analytics.NewRedisCounter(app.Redis, posts.Service)

	publicHandler := public.NewHandler(
		posts.Service,
		editors.Service,
		categories.Service,
		locales.Service,
		ads.Service,
		views,
	)

	storageHandler := DefaultStorageHandler()

	router := NewRouter(sessions)

	router.AddPostRoutes(posts.Handler)
	router.AddUserRoutes(users.Handler)
	router.AddCategoryRoutes(categories.Handler)
	router.AddAdsRoutes(ads.Handler)
	router.AddEditorRoutes(editors.Handler)
	router.AddLocaleRoutes(locales.Handler)
	router.AddStorageRoutes(storageHandler)
	router.AddPublicoutes(publicHandler)

	server := runner.NewHttp(address, port, router.mux)

	dispatcher := outbox.NewDirectPublisher()
	relay := outbox.NewRelay(db, dispatcher)

	app.DB = db
	app.Redis = redis
	app.Meilisearch = search

	app.Categories = categories
	app.Users = users
	app.Posts = posts
	app.Ads = ads
	app.Editors = editors
	app.Locales = locales

	app.Server = server
	app.ViewsCounter = views
	app.MessageRelay = relay

	app.WireEvents(dispatcher)
	return nil
}

func (app *Container) WireEvents(dispatcher outbox.Publisher) {
	dispatcher.Subscribe("post", app.Posts.Service.SyncSearch)
	dispatcher.Subscribe("user", app.Users.Service.SyncSearch)
	dispatcher.Subscribe("category", app.Categories.Service.SyncSearch)
	dispatcher.Subscribe("editor", app.Editors.Service.SyncSearch)
	dispatcher.Subscribe("ads", app.Ads.Service.SyncSearch)
}

func (app *Container) TearDown() {
	sqlDB, err := app.DB.DB()
	if err != nil {
		logger.Error("Init", "Failed to close DB pool", err)
	}
	err = sqlDB.Close()
	if err != nil {
		logger.Error("Init", "Failed to sql DB connection", err)
	}

	err = app.Redis.Close()
	if err != nil {
		logger.Error("Init", "Failed to close Redis connection", err)
	}
}

func (app *Container) Start(ctx context.Context) {
	app.group.Add(1)
	go app.ViewsCounter.Start(ctx, &app.group, time.Minute*30)
	app.group.Add(1)
	go app.MessageRelay.Start(ctx, &app.group, 100*time.Millisecond)
	app.group.Add(1)
	go app.Server.Start(ctx, &app.group)
}

func (app *Container) Stop(timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	success := true

	if err := app.Server.Stop(ctx); err != nil {
		success = false
	}
	if err := app.ViewsCounter.Stop(ctx); err != nil {
		success = false
	}
	if err := app.MessageRelay.Stop(ctx); err != nil {
		success = false
	}

	app.group.Wait()
	return success
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
