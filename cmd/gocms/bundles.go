package main

import (
	"gocms/internal/ads"
	"gocms/internal/category"
	"gocms/internal/editor"
	"gocms/internal/locale"
	"gocms/internal/post"
	"gocms/internal/user"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

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
	Service           *category.Service
	Handler           *category.Handler
	RedisRepo         *category.RedisRepository
	GORMRepo          *category.GORMRepository
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
	RedisRepo *locale.RedisRepository
	GORMRepo  *locale.GORMRepository
}

func NewUserBundle(db *gorm.DB, search meilisearch.ServiceManager, redis *redis.Client) *UserBundle {
	redisCfg := user.RedisConfig{
		Client:    redis,
		Ttl:       30 * time.Minute,
		KeyFormat: "cms:users:%s",
	}

	repo := user.NewGORMRepository(db)
	sfRepo := user.NewSingleflightRepository(repo, redisCfg.KeyFormat)
	redisRepo := user.NewRedisRepository(sfRepo, &redisCfg)
	searchEngine := user.NewMeilisearchEngine(search, "users")
	consumer := user.NewEventConsumer(repo, searchEngine)
	service := user.NewService(redisRepo, searchEngine)
	handler := user.NewHandler(service)
	return &UserBundle{
		Consumer:          consumer,
		Service:           service,
		Handler:           handler,
		CacheRepo:         redisRepo,
		DedupRepo:         sfRepo,
		DBRepo:            repo,
		MeilisearchEngine: searchEngine,
	}
}

func NewPostBundle(db *gorm.DB, search meilisearch.ServiceManager, redis *redis.Client) *PostBundle {
	redisCfg := post.RedisConfig{
		Client:    redis,
		Ttl:       30 * time.Minute,
		KeyFormat: "cms:posts:%d:%s",
	}

	repo := post.NewGORMRepository(db)
	cacheRepo := post.NewRedisRepository(repo, &redisCfg)
	searchEngine := post.NewMeilisearchEngine(search, "posts")
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

func NewCategoryBundle(db *gorm.DB, search meilisearch.ServiceManager, redis *redis.Client) *CategoryBundle {
	redisCfg := category.RedisConfig{
		Client:    redis,
		Ttl:       120 * time.Minute,
		KeyFormat: "cms:categories:%s:%s",
	}

	repo := category.NewGORMRepository(db)
	cacheRepo := category.NewRedisRepository(repo, &redisCfg)
	searchEngine := category.NewMeilisearchEngine(search, "categories")
	service := category.NewService(cacheRepo, searchEngine)
	handler := category.NewHandler(service)
	return &CategoryBundle{
		Service:           service,
		Handler:           handler,
		RedisRepo:         cacheRepo,
		GORMRepo:          repo,
		MeilisearchEngine: searchEngine,
	}
}

func NewEditorBundle(db *gorm.DB, search meilisearch.ServiceManager, redis *redis.Client) *EditorBundle {
	redisCfg := editor.RedisConfig{
		Client:    redis,
		Ttl:       60 * time.Minute,
		KeyFormat: "cms:editors:%s:%s",
	}

	repo := editor.NewGORMRepository(db)
	cacheRepo := editor.NewRedisRepository(repo, &redisCfg)
	searchEngine := editor.NewMeilisearchEngine(search, "editors")
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
		Ttl:       60 * time.Minute,
		KeyFormat: "cms:ads:%d",
	}

	repo := ads.NewGORMRepository(db)
	cacheRepo := ads.NewRedisRepository(repo, &redisCfg)
	searchEngine := ads.NewMeilisearchEngine(search, "ads")
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

func NewLocaleBundle(db *gorm.DB, redis *redis.Client) *LocaleBundle {
	redisCfg := locale.RedisConfig{
		Client:    redis,
		Ttl:       24 * time.Hour,
		KeyFormat: "cms:locales",
	}

	repo := locale.NewGORMRepository(db)
	cacheRepo := locale.NewRedisRepository(repo, &redisCfg)
	service := locale.NewService(cacheRepo)
	handler := locale.NewHandler(service)
	return &LocaleBundle{
		Service:   service,
		Handler:   handler,
		RedisRepo: cacheRepo,
		GORMRepo:  repo,
	}
}
