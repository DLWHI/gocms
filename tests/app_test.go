package tests

import (
	"context"
	"gocms/internal/ads"
	"gocms/internal/analytics"
	"gocms/internal/category"
	"gocms/internal/database"
	"gocms/internal/domain"
	"gocms/internal/editor"
	"gocms/internal/flow"
	"gocms/internal/locale"
	"gocms/internal/logger"
	"gocms/internal/outbox"
	"gocms/internal/password"
	"gocms/internal/post"
	"gocms/internal/public"
	"gocms/internal/session"
	"gocms/internal/user"
	"log/slog"
	"sync"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
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
	Auth       *AuthBundle

	PasswordService password.PasswordService
	MessageRelay    outbox.MessageRepository
	Router          *Router

	ViewsCounter flow.Runner
	OutboxRunner flow.Runner

	Subscriptions []*outbox.Subscription
}

func NewContainer() *Container {
	return &Container{}
}

func (app *Container) Init(ctx context.Context) error {
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

	passwords := password.NewBCryptService(12, 8)
	sessions := session.NewRedisRepository(redis, 24*time.Hour)

	categories := NewCategoryBundle(db, search, redis)
	users := NewUserBundle(db, search, redis, passwords)
	posts := NewPostBundle(db, search, redis)
	ads := NewAdsBundle(db, search, redis)
	editors := NewEditorBundle(db, search, redis)
	locales := NewLocaleBundle(db, redis)
	auth := NewAuthBundle(users.Service, sessions)

	views := analytics.NewRedisCounter(redis, posts.Service)

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

	router.AddAuthRoutes(auth.Handler)
	router.AddPostRoutes(posts.Handler)
	router.AddUserRoutes(users.Handler)
	router.AddCategoryRoutes(categories.Handler)
	router.AddAdsRoutes(ads.Handler)
	router.AddEditorRoutes(editors.Handler)
	router.AddLocaleRoutes(locales.Handler)
	router.AddStorageRoutes(storageHandler)
	router.AddPublicoutes(publicHandler)

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
	app.Auth = auth

	app.PasswordService = passwords
	app.MessageRelay = relay
	app.Router = router

	app.ViewsCounter = views
	app.OutboxRunner = relay

	app.WireEvents(dispatcher)
	if err := Migrate(ctx, app); err != nil {
		logger.Error("Init", "Failed to migrate schema", err)
		return err
	}

	return nil
}

func (app *Container) WireEvents(dispatcher outbox.Publisher) {
	app.Subscriptions = make([]*outbox.Subscription, 0, 6)
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe("post", app.Posts.Service.SyncSearch))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe(domain.UserEventName, app.Users.Consumer.SyncSearch))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe("user", app.Auth.Consumer.RevokeSessions))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe("category", app.Categories.Service.SyncSearch))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe("editor", app.Editors.Service.SyncSearch))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe("ads", app.Ads.Service.SyncSearch))
}

func (app *Container) TearDown() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	logger.Info("Wipe", "Tearing down application environment")
	for _, sub := range app.Subscriptions {
		sub.Unsubscribe()
	}

	err := app.DB.Migrator().DropTable("post_categories")
	if err != nil {
		logger.Error("Wipe", "Failed to drop table records", err)
	}
	if err := app.DB.WithContext(ctx).Migrator().DropTable(&ads.Model{},
		&user.Record{},
		&editor.Model{},
		&post.Meta{},
		&post.Translation{},
		&category.Model{},
		&category.Translation{},
		&locale.Locale{},
		&outbox.Message{},
	); err != nil {
		logger.Error("Wipe", "Failed to drop table records", err)
	}

	if err := app.Posts.MeilisearchEngine.WipeMeilisearchIndex(ctx); err != nil {
		logger.Error("Wipe", "Failed to delete post search index", err)
	}
	if err := app.Users.MeilisearchEngine.WipeMeilisearchIndex(ctx); err != nil {
		logger.Error("Wipe", "Failed to delete user search index", err)
	}
	if err := app.Categories.MeilisearchEngine.WipeMeilisearchIndex(ctx); err != nil {
		logger.Error("Wipe", "Failed to delete category search index", err)
	}
	if err := app.Editors.MeilisearchEngine.WipeMeilisearchIndex(ctx); err != nil {
		logger.Error("Wipe", "Failed to delete editor search index", err)
	}
	if err := app.Ads.MeilisearchEngine.WipeMeilisearchIndex(ctx); err != nil {
		logger.Error("Wipe", "Failed to delete ads search index", err)
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
	go app.OutboxRunner.Start(ctx, &app.group, 100*time.Millisecond)
}

func (app *Container) Stop(timeout time.Duration) bool {
	logger.Info("Wipe", "Stopping application")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	success := true

	if err := app.ViewsCounter.Stop(ctx); err != nil {
		success = false
	}
	if err := app.OutboxRunner.Stop(ctx); err != nil {
		success = false
	}

	app.group.Wait()

	return success
}

func Migrate(ctx context.Context, app *Container) error {
	if err := database.SoftMigrate(app.DB); err != nil {
		logger.Error("Wipe", "Failed to migrate DB", err)
		return err
	}
	if err := app.Posts.MeilisearchEngine.ConfigureMeilisearchIndex(ctx); err != nil {
		logger.Error("Wipe", "Failed to configure post search index", err)
		return err
	}
	if err := app.Users.MeilisearchEngine.ConfigureMeilisearchIndex(ctx); err != nil {
		logger.Error("Wipe", "Failed to configure user search index", err)
		return err
	}
	if err := app.Categories.MeilisearchEngine.ConfigureMeilisearchIndex(ctx); err != nil {
		logger.Error("Wipe", "Failed to configure category search index", err)
		return err
	}
	if err := app.Editors.MeilisearchEngine.ConfigureMeilisearchIndex(ctx); err != nil {
		logger.Error("Wipe", "Failed to configure editor search index", err)
		return err
	}
	if err := app.Ads.MeilisearchEngine.ConfigureMeilisearchIndex(ctx); err != nil {
		logger.Error("Wipe", "Failed to configure ads search index", err)
		return err
	}
	logger.Info("Migrate", "Migration successful")
	return nil
}
