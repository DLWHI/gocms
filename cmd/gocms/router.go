package main

import (
	"gocms/internal/advertisement"
	"gocms/internal/auth"
	"gocms/internal/category"
	"gocms/internal/editor"
	"gocms/internal/lang"
	"gocms/internal/post"
	"gocms/internal/public"
	"gocms/internal/server"
	"gocms/internal/storage"
	"gocms/internal/user"
	"net/http"
)

const (
	RequestLimit = 4 * 1024          // 4 KiB
	ContentLimit = 64 * 1024         // 64 KiB
	FileLimit    = 512 * 1024 * 1024 // 512 MiB
)

type Router struct {
	mux  *http.ServeMux
	auth server.Middleware
	lang server.Middleware
}

func NewRouter(langs lang.Service, sessions auth.SessionRepository) *Router {
	return &Router{
		mux:  http.NewServeMux(),
		auth: auth.NewSessionMiddleware(sessions).Wrap,
		lang: lang.NewMiddleware(langs).Wrap,
	}
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	router.mux.ServeHTTP(w, r)
}

func (r *Router) AddAuthRoutes(handler *auth.Handler) {
	limiter := server.NewRequestLimiter(RequestLimit)
	m := server.ChainMiddleware(
		limiter.Wrap,
		r.auth,
		server.RequireJSON,
	)

	r.mux.Handle("POST /cms/login", limiter.Wrap(server.RequireJSON(handler.Login)))
	r.mux.Handle("PATCH /cms/passwd", m(handler.ChangePasswd))
	r.mux.Handle("POST /cms/logout", m(handler.Logout))
	r.mux.Handle("GET /cms/profile", m(handler.Profile))
	r.mux.Handle("GET /cms/permissions", m(handler.Permissions))
}

func (r *Router) AddUserRoutes(handler *user.Handler) {
	limiter := server.NewRequestLimiter(RequestLimit)
	m := server.ChainMiddleware(
		limiter.Wrap,
		r.auth,
		server.RequireJSON,
	)

	r.mux.Handle("POST /cms/users", m(handler.Create))
	r.mux.Handle("POST /cms/users/search", m(handler.Search))
	r.mux.Handle("GET /cms/users", m(handler.All))
	r.mux.Handle("GET /cms/users/{username}", m(handler.Get))
	r.mux.Handle("PATCH /cms/users/{username}/email", m(handler.ChangeEmail))
	r.mux.Handle("PATCH /cms/users/{username}/password", m(handler.ResetPassword))
	r.mux.Handle("PATCH /cms/users/{username}/permissions", m(handler.UpdatePermissions))
	r.mux.Handle("DELETE /cms/users/{username}", m(handler.Suspend))
}

func (r *Router) AddEditorRoutes(handler *editor.Handler) {
	streamLimiter := server.NewRequestLimiter(ContentLimit)
	m := server.ChainMiddleware(
		streamLimiter.Wrap,
		r.auth,
		server.RequireJSON,
		r.lang,
	)

	r.mux.Handle("GET /cms/editors/{lang}", m(handler.All))
	r.mux.Handle("GET /cms/editors/{lang}/{slug}", m(handler.Get))
	r.mux.Handle("POST /cms/editors/{lang}", m(handler.Create))
	r.mux.Handle("POST /cms/editors/{lang}/search", m(handler.Search))
	r.mux.Handle("PATCH /cms/editors/{lang}/{slug}", m(handler.Update))
	r.mux.Handle("DELETE /cms/editors/{slug}", m(handler.Delete))
}

func (r *Router) AddLangRoutes(handler *lang.Handler) {
	limiter := server.NewRequestLimiter(RequestLimit)
	m := server.ChainMiddleware(
		limiter.Wrap,
		r.auth,
		server.RequireJSON,
	)

	r.mux.Handle("GET /cms/languages", m(handler.Get))
	r.mux.Handle("POST /cms/languages", m(handler.Create))
	r.mux.Handle("DELETE /cms/languages/{code}", m(handler.Delete))
}

