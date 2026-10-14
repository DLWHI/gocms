package post

import (
	"encoding/json"
	"gocms/internal/errs"
	"net/http"
	"strconv"
)

// ------------------------------------------------------------------
// POST /posts - Create post translation
// POST /posts/search - Get/search posts
// GET /posts/{id} - Get post by id
// PATCH /posts/{id} - Update post translation
// PATCH /posts/{id}/views - Override posts views
// DELETE /posts/{id} - Delete post with all translations
// DELETE /posts/{id}/{locale} - Delete post translation
//
// ------------------------------------------------------------------

type Handler struct {
	Create       http.Handler
	Search       http.Handler
	Get          http.Handler
	Update       http.Handler
	UpdateViews  http.Handler
	Delete       http.Handler
	DeleteLocale http.Handler
	postService  *Service
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
	hndl.DeleteLocale = http.HandlerFunc(hndl.deleteLocale)
	return &hndl
}

// POST /posts - Create post translation
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var post Doc

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	err := h.postService.Create(r.Context(), &post)
	if err != nil {
		errs.ServeMessage(w, "Post", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(&post)
}

// POST /posts/search - Get/search posts
func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	var filters SearchParams

	if err := json.NewDecoder(r.Body).Decode(&filters); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	posts, err := h.postService.Search(r.Context(), &filters)
	if err != nil {
		errs.ServeMessage(w, "Post", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&posts)
}

// GET /posts/{id} - Get post by id
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	post, err := h.postService.Get(r.Context(), uint(id))
	if err != nil {
		errs.ServeMessage(w, "Post", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(post)
}

// PATCH /posts/{id} - Update post translation
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	var post Doc

	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	post.PostID = uint(id)

	if _, err := h.postService.Edit(r.Context(), &post); err != nil {
		errs.ServeMessage(w, "Post", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// PATCH /posts/{id}/views - Override posts views
func (h *Handler) updateViews(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	var value uint

	if err := json.NewDecoder(r.Body).Decode(&value); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := h.postService.SetViews(r.Context(), uint(id), value); err != nil {
		errs.ServeMessage(w, "Post", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DELETE /posts/{id} - Delete post with all translations
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.postService.Delete(r.Context(), uint(id)); err != nil {
		errs.ServeMessage(w, "Post", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DELETE /posts/{id}/{locale} - Delete post translation
func (h *Handler) deleteLocale(w http.ResponseWriter, r *http.Request) {
	locale := r.PathValue("locale")
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.postService.DeleteTranslation(r.Context(), uint(id), locale); err != nil {
		errs.ServeMessage(w, "Post translation", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
