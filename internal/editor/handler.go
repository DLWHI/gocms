package editor

import (
	"encoding/json"
	"gocms/internal/domain"
	"gocms/internal/server"
	"net/http"
)

// ------------------------------------------------------------------
// POST /editors/search - Get/search editors
// GET /editors - Get all editors
// POST /editors - Create editor
// GET /editors/{username} - Get editor
// PATCH /editors/{username} - Update editor
// DELETE /editors/{username} - Delete editor

// ------------------------------------------------------------------
type Handler struct {
	Search        http.Handler
	All           http.Handler
	Create        http.Handler
	Get           http.Handler
	Update        http.Handler
	Delete        http.Handler
	editorService *Service
}

func NewHandler(s *Service) *Handler {
	var hndl Handler
	hndl.editorService = s
	hndl.Search = http.HandlerFunc(hndl.search)
	hndl.All = http.HandlerFunc(hndl.all)
	hndl.Create = http.HandlerFunc(hndl.create)
	hndl.Get = http.HandlerFunc(hndl.get)
	hndl.Update = http.HandlerFunc(hndl.update)
	hndl.Delete = http.HandlerFunc(hndl.delete)
	return &hndl
}

// POST /editors/search - Get/search editors
func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	var filters SearchParams

	if err := json.NewDecoder(r.Body).Decode(&filters); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	filters.Select = []string{
		"name",
		"slug",
		"locale",
		"deleted_at",
	}

	editors, err := h.editorService.Search(r.Context(), &filters)
	if err != nil {
		server.WriteServerError(w, "Editor", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&editors)
}

// GET /editors - Get all editors
func (h *Handler) all(w http.ResponseWriter, r *http.Request) {
	editors, err := h.editorService.All(r.Context())
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

// POST /editors - Create editor
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var editor domain.Editor
	if err := json.NewDecoder(r.Body).Decode(&editor); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	if _, err := h.editorService.Create(r.Context(), &editor); err != nil {
		server.WriteServerError(w, "Editor", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GET /editors/{slug} - Get editor
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		server.WritePathError(w, "Missing editor slug")
		return
	}

	editor, err := h.editorService.Get(r.Context(), slug)
	if err != nil {
		server.WriteServerError(w, "Editor", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(editor)
}

// PATCH /editors/{slug} - Update editor
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

	if _, err := h.editorService.Edit(r.Context(), editor, &diff); err != nil {
		server.WriteServerError(w, "Editor", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /editors/{slug} - Delete editor user
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	editor := r.PathValue("username")
	if editor == "" {
		server.WritePathError(w, "Missing editor slug")
		return
	}

	if err := h.editorService.Delete(r.Context(), editor); err != nil {
		server.WriteServerError(w, "Editor", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
