//go:build windows

package repair

import (
	"os"
	"path/filepath"
)

func chownTree(path string) error {
	_ = path
	return nil
}

func clearCache(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(path, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func checkOwnership(path string) error { return nil }
