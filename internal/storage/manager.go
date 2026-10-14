package storage

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gocms/internal/errs"
	"gocms/internal/logger"
)

type Manager struct {
	root   string
	wMutex sync.Mutex
}

func NewManager(rootPath string) *Manager {
	return &Manager{
		root: filepath.Clean(rootPath),
	}
}

func (s *Manager) safePath(path string) (string, string, error) {
	cleaned := filepath.Clean(path)

	fullPath := filepath.Join(s.root, cleaned)

	rel, err := filepath.Rel(s.root, fullPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", "", errs.ErrInval
	}
	return fullPath, cleaned, nil
}

func (s *Manager) writeTemp(dir string, content io.Reader) (string, error) {
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, content)
	if err != nil {
		return tmpFile.Name(), err
	}

	if err := tmpFile.Sync(); err != nil {
		return tmpFile.Name(), err
	}
	return tmpFile.Name(), nil
}

func (s *Manager) syncDir(dir string) error {
	dirFD, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer dirFD.Close()

	if err := dirFD.Sync(); err != nil {
		return err
	}
	return err
}

func (s *Manager) detectWithMIMEType(path string) (ItemType, error) {
	file, err := os.Open(filepath.Join(s.root, path))
	if err != nil {
		return TypeFile, err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return TypeFile, err
	}

	// Sniff magic bytes (e.g., JPEG, PNG, MP4 headers)
	mimeType := http.DetectContentType(buffer[:n])

	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return TypeImage, nil
	case strings.HasPrefix(mimeType, "video/"):
		return TypeVideo, nil
	default:
		return TypeFile, nil
	}
}

func (s *Manager) GetFiles(path string) ([]Item, error) {
	fullPath, rel, err := s.safePath(path)
	if err != nil {
		logger.Error("Storage.Manager", "Failed to list files.", err)
		return nil, err
	}

	files, err := os.ReadDir(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errs.ErrNotFound
		}
		logger.Error("Storage.Manager", "Failed to list files.", err)
		return nil, err
	}

	fileList := make([]Item, 0, len(files))

	for _, entry := range files {
		var file Item
		info, err := entry.Info()
		if err != nil {
			logger.Error("Storage.Manager", "Failed to list files.", err)
			return nil, err
		}

		file.Path = filepath.Join(rel, entry.Name())
		file.Size = info.Size()
		file.Date = info.ModTime()

		if info.IsDir() {
			file.Type = TypeFolder
		} else if !info.Mode().IsRegular() {
			file.Type = TypeFile
		} else {
			file.Type, err = s.detectWithMIMEType(file.Path)
			if err != nil {
				logger.Error("Storage.Manager", "Failed to list files.", err)
			}
		}

		fileList = append(fileList, file)
	}

	return fileList, nil
}

func (s *Manager) GetTree(path string) ([]Item, error) {
	fullPath, rel, err := s.safePath(path)
	if err != nil {
		logger.Error("Storage.Manager", "Failed to build the file tree.", err)
		return nil, err
	}

	var fileList []Item

	err = filepath.WalkDir(fullPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logger.Error("Storage.Manager", "Failed to build the file tree.", err)
			return err
		}

		if path == s.root {
			return nil
		}

		var file Item
		info, err := d.Info()
		if err != nil {
			logger.Error("Storage.Manager", "Failed to build the file tree.", err)
			return err
		}

		file.Path = filepath.Join(rel, d.Name())
		file.Size = info.Size()
		file.Date = info.ModTime()

		if info.IsDir() {
			file.Type = TypeFolder
		} else if !info.Mode().IsRegular() {
			file.Type = TypeFile
		} else {
			file.Type, err = s.detectWithMIMEType(file.Path)
			if err != nil {
				logger.Error("Storage.Manager", "Failed to build the file tree.", err)
			}
		}
		fileList = append(fileList, file)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return fileList, nil
}

func (s *Manager) Create(path string) error {
	fullPath, _, err := s.safePath(path)
	if err != nil {
		logger.Error("Storage.Manager", "Failed to create a file.", err)
		return err
	}

	dir := filepath.Dir(fullPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		logger.Error("Storage.Manager", "Failed to create a file.", err)
		return err
	}

	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0660)
	if err != nil {
		logger.Error("Storage.Manager", "Failed to create a file.", err)
		return err
	}
	file.Close()

	return nil
}

func (s *Manager) CreateDir(path string) error {
	fullPath, _, err := s.safePath(path)
	if err != nil {
		logger.Error("Storage.Manager", "Failed to create a directory.", err)
		return err
	}

	if err := os.Mkdir(fullPath, 0755); err != nil {
		logger.Error("Storage.Manager", "Failed to create a directory.", err)
		return err
	}

	return nil
}

func (s *Manager) Save(path string, content io.Reader) error {
	fullPath, _, err := s.safePath(path)
	if err != nil {
		logger.Error("Storage.Manager", "Failed to save a file.", err)
		return err
	}
	dir := filepath.Dir(fullPath)

	s.wMutex.Lock()
	defer s.wMutex.Unlock()

	tmp, err := s.writeTemp(dir, content)
	if err != nil {
		os.Remove(tmp)
		logger.Error("Storage.Manager", "Failed to write a temporary file.", err)
		return err
	}

	if err := os.Rename(tmp, fullPath); err != nil {
		os.Remove(tmp)
		logger.Error("Storage.Manager", "Failed to replace the stored file.", err)
		return err
	}

	if err := s.syncDir(dir); err != nil {
		logger.Error("Storage.Manager", "Failed to sync the storage directory.", err)
		return err
	}

	return nil
}

func (s *Manager) Read(path string) (io.ReadCloser, error) {
	fullPath, _, err := s.safePath(path)
	if err != nil {
		logger.Error("Storage.Manager", "Failed to read a file.", err)
		return nil, err
	}

	dst, err := os.Open(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errs.ErrNotFound
		}
		logger.Error("Storage.Manager", "Failed to read a file.", err)
		return nil, err
	}

	return dst, nil
}

func (s *Manager) TypeOf(path string) (ItemType, error) {
	fullPath, _, err := s.safePath(path)
	if err != nil {
		logger.Error("Storage.Manager", "Failed to determine file type.", err)
		return TypeFile, err
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		logger.Error("Storage.Manager", "Failed to determine file type.", err)
		return TypeFile, err
	}

	if info.IsDir() {
		return TypeFolder, nil
	} else if !info.Mode().IsRegular() {
		return TypeFile, nil
	} else {
		return s.detectWithMIMEType(fullPath)
	}
}

func (s *Manager) Exists(path string) bool {
	fullPath, _, err := s.safePath(path)
	if err != nil {
		logger.Error("Storage.Manager", "Failed to check whether a file exists.", err)
		return false
	}

	_, err = os.Stat(fullPath)
	if err == nil {
		return true
	}
	return false
}

func (s *Manager) Delete(path string, recursive bool) error {
	fullPath, _, err := s.safePath(path)
	if err != nil {
		logger.Error("Storage.Manager", "Failed to delete a file.", err)
		return err
	}

	if recursive {
		err = os.RemoveAll(fullPath)
	} else {
		err = os.Remove(fullPath)
	}
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return errs.ErrNotFound
		}
		logger.Error("Storage.Manager", "Failed to delete a file.", err)
		return err
	}
	return nil
}
