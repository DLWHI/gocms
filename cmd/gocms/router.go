package main

import (
	"gocms/internal/ads"
	"gocms/internal/auth"
	"gocms/internal/category"
	"gocms/internal/editor"
	"gocms/internal/locale"
	"gocms/internal/post"
	"gocms/internal/public"
	"gocms/internal/server"
	"gocms/internal/session"
	"gocms/internal/storage"
	"gocms/internal/user"
	"net/http"
)

const (
	RequestLimit = 4 * 1024          // 4 KB
	ContentLimit = 2 * 1024 * 1024   // 2 MB
	FileLimit    = 512 * 1024 * 1024 // 512 MB
)

type Router struct {
	mux  *http.ServeMux
	auth server.Middleware
}

func NewRouter(provider session.Repository) *Router {
	return &Router{
		mux:  http.NewServeMux(),
		auth: auth.NewSessionMiddleware(provider).Wrap,
	}
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.mux.ServeHTTP(w, r)
}

func (router *Router) AddAuthRoutes(handler *auth.Handler) {
	limiter := server.NewRequestLimiter(RequestLimit)
	m := server.ChainMiddleware(
		limiter.Wrap,
		router.auth,
	)

	router.mux.Handle("POST /cms/login", limiter.Wrap(handler.Login))
	router.mux.Handle("PATCH /cms/passwd", m(handler.ChangePasswd))
	router.mux.Handle("POST /cms/logout", m(handler.Logout))
	router.mux.Handle("GET /cms/profile", m(handler.Profile))
	router.mux.Handle("GET /cms/permissions", m(handler.Permissions))
}

func (router *Router) AddUserRoutes(handler *user.Handler) {
	limiter := server.NewRequestLimiter(RequestLimit)
	m := server.ChainMiddleware(
		limiter.Wrap,
		router.auth,
	)

	router.mux.Handle("POST /cms/users", m(handler.Create))
	router.mux.Handle("POST /cms/users/search", m(handler.Search))
	router.mux.Handle("GET /cms/users", m(handler.All))
	router.mux.Handle("GET /cms/users/{username}", m(handler.Get))
	router.mux.Handle("PATCH /cms/users/{username}/email", m(handler.ChangeEmail))
	router.mux.Handle("PATCH /cms/users/{username}/reset", m(handler.ResetPasswd))
	router.mux.Handle("PATCH /cms/users/{username}/permissions", m(handler.UpdatePermissions))
	router.mux.Handle("DELETE /cms/users/{username}", m(handler.Suspend))
}

func (router *Router) AddEditorRoutes(handler *editor.Handler) {
	streamLimiter := server.NewRequestLimiter(ContentLimit)
	m := server.ChainMiddleware(
		streamLimiter.Wrap,
		router.auth,
		locale.Middleware,
	)

	router.mux.Handle("POST /cms/editors/search", m(handler.Search))
	router.mux.Handle("GET /cms/editors", m(handler.All))
	router.mux.Handle("POST /cms/editors", m(handler.Create))
	router.mux.Handle("GET /cms/editors/{username}", m(handler.Get))
	router.mux.Handle("PATCH /cms/editors/{username}", m(handler.Update))
	router.mux.Handle("DELETE /cms/editors/{username}", m(handler.Delete))
}

func (router *Router) AddLocaleRoutes(handler *locale.Handler) {
	limiter := server.NewRequestLimiter(RequestLimit)
	m := server.ChainMiddleware(
		limiter.Wrap,
		router.auth,
	)

	router.mux.Handle("GET /cms/locales", m(handler.Get))
	router.mux.Handle("POST /cms/locales", m(handler.Create))
	router.mux.Handle("DELETE /cms/locales/{code}", m(handler.Delete))
}

