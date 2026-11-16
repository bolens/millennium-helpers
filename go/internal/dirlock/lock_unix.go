//go:build unix

package dirlock

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

// Acquire locks an existing directory without creating persistent lock state.
// Closing the returned file, or exiting the process, releases the lock.
func Acquire(dir string) (*os.File, error) {
	fd, err := unix.Open(dir, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	var lockErr error
	for {
		lockErr = unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB)
		if !errors.Is(lockErr, unix.EINTR) {
			break
		}
	}
	if lockErr != nil {
		_ = unix.Close(fd)
		if errors.Is(lockErr, unix.EWOULDBLOCK) || errors.Is(lockErr, unix.EAGAIN) {
			return nil, ErrBusy
		}
		return nil, lockErr
	}
	return os.NewFile(uintptr(fd), dir), nil
}
