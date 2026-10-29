package category

import (
	"encoding/json"
	"gocms/internal/domain"
	"gocms/internal/server"
	"net/http"
)

// ------------------------------------------------------------------
// POST /categories/search - Get/search categories
// GET /categories - Get all categories
// POST /categories - Create category
// GET /categories/{slug} - Get category
// PATCH /categories/{slug} - Update category
// DELETE /categories/{slug} - Delete category
// ------------------------------------------------------------------

type Handler struct {
	Search          http.Handler
	All             http.Handler
	Create          http.Handler
	Get             http.Handler
	Update          http.Handler
	Delete          http.Handler
	categoryService *Service
}

func NewHandler(s *Service) *Handler {
	var hndl Handler
	hndl.categoryService = s
	hndl.Search = http.HandlerFunc(hndl.search)
	hndl.All = http.HandlerFunc(hndl.all)
	hndl.Create = http.HandlerFunc(hndl.create)
	hndl.Get = http.HandlerFunc(hndl.get)
	hndl.Update = http.HandlerFunc(hndl.update)
	hndl.Delete = http.HandlerFunc(hndl.delete)
	return &hndl
}

// POST /categories/search - Get/search categories
func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	var filters SearchParams

	if err := json.NewDecoder(r.Body).Decode(&filters); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	categories, err := h.categoryService.Search(r.Context(), &filters)
	if err != nil {
		server.WriteServerError(w, "Category", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(categories)
}

// GET /categories - Get all categories
func (h *Handler) all(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categoryService.All(r.Context())
	if err != nil {
		server.WriteServerError(w, "Category", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(categories)
}

// POST /categories - Create category
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var category domain.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	if _, err := h.categoryService.Create(r.Context(), &category); err != nil {
		server.WriteServerError(w, "Category", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

// GET /categories/{slug} - Get get
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		server.WritePathError(w, "Missing slug")
		return
	}

	category, err := h.categoryService.Get(r.Context(), slug)
	if err != nil {
		server.WriteServerError(w, "Category", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(category)
}

// PATCH /categories/{slug} - Update category
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

	if _, err := h.categoryService.Edit(r.Context(), slug, &diff); err != nil {
		server.WriteServerError(w, "Category", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(diff)
}

// DELETE /categories/{slug} - Delete category
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		server.WritePathError(w, "Missing slug")
		return
	}

	err := h.categoryService.Delete(r.Context(), slug)
	if err != nil {
		server.WriteServerError(w, "Category", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
