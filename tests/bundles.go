package tests

// type UserBundle struct {
// 	Service           *user.Service
// 	Handler           *user.Handler
// 	RedisRepo         *user.RedisRepository
// 	GORMRepo          *user.GORMRepository
// 	MeilisearchEngine *user.MeilisearchEngine
// }

// type PostBundle struct {
// 	Service           *post.Service
// 	Handler           *post.Handler
// 	RedisRepo         *post.RedisRepository
// 	GORMRepo          *post.GORMRepository
// 	MeilisearchEngine *post.MeilisearchEngine
// }

// type EditorBundle struct {
// 	Service           *editor.Service
// 	Handler           *editor.Handler
// 	RedisRepo         *editor.RedisRepository
// 	GORMRepo          *editor.GORMRepository
// 	MeilisearchEngine *editor.MeilisearchEngine
// }

// type CategoryBundle struct {
// 	Service           *category.Service
// 	Handler           *category.Handler
// 	RedisRepo         *category.RedisRepository
// 	GORMRepo          *category.GORMRepository
// 	MeilisearchEngine *category.MeilisearchEngine
// }

// type AdsBundle struct {
// 	Service           *ads.Service
// 	Handler           *ads.Handler
// 	RedisRepo         *ads.RedisRepository
// 	GORMRepo          *ads.GORMRepository
// 	MeilisearchEngine *ads.MeilisearchEngine
// }

// type LocaleBundle struct {
// 	Service   *locale.Service
// 	Handler   *locale.Handler
// 	RedisRepo *locale.RedisRepository
// 	GORMRepo  *locale.GORMRepository
// }

// func NewUserBundle(db *gorm.DB, search meilisearch.ServiceManager, redis *redis.Client, sessions user.SessionManager) *UserBundle {
// 	redisCfg := user.RedisConfig{
// 		Client:    redis,
// 		TTL:       30 * time.Minute,
// 		KeyFormat: "test:cms:users:%s",
// 	}

// 	repo := user.NewGORMRepository(db)
// 	cacheRepo := user.NewRedisRepository(repo, &redisCfg)
// 	searchEngine := user.NewMeilisearchEngine(search, "test_users")
// 	service := user.NewService(cacheRepo, searchEngine, sessions)
// 	handler := user.NewHandler(service)
// 	return &UserBundle{
// 		Service:           service,
// 		Handler:           handler,
// 		RedisRepo:         cacheRepo,
// 		GORMRepo:          repo,
// 		MeilisearchEngine: searchEngine,
// 	}
// }

// func NewPostBundle(db *gorm.DB, search meilisearch.ServiceManager, redis *redis.Client) *PostBundle {
// 	redisCfg := post.RedisConfig{
// 		Client:    redis,
// 		TTL:       30 * time.Minute,
// 		KeyFormat: "test:cms:posts:%d:%s",
// 	}

// 	repo := post.NewGORMRepository(db)
// 	cacheRepo := post.NewRedisRepository(repo, &redisCfg)
// 	searchEngine := post.NewMeilisearchEngine(search, "test_posts")
// 	service := post.NewService(cacheRepo, searchEngine)
// 	handler := post.NewHandler(service)
// 	return &PostBundle{
// 		Service:           service,
// 		Handler:           handler,
// 		RedisRepo:         cacheRepo,
// 		GORMRepo:          repo,
// 		MeilisearchEngine: searchEngine,
// 	}
// }

// func NewCategoryBundle(db *gorm.DB, search meilisearch.ServiceManager, redis *redis.Client) *CategoryBundle {
// 	redisCfg := category.RedisConfig{
// 		Client:    redis,
// 		TTL:       120 * time.Minute,
// 		KeyFormat: "test:cms:categories:%s:%s",
// 	}

// 	repo := category.NewGORMRepository(db)
// 	cacheRepo := category.NewRedisRepository(repo, &redisCfg)
// 	searchEngine := category.NewMeilisearchEngine(search, "test_categories")
// 	service := category.NewService(cacheRepo, searchEngine)
// 	handler := category.NewHandler(service)
// 	return &CategoryBundle{
// 		Service:           service,
// 		Handler:           handler,
// 		RedisRepo:         cacheRepo,
// 		GORMRepo:          repo,
// 		MeilisearchEngine: searchEngine,
// 	}
// }

// func NewEditorBundle(db *gorm.DB, search meilisearch.ServiceManager, redis *redis.Client) *EditorBundle {
// 	redisCfg := editor.RedisConfig{
// 		Client:    redis,
// 		TTL:       60 * time.Minute,
// 		KeyFormat: "test:cms:editors:%s:%s",
// 	}

// 	repo := editor.NewGORMRepository(db)
// 	cacheRepo := editor.NewRedisRepository(repo, &redisCfg)
// 	searchEngine := editor.NewMeilisearchEngine(search, "test_editors")
// 	service := editor.NewService(cacheRepo, searchEngine)
// 	handler := editor.NewHandler(service)
// 	return &EditorBundle{
// 		Service:           service,
// 		Handler:           handler,
// 		RedisRepo:         cacheRepo,
// 		GORMRepo:          repo,
// 		MeilisearchEngine: searchEngine,
// 	}
// }

// func NewAdsBundle(db *gorm.DB, search meilisearch.ServiceManager, redis *redis.Client) *AdsBundle {
// 	redisCfg := ads.RedisConfig{
// 		Client:    redis,
// 		TTL:       60 * time.Minute,
// 		KeyFormat: "test:cms:ads:%d",
// 	}

// 	repo := ads.NewGORMRepository(db)
// 	cacheRepo := ads.NewRedisRepository(repo, &redisCfg)
// 	searchEngine := ads.NewMeilisearchEngine(search, "test_ads")
// 	service := ads.NewService(cacheRepo, searchEngine)
// 	handler := ads.NewHandler(service)
// 	return &AdsBundle{
// 		Service:           service,
// 		Handler:           handler,
// 		RedisRepo:         cacheRepo,
// 		GORMRepo:          repo,
// 		MeilisearchEngine: searchEngine,
// 	}
// }

// func NewLocaleBundle(db *gorm.DB, redis *redis.Client) *LocaleBundle {
// 	redisCfg := locale.RedisConfig{
// 		Client:    redis,
// 		TTL:       24 * time.Hour,
// 		KeyFormat: "test:cms:locales",
// 	}

// 	repo := locale.NewGORMRepository(db)
// 	cacheRepo := locale.NewRedisRepository(repo, &redisCfg)
// 	service := locale.NewService(cacheRepo)
// 	handler := locale.NewHandler(service)
// 	return &LocaleBundle{
// 		Service:   service,
// 		Handler:   handler,
// 		RedisRepo: cacheRepo,
// 		GORMRepo:  repo,
// 	}
// }
