package config

import (
	"fmt"
	"path/filepath"

	"github.com/bolens/millennium-helpers/internal/dirlock"
)

// Update reads, changes, and writes one configuration while excluding other
// helper writers. Contention fails before invoking change or loading a snapshot.
func Update(change func(Data) error) error {
	path := Path()
	if path == "" {
		return fmt.Errorf("cannot resolve configuration directory")
	}
	if err := prepareDirectory(path); err != nil {
		return err
	}
	lock, err := dirlock.Acquire(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer func() { _ = lock.Close() }()
	data, err := Load()
	if err != nil {
		return err
	}
	if err := change(data); err != nil {
		return err
	}
	return Save(data)
}
