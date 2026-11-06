package tests

import (
	"context"
	"gocms/internal/advertisement"
	"gocms/internal/auth"
	"gocms/internal/cache"
	"gocms/internal/category"
	"gocms/internal/editor"
	"gocms/internal/lang"
	"gocms/internal/password"
	"gocms/internal/post"
	"gocms/internal/user"
	"time"

	"github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	UsersSearchIndexName      = "test_users"
	PostsSearchIndexName      = "test_posts"
	EditorsSearchIndexName    = "test_editors"
	CategoriesSearchIndexName = "test_categories"
	AdsSearchIndexName        = "test_ads"
)

type AuthModule struct {
	Consumer    *auth.EventConsumer
	Service     *auth.SessionService
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
	Consumer          *post.EventConsumer
	Service           *post.Service
	Handler           *post.Handler
	DedupRepo         post.Repository
	CacheRepo         post.Repository
	DBRepo            post.Repository
	MeilisearchEngine post.SearchEngine
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
	Consumer          *advertisement.EventConsumer
	Service           *advertisement.Service
	Handler           *advertisement.Handler
	DedupRepo         advertisement.Repository
	CacheRepo         advertisement.Repository
	DBRepo            advertisement.Repository
	MeilisearchEngine advertisement.SearchEngine
}

type LangModule struct {
	Consumer *lang.LocalService
	Service  lang.Service
	Handler  *lang.Handler
	DBRepo   lang.Repository
}

func NewAuthModule(db *gorm.DB, redis *redis.Client, passwords password.Service) (*AuthModule, error) {
	keys, err := cache.NewKeyspace(cache.KeyConfig{
		Prefix: []string{"cms", "test"},
	}, "sessions")
	if err != nil {
		return nil, err
	}

	userRepo := user.NewGORMRepository(db)
	sessionRepo := auth.NewRedisSessionRepository(redis, keys, 5*time.Minute)
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
	creds auth.Service,
	passwords password.Service,
) (*UserModule, error) {
	keys, err := cache.NewKeyspace(cache.KeyConfig{
		Prefix: []string{"cms", "test"},
	}, "users")
	if err != nil {
		return nil, err
	}
	gormRepo := user.NewGORMRepository(db)
	flowRepo := user.NewFlowRepository(gormRepo, 20*time.Second, keys)
	redisRepo := user.NewRedisRepository(flowRepo, redis, 5*time.Minute, keys)
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

func NewPostModule(
	db *gorm.DB,
	search meilisearch.ServiceManager,
	redis *redis.Client,
	creds auth.Service,
) (*PostModule, error) {
	keys, err := cache.NewKeyspace(cache.KeyConfig{
		Prefix: []string{"cms"},
	}, "posts")
	if err != nil {
		return nil, err
	}

	gormRepo := post.NewGORMRepository(db)
	flowRepo := post.NewFlowRepository(gormRepo, 20*time.Second, keys)
	redisRepo := post.NewRedisRepository(flowRepo, redis, 30*time.Minute, keys)
	searchEngine := post.NewMeilisearchEngine(search, PostsSearchIndexName)
	consumer := post.NewEventConsumer(gormRepo, searchEngine)
	service := post.NewService(redisRepo, searchEngine, creds)
	handler := post.NewHandler(service)
	return &PostModule{
		Consumer:          consumer,
		Service:           service,
		Handler:           handler,
		CacheRepo:         redisRepo,
		DBRepo:            gormRepo,
		DedupRepo:         flowRepo,
		MeilisearchEngine: searchEngine,
	}, nil
}

func NewCategoryModule(
	db *gorm.DB,
	search meilisearch.ServiceManager,
	redis *redis.Client,
	creds auth.Service,
) (*CategoryModule, error) {
	keys, err := cache.NewKeyspace(cache.KeyConfig{
		Prefix: []string{"cms", "test"},
	}, "categories")
	if err != nil {
		return nil, err
	}

	gormRepo := category.NewGORMRepository(db)
	flowRepo := category.NewFlowRepository(gormRepo, 20*time.Second, keys)
	redisRepo := category.NewRedisRepository(flowRepo, redis, 5*time.Minute, keys)
	searchEngine := category.NewMeilisearchEngine(search, CategoriesSearchIndexName)
	consumer := category.NewEventConsumer(gormRepo, searchEngine)
	service := category.NewService(redisRepo, searchEngine, creds)
	handler := category.NewHandler(service)
	return &CategoryModule{
		Consumer:          consumer,
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
	creds auth.Service,
) (*EditorModule, error) {
	keys, err := cache.NewKeyspace(cache.KeyConfig{
		Prefix: []string{"cms", "test"},
	}, "editors")
	if err != nil {
		return nil, err
	}

	gormRepo := editor.NewGORMRepository(db)
	flowRepo := editor.NewFlowRepository(gormRepo, 20*time.Second, keys)
	redisRepo := editor.NewRedisRepository(flowRepo, redis, 5*time.Minute, keys)
	searchEngine := editor.NewMeilisearchEngine(search, EditorsSearchIndexName)
	consumer := editor.NewEventConsumer(gormRepo, searchEngine)
	service := editor.NewService(redisRepo, searchEngine, creds)
	handler := editor.NewHandler(service)
	return &EditorModule{
		Consumer:          consumer,
		Service:           service,
		Handler:           handler,
		CacheRepoRepo:     redisRepo,
		DBRepo:            gormRepo,
		DedupRepo:         flowRepo,
		MeilisearchEngine: searchEngine,
	}, nil
}

func NewAdsModule(
	db *gorm.DB,
	search meilisearch.ServiceManager,
	redis *redis.Client,
	creds auth.Service,
) (*AdsModule, error) {
	keys, err := cache.NewKeyspace(cache.KeyConfig{
		Prefix: []string{"cms", "test"},
	}, "ads")
	if err != nil {
		return nil, err
	}

	gormRepo := advertisement.NewGORMRepository(db)
	flowRepo := advertisement.NewFlowRepository(gormRepo, 20*time.Second, keys)
	redisRepo := advertisement.NewRedisRepository(flowRepo, redis, 12*time.Hour, keys)
	searchEngine := advertisement.NewMeilisearchEngine(search, AdsSearchIndexName)
	consume := advertisement.NewEventConsumer(gormRepo, searchEngine)
	service := advertisement.NewService(redisRepo, searchEngine)
	handler := advertisement.NewHandler(service)
	return &AdsModule{
		Consumer:          consume,
		Service:           service,
		Handler:           handler,
		CacheRepo:         redisRepo,
		DBRepo:            gormRepo,
		MeilisearchEngine: searchEngine,
		DedupRepo:         flowRepo,
	}, nil
}

func NewLangModule(ctx context.Context, db *gorm.DB, creds auth.Service) (*LangModule, error) {
	gormRepo := lang.NewGORMRepository(db)
	service, err := lang.NewLocalService(ctx, gormRepo, creds)
	if err != nil {
		return nil, err
	}
	handler := lang.NewHandler(service)
	return &LangModule{
		Consumer: service,
		Service:  service,
		Handler:  handler,
		DBRepo:   gormRepo,
	}, nil
}
