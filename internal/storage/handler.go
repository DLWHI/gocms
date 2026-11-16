package storage

import (
	"encoding/json"
	"gocms/internal/auth"
	"gocms/internal/server"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// ------------------------------------------------------------------
// GET /storage/tree/{path...} - Get storage
// GET /storage/media/{path...} - Get file on specific path
// GET /storage/preview/{path...}?w={width}&h={height} - Get preview of image on specific path
// PUT /storage/media/{path...} - Upload a file
// DELETE /storage/media/{path...}?recursive={true|false} - Delete a file/folder
// PUT /storage/folders/{path...} - Create a folder
// ------------------------------------------------------------------
type Handler struct {
	service *Service
	limiter server.Middleware
	auth    server.Middleware
}

func NewHandler(service *Service, sessions auth.SessionRepository) *Handler {
	var hndl Handler
	hndl.service = service
	hndl.limiter = server.NewRequestLimiter(512 * 1024 * 1024).Wrap // 512 MiB
	hndl.auth = auth.NewSessionMiddleware(sessions).Wrap
	return &hndl
}

func (h *Handler) RegisterRoutes(r *http.ServeMux) {
	mw := server.ChainMiddleware(
		h.limiter,
		h.auth,
	)
	tree := http.HandlerFunc(h.tree)
	get := http.HandlerFunc(h.get)
	upload := http.HandlerFunc(h.upload)
	delete := http.HandlerFunc(h.delete)
	folder := http.HandlerFunc(h.folder)
	preview := http.HandlerFunc(h.preview)

	r.Handle("GET /tree/{path...}", mw(tree))
	r.Handle("GET /media/{path...}", h.limiter(get))
	r.Handle("GET /preview/{path...}", h.limiter(preview))
	r.Handle("PUT /media/{path...}", mw(upload))
	r.Handle("PUT /folders/{path...}", mw(folder))
	r.Handle("DELETE /media/{path...}", mw(delete))
}

// GET /storage/tree/{path...} - Get storage
func (h *Handler) tree(w http.ResponseWriter, r *http.Request) {
	path, err := url.PathUnescape(r.PathValue("path"))
	if err != nil {
		http.Error(w, "Invalid path provided", http.StatusBadRequest)
		return
	}

	var response struct {
		Files []Item `json:"files"`
		Info  Info   `json:"info"`
	}

	response.Files, err = h.service.GetFiles(path)
	if err != nil {
		server.WriteServerError(w, "Path", err)
		return
	}

	response.Info, err = h.service.GetInfo()
	if err != nil {
		server.WriteServerError(w, "Storage", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&response)
}

// GET /storage/media/{path...} - Get file on specific path
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	path, err := url.PathUnescape(r.PathValue("path"))
	if err != nil {
		http.Error(w, "Invalid path provided", http.StatusBadRequest)
		return
	}

	file_out, mime, err := h.service.Read(path)
	if err != nil {
		server.WriteServerError(w, "File", err)
		return
	}
	defer file_out.Close()

	w.Header().Add("Content-Type", mime)
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, file_out)
}

// GET /storage/preview/{path...}?w={width}&h={height} - Get preview of image on specific path
func (h *Handler) preview(w http.ResponseWriter, r *http.Request) {
	path, err := url.PathUnescape(r.PathValue("path"))
	if err != nil {
		http.Error(w, "Invalid path provided", http.StatusBadRequest)
		return
	}

	width := 360
	height := 360

	if r.URL.Query().Has("w") {
		width, err := strconv.Atoi(r.URL.Query().Get("w"))
		if err != nil || width < 0 {
			http.Error(w, "Invalid width value", http.StatusBadRequest)
			return
		}
	}
	if r.URL.Query().Has("h") {
		height, err := strconv.Atoi(r.URL.Query().Get("h"))
		if err != nil || height < 0 {
			http.Error(w, "Invalid width value", http.StatusBadRequest)
			return
		}
	}

	file, _, err := h.service.Read(path)
	if err != nil {
		server.WriteServerError(w, "File", err)
		return
	}
	defer file.Close()

	preview, err := GeneratePreview(file, width, height)
	if err != nil {
		server.WriteServerError(w, "File", err)
		return
	}

	w.Header().Set("Content-Type", "image/webp")
	w.WriteHeader(http.StatusOK)
	w.Write(preview)
}

// PUT /storage/media/{path...} - Upload a file
func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	path, err := url.PathUnescape(r.PathValue("path"))
	if err != nil {
		http.Error(w, "Invalid path provided", http.StatusBadRequest)
		return
	}

	err = h.service.Create(path)
	if err != nil {
		http.Error(w, "Path does not exist or file already created", http.StatusBadRequest)
		return
	}

	if err := h.service.Save(path, r.Body); err != nil {
		server.WriteServerError(w, "File", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DELETE /storage/media/{path...}?recursive={true|false} - Delete a file/folder
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	path, err := url.PathUnescape(r.PathValue("path"))
	if err != nil {
		http.Error(w, "Invalid path provided", http.StatusBadRequest)
		return
	}

	recursive := false
	if r.URL.Query().Get("recursive") == "true" {
		recursive = true
	} else if r.URL.Query().Get("recursive") != "false" {
		http.Error(w, "Invalid parameter value: recursive", http.StatusBadRequest)
		return
	}

	err = h.service.Delete(path, recursive)
	if err != nil {
		server.WriteServerError(w, "Path", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// PUT /storage/folders/{path...} - Create a folder
func (h *Handler) folder(w http.ResponseWriter, r *http.Request) {
	path, err := url.PathUnescape(r.PathValue("path"))
	if err != nil {
		http.Error(w, "Invalid path provided", http.StatusBadRequest)
		return
	}

	err = h.service.CreateDir(path)
	if err != nil {
		server.WriteServerError(w, "Folder", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
