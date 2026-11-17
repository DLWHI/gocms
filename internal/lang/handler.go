package lang

import (
	"encoding/json"
	"gocms/internal/auth"
	"gocms/internal/server"
	"net/http"
)

// ------------------------------------------------------------------
// GET /languages - Get all languages
// POST /languages - Add languages
// DELETE /languages/{code} - Delete languages
// ------------------------------------------------------------------
type Handler struct {
	service *LocalService
	limiter server.Middleware
	auth    server.Middleware
}

func NewHandler(service *LocalService, sessions auth.SessionRepository) *Handler {
	var hndl Handler
	hndl.service = service
	hndl.limiter = server.NewRequestLimiter(4 * 1024).Wrap // 4 KiB
	hndl.auth = auth.NewSessionMiddleware(sessions).Wrap
	return &hndl
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mw := server.ChainMiddleware(
		h.limiter,
		server.RequireJSON,
		h.auth,
	)
	get := http.HandlerFunc(h.get)
	create := http.HandlerFunc(h.create)
	delete := http.HandlerFunc(h.delete)

	mux.Handle("GET /languages", mw(get))
	mux.Handle("POST /languages", mw(create))
	mux.Handle("DELETE /languages/{code}", mw(delete))
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
