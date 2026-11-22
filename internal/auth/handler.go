package auth

import (
	"encoding/json"
	"gocms/internal/domain"
	"gocms/internal/password"
	"gocms/internal/server"
	"net/http"
)

// ------------------------------------------------------------------
// GET /profile - Get current user/Validate session
// GET /permissions - Get all of current permissions
// POST /login - Login
// POST /logout - Log out
// PATCH /passwd - Update user password
// ------------------------------------------------------------------

type Handler struct {
	sessions *SessionService
	auth     server.Middleware
	limiter  server.Middleware
}

func NewHandler(svc *SessionService, sessions SessionRepository) *Handler {
	var hndl Handler
	hndl.sessions = svc
	hndl.limiter = server.NewRequestLimiter(4 * 1024).Wrap // 4 KiB
	hndl.auth = NewSessionMiddleware(sessions).Wrap

	return &hndl
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mw := server.ChainMiddleware(
		h.limiter,
		h.auth,
		server.RequireJSON,
	)
	profile := http.HandlerFunc(h.profile)
	permissions := http.HandlerFunc(h.perms)
	login := http.HandlerFunc(h.login)
	logout := http.HandlerFunc(h.logout)
	changePasswd := http.HandlerFunc(h.changePasswd)

	mux.Handle("GET /profile", mw(profile))
	mux.Handle("GET /permissions", mw(permissions))
	mux.Handle("POST /login", h.limiter(server.RequireJSON(login)))
	mux.Handle("POST /logout", mw(logout))
	mux.Handle("PATCH /passwd", mw(changePasswd))
}

// GET /profile - Get current user/Validate session
func (h *Handler) profile(w http.ResponseWriter, r *http.Request) {
	user, err := h.sessions.CurrentUser(r.Context())
	if err != nil {
		server.WriteServerError(w, "User", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

// GET /permissions - Get all of current permissions
func (h *Handler) perms(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	permissionData := struct {
		Entities map[string]uint32 `json:"entities"`
		Actions  map[string]uint32 `json:"actions"`
	}{
		Entities: domain.GetEntities(),
		Actions:  domain.GetActions(),
	}

	_ = json.NewEncoder(w).Encode(&permissionData)
}

// POST /login - Login
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Username string            `json:"username"`
		Passwd   password.Password `json:"passwd"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	token, err := h.sessions.Login(r.Context(), payload.Username, payload.Passwd)
	if err != nil {
		server.WriteServerError(w, "Auth", err)
		return
	}

	res := map[string]string{
		"token": token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&res)
}

// POST /logout - Log out
func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if err := h.sessions.Logout(r.Context()); err != nil {
		server.WriteServerError(w, "Auth", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PATCH /passwd - Update user password
func (h *Handler) changePasswd(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		OldPasswd password.Password `json:"old_password"`
		NewPaswd  password.Password `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	if err := h.sessions.ChangePasswd(r.Context(), payload.OldPasswd, payload.NewPaswd); err != nil {
		server.WriteServerError(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
