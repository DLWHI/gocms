package user

import (
	"encoding/json"
	"gocms/internal/domain"
	"gocms/internal/password"
	"gocms/internal/server"
	"net/http"
)

// ------------------------------------------------------------------
// POST /users - Register new user
// POST /users/search - Search users
// GET /users - Get all users
// GET /users/{username} - Get user
// PATCH /users/{username}/email - Update user's email
// PATCH /users/{username}/password - Reset user password
// PATCH /users/{username}/permissions - Update user's permissions
// DELETE /users/{username} - Suspend user
// ------------------------------------------------------------------

type Handler struct {
	Create            http.Handler
	Search            http.Handler
	All               http.Handler
	Get               http.Handler
	ChangeEmail       http.Handler
	ResetPassword     http.Handler
	UpdatePermissions http.Handler
	Suspend           http.Handler
	userService       *Service
}

func NewHandler(s *Service) *Handler {
	var hndl Handler
	hndl.userService = s
	hndl.Create = http.HandlerFunc(hndl.create)
	hndl.Search = http.HandlerFunc(hndl.search)
	hndl.All = http.HandlerFunc(hndl.all)
	hndl.Get = http.HandlerFunc(hndl.get)
	hndl.Suspend = http.HandlerFunc(hndl.suspend)
	hndl.ChangeEmail = http.HandlerFunc(hndl.changeEmail)
	hndl.ResetPassword = http.HandlerFunc(hndl.resetPassword)
	hndl.UpdatePermissions = http.HandlerFunc(hndl.updatePermissions)
	return &hndl
}

// POST /users - Register new user
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		domain.User
		Password password.Password `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		clear(payload.Password)
		server.WriteDecodeError(w, err)
		return
	}
	defer clear(payload.Password)

	if _, err := h.userService.Create(r.Context(), &payload.User, payload.Password); err != nil {
		server.WriteServerError(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GET /users - Get all users
func (h *Handler) all(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.AllActive(r.Context())
	if err != nil {
		server.WriteServerError(w, "User", err)
		return
	}

	res := map[string][]domain.User{
		"users": users,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&res)
}

// POST /users/search - Search users
func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	var filters SearchParams

	if err := json.NewDecoder(r.Body).Decode(&filters); err != nil {
		server.WriteDecodeError(w, err)
		return
	}
	users, err := h.userService.Search(r.Context(), &filters)
	if err != nil {
		server.WriteServerError(w, "User", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&users)
}

// GET /users/{username} - Get user
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	if username == "" {
		server.WritePathError(w, "Invalid username provided")
		return
	}

	user, err := h.userService.Get(r.Context(), username)
	if err != nil {
		server.WriteServerError(w, "User", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

// PATCH /users/{username}/email - Update user's email
func (h *Handler) changeEmail(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		server.WritePathError(w, "Invalid username provided")
		return
	}

	var payload struct {
		Email *string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Email == nil {
		server.WriteDecodeError(w, err)
		return
	}

	if err := h.userService.ChangeEmail(r.Context(), username, *payload.Email); err != nil {
		server.WriteServerError(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PATCH /users/{username}/password - Reset user password
func (h *Handler) resetPassword(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		server.WritePathError(w, "Invalid username provided")
		return
	}

	var payload struct {
		Password password.Password `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Password == nil {
		clear(payload.Password)
		server.WriteDecodeError(w, err)
		return
	}
	defer clear(payload.Password)

	if err := h.userService.OverridePassword(r.Context(), username, payload.Password); err != nil {
		server.WriteServerError(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PATCH /users/{username}/permissions - Update user's permissions
func (h *Handler) updatePermissions(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		server.WritePathError(w, "Invalid username provided")
		return
	}

	var payload struct {
		Permissions *uint32 `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil || payload.Permissions == nil {
		server.WriteDecodeError(w, err)
		return
	}

	if err := h.userService.ChangePermissions(r.Context(), username, *payload.Permissions); err != nil {
		server.WriteServerError(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /users/{username} - Suspend user
func (h *Handler) suspend(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		server.WritePathError(w, "Invalid username provided")
		return
	}

	if err := h.userService.Suspend(r.Context(), username); err != nil {
		server.WriteServerError(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
