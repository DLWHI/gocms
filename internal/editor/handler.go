package editor

import (
	"encoding/json"
	"gocms/internal/errs"
	"io"
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
// GET /editors/{username}/bio - Get editor bio
// PATCH /editors/{username}/bio - Update editor bio

// ------------------------------------------------------------------
type Handler struct {
	Search        http.Handler
	All           http.Handler
	Create        http.Handler
	Get           http.Handler
	Update        http.Handler
	Delete        http.Handler
	Bio           http.Handler
	EditBio       http.Handler
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
	hndl.Bio = http.HandlerFunc(hndl.bio)
	hndl.EditBio = http.HandlerFunc(hndl.editBio)
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
		errs.ServeMessage(w, "Editor", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&editors)
}

// GET /editors - Get all editors
func (h *Handler) all(w http.ResponseWriter, r *http.Request) {
	editors, err := h.editorService.All(r.Context())
	if err != nil {
		errs.ServeMessage(w, "Editor", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&editors)
}

// POST /editors - Create editor
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var editor Doc
	if err := json.NewDecoder(r.Body).Decode(&editor); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := h.editorService.Create(r.Context(), &editor); err != nil {
		errs.ServeMessage(w, "Editor", err)
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
		errs.ServeMessage(w, "Editor", err)
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

	var editor Doc
	if err := json.NewDecoder(r.Body).Decode(&editor); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if editor.User != username {
		http.Error(w, "User mismatch, are you ebanuti?", http.StatusBadRequest)
		return
	}

	if err := h.editorService.Edit(r.Context(), &editor); err != nil {
		errs.ServeMessage(w, "Editor", err)
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
		errs.ServeMessage(w, "Editor", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GET /users/{username}/bio - Get user's bio
func (h *Handler) bio(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "Username was not provided", http.StatusBadRequest)
		return
	}

	content, err := h.editorService.ReadBio(r.Context(), username)
	if err != nil {
		errs.ServeMessage(w, "Editor's bio", err)
		return
	}
	defer content.Close()

	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, content)
}

// PATCH /editors/{username}/bio - Update editor bio
func (h *Handler) editBio(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "Username was not provided", http.StatusInternalServerError)
		return
	}

	if err := h.editorService.EditBio(r.Context(), username, r.Body); err != nil {
		errs.ServeMessage(w, "Editor's bio", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
