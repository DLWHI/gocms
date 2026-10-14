package user

import (
	"encoding/json"
	"errors"
	"gocms/internal/errs"
	"gocms/internal/permissions"
	"net/http"
)

// ------------------------------------------------------------------
// POST /login - Login
// POST /register - Register new user
// POST /logout - Log out
// GET /profile - Get current user/Validate session
// GET /permissions - Get all of current permissions
// POST /users/search - Search users
// GET /users?suspended={boolean} - Get all users
// GET /users/{username} - Get user
// PATCH /users/{username}/email - Update user's email
// PATCH /users/{username}/change - Update user password
// PATCH /users/{username}/reset - Reset user password
// PATCH /users/{username}/permissions - Update user's permissions
// DELETE /users/{username} - Suspend user
// ------------------------------------------------------------------

type Handler struct {
	Login             http.Handler
	Register          http.Handler
	Logout            http.Handler
	Profile           http.Handler
	Permissions       http.Handler
	Search            http.Handler
	All               http.Handler
	Get               http.Handler
	ChangeEmail       http.Handler
	ChangePasswd      http.Handler
	ResetPasswd       http.Handler
	UpdatePermissions http.Handler
	Suspend           http.Handler
	userService       *Service
}

func NewHandler(s *Service) *Handler {
	var hndl Handler
	hndl.userService = s
	hndl.Login = http.HandlerFunc(hndl.login)
	hndl.Register = http.HandlerFunc(hndl.register)
	hndl.Logout = http.HandlerFunc(hndl.logout)
	hndl.Profile = http.HandlerFunc(hndl.profile)
	hndl.Search = http.HandlerFunc(hndl.search)
	hndl.All = http.HandlerFunc(hndl.all)
	hndl.Get = http.HandlerFunc(hndl.get)
	hndl.Permissions = http.HandlerFunc(hndl.perms)
	hndl.Suspend = http.HandlerFunc(hndl.suspend)
	hndl.ChangeEmail = http.HandlerFunc(hndl.changeEmail)
	hndl.ChangePasswd = http.HandlerFunc(hndl.changePasswd)
	hndl.ResetPasswd = http.HandlerFunc(hndl.resetPasswd)
	hndl.UpdatePermissions = http.HandlerFunc(hndl.updatePermissions)
	return &hndl
}

// POST /login - Login
func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Username string `json:"username"`
		Passwd   Bytes  `json:"passwd"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	session, err := h.userService.Login(r.Context(), payload.Username, payload.Passwd)
	if err != nil {
		if errors.Is(err, errs.ErrCred) {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&session)
}

// POST /register - Register new user
func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var userData Model

	if err := json.NewDecoder(r.Body).Decode(&userData); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := h.userService.Register(r.Context(), &userData); err != nil {
		errs.ServeMessage(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
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
		errs.ServeMessage(w, "User", err)
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
		Entities: permissions.GetEntities(),
		Actions:  permissions.GetActions(),
	}

	_ = json.NewEncoder(w).Encode(&permissionData)
}

// GET /users - Get all users
func (h *Handler) all(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.All(r.Context())
	if err != nil {
		errs.ServeMessage(w, "User", err)
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
		errs.ServeMessage(w, "User", err)
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
		errs.ServeMessage(w, "User", err)
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
		errs.ServeMessage(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// PATCH /users/{username}/change - Update user password
func (h *Handler) changePasswd(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "Username was not provided", http.StatusBadRequest)
		return
	}

	var payload struct {
		OldPasswd Bytes `json:"old_password"`
		NewPaswd  Bytes `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := h.userService.ChangePasswd(r.Context(), username, payload.OldPasswd, payload.NewPaswd); err != nil {
		errs.ServeMessage(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// PATCH /users/{username}/reset - Update user password
func (h *Handler) resetPasswd(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "Username was not provided", http.StatusBadRequest)
		return
	}

	var paswd Bytes

	if err := json.NewDecoder(r.Body).Decode(&paswd); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := h.userService.ResetPasswd(r.Context(), username, paswd); err != nil {
		errs.ServeMessage(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// PATCH /users/{username}/permissions - Update user's permissions
func (h *Handler) updatePermissions(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "Username was not provided", http.StatusInternalServerError)
		return
	}

	var permissions uint32
	if err := json.NewDecoder(r.Body).Decode(&permissions); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := h.userService.ChangePermissions(r.Context(), username, permissions); err != nil {
		errs.ServeMessage(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DELETE /users/{username} - Suspend user
func (h *Handler) suspend(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "Unknown user", http.StatusBadRequest)
		return
	}

	if err := h.userService.Suspend(r.Context(), username); err != nil {
		errs.ServeMessage(w, "User", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
