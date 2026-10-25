package locale

import (
	"encoding/json"
	"gocms/internal/server"
	"net/http"
)

// ------------------------------------------------------------------
// GET /locales - Get all locales
// POST /locales - Add locale
// DELETE /locales/{code} - Delete locale
// ------------------------------------------------------------------
type Handler struct {
	Get           http.Handler
	Create        http.Handler
	Delete        http.Handler
	localeService *Service
}

func NewHandler(s *Service) *Handler {
	var hndl Handler
	hndl.localeService = s
	hndl.Get = http.HandlerFunc(hndl.get)
	hndl.Create = http.HandlerFunc(hndl.create)
	hndl.Delete = http.HandlerFunc(hndl.delete)
	return &hndl
}

// POST /locales - Add locale
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var locale Locale
	if err := json.NewDecoder(r.Body).Decode(&locale); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := h.localeService.Create(r.Context(), &locale); err != nil {
		server.WriteServerError(w, "Locale", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GET /locales - Get all locales
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	locales, err := h.localeService.GetAll(r.Context())
	if err != nil {
		server.WriteServerError(w, "Locale", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(locales)
}

// DELETE /locales/{code} - Delete locale
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	if code == "" {
		http.Error(w, "Missing or invalid locale code", http.StatusBadRequest)
		return
	}

	err := h.localeService.Delete(r.Context(), code)
	if err != nil {
		server.WriteServerError(w, "Locale", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
