package tests

import (
	"context"
	"fmt"
	"gocms/internal/advertisement"
	"gocms/internal/analytics"
	"gocms/internal/category"
	"gocms/internal/domain"
	"gocms/internal/editor"
	"gocms/internal/lang"
	"gocms/internal/logger"
	"gocms/internal/outbox"
	"gocms/internal/password"
	"gocms/internal/post"
	"gocms/internal/public"
	"gocms/internal/search"
	"gocms/internal/server"
	"gocms/internal/user"
	"log/slog"
	"net/http"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Container struct {
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

	Router       http.Handler
	ViewsCounter analytics.PostViewCounterService
	MessageRelay *outbox.Relay

	Subscriptions []*outbox.Subscription
}

func NewContainer() *Container {
	return &Container{}
}

func (app *Container) Init(ctx context.Context) error {
	db, err := MySQLDatabase(app)
	if err != nil {
		return fmt.Errorf("Failed to connect to database: %w", err)
	}
	redis, err := RedisClient(ctx, app)
	if err != nil {
		return fmt.Errorf("Failed to connect to redis: %w", err)
	}
	search, err := MeilisearchClient(app)
	if err != nil {
		return fmt.Errorf("Failed to connect to meilisearch: %w", err)
	}
	passwords := password.NewBCryptService(12, 8)

	app.DB = db
	app.Redis = redis
	app.Meilisearch = search
	app.Passwords = passwords

	if err := softMigrate(db); err != nil {
		return err
	}

	if err := app.wireModules(ctx); err != nil {
		return err
	}

	views := analytics.NewRedisCounter(app.Redis, app.Posts.Service)
	app.ViewsCounter = views

	app.wireRoutes(views)

	dispatcher := outbox.NewDirectPublisher()
	relay := outbox.NewRelay(db, dispatcher)
	app.MessageRelay = relay

	destroyIndexes(ctx, app.Meilisearch)
	if err := createIndexes(ctx, app.Meilisearch); err != nil {
		return err
	}
	app.wireEvents(dispatcher)

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

func (app *Container) wireRoutes(vc analytics.PostViewCounterService) {
	publicHandler := public.NewHandler(
		app.Posts.Service,
		app.Editors.Service,
		app.Categories.Service,
		app.Langs.Service,
		app.Ads.Service,
		vc,
	)

	storageHandler := DefaultStorageHandler(app.Auth.SessionRepo)

	mw := server.ChainMiddleware(
		server.CORS,
		server.RequireJSON,
	)
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

	app.Router = mw(router)
}

func (app *Container) TearDown() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	logger.Info("Wipe", "Tearing down application environment")
	for _, sub := range app.Subscriptions {
		sub.Unsubscribe()
	}

	if err := dropAll(app.DB); err != nil {
		logger.Error("Wipe", "Failed to drop table records", err)
	}

	if err := destroyIndexes(ctx, app.Meilisearch); err != nil {
		logger.Error("Wipe", "Failed to delete search indexes", err)
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

func softMigrate(db *gorm.DB) error {

	err := db.AutoMigrate(
		&advertisement.Record{},
		&user.Record{},
		&editor.Record{},
		&editor.RecordTranslation{},
		&post.Record{},
		&post.RecordTranslation{},
		&category.Record{},
		&category.RecordTranslation{},
		&lang.Language{},
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

func createIndexes(ctx context.Context, client meilisearch.ServiceManager) error {
	indexes := []search.ConfigurableIndex{
		user.NewMeilisearchEngine(client, UsersSearchIndexName),
		category.NewMeilisearchEngine(client, CategoriesSearchIndexName),
		editor.NewMeilisearchEngine(client, EditorsSearchIndexName),
		post.NewMeilisearchEngine(client, PostsSearchIndexName),
		advertisement.NewMeilisearchEngine(client, AdsSearchIndexName),
	}

	for _, index := range indexes {
		if err := index.Create(ctx); err != nil {
			return err
		}
	}

	return nil
}

func dropAll(db *gorm.DB) error {
	err := db.Migrator().DropTable("post_categories")
	if err != nil {
		return err
	}
	if err := db.Migrator().DropTable(
		&advertisement.Record{},
		&user.Record{},
		&editor.Record{},
		&editor.RecordTranslation{},
		&post.Record{},
		&post.RecordTranslation{},
		&category.Record{},
		&category.RecordTranslation{},
		&lang.Language{},
		&outbox.Message{},
	); err != nil {
		return err
	}
	return nil
}

func destroyIndexes(ctx context.Context, client meilisearch.ServiceManager) error {
	indexes := []search.ConfigurableIndex{
		user.NewMeilisearchEngine(client, UsersSearchIndexName),
		category.NewMeilisearchEngine(client, CategoriesSearchIndexName),
		editor.NewMeilisearchEngine(client, EditorsSearchIndexName),
		post.NewMeilisearchEngine(client, PostsSearchIndexName),
		advertisement.NewMeilisearchEngine(client, AdsSearchIndexName),
	}

	for _, index := range indexes {
		if err := index.Destroy(ctx); err != nil {
			return err
		}
	}

	return nil
}
