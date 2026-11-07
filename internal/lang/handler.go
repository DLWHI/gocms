package lang

import (
	"encoding/json"
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
	service *LocalService
}

func NewHandler(service *LocalService) *Handler {
	var hndl Handler
	hndl.service = service
	hndl.Get = http.HandlerFunc(hndl.get)
	hndl.Create = http.HandlerFunc(hndl.create)
	hndl.Delete = http.HandlerFunc(hndl.delete)
	return &hndl
}

// POST /languages - Add language
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var payload Language
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	if err := h.service.Add(r.Context(), &payload); err != nil {
		server.WriteServerError(w, "Language", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GET /languages - Get all languages
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	res := map[string][]Language{
		"languages": h.service.GetSupported(),
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

	err := h.service.Remove(r.Context(), code)
	if err != nil {
		server.WriteServerError(w, "Language", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
