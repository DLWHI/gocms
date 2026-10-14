//go:build windows

// storage_windows.go
package storage

import (
	"golang.org/x/sys/windows"
)

func (s *Manager) GetInfo() (Info, error) {
	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64

	pathPtr, err := windows.UTF16PtrFromString(s.root)
	if err != nil {
		return Info{}, err
	}

	err = windows.GetDiskFreeSpaceEx(
		pathPtr,
		&freeBytesAvailable,
		&totalNumberOfBytes,
		&totalNumberOfFreeBytes,
	)
	if err != nil {
		return Info{}, err
	}

	return Info{
		Total: totalNumberOfBytes,
		Free:  freeBytesAvailable,
		Used:  totalNumberOfBytes - freeBytesAvailable,
	}, nil
}
