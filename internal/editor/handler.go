package editor

import (
	"encoding/json"
	"gocms/internal/server"
	"net/http"
)

// ------------------------------------------------------------------
// POST /editors/search - Get/search editors
// GET /editors - Get all editors
// POST /editors - Create editor
// GET /editors/{username} - Get editor
// PATCH /editors/{username} - Update editor
// PATCH /editors/{username}/permission - Update editors permissions
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
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	editors, err := h.editorService.Search(r.Context(), &filters)
	if err != nil {
		server.ServeMessage(w, "Editor", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&editors)
}

// GET /editors - Get all editors
func (h *Handler) all(w http.ResponseWriter, r *http.Request) {
	editors, err := h.editorService.All(r.Context(), nil)
	if err != nil {
		server.ServeMessage(w, "Editor", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&editors)
}

// POST /editors - Create editor
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var editor Document
	if err := json.NewDecoder(r.Body).Decode(&editor); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := h.editorService.Create(r.Context(), &editor); err != nil {
		server.ServeMessage(w, "Editor", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GET /editors/{username} - Get editor
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	if username == "" {
		http.Error(w, "Invalid user", http.StatusBadRequest)
	}

	editor, err := h.editorService.Get(r.Context(), username)
	if err != nil {
		server.ServeMessage(w, "Editor", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(editor)
}

// PATCH /editors/{username} - Update editor
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "Unknown user", http.StatusBadRequest)
		return
	}

	var editor Document
	if err := json.NewDecoder(r.Body).Decode(&editor); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if editor.User != username {
		http.Error(w, "User mismatch, are you ebanuti?", http.StatusBadRequest)
		return
	}

	if err := h.editorService.Edit(r.Context(), &editor); err != nil {
		server.ServeMessage(w, "Editor", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DELETE /editors/{username} - Delete editor user
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "Unknown user", http.StatusBadRequest)
		return
	}

	if err := h.editorService.Suspend(r.Context(), username); err != nil {
		server.ServeMessage(w, "Editor", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
