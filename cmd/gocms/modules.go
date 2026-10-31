package main

import (
	"context"
	"gocms/internal/ads"
	"gocms/internal/auth"
	"gocms/internal/cache"
	"gocms/internal/category"
	"gocms/internal/domain"
	"gocms/internal/editor"
	"gocms/internal/locale"
	"gocms/internal/password"
	"gocms/internal/post"
	"gocms/internal/user"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	UsersSearchIndexName      = "users"
	PostsSearchIndexName      = "posts"
	EditorsSearchIndexName    = "editor"
	CategoriesSearchIndexName = "categories"
	AdsSearchIndexName        = "ads"
)

type AuthModule struct {
	Consumer    *auth.EventConsumer
	Service     *auth.Service
	Handler     *auth.Handler
	UserRepo    auth.UserRepository
	SessionRepo auth.SessionRepository
}

type UserModule struct {
	Consumer          *user.EventConsumer
	Service           *user.Service
	Handler           *user.Handler
	CacheRepo         user.Repository
	DedupRepo         user.Repository
	DBRepo            user.Repository
	MeilisearchEngine *user.MeilisearchEngine
}

type PostModule struct {
	Service           *post.Service
	Handler           *post.Handler
	RedisRepo         *post.RedisRepository
	GORMRepo          *post.GORMRepository
	MeilisearchEngine *post.MeilisearchEngine
}

type EditorModule struct {
	Consumer          *editor.EventConsumer
	Service           *editor.Service
	Handler           *editor.Handler
	CacheRepoRepo     editor.Repository
	DBRepo            editor.Repository
	DedupRepo         editor.Repository
	MeilisearchEngine *editor.MeilisearchEngine
}

type CategoryModule struct {
	Consumer          *category.EventConsumer
	Service           *category.Service
	Handler           *category.Handler
	DedupRepo         category.Repository
	CacheRepo         category.Repository
	DBRepo            category.Repository
	MeilisearchEngine *category.MeilisearchEngine
}

type AdsModule struct {
	Service           *ads.Service
	Handler           *ads.Handler
	RedisRepo         *ads.RedisRepository
	GORMRepo          *ads.GORMRepository
	MeilisearchEngine *ads.MeilisearchEngine
}

type LocaleModule struct {
	Service   *locale.Service
	Handler   *locale.Handler
	CacheRepo locale.Repository
	DBRepo    locale.Repository
	DedupRepo locale.Repository
}

func NewAuthModule(db *gorm.DB, redis *redis.Client, passwords password.PasswordService) (*AuthModule, error) {
	keys, err := cache.NewKeyspace(cache.KeyConfig{
		Prefix: []string{"cms"},
	}, "sessions")
	if err != nil {
		return nil, err
	}

	userRepo := user.NewGORMRepository(db)
	sessionRepo := auth.NewRedisSessionRepository(redis, keys, 24*time.Hour)
	consumer := auth.NewEventConsumer(sessionRepo, userRepo)
	service := auth.NewService(sessionRepo, userRepo, passwords)
	handler := auth.NewHandler(service)
	return &AuthModule{
		Consumer:    consumer,
		Service:     service,
		Handler:     handler,
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
	}, nil
}

func NewUserModule(
	db *gorm.DB,
	search meilisearch.ServiceManager,
	redis *redis.Client,
	creds domain.AuthService,
	passwords password.PasswordService,
) (*UserModule, error) {
	keys, err := cache.NewKeyspace(cache.KeyConfig{
		Prefix: []string{"cms"},
	}, "users")
	if err != nil {
		return nil, err
	}
	gormRepo := user.NewGORMRepository(db)
	flowRepo := user.NewFlowRepository(gormRepo, 20*time.Second, keys)
	redisRepo := user.NewRedisRepository(flowRepo, redis, 30*time.Minute, keys)
	searchEngine := user.NewMeilisearchEngine(search, UsersSearchIndexName)
	consumer := user.NewEventConsumer(gormRepo, searchEngine)
	service := user.NewService(redisRepo, searchEngine, creds, passwords)
	handler := user.NewHandler(service)
	return &UserModule{
		Consumer:          consumer,
		Service:           service,
		Handler:           handler,
		CacheRepo:         redisRepo,
		DedupRepo:         flowRepo,
		DBRepo:            gormRepo,
		MeilisearchEngine: searchEngine,
	}, nil
}

