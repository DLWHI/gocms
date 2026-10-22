package user

import (
	"encoding/json"
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
	ResetPasswd       http.Handler
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
	hndl.ResetPasswd = http.HandlerFunc(hndl.resetPasswd)
	hndl.UpdatePermissions = http.HandlerFunc(hndl.updatePermissions)
	return &hndl
}

// POST /users - Register new user
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var userData Record

	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := h.userService.Create(r.Context(), &userData); err != nil {
		server.ServeMessage(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GET /users - Get all users
func (h *Handler) all(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.AllActive(r.Context())
	if err != nil {
		server.ServeMessage(w, "User", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&users)
}

// POST /users/search - Search users
func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	var filters SearchParams

	if err := json.NewDecoder(r.Body).Decode(&filters); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	users, err := h.userService.Search(r.Context(), &filters)
	if err != nil {
		server.ServeMessage(w, "User", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&users)
}

// GET /users/{username} - Get user
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")

	if username == "" {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Get(r.Context(), username)
	if err != nil {
		server.ServeMessage(w, "User", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

// PATCH /users/{username}/email - Update user's email
func (h *Handler) changeEmail(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "Username was not provided", http.StatusBadRequest)
		return
	}

	var email string

	if err := json.NewDecoder(r.Body).Decode(&email); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := h.userService.ChangeEmail(r.Context(), username, email); err != nil {
		server.ServeMessage(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PATCH /users/{username}/reset - Reset user password
func (h *Handler) resetPasswd(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "Username was not provided", http.StatusBadRequest)
		return
	}

	var paswd Password

	if err := json.NewDecoder(r.Body).Decode(&paswd); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := h.userService.ChangePasswd(r.Context(), username, paswd); err != nil {
		server.ServeMessage(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PATCH /users/{username}/permissions - Update user's permissions
func (h *Handler) updatePermissions(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "Username was not provided", http.StatusBadRequest)
		return
	}

	var permissions uint32
	if err := json.NewDecoder(r.Body).Decode(&permissions); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := h.userService.ChangePermissions(r.Context(), username, permissions); err != nil {
		server.ServeMessage(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /users/{username} - Suspend user
func (h *Handler) suspend(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "Unknown user", http.StatusBadRequest)
		return
	}

	if err := h.userService.Suspend(r.Context(), username); err != nil {
		server.ServeMessage(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
