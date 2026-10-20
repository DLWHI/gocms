package ads

import (
	"encoding/json"
	"gocms/internal/errs"
	"net/http"
	"strconv"
)

// ------------------------------------------------------------------
// POST /ads/search - Get/search ads
// POST /ads - Create ad
// GET /ads/{id} - Get ad
// PATCH /ads/{id} - Update ad
// DELETE /ads/{id} - Delete ad
// GET /ads/groups - Get ad groups
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
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	ads, err := h.adService.Search(r.Context(), &filters)
	if err != nil {
		errs.ServeMessage(w, "Advertisement", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&ads)
}

// POST /ads/create - Create ad
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var ad Doc

	if err := json.NewDecoder(r.Body).Decode(&ad); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	err := h.adService.Create(r.Context(), &ad)
	if err != nil {
		errs.ServeMessage(w, "Advertisement", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GET /ads/{id} - Get ad
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 0)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	ads, err := h.adService.Get(r.Context(), uint(id))
	if err != nil {
		errs.ServeMessage(w, "Advertisement", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ads)
}

// PATCH /ads/{id} - Update ad
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 0)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	var ad Doc

	if err := json.NewDecoder(r.Body).Decode(&ad); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	ad.ID = uint(id)

	_, err = h.adService.Update(r.Context(), &ad)
	if err != nil {
		errs.ServeMessage(w, "Advertisement", err)
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
		errs.ServeMessage(w, "Advertisement", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GET /ads/groups - Get ad groups
func (h *Handler) groups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.adService.AllGroups(r.Context())
	if err != nil {
		errs.ServeMessage(w, "Advertisement", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&groups)
}

// GET /ads/groups/{group} - Get ad group
func (h *Handler) getGroup(w http.ResponseWriter, r *http.Request) {
	group := r.PathValue("group")
	if group == "" {
		http.Error(w, "Missing groups", http.StatusBadRequest)
		return
	}

	ads, err := h.adService.GetGroup(r.Context(), group)
	if err != nil {
		errs.ServeMessage(w, "Advertisement", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&ads)
}

// POST /ads/groups/search - Get multiple ad groups
func (h *Handler) getGroups(w http.ResponseWriter, r *http.Request) {
	var groups []string

	if err := json.NewDecoder(r.Body).Decode(&groups); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if (len(groups)) == 0 {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(&[]string{})
		return
	}

	advertisements := make(map[string][]Doc, len(groups))

	for _, group := range groups {
		ads, err := h.adService.GetGroup(r.Context(), group)
		if err != nil {
			errs.ServeMessage(w, "Advertisement", err)
			return
		}
		advertisements[group] = ads
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&advertisements)
}
