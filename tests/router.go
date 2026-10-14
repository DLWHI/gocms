package tests

import (
	"gocms/internal/ads"
	"gocms/internal/category"
	"gocms/internal/editor"
	"gocms/internal/locale"
	"gocms/internal/middleware"
	"gocms/internal/post"
	"gocms/internal/user"
	"net/http"
)

type Router struct {
	defaultMiddleware middleware.Middleware
}

func NewRouter() *Router {
	var router Router
	router.defaultMiddleware = middleware.Chain(
		middleware.Locale,
	)
	return &router
}

func (router *Router) AddUserRoutes(mux *http.ServeMux, handler *user.Handler) {
	m := router.defaultMiddleware

	mux.Handle("POST /cms/login", m(handler.Login))
	mux.Handle("POST /cms/register", m(handler.Register))
	mux.Handle("POST /cms/logout", m(handler.Logout))
	mux.Handle("GET /cms/profile", m(handler.Profile))
	mux.Handle("GET /cms/permissions", m(handler.Permissions))
	mux.Handle("GET /cms/users", m(handler.Search))
	mux.Handle("GET /cms/users/{username}", m(handler.Get))
	mux.Handle("PATCH /cms/users/{username}/change", m(handler.ChangePasswd))
	mux.Handle("PATCH /cms/users/{username}/reset", m(handler.ResetPasswd))
	mux.Handle("PATCH /cms/users/{username}/permissions", m(handler.UpdatePermissions))
	mux.Handle("DELETE /cms/users/{username}", m(handler.Suspend))
}

func (router *Router) AddEditorRoutes(mux *http.ServeMux, handler *editor.Handler) {
	m := router.defaultMiddleware

	mux.Handle("GET /cms/editors", m(handler.Search))
	mux.Handle("POST /cms/editors", m(handler.Create))
	mux.Handle("GET /cms/editors/{username}", m(handler.Get))
	mux.Handle("PATCH /cms/editors/{username}", m(handler.Update))
	mux.Handle("DELETE /cms/editors/{username}", m(handler.Delete))
	mux.Handle("GET /cms/editors/{username}/bio", m(handler.Bio))
	mux.Handle("PATCH /cms/editors/{username}/bio", m(handler.EditBio))
}

func (router *Router) AddLocaleRoutes(mux *http.ServeMux, handler *locale.Handler) {
	m := router.defaultMiddleware

	mux.Handle("GET /cms/locales", m(handler.Get))
	mux.Handle("POST /cms/locales", m(handler.Create))
	mux.Handle("DELETE /cms/locales", m(handler.Delete))
}

func (router *Router) AddPostRoutes(mux *http.ServeMux, handler *post.Handler) {
	m := router.defaultMiddleware

	mux.Handle("POST /cms/posts", m(handler.Create))
	mux.Handle("GET /cms/posts", m(handler.Search))
	mux.Handle("GET /cms/posts/{id}", m(handler.Get))
	mux.Handle("PATCH /cms/posts/{id}", m(handler.Update))
	mux.Handle("DELETE /cms/posts/{id}", m(handler.Delete))
	mux.Handle("DELETE /cms/posts/{id}/{locale}", m(handler.DeleteLocale))
	mux.Handle("GET /cms/posts/{id}/content", m(handler.GetContent))
	mux.Handle("PATCH /cms/posts/{id}/content", m(handler.UpdateContent))
}

func (router *Router) AddCategoryRoutes(mux *http.ServeMux, handler *category.Handler) {
	m := router.defaultMiddleware

	mux.Handle("GET /cms/categories", m(handler.Search))
	mux.Handle("POST /cms/categories", m(handler.Create))
	mux.Handle("GET /cms/categories/{slug}", m(handler.Get))
	mux.Handle("PATCH /cms/categories/{slug}", m(handler.Update))
	mux.Handle("DELETE /cms/categories/{slug}", m(handler.Delete))
}

func (router *Router) AddAdsRoutes(mux *http.ServeMux, handler *ads.Handler) {
	m := router.defaultMiddleware

	mux.Handle("GET /cms/ads", m(handler.Search))
	mux.Handle("POST /cms/ads", m(handler.Create))
	mux.Handle("GET /cms/ads/{id}", m(handler.Get))
	mux.Handle("PATCH /cms/ads/{id}", m(handler.Update))
	mux.Handle("DELETE /cms/ads/{id}", m(handler.Delete))
	mux.Handle("GET /cms/ads/groups/{group}", m(handler.GetGroup))
	mux.Handle("GET /cms/partners", m(handler.Partners))
}
