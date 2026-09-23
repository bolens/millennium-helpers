//go:build unix

package diag

import (
	"fmt"
	"os"

	"github.com/bolens/millennium-helpers/internal/steam"
)

func doctorCloseSteam(yes bool) (relaunch bool, err error) {
	if !steam.IsSteamRunning() {
		return false, nil
	}
	username, home, err := steam.TargetUser()
	if err != nil {
		return false, err
	}
	if err := steam.CaptureEnv(username, home); err != nil {
		return false, err
	}
	fmt.Println("Steam is currently running and must be closed to apply repairs to hooks/binaries.")
	if err := steam.ConfirmClose(yes); err != nil {
		return false, err
	}
	return true, nil
}

func doctorRelaunchSteam() {
	username, home, err := steam.TargetUser()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if _, err := steam.RelaunchFromState(username, home); err != nil {
		fmt.Fprintln(os.Stderr, "Could not relaunch Steam for the invoking user")
	}
}
