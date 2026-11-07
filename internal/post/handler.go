package post

import (
	"encoding/json"
	"fmt"
	"gocms/internal/domain"
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
	Create      http.Handler
	Search      http.Handler
	Get         http.Handler
	Update      http.Handler
	UpdateViews http.Handler
	Delete      http.Handler
	postService *Service
}

func NewHandler(s *Service) *Handler {
	var hndl Handler
	hndl.postService = s
	hndl.Create = http.HandlerFunc(hndl.create)
	hndl.Search = http.HandlerFunc(hndl.search)
	hndl.Get = http.HandlerFunc(hndl.get)
	hndl.Update = http.HandlerFunc(hndl.update)
	hndl.UpdateViews = http.HandlerFunc(hndl.updateViews)
	hndl.Delete = http.HandlerFunc(hndl.delete)
	return &hndl
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
	if err := h.postService.Create(r.Context(), post); err != nil {
		server.WriteServerError(w, "Post", err)
		return
	}

	res := map[string]any{
		"id":       post.PostID,
		"location": fmt.Sprintf("/cms/posts/%s/%d", post.Language, post.PostID),
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

	posts, err := h.postService.Search(r.Context(), &filters)
	if err != nil {
		server.WriteServerError(w, "Post", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&posts)
}

// GET /posts/{lang}/{id} - Get post by id
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		server.WritePathError(w, "Invalid ID")
		return
	}

	post, err := h.postService.Get(r.Context(), uint(id))
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

	if err := h.postService.Edit(r.Context(), uint(id), &post); err != nil {
		server.WriteServerError(w, "Post", err)
		return
	}

	w.WriteHeader(http.StatusOK)
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

	if err := h.postService.SetViews(r.Context(), uint(id), *payload.Views); err != nil {
		server.WriteServerError(w, "Post", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DELETE /posts/{id} - Delete post with all translations
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		server.WritePathError(w, "Invalid ID")
		return
	}

	if err := h.postService.Delete(r.Context(), uint(id)); err != nil {
		server.WriteServerError(w, "Post", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
