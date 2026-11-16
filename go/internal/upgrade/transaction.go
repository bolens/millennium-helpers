package upgrade

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bolens/millennium-helpers/internal/clientfiles"
	"github.com/bolens/millennium-helpers/internal/dirlock"
)

func withInstallLock(work func() error) error {
	root := InstallRoot()
	if root == "" {
		return fmt.Errorf("installation root is unavailable")
	}
	parent := filepath.Dir(root)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return err
	}
	lock, err := dirlock.Acquire(parent)
	if err != nil {
		return err
	}
	defer func() { _ = lock.Close() }()
	return work()
}

func installPlatform(archivePath, version string, o Options) error {
	if !clientfiles.ValidVersion(version) {
		return fmt.Errorf("invalid client version label")
	}
	return withInstallLock(func() error { return installPlatformLocked(archivePath, version, o) })
}

func rollbackPlatform(backupName string, o Options) error {
	return withInstallLock(func() error { return rollbackPlatformLocked(backupName, o) })
}