func (r *Router) AddPostRoutes(handler *post.Handler) {
	streamLimiter := server.NewRequestLimiter(ContentLimit)
	m := server.ChainMiddleware(
		streamLimiter.Wrap,
		r.auth,
		server.RequireJSON,
		r.lang,
	)

	r.mux.Handle("POST /cms/posts/{lang}", m(handler.Create))
	r.mux.Handle("POST /cms/posts/{lang}/search", m(handler.Search))
	r.mux.Handle("GET /cms/posts/{lang}/{id}", m(handler.Get))
	r.mux.Handle("PATCH /cms/posts/{lang}/{id}", m(handler.Update))
	r.mux.Handle("PUT /cms/posts/views/{id}", m(handler.UpdateViews))
	r.mux.Handle("DELETE /cms/posts/{id}", m(handler.Delete))
}

func (r *Router) AddCategoryRoutes(handler *category.Handler) {
	limiter := server.NewRequestLimiter(RequestLimit)
	m := server.ChainMiddleware(
		limiter.Wrap,
		r.auth,
		server.RequireJSON,
		r.lang,
	)

	r.mux.Handle("GET /cms/categories/{lang}", m(handler.All))
	r.mux.Handle("GET /cms/categories/{lang}/{slug}", m(handler.Get))
	r.mux.Handle("POST /cms/{lang}/categories", r.lang(m(handler.Create)))
	r.mux.Handle("POST /cms/categories/{lang}/search", m(handler.Search))
	r.mux.Handle("PATCH /cms/categories/{lang}/{slug}", m(handler.Update))
	r.mux.Handle("DELETE /cms/categories/{slug}", m(handler.Delete))
}

func (r *Router) AddAdsRoutes(handler *advertisement.Handler) {
	limiter := server.NewRequestLimiter(RequestLimit)
	m := server.ChainMiddleware(
		limiter.Wrap,
		r.auth,
		server.RequireJSON,
	)

	r.mux.Handle("GET /cms/ads/{id}", m(handler.Get))
	r.mux.Handle("GET /cms/ads/groups", m(handler.Groups))
	r.mux.Handle("POST /cms/ads", m(handler.Create))
	r.mux.Handle("POST /cms/ads/search", m(handler.Search))
	r.mux.Handle("PATCH /cms/ads/{id}", m(handler.Update))
	r.mux.Handle("DELETE /cms/ads/{id}", m(handler.Delete))
}

func (r *Router) AddStorageRoutes(handler *storage.Handler) {
	requestLimiter := server.NewRequestLimiter(RequestLimit)
	streamLimiter := server.NewRequestLimiter(FileLimit)
	m := server.ChainMiddleware(
		r.auth,
		requestLimiter.Wrap,
	)
	s := server.ChainMiddleware(
		r.auth,
		streamLimiter.Wrap,
	)

	r.mux.Handle("GET /storage/tree/{path...}", m(handler.Tree))
	r.mux.Handle("GET /storage/media/{path...}", requestLimiter.Wrap(handler.Get))
	r.mux.Handle("PUT /storage/media/{path...}", s(handler.Upload))
	r.mux.Handle("DELETE /storage/media/{path...}", m(handler.Delete))
	r.mux.Handle("PUT /storage/folders/{path...}", m(handler.Folder))
	r.mux.Handle("GET /storage/preview/{path...}", requestLimiter.Wrap(handler.Preview))
}

func (r *Router) AddPublicoutes(handler *public.Handler) {
	limiter := server.NewRequestLimiter(RequestLimit)
	m := server.ChainMiddleware(
		limiter.Wrap,
		server.RequireJSON,
		r.lang,
	)

	r.mux.Handle("GET /public/languages", limiter.Wrap(handler.LanguagesAll))
	r.mux.Handle("POST /public/posts", m(handler.PostSearch))
	r.mux.Handle("GET /public/posts/{id}", m(handler.PostGet))
	r.mux.Handle("GET /public/categories", m(handler.CategoriesAll))
	r.mux.Handle("GET /public/editors", m(handler.EditorAll))
	r.mux.Handle("GET /public/editors/{username}", m(handler.EditorGet))
	r.mux.Handle("GET /public/advertisements", m(handler.Ads))
	r.mux.Handle("GET /cms/partners", m(handler.Partners))
}
