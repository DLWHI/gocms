package advertisement

import (
	"encoding/json"
	"fmt"
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
	Search    http.Handler
	Create    http.Handler
	Get       http.Handler
	Update    http.Handler
	Delete    http.Handler
	Groups    http.Handler
	adService *Service
}

func NewHandler(s *Service) *Handler {
	var hndl Handler
	hndl.adService = s
	hndl.Search = http.HandlerFunc(hndl.ads)
	hndl.Create = http.HandlerFunc(hndl.create)
	hndl.Get = http.HandlerFunc(hndl.get)
	hndl.Update = http.HandlerFunc(hndl.update)
	hndl.Delete = http.HandlerFunc(hndl.delete)
	hndl.Groups = http.HandlerFunc(hndl.groups)
	return &hndl
}

// POST /ads/search - Get/search ads
func (h *Handler) ads(w http.ResponseWriter, r *http.Request) {
	var filters SearchParams

	if err := json.NewDecoder(r.Body).Decode(&filters); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	ads, err := h.adService.Search(r.Context(), &filters)
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
	if err := h.adService.Create(r.Context(), ad); err != nil {
		server.WriteServerError(w, "Advertisement", err)
		return
	}

	res := map[string]any{
		"id":       ad.ID,
		"location": fmt.Sprintf("/cms/ads/%d", ad.ID),
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&res)
}

// GET /ads/{id} - Get ad
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 0)
	if err != nil {
		server.WritePathError(w, "Invalid ID")
		return
	}

	ads, err := h.adService.Get(r.Context(), uint(id))
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
	if err != nil {
		server.WritePathError(w, "Invalid ID")
		return
	}
	var ad domain.AdvertisementDiff
	if err := json.NewDecoder(r.Body).Decode(&ad); err != nil {
		server.WriteDecodeError(w, err)
		return
	}

	if err = h.adService.Update(r.Context(), uint(id), &ad); err != nil {
		server.WriteServerError(w, "Advertisement", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DELETE /ads/{id} - Delete ad
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 0)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.adService.Delete(r.Context(), uint(id)); err != nil {
		server.WriteServerError(w, "Advertisement", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GET /ads/groups - Get ad groups
func (h *Handler) groups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.adService.AllGroups(r.Context())
	if err != nil {
		server.WriteServerError(w, "Advertisement", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&groups)
}

// GET /ads/groups/{group} - Get ad group
func (h *Handler) getGroup(w http.ResponseWriter, r *http.Request) {
	group := r.PathValue("group")
	if group == "" {
		server.WritePathError(w, "Invalid ID")
		return
	}

	ads, err := h.adService.GetGroup(r.Context(), group)
	if err != nil {
		server.WriteServerError(w, "Advertisement", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&ads)
}
