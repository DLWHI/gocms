package main

import (
	"gocms/internal/ads"
	"gocms/internal/auth"
	"gocms/internal/cache"
	"gocms/internal/category"
	"gocms/internal/editor"
	"gocms/internal/locale"
	"gocms/internal/password"
	"gocms/internal/post"
	"gocms/internal/session"
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

type AuthBundle struct {
	Consumer *auth.EventConsumer
	Service  *auth.Service
	Handler  *auth.Handler
}

type UserBundle struct {
	Consumer          *user.EventConsumer
	Service           *user.Service
	Handler           *user.Handler
	CacheRepo         user.Repository
	DedupRepo         user.Repository
	DBRepo            user.Repository
	MeilisearchEngine *user.MeilisearchEngine
}

type PostBundle struct {
	Service           *post.Service
	Handler           *post.Handler
	RedisRepo         *post.RedisRepository
	GORMRepo          *post.GORMRepository
	MeilisearchEngine *post.MeilisearchEngine
}

type EditorBundle struct {
	Service           *editor.Service
	Handler           *editor.Handler
	RedisRepo         *editor.RedisRepository
	GORMRepo          *editor.GORMRepository
	MeilisearchEngine *editor.MeilisearchEngine
}

type CategoryBundle struct {
	Consumer          *category.EventConsumer
	Service           *category.Service
	Handler           *category.Handler
	DedupRepo         category.Repository
	CacheRepo         category.Repository
	DBRepo            category.Repository
	MeilisearchEngine *category.MeilisearchEngine
}

type AdsBundle struct {
	Service           *ads.Service
	Handler           *ads.Handler
	RedisRepo         *ads.RedisRepository
	GORMRepo          *ads.GORMRepository
	MeilisearchEngine *ads.MeilisearchEngine
}

type LocaleBundle struct {
	Service   *locale.Service
	Handler   *locale.Handler
	CacheRepo locale.Repository
	DBRepo    locale.Repository
	DedupRepo locale.Repository
}

func NewAuthBundle(users auth.UserManager, sessions session.Repository) *AuthBundle {
	consumer := auth.NewEventConsumer(sessions, users)
	service := auth.NewService(users, sessions)
	handler := auth.NewHandler(service)
	return &AuthBundle{
		Consumer: consumer,
		Service:  service,
		Handler:  handler,
	}
}

func NewUserBundle(db *gorm.DB, search meilisearch.ServiceManager, redis *redis.Client, pwd password.PasswordService) (*UserBundle, error) {
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
	service := user.NewService(redisRepo, searchEngine, pwd)
	handler := user.NewHandler(service)
	return &UserBundle{
		Consumer:          consumer,
		Service:           service,
		Handler:           handler,
		CacheRepo:         redisRepo,
		DedupRepo:         flowRepo,
		DBRepo:            gormRepo,
		MeilisearchEngine: searchEngine,
	}, nil
}

func NewPostBundle(db *gorm.DB, search meilisearch.ServiceManager, redis *redis.Client) *PostBundle {
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
	return &PostBundle{
		Service:           service,
		Handler:           handler,
		RedisRepo:         cacheRepo,
		GORMRepo:          repo,
		MeilisearchEngine: searchEngine,
	}
}

func NewCategoryBundle(db *gorm.DB, search meilisearch.ServiceManager, redis *redis.Client) (*CategoryBundle, error) {
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
	service := category.NewService(redisRepo, searchEngine)
	handler := category.NewHandler(service)
	return &CategoryBundle{
		Service:           service,
		Handler:           handler,
		CacheRepo:         redisRepo,
		DBRepo:            gormRepo,
		MeilisearchEngine: searchEngine,
	}, nil
}

func NewEditorBundle(db *gorm.DB, search meilisearch.ServiceManager, redis *redis.Client) *EditorBundle {
	redisCfg := editor.RedisConfig{
		Client:    redis,
		TTL:       60 * time.Minute,
		KeyFormat: "cms:editors:%s:%s",
	}

	repo := editor.NewGORMRepository(db)
	cacheRepo := editor.NewRedisRepository(repo, &redisCfg)
	searchEngine := editor.NewMeilisearchEngine(search, EditorsSearchIndexName)
	service := editor.NewService(cacheRepo, searchEngine)
	handler := editor.NewHandler(service)
	return &EditorBundle{
		Service:           service,
		Handler:           handler,
		RedisRepo:         cacheRepo,
		GORMRepo:          repo,
		MeilisearchEngine: searchEngine,
	}
}

func NewAdsBundle(db *gorm.DB, search meilisearch.ServiceManager, redis *redis.Client) *AdsBundle {
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
	return &AdsBundle{
		Service:           service,
		Handler:           handler,
		RedisRepo:         cacheRepo,
		GORMRepo:          repo,
		MeilisearchEngine: searchEngine,
	}
}

func NewLocaleBundle(db *gorm.DB, redis *redis.Client) (*LocaleBundle, error) {
	keys, err := cache.NewKeyspace(cache.KeyConfig{
		Prefix: []string{"cms"},
	}, "locales")
	if err != nil {
		return nil, err
	}

	gormRepo := locale.NewGORMRepository(db)
	flowRepo := locale.NewFlowRepository(gormRepo, 20*time.Second, keys)
	redisRepo := locale.NewRedisRepository(flowRepo, redis, 24*time.Hour, keys)
	service := locale.NewService(redisRepo)
	handler := locale.NewHandler(service)
	return &LocaleBundle{
		Service:   service,
		Handler:   handler,
		CacheRepo: redisRepo,
		DBRepo:    gormRepo,
		DedupRepo: flowRepo,
	}, nil
}