func (router *Router) AddPostRoutes(handler *post.Handler) {
	streamLimiter := server.NewRequestLimiter(ContentLimit)
	m := server.ChainMiddleware(
		streamLimiter.Wrap,
		router.auth,
		locale.Middleware,
	)

	router.mux.Handle("POST /cms/posts", m(handler.Create))
	router.mux.Handle("POST /cms/posts/search", m(handler.Search))
	router.mux.Handle("GET /cms/posts/{id}", m(handler.Get))
	router.mux.Handle("PATCH /cms/posts/{id}", m(handler.Update))
	router.mux.Handle("PATCH /cms/posts/{id}/views", m(handler.UpdateViews))
	router.mux.Handle("DELETE /cms/posts/{id}", m(handler.Delete))
	router.mux.Handle("DELETE /cms/posts/{id}/{locale}", m(handler.DeleteLocale))
}

func (router *Router) AddCategoryRoutes(handler *category.Handler) {
	limiter := server.NewRequestLimiter(RequestLimit)
	m := server.ChainMiddleware(
		limiter.Wrap,
		router.auth,
		locale.Middleware,
	)

	router.mux.Handle("POST /cms/categories/search", m(handler.Search))
	router.mux.Handle("GET /cms/categories", m(handler.All))
	router.mux.Handle("POST /cms/categories", m(handler.Create))
	router.mux.Handle("GET /cms/categories/{slug}", m(handler.Get))
	router.mux.Handle("PATCH /cms/categories/{slug}", m(handler.Update))
	router.mux.Handle("DELETE /cms/categories/{slug}", m(handler.Delete))
}

func (router *Router) AddAdsRoutes(handler *ads.Handler) {
	limiter := server.NewRequestLimiter(RequestLimit)
	m := server.ChainMiddleware(
		limiter.Wrap,
		router.auth,
	)

	router.mux.Handle("POST /cms/ads/search", m(handler.Search))
	router.mux.Handle("POST /cms/ads", m(handler.Create))
	router.mux.Handle("GET /cms/ads/{id}", m(handler.Get))
	router.mux.Handle("PATCH /cms/ads/{id}", m(handler.Update))
	router.mux.Handle("DELETE /cms/ads/{id}", m(handler.Delete))
	router.mux.Handle("GET /cms/ads/groups", m(handler.Groups))
}

func (router *Router) AddStorageRoutes(handler *storage.Handler) {
	requestLimiter := server.NewRequestLimiter(RequestLimit)
	streamLimiter := server.NewRequestLimiter(FileLimit)
	m := server.ChainMiddleware(
		requestLimiter.Wrap,
		router.auth,
	)
	s := server.ChainMiddleware(
		streamLimiter.Wrap,
		router.auth,
	)

	router.mux.Handle("GET /storage/tree/{path...}", m(handler.Tree))
	router.mux.Handle("GET /storage/media/{path...}", requestLimiter.Wrap(handler.Get))
	router.mux.Handle("PUT /storage/media/{path...}", s(handler.Upload))
	router.mux.Handle("DELETE /storage/media/{path...}", m(handler.Delete))
	router.mux.Handle("PUT /storage/folders/{path...}", m(handler.Folder))
	router.mux.Handle("GET /storage/preview/{path...}", requestLimiter.Wrap(handler.Preview))
}

func (router *Router) AddPublicoutes(handler *public.Handler) {
	limiter := server.NewRequestLimiter(RequestLimit)
	m := server.ChainMiddleware(
		limiter.Wrap,
		locale.Middleware,
	)

	router.mux.Handle("GET /public/locales", limiter.Wrap(handler.LocalesAll))
	router.mux.Handle("POST /public/posts", m(handler.PostSearch))
	router.mux.Handle("GET /public/posts/{id}", m(handler.PostGet))
	router.mux.Handle("GET /public/categories", m(handler.CategoriesAll))
	router.mux.Handle("GET /public/editors", m(handler.EditorAll))
	router.mux.Handle("GET /public/editors/{username}", m(handler.EditorGet))
	router.mux.Handle("GET /public/advertisements", m(handler.Ads))
	router.mux.Handle("GET /cms/partners", m(handler.Partners))
}
