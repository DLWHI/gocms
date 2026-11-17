package advertisement

import (
	"encoding/json"
	"fmt"
	"gocms/internal/auth"
	"gocms/internal/domain"
	"gocms/internal/server"
	"net/http"
	"strconv"
)

// ------------------------------------------------------------------
// GET /ads/{id} - Get ad
// GET /ads/groups - Get ad groups
// POST /ads - Create ad
// POST /ads/search - Get/search ads
// PATCH /ads/{id} - Update ad
// DELETE /ads/{id} - Delete ad
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

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mw := server.ChainMiddleware(
		h.limiter,
		server.RequireJSON,
		h.auth,
	)
	search := http.HandlerFunc(h.ads)
	create := http.HandlerFunc(h.create)
	get := http.HandlerFunc(h.get)
	update := http.HandlerFunc(h.update)
	delete := http.HandlerFunc(h.delete)
	groups := http.HandlerFunc(h.groups)

	mux.Handle("GET /ads/{id}", mw(get))
	mux.Handle("GET /ads/groups", mw(groups))
	mux.Handle("POST /ads", mw(create))
	mux.Handle("POST /ads/search", mw(search))
	mux.Handle("PATCH /ads/{id}", mw(update))
	mux.Handle("DELETE /ads/{id}", mw(delete))
}

// POST /ads/search - Get/search ads
func (h *Handler) ads(w http.ResponseWriter, r *http.Request) {
	var filters SearchParams

	if err := json.NewDecoder(r.Body).Decode(&filters); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	ads, err := h.service.Search(r.Context(), &filters)
	if err != nil {
		server.WriteServerError(w, "Advertisement", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&ads)
}

// POST /ads/create - Create ad
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var payload domain.AdvertisementDiff
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	ad := &domain.Advertisement{}
	payload.Apply(ad)
	if err := h.service.Create(r.Context(), ad); err != nil {
		server.WriteServerError(w, "Advertisement", err)
		return
	}

	res := map[string]any{
		"id":       ad.ID,
		"location": fmt.Sprintf("/cms/ads/%d", ad.ID),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(&res)
}

// GET /ads/{id} - Get ad
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 0)
	if err != nil || id == 0 {
		server.WritePathError(w, "Invalid ID")
		return
	}

	ads, err := h.service.Get(r.Context(), uint(id))
	if err != nil {
		server.WriteServerError(w, "Advertisement", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ads)
}

// PATCH /ads/{id} - Update ad
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 0)
	if err != nil || id == 0 {
		server.WritePathError(w, "Invalid ID")
		return
	}
	var ad domain.AdvertisementDiff
	if err := json.NewDecoder(r.Body).Decode(&ad); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	if err = h.service.Update(r.Context(), uint(id), &ad); err != nil {
		server.WriteServerError(w, "Advertisement", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DELETE /ads/{id} - Delete ad
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 0)
	if err != nil || id == 0 {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), uint(id)); err != nil {
		server.WriteServerError(w, "Advertisement", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /ads/groups - Get ad groups
func (h *Handler) groups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.service.AllGroups(r.Context())
	if err != nil {
		server.WriteServerError(w, "Advertisement", err)
		return
	}

	res := map[string][]string{
		"groups": groups,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&res)
}

// GET /ads/groups/{group} - Get ad group
func (h *Handler) getGroup(w http.ResponseWriter, r *http.Request) {
	group := r.PathValue("group")
	if group == "" {
		server.WritePathError(w, "Invalid ID")
		return
	}

	ads, err := h.service.GetGroup(r.Context(), group)
	if err != nil {
		server.WriteServerError(w, "Advertisement", err)
		return
	}

	res := map[string][]domain.Advertisement{
		group: ads,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&res)
}
