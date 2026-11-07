package editor

import (
	"encoding/json"
	"fmt"
	"gocms/internal/domain"
	"gocms/internal/server"
	"net/http"
)

// ------------------------------------------------------------------
// GET /editors/{lang} - Get all editors
// GET /editors/{lang}/{username} - Get editor
// POST /editors/{lang} - Create editor
// POST /editors/{lang}/search - Get/search editors
// PATCH /editors/{lang}/{username} - Update editor
// DELETE /editors/{username} - Delete editor

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

// POST /editors/search/{lang} - Get/search editors
func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	var filters SearchParams

	if err := json.NewDecoder(r.Body).Decode(&filters); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	editors, err := h.service.Search(r.Context(), &filters)
	if err != nil {
		server.WriteServerError(w, "Editor", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&editors)
}

// GET /editors/{lang} - Get all editors
func (h *Handler) all(w http.ResponseWriter, r *http.Request) {
	editors, err := h.service.All(r.Context())
	if err != nil {
		server.WriteServerError(w, "Editor", err)
		return
	}

	res := map[string][]domain.Editor{
		"editors": editors,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&res)
}

// POST /editors/{lang} - Create editor
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var payload domain.EditorDiff
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	editor := &domain.Editor{}
	payload.Apply(editor)
	if err := h.service.Create(r.Context(), editor); err != nil {
		server.WriteServerError(w, "Editor", err)
		return
	}

	res := map[string]any{
		"slug":     editor.Slug,
		"location": fmt.Sprintf("/cms/editors/%s/%s", editor.Lang, editor.Slug),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(&res)
}

// GET /editors/{lang}/{slug} - Get editor
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		server.WritePathError(w, "Missing editor slug")
		return
	}

	editor, err := h.service.Get(r.Context(), slug)
	if err != nil {
		server.WriteServerError(w, "Editor", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(editor)
}

// PATCH /editors/{lang}/{slug} - Update editor
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	editor := r.PathValue("slug")
	if editor == "" {
		server.WritePathError(w, "Missing editor slug")
		return
	}

	var diff domain.EditorDiff
	if err := json.NewDecoder(r.Body).Decode(&diff); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	if _, err := h.service.Edit(r.Context(), editor, &diff); err != nil {
		server.WriteServerError(w, "Editor", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /editors/{slug} - Delete editor user
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	editor := r.PathValue("slug")
	if editor == "" {
		server.WritePathError(w, "Missing editor slug")
		return
	}

	if err := h.service.Delete(r.Context(), editor); err != nil {
		server.WriteServerError(w, "Editor", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
