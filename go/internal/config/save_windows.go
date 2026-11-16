package config

import (
	"os"
	"path/filepath"
)

func saveFile(path string, data []byte) error { return saveLocalFile(path, data) }

func prepareDirectory(path string) error { return os.MkdirAll(filepath.Dir(path), 0700) }
