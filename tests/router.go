package tests

// const (
// 	RequestLimit = 4 * 1024          // 4 KB
// 	ContentLimit = 2 * 1024 * 1024   // 2 MB
// 	FileLimit    = 512 * 1024 * 1024 // 512 MB
// )

// type Router struct {
// 	mux  *http.ServeMux
// 	auth middleware.Middleware
// }

// func NewRouter(provider user.SessionManager) *Router {
// 	return &Router{
// 		mux:  http.NewServeMux(),
// 		auth: middleware.NewSessionAuth(provider).Wrap,
// 	}
// }

// func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	router.mux.ServeHTTP(w, r)
// }

// func (router *Router) AddUserRoutes(handler *user.Handler) {
// 	limiter := middleware.NewRequestLimiter(RequestLimit)
// 	m := middleware.Chain(
// 		limiter.Wrap,
// 		router.auth,
// 	)

// 	router.mux.Handle("POST /cms/login", limiter.Wrap(handler.Login))
// 	router.mux.Handle("POST /cms/register", m(handler.Register))
// 	router.mux.Handle("POST /cms/logout", m(handler.Logout))
// 	router.mux.Handle("GET /cms/profile", m(handler.Profile))
// 	router.mux.Handle("GET /cms/permissions", m(handler.Permissions))
// 	router.mux.Handle("POST /cms/users/search", m(handler.Search))
// 	router.mux.Handle("GET /cms/users", m(handler.All))
// 	router.mux.Handle("GET /cms/users/{username}", m(handler.Get))
// 	router.mux.Handle("PATCH /cms/users/{username}/email", m(handler.ChangeEmail))
// 	router.mux.Handle("PATCH /cms/users/{username}/change", m(handler.ChangePasswd))
// 	router.mux.Handle("PATCH /cms/users/{username}/reset", m(handler.ResetPassword))
// 	router.mux.Handle("PATCH /cms/users/{username}/permissions", m(handler.UpdatePermissions))
// 	router.mux.Handle("DELETE /cms/users/{username}", m(handler.Suspend))
// }

// func (router *Router) AddEditorRoutes(handler *editor.Handler) {
// 	streamLimiter := middleware.NewRequestLimiter(ContentLimit)
// 	m := middleware.Chain(
// 		streamLimiter.Wrap,
// 		router.auth,
// 		middleware.Locale,
// 	)

// 	router.mux.Handle("POST /cms/editors/search", m(handler.Search))
// 	router.mux.Handle("GET /cms/editors", m(handler.All))
// 	router.mux.Handle("POST /cms/editors", m(handler.Create))
// 	router.mux.Handle("GET /cms/editors/{username}", m(handler.Get))
// 	router.mux.Handle("PATCH /cms/editors/{username}", m(handler.Update))
// 	router.mux.Handle("DELETE /cms/editors/{username}", m(handler.Delete))
// }

// func (router *Router) AddLocaleRoutes(handler *locale.Handler) {
// 	limiter := middleware.NewRequestLimiter(RequestLimit)
// 	m := middleware.Chain(
// 		limiter.Wrap,
// 		router.auth,
// 	)

// 	router.mux.Handle("GET /cms/locales", m(handler.Get))
// 	router.mux.Handle("POST /cms/locales", m(handler.Create))
// 	router.mux.Handle("DELETE /cms/locales/{code}", m(handler.Delete))
// }

// func (router *Router) AddPostRoutes(handler *post.Handler) {
// 	streamLimiter := middleware.NewRequestLimiter(ContentLimit)
// 	m := middleware.Chain(
// 		streamLimiter.Wrap,
// 		router.auth,
// 		middleware.Locale,
// 	)

// 	router.mux.Handle("POST /cms/posts", m(handler.Create))
// 	router.mux.Handle("POST /cms/posts/search", m(handler.Search))
// 	router.mux.Handle("GET /cms/posts/{id}", m(handler.Get))
// 	router.mux.Handle("PATCH /cms/posts/{id}", m(handler.Update))
// 	router.mux.Handle("PATCH /cms/posts/{id}/views", m(handler.UpdateViews))
// 	router.mux.Handle("DELETE /cms/posts/{id}", m(handler.Delete))
// 	router.mux.Handle("DELETE /cms/posts/{id}/{locale}", m(handler.DeleteLocale))
// }

