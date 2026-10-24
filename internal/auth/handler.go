package auth

import (
	"encoding/json"
	"errors"
	"gocms/internal/domain"
	"gocms/internal/password"
	"gocms/internal/server"
	"net/http"
)

// ------------------------------------------------------------------
// POST /login - Login
// POST /logout - Log out
// GET /profile - Get current user/Validate session
// GET /permissions - Get all of current permissions
// PATCH /users/{username}/change - Update user password
// ------------------------------------------------------------------

type Handler struct {
	Login        http.Handler
	Logout       http.Handler
	Profile      http.Handler
	Permissions  http.Handler
	ChangePasswd http.Handler
	userService  *Service
}

func NewHandler(s *Service) *Handler {
	var hndl Handler
	hndl.userService = s
	hndl.Login = http.HandlerFunc(hndl.login)
	hndl.Logout = http.HandlerFunc(hndl.logout)
	hndl.Profile = http.HandlerFunc(hndl.profile)
	hndl.Permissions = http.HandlerFunc(hndl.perms)
	hndl.ChangePasswd = http.HandlerFunc(hndl.changePasswd)
	return &hndl
}

// POST /login - Login
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Username string            `json:"username"`
		Passwd   password.Password `json:"passwd"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	session, err := h.userService.Login(r.Context(), payload.Username, payload.Passwd)
	if err != nil {
		if errors.Is(err, domain.ErrCred) {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&session)
}

// POST /logout - Log out
func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if err := h.userService.Logout(r.Context()); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GET /profile - Get current user/Validate session
func (h *Handler) profile(w http.ResponseWriter, r *http.Request) {
	user, err := h.userService.CurrentUser(r.Context())
	if err != nil {
		server.ServeMessage(w, "User", err)
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

// PATCH /users/{username}/change - Update user password
func (h *Handler) changePasswd(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "Username was not provided", http.StatusBadRequest)
		return
	}

	var payload struct {
		OldPasswd password.Password `json:"old_password"`
		NewPaswd  password.Password `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := h.userService.ChangePasswd(r.Context(), username, payload.OldPasswd, payload.NewPaswd); err != nil {
		server.ServeMessage(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
