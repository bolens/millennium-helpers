package clientfiles

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidVersion accepts a bounded display label that is also safe in a basename.
func ValidVersion(v string) bool {
	if len(v) == 0 || len(v) > 128 || v == "." || v == ".." {
		return false
	}
	for _, c := range v {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '.', c == '-', c == '_':
		default:
			return false
		}
	}
	return true
}

// Version requires regular, bounded version metadata without path components.
func Version(root string) (string, error) {
	path := filepath.Join(root, "version.txt")
	st, err := os.Lstat(path)
	if errors.Is(err, os.ErrPermission) {
		return "", fmt.Errorf("cannot inspect version.txt: %w", os.ErrPermission)
	}
	if err != nil || !st.Mode().IsRegular() || st.Size() > 130 {
		return "", fmt.Errorf("version.txt missing or invalid")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return "", fmt.Errorf("cannot read version.txt: %w", os.ErrPermission)
		}
		return "", fmt.Errorf("cannot read version.txt")
	}
	v := strings.TrimSpace(string(b))
	if !ValidVersion(v) {
		return "", fmt.Errorf("invalid version.txt label")
	}
	return v, nil
}