// func (router *Router) AddCategoryRoutes(handler *category.Handler) {
// 	limiter := middleware.NewRequestLimiter(RequestLimit)
// 	m := middleware.Chain(
// 		limiter.Wrap,
// 		router.auth,
// 		middleware.Locale,
// 	)

// 	router.mux.Handle("POST /cms/categories/search", m(handler.Search))
// 	router.mux.Handle("GET /cms/categories", m(handler.All))
// 	router.mux.Handle("POST /cms/categories", m(handler.Create))
// 	router.mux.Handle("GET /cms/categories/{slug}", m(handler.Get))
// 	router.mux.Handle("PATCH /cms/categories/{slug}", m(handler.Update))
// 	router.mux.Handle("DELETE /cms/categories/{slug}", m(handler.Delete))
// }

// func (router *Router) AddAdsRoutes(handler *ads.Handler) {
// 	limiter := middleware.NewRequestLimiter(RequestLimit)
// 	m := middleware.Chain(
// 		limiter.Wrap,
// 		router.auth,
// 	)

// 	router.mux.Handle("POST /cms/ads/search", m(handler.Search))
// 	router.mux.Handle("POST /cms/ads", m(handler.Create))
// 	router.mux.Handle("GET /cms/ads/{id}", m(handler.Get))
// 	router.mux.Handle("PATCH /cms/ads/{id}", m(handler.Update))
// 	router.mux.Handle("DELETE /cms/ads/{id}", m(handler.Delete))
// 	router.mux.Handle("GET /cms/ads/groups", m(handler.Groups))
// 	router.mux.Handle("GET /cms/ads/groups/{group}", m(handler.GetGroup))
// 	router.mux.Handle("POST /cms/groups/search", m(handler.GetGroups))
// 	router.mux.Handle("GET /cms/partners", m(handler.Partners))
// }

// func (router *Router) AddStorageRoutes(handler *storage.Handler) {
// 	requestLimiter := middleware.NewRequestLimiter(RequestLimit)
// 	streamLimiter := middleware.NewRequestLimiter(FileLimit)
// 	m := middleware.Chain(
// 		requestLimiter.Wrap,
// 		router.auth,
// 	)
// 	s := middleware.Chain(
// 		streamLimiter.Wrap,
// 		router.auth,
// 	)

// 	router.mux.Handle("GET /storage/tree/{path...}", m(handler.Tree))
// 	router.mux.Handle("GET /storage/media/{path...}", requestLimiter.Wrap(handler.Get))
// 	router.mux.Handle("PUT /storage/media/{path...}", s(handler.Upload))
// 	router.mux.Handle("DELETE /storage/media/{path...}", m(handler.Delete))
// 	router.mux.Handle("PUT /storage/folders/{path...}", m(handler.Folder))
// 	router.mux.Handle("GET /storage/preview/{path...}", requestLimiter.Wrap(handler.Preview))
// }

// func (router *Router) AddPublicoutes(handler *public.Handler) {
// 	limiter := middleware.NewRequestLimiter(RequestLimit)
// 	m := middleware.Chain(
// 		limiter.Wrap,
// 		middleware.Locale,
// 	)

// 	router.mux.Handle("GET /public/locales", limiter.Wrap(handler.LocalesAll))
// 	router.mux.Handle("POST /public/posts", m(handler.PostSearch))
// 	router.mux.Handle("GET /public/posts/{id}", m(handler.PostGet))
// 	router.mux.Handle("GET /public/categories", m(handler.CategoriesAll))
// 	router.mux.Handle("GET /public/editors", m(handler.EditorAll))
// 	router.mux.Handle("GET /public/editors/{username}", m(handler.EditorGet))
// 	router.mux.Handle("GET /public/advertisements", m(handler.Ads))
// }
