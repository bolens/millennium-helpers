//go:build unix

package upgrade

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bolens/millennium-helpers/internal/clientfiles"
)

func rollbackPlatform(backupName string, o Options) error {
	lib := LibDir()
	backupPath := filepath.Join(lib, backupName)
	dest := filepath.Join(lib, "millennium")
	if st, err := os.Lstat(backupPath); err != nil || !st.IsDir() {
		return fmt.Errorf("Error: Backup '%s' not found.", backupName)
	}

	if runtime.GOOS == "linux" {
		if err := clientfiles.Verify(backupPath); err != nil {
			return fmt.Errorf("backup validation failed: %w", err)
		}
		if err := clientfiles.NormalizeHelpers(backupPath); err != nil {
			return fmt.Errorf("backup permissions failed: %w", err)
		}
	}

	saved := ""
	if st, err := os.Lstat(dest); err == nil {
		if !st.IsDir() {
			return fmt.Errorf("active client must be a real directory")
		}
		saved, err = reserveBackup(lib, dest)
		if err != nil {
			return err
		}
		if err = os.Rename(dest, saved); err != nil {
			_ = os.Remove(saved)
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(backupPath, dest); err != nil {
		if saved != "" {
			if restoreErr := os.Rename(saved, dest); restoreErr != nil {
				return fmt.Errorf("activation and recovery failed: %w", restoreErr)
			}
		}
		return fmt.Errorf("failed to activate backup: %w", err)
	}
	label := strings.TrimPrefix(backupName, "millennium.bak_")
	if !o.Quiet {
		fmt.Printf("Rollback successful! Backup %s is now active.\n", label)
	}
	if saved != "" && !o.Quiet {
		fmt.Printf("Saved previous installation to %s\n", filepath.Base(saved))
	}
	return nil
}
