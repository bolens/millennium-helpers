//go:build unix

package config

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/bolens/millennium-helpers/internal/safefs"
	"github.com/bolens/millennium-helpers/internal/usercontext"
	"golang.org/x/sys/unix"
)

func saveFile(path string, data []byte) error {
	if os.Geteuid() != 0 || os.Getenv("SUDO_USER") == "" {
		return saveLocalFile(path, data)
	}
	ctx, err := usercontext.Resolve()
	if err != nil {
		return err
	}
	uid, err := strconv.Atoi(ctx.UID)
	if err != nil {
		return err
	}
	gid, err := strconv.Atoi(ctx.GID)
	if err != nil {
		return err
	}
	manageDir := os.Getenv("MILLENNIUM_CONFIG_FILE") == "" && os.Getenv("MILLENNIUM_CONFIG_DIR") == ""
	return saveOwnedFile(path, data, uid, gid, manageDir)
}

func saveOwnedFile(path string, data []byte, uid, gid int, manageDir bool) error {
	dir, err := safefs.OpenDir(filepath.Dir(path), true, 0o700, uid, gid)
	if err != nil {
		return err
	}
	defer unix.Close(dir)
	// Only the default dedicated directory belongs to the helpers. An override
	// may be inside a shared directory whose owner and mode must be preserved.
	if manageDir {
		if err := unix.Fchown(dir, uid, gid); err != nil {
			return err
		}
		if err := unix.Fchmod(dir, 0o700); err != nil {
			return err
		}
	}
	return safefs.WriteFile(dir, filepath.Base(path), data, 0o600, uid, gid)
}
