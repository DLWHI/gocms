package tests

import (
	"context"
	"fmt"
	"gocms/internal/ads"
	"gocms/internal/category"
	"gocms/internal/domain"
	"gocms/internal/editor"
	"gocms/internal/flow"
	"gocms/internal/lang"
	"gocms/internal/logger"
	"gocms/internal/outbox"
	"gocms/internal/password"
	"gocms/internal/post"
	"gocms/internal/search"
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

	Categories *CategoryModule
	Users      *UserModule
	Posts      *PostModule
	Ads        *AdsModule
	Editors    *EditorModule
	Langs      *LangModule
	Auth       *AuthModule

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

	auth, err := NewAuthModule(db, redis, passwords)
	if err != nil {
		return err
	}
	langs, err := NewLangModule(ctx, db, auth.Service)
	if err != nil {
		return err
	}
	categories, err := NewCategoryModule(db, search, redis, langs.Service, auth.Service)
	if err != nil {
		return err
	}
	users, err := NewUserModule(db, search, redis, auth.Service, passwords)
	if err != nil {
		return err
	}
	posts := NewPostModule(db, search, redis)
	ads := NewAdsModule(db, search, redis)
	editors, err := NewEditorModule(db, search, redis, langs.Service, auth.Service)
	if err != nil {
		return err
	}

	// views := analytics.NewRedisCounter(redis, posts.Service)

	// publicHandler := public.NewHandler(
	// 	posts.Service,
	// 	editors.Service,
	// 	categories.Service,
	// 	langs.Service,
	// 	ads.Service,
	// 	views,
	// )

	storageHandler := DefaultStorageHandler()

	router := NewRouter(auth.SessionRepo)

	router.AddAuthRoutes(auth.Handler)
	// router.AddPostRoutes(posts.Handler)
	router.AddUserRoutes(users.Handler)
	router.AddCategoryRoutes(categories.Handler)
	// router.AddAdsRoutes(ads.Handler)
	router.AddEditorRoutes(editors.Handler)
	router.AddLangRoutes(langs.Handler)
	router.AddStorageRoutes(storageHandler)
	// router.AddPublicoutes(publicHandler)

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
	app.Langs = langs
	app.Auth = auth

	app.PasswordService = passwords
	app.MessageRelay = relay
	app.Router = router

	// app.ViewsCounter = views
	app.OutboxRunner = relay

	if err := Migrate(ctx, app); err != nil {
		return fmt.Errorf("Failed to migrate schema: %w", err)
	}
	app.WireEvents(dispatcher)

	return nil
}

func (app *Container) WireEvents(dispatcher outbox.Publisher) {
	app.Subscriptions = make([]*outbox.Subscription, 0, 6)
	// app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe("post", app.Posts.Service.SyncSearch))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe(domain.UserEventName, app.Users.Consumer.SyncSearch))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe(domain.UserEventName, app.Auth.Consumer.RevokeSessions))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe(domain.CategoryEventName, app.Categories.Consumer.SyncSearch))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe(domain.EditorEventName, app.Editors.Consumer.SyncSearch))
	// app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe("ads", app.Ads.Service.SyncSearch))
	app.Subscriptions = append(app.Subscriptions, dispatcher.Subscribe(domain.LangUpdateEvent, app.Langs.Consumer.Reload))
}

func (app *Container) TearDown() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	logger.Info("Wipe", "Tearing down application environment")
	for _, sub := range app.Subscriptions {
		sub.Unsubscribe()
	}

	if err := DropAll(app.DB); err != nil {
		logger.Error("Wipe", "Failed to drop table records", err)
	}

	if err := DestroyIndexes(ctx, app.Meilisearch); err != nil {
		logger.Error("Wipe", "Failed to drop table records", err)
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

func Migrate(ctx context.Context, app *Container) error {
	if err := SoftMigrate(app.DB); err != nil {
		return err
	}

	if err := CreateIndexes(ctx, app.Meilisearch); err != nil {
		return err
	}
	return nil
}

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

func CreateIndexes(ctx context.Context, client meilisearch.ServiceManager) error {
	indexes := []search.ConfigurableIndex{
		user.NewMeilisearchEngine(client, UsersSearchIndexName),
		category.NewMeilisearchEngine(client, CategoriesSearchIndexName),
		editor.NewMeilisearchEngine(client, EditorsSearchIndexName),
		post.NewMeilisearchEngine(client, PostsSearchIndexName),
		ads.NewMeilisearchEngine(client, AdsSearchIndexName),
	}

	for _, index := range indexes {
		if err := index.Create(ctx); err != nil {
			return err
		}
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
		&lang.Language{},
		&outbox.Message{},
	); err != nil {
		return err
	}
	return nil
}

func DestroyIndexes(ctx context.Context, client meilisearch.ServiceManager) error {
	indexes := []search.ConfigurableIndex{
		user.NewMeilisearchEngine(client, UsersSearchIndexName),
		category.NewMeilisearchEngine(client, CategoriesSearchIndexName),
		editor.NewMeilisearchEngine(client, EditorsSearchIndexName),
		post.NewMeilisearchEngine(client, PostsSearchIndexName),
		ads.NewMeilisearchEngine(client, AdsSearchIndexName),
	}

	for _, index := range indexes {
		if err := index.Destroy(ctx); err != nil {
			return err
		}
	}

	return nil
}
