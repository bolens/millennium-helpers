package dirlock

import (
	"errors"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

// Acquire keeps a stable lock file. Deleting it on unlock would allow contenders
// to lock different file objects. Windows releases the byte lock on handle close.
func Acquire(dir string) (*os.File, error) {
	f, err := os.OpenFile(filepath.Join(dir, ".millennium.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	var overlap windows.Overlapped
	err = windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &overlap)
	if err != nil {
		_ = f.Close()
		if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
			return nil, ErrBusy
		}
		return nil, err
	}
	return f, nil
}
