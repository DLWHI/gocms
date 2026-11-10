//go:build windows

// error_windows.go
package storage

import (
	"errors"
	"fmt"
	"gocms/internal/domain"
	"os"
	"syscall"
)

func MapError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, os.ErrNotExist):
		return domain.ErrNotFound
	case errors.Is(err, syscall.ERROR_DIR_NOT_EMPTY):
		return fmt.Errorf("%w: directory not empty", domain.ErrInval)
	default:
		return err
	}
}
