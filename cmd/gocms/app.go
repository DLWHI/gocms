package main

import (
	"context"
	"gocms/internal/analytics"
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

type Config struct {
	Init    bool
	User    bool
	Wipe    bool
	Port    string
	Address string
	Storage string
}

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

func (app *Container) Init(ctx context.Context, cfg Config) error {
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

	storageHandler := DefaultStorageHandler(cfg)

	router := NewRouter(sessions)

	router.AddPostRoutes(posts.Handler)
	router.AddUserRoutes(users.Handler)
	router.AddCategoryRoutes(categories.Handler)
	router.AddAdsRoutes(ads.Handler)
	router.AddEditorRoutes(editors.Handler)
	router.AddLocaleRoutes(locales.Handler)
	router.AddStorageRoutes(storageHandler)
	router.AddPublicoutes(publicHandler)

	server := runner.NewHttp(cfg.Address, cfg.Port, router.mux)

	dispatcher := outbox.NewMessageDispatcher()
	relay := outbox.NewRelay(app.DB, dispatcher)

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

	return nil
}

func (app *Container) WireEvents(dispatcher outbox.PublishDriver) {
	dispatcher.Subscribe("post", app.Posts.Service.SyncSearch)
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

	app.group.Wait()
	return success
}
