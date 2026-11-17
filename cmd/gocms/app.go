package main

import (
	"context"
	"gocms/internal/analytics"
	"gocms/internal/domain"
	"gocms/internal/flow"
	"gocms/internal/lang"
	"gocms/internal/logger"
	"gocms/internal/outbox"
	"gocms/internal/password"
	"gocms/internal/public"
	"gocms/internal/server"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Config struct {
	Init      bool
	User      bool
	HardReset bool
	Port      string
	Address   string
	Storage   string
}

type Container struct {
	group  sync.WaitGroup
	Logger *slog.Logger

	DB          *gorm.DB
	Redis       *redis.Client
	Meilisearch meilisearch.ServiceManager
	Passwords   password.Service

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
	app.DB = db
	app.Redis = redis
	app.Meilisearch = search
	app.Passwords = passwords

	if err := app.wireModules(ctx); err != nil {
		return err
	}

	views := analytics.NewRedisCounter(app.Redis, app.Posts.Service)
	app.ViewsCounter = views
	app.wireRoutes(cfg, views)

	dispatcher := outbox.NewDirectPublisher()
	app.wireEvents(dispatcher)
	relay := outbox.NewRelay(db, dispatcher)

	app.MessageRelay = relay

	return nil
}

func (app *Container) wireEvents(dispatcher outbox.Publisher) {
	app.Subscriptions = make([]*outbox.Subscription, 0, 6)
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe(domain.PostEvent, app.Posts.Consumer.SyncSearch))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe(domain.UserEvent, app.Users.Consumer.SyncSearch))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe(domain.UserEvent, app.Auth.Consumer.RevokeSessions))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe(domain.CategoryEvent, app.Categories.Consumer.SyncSearch))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe(domain.CategoryEvent, app.Posts.Consumer.SyncCategories))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe(domain.EditorEvent, app.Editors.Consumer.SyncSearch))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe(domain.AdvertisementEvent, app.Ads.Consumer.SyncSearch))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe(lang.Event, app.Langs.Consumer.Reload))
}

func (app *Container) wireModules(ctx context.Context) error {
	auths, err := NewAuthModule(app.DB, app.Redis, app.Passwords)
	if err != nil {
		return err
	}
	langs, err := NewLangModule(
		ctx,
		app.DB,
		auths.Service,
		auths.SessionRepo,
	)
	if err != nil {
		return err
	}
	categories, err := NewCategoryModule(
		app.DB,
		app.Meilisearch,
		app.Redis,
		langs.Service,
		auths.Service,
		auths.SessionRepo,
	)
	if err != nil {
		return err
	}
	users, err := NewUserModule(
		app.DB,
		app.Meilisearch,
		app.Redis,
		auths.Service,
		app.Passwords,
		auths.SessionRepo,
	)
	if err != nil {
		return err
	}
	posts, err := NewPostModule(
		app.DB,
		app.Meilisearch,
		app.Redis,
		langs.Service,
		auths.Service,
		auths.SessionRepo,
	)
	if err != nil {
		return err
	}
	ads, err := NewAdsModule(
		app.DB,
		app.Meilisearch,
		app.Redis,
		auths.Service,
		auths.SessionRepo,
	)
	if err != nil {
		return err
	}
	editors, err := NewEditorModule(
		app.DB,
		app.Meilisearch,
		app.Redis,
		langs.Service,
		auths.Service,
		auths.SessionRepo,
	)
	if err != nil {
		return err
	}

	app.Categories = categories
	app.Users = users
	app.Posts = posts
	app.Ads = ads
	app.Editors = editors
	app.Langs = langs
	app.Auth = auths
	return nil
}

func (app *Container) wireRoutes(cfg Config, vc analytics.PostViewCounterService) {
	publicHandler := public.NewHandler(
		app.Posts.Service,
		app.Editors.Service,
		app.Categories.Service,
		app.Langs.Service,
		app.Ads.Service,
		vc,
	)

	storageHandler := DefaultStorageHandler(cfg, app.Auth.SessionRepo)

	cmsRouter := http.NewServeMux()
	storageRouter := http.NewServeMux()
	publicRouter := http.NewServeMux()

	app.Categories.Handler.RegisterRoutes(cmsRouter)
	app.Users.Handler.RegisterRoutes(cmsRouter)
	app.Posts.Handler.RegisterRoutes(cmsRouter)
	app.Ads.Handler.RegisterRoutes(cmsRouter)
	app.Editors.Handler.RegisterRoutes(cmsRouter)
	app.Langs.Handler.RegisterRoutes(cmsRouter)
	app.Auth.Handler.RegisterRoutes(cmsRouter)

	publicHandler.RegisterRoutes(publicRouter)
	storageHandler.RegisterRoutes(storageRouter)

	router := http.NewServeMux()
	router.Handle("/cms/", http.StripPrefix("/cms", cmsRouter))
	router.Handle("/storage/", http.StripPrefix("/storage", storageRouter))
	router.Handle("/public/", http.StripPrefix("/public", publicRouter))

	app.Server = server.NewHttpRunner(cfg.Address, cfg.Port, server.CORS(router))
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
