package category

import (
	"encoding/json"
	"fmt"
	"gocms/internal/domain"
	"gocms/internal/server"
	"net/http"
)

// ------------------------------------------------------------------
// GET /categories/{lang} - Get all categories
// GET /categories/{lang}/{slug} - Get category
// POST /categories/{lang}/search - Get/search categories
// POST /categories/{lang} - Create category
// PATCH /categories/{lang}/{slug} - Update category
// DELETE /categories/{slug} - Delete category
// ------------------------------------------------------------------

type Handler struct {
	Search  http.Handler
	All     http.Handler
	Create  http.Handler
	Get     http.Handler
	Update  http.Handler
	Delete  http.Handler
	service *Service
}

func NewHandler(service *Service) *Handler {
	var hndl Handler
	hndl.service = service
	hndl.Search = http.HandlerFunc(hndl.search)
	hndl.All = http.HandlerFunc(hndl.all)
	hndl.Create = http.HandlerFunc(hndl.create)
	hndl.Get = http.HandlerFunc(hndl.get)
	hndl.Update = http.HandlerFunc(hndl.update)
	hndl.Delete = http.HandlerFunc(hndl.delete)
	return &hndl
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

// GET /categories/{lang}/{slug} - Get get
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
