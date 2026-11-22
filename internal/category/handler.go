package category

import (
	"encoding/json"
	"fmt"
	"gocms/internal/auth"
	"gocms/internal/domain"
	"gocms/internal/lang"
	"gocms/internal/server"
	"net/http"
)

// ------------------------------------------------------------------
// GET /categories/{lang} - Get all categories
// GET /categories/{lang}/{slug} - Get category
// POST /categories/{lang} - Create category
// POST /categories/{lang}/search - Get/search categories
// PATCH /categories/{lang}/{slug} - Update category
// DELETE /categories/{slug} - Delete category
// ------------------------------------------------------------------

type Handler struct {
	service *Service
	limiter server.Middleware
	langs   server.Middleware
	auth    server.Middleware
}

func NewHandler(service *Service, sessions auth.SessionRepository, langs lang.Service) *Handler {
	var hndl Handler
	hndl.service = service
	hndl.limiter = server.NewRequestLimiter(4 * 1024).Wrap // 4 KiB
	hndl.langs = lang.NewPathMiddleware(langs).Wrap
	hndl.auth = auth.NewSessionMiddleware(sessions).Wrap
	hndl.service = service

	return &hndl
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mw := server.ChainMiddleware(
		h.limiter,
		h.auth,
		server.RequireJSON,
		h.langs,
	)
	all := http.HandlerFunc(h.all)
	get := http.HandlerFunc(h.get)
	create := http.HandlerFunc(h.create)
	search := http.HandlerFunc(h.search)
	update := http.HandlerFunc(h.update)
	delete := http.HandlerFunc(h.delete)

	mux.Handle("GET /categories/{lang}", mw(all))
	mux.Handle("GET /categories/{lang}/{slug}", mw(get))
	mux.Handle("POST /categories/{lang}", mw(create))
	mux.Handle("POST /categories/{lang}/search", mw(search))
	mux.Handle("PATCH /categories/{lang}/{slug}", mw(update))
	mux.Handle("DELETE /categories/{slug}", mw(delete))
}

// GET /categories/{lang} - Get all categories
func (h *Handler) all(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.All(r.Context())
	if err != nil {
		server.WriteServerError(w, "Category", err)
		return
	}

	res := map[string][]domain.Category{
		"categories": categories,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// GET /categories/{lang}/{slug} - Get category
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		server.WritePathError(w, "Missing slug")
		return
	}

	category, err := h.service.Get(r.Context(), slug)
	if err != nil {
		server.WriteServerError(w, "Category", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(category)
}

// POST /categories/{lang} - Create category
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var payload domain.CategoryDiff
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	category := &domain.Category{}
	payload.Apply(category)
	if err := h.service.Create(r.Context(), category); err != nil {
		server.WriteServerError(w, "Category", err)
		return
	}

	res := map[string]any{
		"slug":     category.Slug,
		"location": fmt.Sprintf("/cms/categories/%s/%s", category.Lang, category.Slug),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(&res)
}

// POST /categories/{lang}/search - Get/search categories
func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	var filters SearchParams

	if err := json.NewDecoder(r.Body).Decode(&filters); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	categories, err := h.service.Search(r.Context(), &filters)
	if err != nil {
		server.WriteServerError(w, "Category", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(categories)
}

// PATCH /categories/{lang}/{slug} - Update category
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		server.WritePathError(w, "Missing slug")
		return
	}

	var diff domain.CategoryDiff
	if err := json.NewDecoder(r.Body).Decode(&diff); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	if _, err := h.service.Edit(r.Context(), slug, &diff); err != nil {
		server.WriteServerError(w, "Category", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /categories/{slug} - Delete category
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		server.WritePathError(w, "Missing slug")
		return
	}

	err := h.service.Delete(r.Context(), slug)
	if err != nil {
		server.WriteServerError(w, "Category", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
