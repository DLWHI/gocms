package public

import (
	"encoding/json"
	"gocms/internal/advertisement"
	"gocms/internal/analytics"
	"gocms/internal/category"
	"gocms/internal/domain"
	"gocms/internal/editor"
	"gocms/internal/lang"
	"gocms/internal/post"
	"gocms/internal/server"
	"net/http"
	"strconv"
	"strings"
)

// ------------------------------------------------------------------
// GET /languages - Get supported languages
//
// POST /posts - Get/search published posts
// GET /posts/{id} - Get published post
//
// GET /categories - Get all not hidden categories
//
// GET /editors - Get all editors
// GET /editors/{username} - Get editor
//
// GET /advertisements - Get ad groups
// GET /partners - Get partners group
// ------------------------------------------------------------------

type Handler struct {
	PostSearch    http.Handler
	PostGet       http.Handler
	CategoriesAll http.Handler
	LanguagesAll  http.Handler
	EditorAll     http.Handler
	EditorGet     http.Handler
	EditorBio     http.Handler
	Ads           http.Handler
	Partners      http.Handler
	posts         *post.Service
	editors       *editor.Service
	categories    *category.Service
	languages     lang.Service
	ads           *advertisement.Service
	views         analytics.PostViewCounterService
}

func NewHandler(ps *post.Service, es *editor.Service, cs *category.Service, ls lang.Service, as *advertisement.Service, vc analytics.PostViewCounterService) *Handler {
	var hndl Handler
	hndl.posts = ps
	hndl.editors = es
	hndl.categories = cs
	hndl.languages = ls
	hndl.ads = as
	hndl.views = vc
	hndl.PostSearch = http.HandlerFunc(hndl.postsSearch)
	hndl.PostGet = http.HandlerFunc(hndl.postsGet)
	hndl.CategoriesAll = http.HandlerFunc(hndl.categoriesAll)
	hndl.LanguagesAll = http.HandlerFunc(hndl.languagesAll)
	hndl.EditorAll = http.HandlerFunc(hndl.editorsAll)
	hndl.EditorGet = http.HandlerFunc(hndl.editorsGet)
	hndl.Ads = http.HandlerFunc(hndl.adsGet)
	hndl.Partners = http.HandlerFunc(hndl.partners)
	return &hndl
}

// POST /posts - Get/search published posts
func (h *Handler) postsSearch(w http.ResponseWriter, r *http.Request) {
	var params post.SearchParams

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	trueVal := true
	params.Published = &trueVal
	params.CreatedAfter = nil
	params.CreatedBefore = nil
	params.Select = []string{
		"post_id",
		"title",
		"description",
		"slug",
		"language",
		"published_at",
		"views",
		"preview",
		"author",
		"categories",
	}

	res, err := h.posts.Search(r.Context(), &params)
	if err != nil {
		server.WriteServerError(w, "Post", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&res)
}

// GET /posts/{id} - Get published post by id
func (h *Handler) postsGet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		server.WritePathError(w, "Invalid ID")
		return
	}

	ctx := r.Context()

	post, err := h.posts.GetPublished(ctx, uint(id))
	if err != nil {
		server.WriteServerError(w, "Post", err)
		return
	}

	post.Views, err = h.views.Get(ctx, uint(id))
	if err != nil {
		server.WriteServerError(w, "Post", err)
		return
	}

	if err := h.views.Increment(ctx, uint(id)); err != nil {
		server.WriteServerError(w, "Post", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&post)
}

// GET /categories - Get all not hidden categories
func (h *Handler) categoriesAll(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categories.AllNotHidden(r.Context())
	if err != nil {
		server.WriteServerError(w, "Categories", err)
		return
	}

	res := map[string][]domain.Category{
		"categories": categories,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// GET /languages - Get supported languages
func (h *Handler) languagesAll(w http.ResponseWriter, r *http.Request) {
	res := map[string][]lang.Language{
		"languages": h.languages.GetSupported(),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// GET /editors - Get all editors
func (h *Handler) editorsAll(w http.ResponseWriter, r *http.Request) {
	editors, err := h.editors.All(r.Context())
	if err != nil {
		server.WriteServerError(w, "Editor", err)
		return
	}

	res := map[string][]domain.Editor{
		"editors": editors,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

// GET /editors/{username} - Get editor
func (h *Handler) editorsGet(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "Unknown editor", http.StatusBadRequest)
		return
	}

	editor, err := h.editors.Get(r.Context(), username)
	if err != nil {
		server.WriteServerError(w, "Editor", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&editor)
}

// GET /advertisements - Get ad groups
func (h *Handler) adsGet(w http.ResponseWriter, r *http.Request) {
	var groups []string

	values := r.URL.Query()["groups"]
	if len(values) == 0 {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(&map[string][]domain.Advertisement{})
		return
	}
	for _, val := range values {
		for _, part := range strings.Split(val, ",") {
			groups = append(groups, part)
		}
	}

	advertisements := make(map[string][]domain.Advertisement, len(groups))

	for _, group := range groups {
		ads, err := h.ads.GetGroup(r.Context(), group)
		if err != nil {
			server.WriteServerError(w, "Advertisement", err)
			return
		}
		advertisements[group] = ads
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&advertisements)
}

// GET /partners - Get partners group
func (h *Handler) partners(w http.ResponseWriter, r *http.Request) {
	ads, err := h.ads.GetGroup(r.Context(), "partners")
	if err != nil {
		server.WriteServerError(w, "Advertisement", err)
		return
	}

	res := map[string][]domain.Advertisement{
		"partners": ads,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&ads)
}
