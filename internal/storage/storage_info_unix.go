//go:build linux || darwin

// storage_info_unix.go
package storage

import (
	"golang.org/x/sys/unix"
)

func (s *Service) GetInfo() (Info, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(s.root, &stat); err != nil {
		return Info{}, err
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free

	return Info{
		Total: total,
		Free:  free,
		Used:  used,
	}, nil
}
