//go:build unix

package repair

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/bolens/millennium-helpers/internal/safefs"
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
	fd, err := safefs.OpenDir(filepath.Dir(abs), false, 0, -1, -1)
	if err != nil {
		return -1, "", fmt.Errorf("repair refuses inaccessible or symlinked parent: %w", err)
	}
	return fd, filepath.Base(abs), nil
}

// Anchor each descent to an open directory. Never follow symlinks supplied by
// the user, including links substituted while a privileged repair is running.
func chownEntry(parent int, name string, uid, gid int) error {
	return walkOwnership(parent, name, uid, gid, true)
}

func walkOwnership(parent int, name string, uid, gid int, fix bool) error {
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
	if fix {
		if err := f.Chown(uid, gid); err != nil {
			return err
		}
	} else {
		if err := unix.Fstat(fd, &st); err != nil {
			return err
		}
		if int(st.Uid) != uid || int(st.Gid) != gid {
			return fmt.Errorf("ownership differs from invoking user")
		}
	}
	if !info.IsDir() {
		return nil
	}
	entries, err := f.ReadDir(-1)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := walkOwnership(fd, entry.Name(), uid, gid, fix); err != nil {
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

func checkOwnership(path string) error {
	uid, gid, err := repairOwnerIDs()
	if err != nil {
		return err
	}
	fd, name, err := openRepairParent(path)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	return walkOwnership(fd, name, uid, gid, false)
}
