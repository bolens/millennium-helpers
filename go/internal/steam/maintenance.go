package steam

import (
	"errors"
	"fmt"
)

// WithClosedClient captures a running client's session, confirms shutdown, and
// always attempts to resume it after work, including when work fails.
func WithClosedClient(yes bool, work func() error) error {
	if !IsSteamRunning() {
		return work()
	}
	name, home, err := TargetUser()
	if err != nil {
		return err
	}
	return withClosedClient(
		func() error { return CaptureEnv(name, home) },
		func() error { return ConfirmClose(yes) },
		func() error {
			attempted, err := RelaunchFromState(name, home)
			if err != nil {
				return fmt.Errorf("Steam relaunch failed: %w", err)
			}
			if !attempted {
				return fmt.Errorf("Steam relaunch state unavailable")
			}
			return nil
		}, work)
}

func withClosedClient(capture, close, resume, work func() error) (err error) {
	if err = capture(); err != nil {
		return err
	}
	if err = close(); err != nil {
		return err
	}
	defer func() { err = errors.Join(err, resume()) }()
	return work()
}
