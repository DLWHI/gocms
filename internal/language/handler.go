package language

import (
	"encoding/json"
	"gocms/internal/domain"
	"gocms/internal/server"
	"net/http"
)

// ------------------------------------------------------------------
// GET /languages - Get all languages
// POST /languages - Add languages
// DELETE /languages/{code} - Delete languages
// ------------------------------------------------------------------
type Handler struct {
	Get     http.Handler
	Create  http.Handler
	Delete  http.Handler
	service *Service
}

func NewHandler(s *Service) *Handler {
	var hndl Handler
	hndl.service = s
	hndl.Get = http.HandlerFunc(hndl.get)
	hndl.Create = http.HandlerFunc(hndl.create)
	hndl.Delete = http.HandlerFunc(hndl.delete)
	return &hndl
}

// POST /languages - Add language
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var payload domain.Language
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	if err := h.service.Create(r.Context(), &payload); err != nil {
		server.WriteServerError(w, "Language", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GET /languages - Get all languages
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	languages, err := h.service.GetAll(r.Context())
	if err != nil {
		server.WriteServerError(w, "Language", err)
		return
	}

	res := map[string][]domain.Language{
		"languages": languages,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// DELETE /languages/{code} - Delete language
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		server.WritePathError(w, "Empty language code")
		return
	}

	err := h.service.Delete(r.Context(), code)
	if err != nil {
		server.WriteServerError(w, "Language", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
