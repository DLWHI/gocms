package post

import (
	"encoding/json"
	"fmt"
	"gocms/internal/auth"
	"gocms/internal/domain"
	"gocms/internal/lang"
	"gocms/internal/server"
	"net/http"
	"strconv"
)

// ------------------------------------------------------------------
// GET /posts/{lang}/{id} - Get post by id
// POST /posts/{lang} - Create post translation
// POST /posts/{lang}/search - Get/search posts
// PATCH /posts/{lang}/{id} - Update post translation
// PUT /posts/views/{id} - Override posts views
// DELETE /posts/{id} - Delete post with all translations
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
	hndl.limiter = server.NewRequestLimiter(64 * 1024).Wrap // 64 KiB
	hndl.langs = lang.NewPathMiddleware(langs).Wrap
	hndl.auth = auth.NewSessionMiddleware(sessions).Wrap
	return &hndl
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mw := server.ChainMiddleware(
		h.limiter,
		h.auth,
		server.RequireJSON,
		h.langs,
	)
	create := http.HandlerFunc(h.create)
	search := http.HandlerFunc(h.search)
	get := http.HandlerFunc(h.get)
	update := http.HandlerFunc(h.update)
	updateViews := http.HandlerFunc(h.updateViews)
	delete := http.HandlerFunc(h.delete)

	mux.Handle("GET /posts/{lang}/{id}", mw(get))
	mux.Handle("POST /posts/{lang}", mw(create))
	mux.Handle("POST /posts/{lang}/search", mw(search))
	mux.Handle("PATCH /posts/{lang}/{id}", mw(update))
	mux.Handle("PUT /posts/views/{id}", mw(updateViews))
	mux.Handle("DELETE /posts/{id}", mw(delete))
}

// POST /posts/{lang} - Create post translation
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var payload domain.PostDiff

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	post := &domain.Post{}
	payload.Apply(post)
	if err := h.service.Create(r.Context(), post); err != nil {
		server.WriteServerError(w, "Post", err)
		return
	}

	res := map[string]any{
		"id":       post.PostID,
		"location": fmt.Sprintf("/cms/posts/%s/%d", post.Lang, post.PostID),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(&res)
}

// POST /posts/{lang}/search - Get/search posts
func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	var filters SearchParams

	if err := json.NewDecoder(r.Body).Decode(&filters); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	posts, err := h.service.Search(r.Context(), &filters)
	if err != nil {
		server.WriteServerError(w, "Post", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ToMetaResults(posts))
}

// GET /posts/{lang}/{id} - Get post by id
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		server.WritePathError(w, "Invalid ID")
		return
	}

	post, err := h.service.Get(r.Context(), uint(id))
	if err != nil {
		server.WriteServerError(w, "Post", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(post)
}

// PATCH /posts/{lang}/{id} - Update post translation
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		server.WritePathError(w, "Invalid ID")
		return
	}
	var post domain.PostDiff

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	if err := h.service.Edit(r.Context(), uint(id), &post); err != nil {
		server.WriteServerError(w, "Post", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PATCH /posts/{id}/views - Override posts views
func (h *Handler) updateViews(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		server.WritePathError(w, "Invalid ID")
		return
	}
	var payload struct {
		Views *uint `json:"views"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Views == nil {
		server.WriteDecodeError(w, err)
		return
	}

	if err := h.service.SetViews(r.Context(), uint(id), *payload.Views); err != nil {
		server.WriteServerError(w, "Post", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /posts/{id} - Delete post with all translations
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		server.WritePathError(w, "Invalid ID")
		return
	}

	if err := h.service.Delete(r.Context(), uint(id)); err != nil {
		server.WriteServerError(w, "Post", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
