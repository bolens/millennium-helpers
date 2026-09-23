//go:build unix

package upgrade

import (
	"os"
	"path/filepath"

	"github.com/bolens/millennium-helpers/internal/clientfiles"
)

// reserveBackup allocates an unused backup name without deleting prior backups.
// Damaged active metadata gets an opaque label so reinstall can still recover it.
func reserveBackup(lib, active string) (string, error) {
	label, err := clientfiles.Version(active)
	if err != nil {
		label = "unknown"
	}
	path := filepath.Join(lib, "millennium.bak_"+label)
	if err := os.Mkdir(path, 0o700); err == nil {
		return path, os.Remove(path)
	} else if !os.IsExist(err) {
		return "", err
	}
	path, err = os.MkdirTemp(lib, "millennium.bak_"+label+"-")
	if err != nil {
		return "", err
	}
	return path, os.Remove(path)
}
