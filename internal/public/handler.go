package public

import (
	"encoding/json"
	"gocms/internal/ads"
	"gocms/internal/analytics"
	"gocms/internal/category"
	"gocms/internal/editor"
	"gocms/internal/errs"
	"gocms/internal/locale"
	"gocms/internal/post"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// ------------------------------------------------------------------
// POST /posts - Get/search published posts
// GET /posts/{id}?slug={slug} - Get published post by id
//
// GET /categories - Get all categories
// GET /locales - Get all locales
//
// GET /editors - Get all editors
// GET /editors/{username} - Get editor
// GET /editors/{username}/bio - Read editor bio
//
// GET /advertisements - Get ad groups
// ------------------------------------------------------------------

type Handler struct {
	PostSearch    http.Handler
	PostGet       http.Handler
	CategoriesAll http.Handler
	LocalesAll    http.Handler
	EditorAll     http.Handler
	EditorGet     http.Handler
	EditorBio     http.Handler
	Ads           http.Handler
	posts         *post.Service
	editors       *editor.Service
	categories    *category.Service
	locales       *locale.Service
	ads           *ads.Service
	views         analytics.PostViewCounterService
}

func NewHandler(ps *post.Service, es *editor.Service, cs *category.Service, ls *locale.Service, as *ads.Service, vc analytics.PostViewCounterService) *Handler {
	var hndl Handler
	hndl.posts = ps
	hndl.editors = es
	hndl.categories = cs
	hndl.locales = ls
	hndl.ads = as
	hndl.views = vc
	hndl.PostSearch = http.HandlerFunc(hndl.postsSearch)
	hndl.PostGet = http.HandlerFunc(hndl.postsGet)
	hndl.CategoriesAll = http.HandlerFunc(hndl.categoriesAll)
	hndl.LocalesAll = http.HandlerFunc(hndl.localesAll)
	hndl.EditorAll = http.HandlerFunc(hndl.editorsAll)
	hndl.EditorGet = http.HandlerFunc(hndl.editorsGet)
	hndl.EditorBio = http.HandlerFunc(hndl.editorsBio)
	hndl.Ads = http.HandlerFunc(hndl.adsGet)
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

	res, err := h.posts.Search(r.Context(), &params)
	if err != nil {
		errs.ServeMessage(w, "Post", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&res)
}

// GET /posts/{id} - Get published post by id
func (h *Handler) postsGet(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		http.Error(w, "Invalid slug", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	ctx := r.Context()

	post, err := h.posts.Read(ctx, uint(id), slug)
	if err != nil {
		errs.ServeMessage(w, "Post", err)
		return
	}

	post.Views, err = h.views.Get(ctx, uint(id))
	if err != nil {
		errs.ServeMessage(w, "Post", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&post)
}

// GET /categories - Get all categories
func (h *Handler) categoriesAll(w http.ResponseWriter, r *http.Request) {
	categories, err := h.categories.AllNotHidden(r.Context())
	if err != nil {
		errs.ServeMessage(w, "Categories", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&categories)
}

// GET /locales - Get all locales
func (h *Handler) localesAll(w http.ResponseWriter, r *http.Request) {
	locales, err := h.locales.GetAll(r.Context())
	if err != nil {
		errs.ServeMessage(w, "Categories", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&locales)
}

// GET /editors - Get all editors
func (h *Handler) editorsAll(w http.ResponseWriter, r *http.Request) {
	res, err := h.editors.All(r.Context())
	if err != nil {
		errs.ServeMessage(w, "Editor", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&res)
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
		errs.ServeMessage(w, "Editor", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&editor)
}

// GET /posts/{id}/content - Read post content
func (h *Handler) editorsBio(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		http.Error(w, "Unknown editor", http.StatusBadRequest)
		return
	}

	content, err := h.editors.ReadBio(r.Context(), username)
	if err != nil {
		errs.ServeMessage(w, "Editor's bio", err)
		return
	}
	defer content.Close()

	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, content)
}

// GET /advertisements - Get ad groups
func (h *Handler) adsGet(w http.ResponseWriter, r *http.Request) {
	var groups []string

	values := r.URL.Query()["groups"]
	if len(values) == 0 {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(&map[string][]ads.Doc{})
		return
	}
	for _, val := range values {
		for _, part := range strings.Split(val, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				groups = append(groups, trimmed)
			}
		}
	}

	advertisements := make(map[string][]ads.Doc, len(groups))

	for _, group := range groups {
		ads, err := h.ads.GetGroupUnexpired(r.Context(), group)
		if err != nil {
			errs.ServeMessage(w, "Advertisement", err)
			return
		}
		advertisements[group] = ads
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&advertisements)
}
