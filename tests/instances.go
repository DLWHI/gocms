package tests

// func MySQLDatabase(app *Container) *gorm.DB {
// 	dbCfg := database.DBConfig{
// 		User:     "test",
// 		Password: "",
// 		Host:     "localhost",
// 		Port:     "3306",
// 		Name:     "",
// 	}

// 	DB, err := database.NewMySQLConnection(&dbCfg)
// 	if err != nil {
// 		fmt.Printf("[WARNING] Failed to connect to MySQL database: %v", err)
// 		return nil
// 	}

// 	return DB
// }

// func RedisClient(app *Container) *redis.Client {
// 	rdb, err := cache.NewRedisConnection("localhost", "6379")
// 	if err != nil {
// 		fmt.Printf("[WARNING] Failed to connect to Redis: %v", err)
// 		return nil
// 	}

// 	_, err = rdb.Ping(app.Context).Result()
// 	if err != nil {
// 		fmt.Printf("[WARNING] Failed to connect to Redis: %v", err)
// 		return nil
// 	}

// 	return rdb
// }

// func MeilisearchClient(app *Container) meilisearch.ServiceManager {
// 	client := search.NewMeilisearchConnection("http://localhost", "7700", "")
// 	return client
// }

// func RedisSessionManager(app *Container) *user.RedisSessionStorage {
// 	return user.NewRedisSessionStorage(app.Redis, 24*time.Hour)
// }

// func DefaultCategoryHandler(app *Container) *category.Handler {
// 	redisCfg := category.RedisConfig{
// 		Client:    app.Redis,
// 		Ttl:       5 * time.Minute,
// 		KeyFormat: "cms:categories:%s:%s",
// 	}

// 	repo := category.NewGORMRepository(app.DB)
// 	cacheRepo := category.NewRedisRepository(repo, &redisCfg)
// 	searchEngine := category.NewMeilisearchEngine(app.Meilisearch, "categories_test")
// 	service := category.NewService(cacheRepo, searchEngine)
// 	handler := category.NewHandler(service)
// 	return handler
// }

// func DefaultUserHandler(app *Container) *user.Handler {
// 	redisCfg := user.RedisConfig{
// 		Client:    app.Redis,
// 		Ttl:       5 * time.Minute,
// 		KeyFormat: "cms:users:%d:%s",
// 	}

// 	repo := user.NewGORMRepository(app.DB)
// 	cacheRepo := user.NewRedisRepository(repo, &redisCfg)
// 	searchEngine := user.NewMeilisearchEngine(app.Meilisearch, "users_test")
// 	service := user.NewService(cacheRepo, searchEngine, app.RedisSessionManager)
// 	handler := user.NewHandler(service)
// 	return handler
// }

// func DefaultEditorHandler(app *Container) *editor.Handler {
// 	redisCfg := editor.RedisConfig{
// 		Client:    app.Redis,
// 		Ttl:       5 * time.Minute,
// 		KeyFormat: "cms:editors:%s",
// 	}

// 	repo := editor.NewGORMRepository(app.DB)
// 	cacheRepo := editor.NewRedisRepository(repo, &redisCfg)
// 	searchEngine := editor.NewMeilisearchEngine(app.Meilisearch, "editors_test")
// 	storage := storage.NewManager("./content/bios")
// 	service := editor.NewService(cacheRepo, searchEngine, storage)
// 	return editor.NewHandler(service)
// }

// func DefaultPostHandler(app *Container) *post.Handler {
// 	redisCfg := post.RedisConfig{
// 		Client:    app.Redis,
// 		Ttl:       5 * time.Minute,
// 		KeyFormat: "cms:posts:%d:%s",
// 	}

// 	repo := post.NewGORMRepository(app.DB)
// 	cacheRepo := post.NewRedisRepository(repo, &redisCfg)
// 	storage := storage.NewManager("./content/posts")
// 	searchEngine := post.NewMeilisearchEngine(app.Meilisearch, "posts_test")
// 	service := post.NewService(cacheRepo, searchEngine, storage)
// 	handler := post.NewHandler(service)
// 	return handler
// }

// func DefaultAdvertisementsHandler(app *Container) *ads.Handler {
// 	redisCfg := ads.RedisConfig{
// 		Client:    app.Redis,
// 		Ttl:       5 * time.Minute,
// 		KeyFormat: "cms:ads:%d",
// 	}

// 	repo := ads.NewGORMRepository(app.DB)
// 	cacheRepo := ads.NewRedisRepository(repo, &redisCfg)
// 	searchEngine := ads.NewMeilisearchEngine(app.Meilisearch, "ads_test")
// 	service := ads.NewService(cacheRepo, searchEngine)
// 	handler := ads.NewHandler(service)
// 	return handler
// }

// func DefaultLocalesHandler(app *Container) *locale.Handler {
// 	repo := locale.NewGORMRepository(app.DB)
// 	service := locale.NewService(repo)
// 	handler := locale.NewHandler(service)
// 	return handler
// }
