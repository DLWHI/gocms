package storage

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"gocms/internal/errs"
	"gocms/internal/logger"
)

// ------------------------------------------------------------------
// GET /storage/app/tree/{path} - Get storage
// GET /storage/app/media/{path} - Get file on specific path
// GET /storage/app/preview/{path}?w={width}&h={height} - Get preview of image on specific path
// POST /storage/app/media/{path} - Upload a file
// DELETE /storage/app/media/{path}?recursive={true|false} - Delete a file/folder
// POST /storage/app/folders - Create a folder
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

// GET /storage/app/tree/{path} - Get storage
func (h *Handler) tree(w http.ResponseWriter, r *http.Request) {
	path, err := url.PathUnescape(r.PathValue("path"))
	if err != nil {
		http.Error(w, "Invalid path provided", http.StatusBadRequest)
	}

	var response struct {
		Files []Item `json:"files"`
		Info  Info   `json:"info"`
	}

	response.Files, err = h.fileService.GetFiles(path)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			http.Error(w, "Path does not exist", http.StatusNotFound)
		} else {
			logger.Error("File.Handler", "Failed to retrieve storage capacity information", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	response.Info, err = h.fileService.GetInfo()
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			http.Error(w, "Path does not exist", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(&response)
}

// GET /storage/app/media/{path...} - Get storage
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	path, err := url.PathUnescape(r.PathValue("path"))
	if err != nil {
		http.Error(w, "Invalid path provided", http.StatusBadRequest)
	}

	file_out, err := h.fileService.Read(path)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, file_out)
}

// GET /storage/app/preview/{path}?w={width}&h={height} - Get preview of image on specific path
func (h *Handler) preview(w http.ResponseWriter, r *http.Request) {
	path, err := url.PathUnescape(r.PathValue("path"))
	if err != nil {
		http.Error(w, "Invalid path provided", http.StatusBadRequest)
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

	file, err := h.fileService.Read(path)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	preview, err := GeneratePreview(file, width, height)
	if err != nil {
		if errors.Is(err, errs.ErrInval) {
			http.Error(w, "File is not a supported image", http.StatusBadRequest)
		} else {
			logger.Error("File.Handler", "Failed to generate image preview", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "image/webp")
	w.WriteHeader(http.StatusOK)
	w.Write(preview)
}

// POST /storage/app/media/{path} - Upload a file
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
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DELETE /storage/app/media/{path}?recursive={true|false} - Delete a file/folder
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	path, err := url.PathUnescape(r.PathValue("path"))
	if err != nil {
		http.Error(w, "Invalid path provided", http.StatusBadRequest)
	}

	recursive := false
	if r.URL.Query().Get("recursive") == "true" {
		recursive = true
	} else if r.URL.Query().Get("recursive") != "false" {
		http.Error(w, "Invalid parameter value: recursive", http.StatusBadRequest)
	}

	err = h.fileService.Delete(path, recursive)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "Path does not exist", http.StatusNotFound)
		} else {
			http.Error(w, "Path blocked or directory is not empty", http.StatusBadRequest)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}

// POST /storage/app/folders - Create a folder
func (h *Handler) folder(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Missing 'path' query param", http.StatusBadRequest)
		return
	}

	err := h.fileService.CreateDir(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "Path does not exist", http.StatusNotFound)
		} else {
			http.Error(w, "Path blocked or directory already exists", http.StatusBadRequest)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}
