//go:build unix

package repair

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bolens/millennium-helpers/internal/steam"
	"github.com/bolens/millennium-helpers/internal/usercontext"
	"golang.org/x/sys/unix"
)

func chownTree(path string) error {
	uid, gid, err := repairOwnerIDs()
	if err != nil {
		return err
	}
	parent, name, err := openRepairParent(path)
	if err != nil {
		return err
	}
	defer unix.Close(parent)
	return chownEntry(parent, name, uid, gid)
}

// openRepairParent refuses symlinks in every ancestor of a repair target.
func openRepairParent(path string) (int, string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return -1, "", err
	}
	parts := strings.Split(strings.TrimPrefix(filepath.Clean(abs), "/"), "/")
	fd, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return -1, "", err
	}
	// Protect the initial root as well as descendants. A user-controlled parent
	// symlink must not redirect an elevated repair to a system directory.
	for _, part := range parts[:len(parts)-1] {
		next, openErr := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		unix.Close(fd)
		if openErr != nil {
			return -1, "", fmt.Errorf("repair refuses inaccessible or symlinked parent: %w", openErr)
		}
		fd = next
	}
	return fd, parts[len(parts)-1], nil
}

// Anchor each descent to an open directory. Never follow symlinks supplied by
// the user, including links substituted while a privileged repair is running.
func chownEntry(parent int, name string, uid, gid int) error {
	var st unix.Stat_t
	if err := unix.Fstatat(parent, name, &st, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return err
	}
	if st.Mode&unix.S_IFMT != unix.S_IFREG && st.Mode&unix.S_IFMT != unix.S_IFDIR {
		return nil
	}
	fd, err := unix.Openat(parent, name, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
	if err != nil {
		return err
	}
	f := os.NewFile(uintptr(fd), name)
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if !info.IsDir() && !info.Mode().IsRegular() {
		return nil
	}
	if err := f.Chown(uid, gid); err != nil {
		return err
	}
	if !info.IsDir() {
		return nil
	}
	entries, err := f.ReadDir(-1)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := chownEntry(fd, entry.Name(), uid, gid); err != nil {
			return err
		}
	}
	return nil
}

func repairOwnerIDs() (int, int, error) {
	ctx, err := usercontext.Resolve()
	if err != nil {
		return 0, 0, err
	}
	uid, err := strconv.Atoi(ctx.UID)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid invoking user ID")
	}
	gid, err := strconv.Atoi(ctx.GID)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid invoking group ID")
	}
	return uid, gid, nil
}

func ensureSteamClosedForRepair(yes bool) (relaunch bool, err error) {
	// Offline / CI seam: do not touch a host Steam client under MOCK_LIB_DIR.
	if os.Getenv("MOCK_LIB_DIR") != "" {
		return false, nil
	}
	if !steam.IsSteamRunning() {
		return false, nil
	}
	username, home, err := steam.TargetUser()
	if err != nil {
		return false, err
	}
	if err := steam.CaptureEnv(username, home); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not capture Steam environment: %v\n", err)
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
	username, home, err := steam.TargetUser()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	_, _ = steam.RelaunchFromState(username, home)
}

// clearCache leaves the cache root in place and unlinks only entries reached
// through its open descriptor. A symlinked root or ancestor is never followed.
func clearCache(path string) error {
	parent, name, err := openRepairParent(path)
	if err != nil {
		return err
	}
	defer unix.Close(parent)
	fd, err := unix.Openat(parent, name, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return err
	}
	return removeCacheChildren(fd)
}

// removeCacheChildren owns and closes fd.
func removeCacheChildren(fd int) error {
	f := os.NewFile(uintptr(fd), "cache")
	defer f.Close()
	entries, err := f.ReadDir(-1)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		var st unix.Stat_t
		if err := unix.Fstatat(fd, entry.Name(), &st, unix.AT_SYMLINK_NOFOLLOW); err != nil {
			return err
		}
		flags := 0
		if st.Mode&unix.S_IFMT == unix.S_IFDIR {
			child, err := unix.Openat(fd, entry.Name(), unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
			if err != nil {
				return err
			}
			if err := removeCacheChildren(child); err != nil {
				return err
			}
			flags = unix.AT_REMOVEDIR
		}
		if err := unix.Unlinkat(fd, entry.Name(), flags); err != nil {
			return err
		}
	}
	return nil
}
