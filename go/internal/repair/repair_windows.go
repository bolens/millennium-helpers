//go:build windows

package repair

import (
	"os"
	"path/filepath"

	"github.com/bolens/millennium-helpers/internal/steam"
)

func chownTree(path string) error {
	_ = path
	return nil
}

func ensureSteamClosedForRepair(yes bool) (relaunch bool, err error) {
	if os.Getenv("MOCK_LIB_DIR") != "" {
		return false, nil
	}
	if !steam.IsSteamRunning() {
		return false, nil
	}
	if err := steam.ConfirmClose(yes); err != nil {
		return false, err
	}
	return true, nil
}

func relaunchSteamAfterRepair() {
	if os.Getenv("MOCK_LIB_DIR") != "" {
		return
	}
	steam.RelaunchBestEffort()
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
