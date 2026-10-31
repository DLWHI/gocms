package main

import (
	"context"
	"gocms/internal/analytics"
	"gocms/internal/domain"
	"gocms/internal/flow"
	"gocms/internal/logger"
	"gocms/internal/outbox"
	"gocms/internal/password"
	"gocms/internal/public"
	"gocms/internal/server"
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
	Passwords   password.PasswordService

	Categories *CategoryModule
	Users      *UserModule
	Posts      *PostModule
	Ads        *AdsModule
	Editors    *EditorModule
	Langs      *LangModule
	Auth       *AuthModule

	Server       flow.Runner
	ViewsCounter flow.Runner
	MessageRelay flow.Runner

	Subscriptions []*outbox.Subscription
}

func NewContainer() *Container {
	return &Container{}
}

func (app *Container) Init(ctx context.Context, cfg Config) error {
	app.Logger = slog.New(logger.NewReasonedHandler(nil))
	slog.SetDefault(app.Logger)

	db, err := MySQLDatabase(app)
	if err != nil {
		return err
	}
	redis, err := RedisClient(ctx, app)
	if err != nil {
		return err
	}

	search, err := MeilisearchClient(app)
	if err != nil {
		return err
	}

	passwords := password.NewBCryptService(12, 8)

	auth, err := NewAuthModule(db, redis, passwords)
	if err != nil {
		return err
	}
	categories, err := NewCategoryModule(db, search, redis, auth.Service)
	if err != nil {
		return err
	}
	users, err := NewUserModule(db, search, redis, auth.Service, passwords)
	if err != nil {
		return err
	}
	posts := NewPostModule(db, search, redis)
	ads := NewAdsModule(db, search, redis)
	editors, err := NewEditorModule(db, search, redis, auth.Service)
	if err != nil {
		return err
	}
	langs, err := NewLangModule(ctx, db, auth.Service)
	if err != nil {
		return err
	}

	views := analytics.NewRedisCounter(app.Redis, posts.Service)

	publicHandler := public.NewHandler(
		posts.Service,
		editors.Service,
		categories.Service,
		langs.Service,
		ads.Service,
		views,
	)

	storageHandler := DefaultStorageHandler(cfg)

	router := NewRouter(auth.SessionRepo)

	router.AddAuthRoutes(auth.Handler)
	router.AddPostRoutes(posts.Handler)
	router.AddUserRoutes(users.Handler)
	router.AddCategoryRoutes(categories.Handler)
	router.AddAdsRoutes(ads.Handler)
	router.AddEditorRoutes(editors.Handler)
	router.AddLangRoutes(langs.Handler)
	router.AddStorageRoutes(storageHandler)
	router.AddPublicoutes(publicHandler)

	server := server.NewHttpRunner(cfg.Address, cfg.Port, router.mux)

	dispatcher := outbox.NewDirectPublisher()
	relay := outbox.NewRelay(db, dispatcher)

	app.DB = db
	app.Redis = redis
	app.Meilisearch = search
	app.Passwords = passwords

	app.Categories = categories
	app.Users = users
	app.Posts = posts
	app.Ads = ads
	app.Editors = editors
	app.Langs = langs
	app.Auth = auth

	app.Server = server
	app.ViewsCounter = views
	app.MessageRelay = relay

	app.WireEvents(dispatcher)
	return nil
}

func (app *Container) WireEvents(dispatcher outbox.Publisher) {
	app.Subscriptions = make([]*outbox.Subscription, 0, 6)
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe("post", app.Posts.Service.SyncSearch))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe(domain.UserEventName, app.Users.Consumer.SyncSearch))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe(domain.UserEventName, app.Auth.Consumer.RevokeSessions))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe(domain.CategoryEventName, app.Categories.Consumer.SyncSearch))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe(domain.EditorEventName, app.Editors.Consumer.SyncSearch))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe("ads", app.Ads.Service.SyncSearch))
}

func (app *Container) TearDown() {
	for _, sub := range app.Subscriptions {
		sub.Unsubscribe()
	}

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