func NewPostModule(db *gorm.DB, search meilisearch.ServiceManager, redis *redis.Client) *PostModule {
	redisCfg := post.RedisConfig{
		Client:    redis,
		TTL:       30 * time.Minute,
		KeyFormat: "cms:posts:%d:%s",
	}

	repo := post.NewGORMRepository(db)
	cacheRepo := post.NewRedisRepository(repo, &redisCfg)
	searchEngine := post.NewMeilisearchEngine(search, PostsSearchIndexName)
	service := post.NewService(cacheRepo, searchEngine)
	handler := post.NewHandler(service)
	return &PostModule{
		Service:           service,
		Handler:           handler,
		RedisRepo:         cacheRepo,
		GORMRepo:          repo,
		MeilisearchEngine: searchEngine,
	}
}

func NewCategoryModule(
	db *gorm.DB,
	search meilisearch.ServiceManager,
	redis *redis.Client,
	creds domain.AuthService,
) (*CategoryModule, error) {
	keys, err := cache.NewKeyspace(cache.KeyConfig{
		Prefix: []string{"cms"},
	}, "categories")
	if err != nil {
		return nil, err
	}

	gormRepo := category.NewGORMRepository(db)
	flowRepo := category.NewFlowRepository(gormRepo, 20*time.Second, keys)
	redisRepo := category.NewRedisRepository(flowRepo, redis, 120*time.Minute, keys)
	searchEngine := category.NewMeilisearchEngine(search, CategoriesSearchIndexName)
	service := category.NewService(redisRepo, searchEngine, creds)
	handler := category.NewHandler(service)
	return &CategoryModule{
		Service:           service,
		Handler:           handler,
		CacheRepo:         redisRepo,
		DBRepo:            gormRepo,
		MeilisearchEngine: searchEngine,
	}, nil
}

func NewEditorModule(
	db *gorm.DB,
	search meilisearch.ServiceManager,
	redis *redis.Client,
	creds domain.AuthService,
) (*EditorModule, error) {
	keys, err := cache.NewKeyspace(cache.KeyConfig{
		Prefix: []string{"cms"},
	}, "categories")
	if err != nil {
		return nil, err
	}

	gormRepo := editor.NewGORMRepository(db)
	flowRepo := editor.NewFlowRepository(gormRepo, 20*time.Second, keys)
	redisRepo := editor.NewRedisRepository(flowRepo, redis, 120*time.Minute, keys)
	searchEngine := editor.NewMeilisearchEngine(search, EditorsSearchIndexName)
	service := editor.NewService(redisRepo, searchEngine, creds)
	handler := editor.NewHandler(service)
	return &EditorModule{
		Consumer:          &editor.EventConsumer{},
		Service:           service,
		Handler:           handler,
		CacheRepoRepo:     redisRepo,
		DBRepo:            gormRepo,
		DedupRepo:         flowRepo,
		MeilisearchEngine: searchEngine,
	}, nil
}

func NewAdsModule(db *gorm.DB, search meilisearch.ServiceManager, redis *redis.Client) *AdsModule {
	redisCfg := ads.RedisConfig{
		Client:    redis,
		TTL:       60 * time.Minute,
		KeyFormat: "cms:ads:%d",
	}

	repo := ads.NewGORMRepository(db)
	cacheRepo := ads.NewRedisRepository(repo, &redisCfg)
	searchEngine := ads.NewMeilisearchEngine(search, AdsSearchIndexName)
	service := ads.NewService(cacheRepo, searchEngine)
	handler := ads.NewHandler(service)
	return &AdsModule{
		Service:           service,
		Handler:           handler,
		RedisRepo:         cacheRepo,
		GORMRepo:          repo,
		MeilisearchEngine: searchEngine,
	}
}

func NewLocaleModule(ctx context.Context, db *gorm.DB, creds domain.AuthService) (*LocaleModule, error) {
	// Enable if redis or flow repo required
	// keys, err := cache.NewKeyspace(cache.KeyConfig{
	// 	Prefix: []string{"cms"},
	// }, "locales")
	// if err != nil {
	// 	return nil, err
	// }

	gormRepo := locale.NewGORMRepository(db)
	// Enable if request deduplication preffered
	// flowRepo := locale.NewFlowRepository(gormRepo, 20*time.Second, keys)
	// Enable if redis preffered over local caching
	// redisRepo := locale.NewRedisRepository(flowRepo, redis, 24*time.Hour, keys)
	localRepo := locale.NewLocalCacheRepository(ctx, gormRepo)
	service := locale.NewService(localRepo, creds)
	handler := locale.NewHandler(service)
	return &LocaleModule{
		Service:   service,
		Handler:   handler,
		CacheRepo: localRepo,
		DBRepo:    gormRepo,
	}, nil
}
