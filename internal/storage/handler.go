package storage

import (
	"encoding/json"
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
	Tree        http.Handler
	Get         http.Handler
	Upload      http.Handler
	Delete      http.Handler
	Folder      http.Handler
	Preview     http.Handler
	fileService *Manager
}

func NewHandler(s *Manager) *Handler {
	var hndl Handler
	hndl.fileService = s
	hndl.Tree = http.HandlerFunc(hndl.tree)
	hndl.Get = http.HandlerFunc(hndl.get)
	hndl.Upload = http.HandlerFunc(hndl.upload)
	hndl.Delete = http.HandlerFunc(hndl.delete)
	hndl.Folder = http.HandlerFunc(hndl.folder)
	hndl.Preview = http.HandlerFunc(hndl.preview)
	return &hndl
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

	response.Files, err = h.fileService.GetFiles(path)
	if err != nil {
		server.ServeMessage(w, "Path", err)
		return
	}

	response.Info, err = h.fileService.GetInfo()
	if err != nil {
		server.ServeMessage(w, "Storage", err)
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

	file_out, mime, err := h.fileService.Read(path)
	if err != nil {
		server.ServeMessage(w, "File", err)
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

	file, _, err := h.fileService.Read(path)
	if err != nil {
		server.ServeMessage(w, "File", err)
		return
	}
	defer file.Close()

	preview, err := GeneratePreview(file, width, height)
	if err != nil {
		server.ServeMessage(w, "File", err)
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

	err = h.fileService.Create(path)
	if err != nil {
		http.Error(w, "Path does not exist or file already created", http.StatusBadRequest)
		return
	}

	if err := h.fileService.Save(path, r.Body); err != nil {
		server.ServeMessage(w, "File", err)
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

	err = h.fileService.Delete(path, recursive)
	if err != nil {
		server.ServeMessage(w, "Path", err)
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

	err = h.fileService.CreateDir(path)
	if err != nil {
		server.ServeMessage(w, "Folder", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
