package user

import (
	"encoding/json"
	"fmt"
	"gocms/internal/auth"
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
	service *Service
	limiter server.Middleware
	auth    server.Middleware
}

func NewHandler(service *Service, sessions auth.SessionRepository) *Handler {
	var hndl Handler
	hndl.service = service
	hndl.limiter = server.NewRequestLimiter(4 * 1024).Wrap // 4 KiB
	hndl.auth = auth.NewSessionMiddleware(sessions).Wrap
	return &hndl
}

func (h *Handler) RegisterRoutes(r *http.ServeMux) {
	mw := server.ChainMiddleware(
		h.limiter,
		h.auth,
	)
	all := http.HandlerFunc(h.all)
	get := http.HandlerFunc(h.get)
	create := http.HandlerFunc(h.create)
	search := http.HandlerFunc(h.search)
	suspend := http.HandlerFunc(h.suspend)
	changeEmail := http.HandlerFunc(h.changeEmail)
	resetPassword := http.HandlerFunc(h.resetPassword)
	updatePermissions := http.HandlerFunc(h.updatePermissions)

	r.Handle("GET /users", mw(all))
	r.Handle("GET /users/{username}", mw(get))
	r.Handle("POST /users", mw(create))
	r.Handle("POST /users/search", mw(search))
	r.Handle("PATCH /users/{username}/email", mw(changeEmail))
	r.Handle("PATCH /users/{username}/password", mw(resetPassword))
	r.Handle("PATCH /users/{username}/permissions", mw(updatePermissions))
	r.Handle("DELETE /users/{username}", mw(suspend))
}

// POST /users - Register new user
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		domain.UserDiff
		Username string            `json:"username"`
		Password password.Password `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		clear(payload.Password)
		server.WriteDecodeError(w, err)
		return
	}
	defer clear(payload.Password)

	user := &domain.User{Username: payload.Username}
	payload.Apply(user)
	if err := h.service.Create(r.Context(), user, payload.Password); err != nil {
		server.WriteServerError(w, "User", err)
		return
	}

	res := map[string]any{
		"username": user.Username,
		"location": fmt.Sprintf("/cms/users/%s", user.Username),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(&res)
}

// GET /users - Get all users
func (h *Handler) all(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.AllActive(r.Context())
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
	users, err := h.service.Search(r.Context(), &filters)
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
		server.WritePathError(w, "Missing username")
		return
	}

	user, err := h.service.Get(r.Context(), username)
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
		server.WritePathError(w, "Missing username")
		return
	}

	var payload struct {
		Email *string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	if err := h.service.ChangeEmail(r.Context(), username, &domain.UserDiff{Email: payload.Email}); err != nil {
		server.WriteServerError(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PATCH /users/{username}/password - Reset user password
func (h *Handler) resetPassword(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		server.WritePathError(w, "Missing username")
		return
	}

	var payload struct {
		Password password.Password `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		clear(payload.Password)
		server.WriteDecodeError(w, err)
		return
	}
	defer clear(payload.Password)

	if err := h.service.OverridePassword(r.Context(), username, payload.Password); err != nil {
		server.WriteServerError(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PATCH /users/{username}/permissions - Update user's permissions
func (h *Handler) updatePermissions(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		server.WritePathError(w, "Missing username")
		return
	}

	var payload struct {
		Permissions *uint32 `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	if err := h.service.ChangePermissions(r.Context(), username, &domain.UserDiff{Permissions: payload.Permissions}); err != nil {
		server.WriteServerError(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /users/{username} - Suspend user
func (h *Handler) suspend(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		server.WritePathError(w, "Missing username")
		return
	}

	if err := h.service.Suspend(r.Context(), username); err != nil {
		server.WriteServerError(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
