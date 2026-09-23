//go:build unix

package safefs

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

// OpenDir opens an absolute directory without following any symlink. When create
// is true, missing components are created and assigned to uid/gid before descent.
// The caller owns the returned descriptor. Negative IDs preserve ownership.
func OpenDir(path string, create bool, mode os.FileMode, uid, gid int) (int, error) {
	if !filepath.IsAbs(path) {
		return -1, fmt.Errorf("directory must be absolute")
	}
	fd, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return -1, err
	}
	for _, part := range strings.Split(strings.TrimPrefix(filepath.Clean(path), "/"), "/") {
		if part == "" {
			continue
		}
		next, err := OpenChildDir(fd, part, create, mode, uid, gid)
		_ = unix.Close(fd)
		if err != nil {
			return -1, err
		}
		fd = next
	}
	return fd, nil
}

func component(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.Contains(name, "/")
}

// OpenChildDir never follows a child symlink, including after Mkdirat races.
func OpenChildDir(parent int, name string, create bool, mode os.FileMode, uid, gid int) (int, error) {
	if !component(name) {
		return -1, fmt.Errorf("invalid path component")
	}
	flags := unix.O_RDONLY | unix.O_DIRECTORY | unix.O_CLOEXEC | unix.O_NOFOLLOW
	fd, err := unix.Openat(parent, name, flags, 0)
	if err != unix.ENOENT || !create {
		return fd, err
	}
	created := false
	if err := unix.Mkdirat(parent, name, uint32(mode.Perm())); err != nil {
		if err != unix.EEXIST {
			return -1, err
		}
	} else {
		created = true
	}
	fd, err = unix.Openat(parent, name, flags, 0)
	if err != nil {
		return -1, err
	}
	if created {
		if err = unix.Fchown(fd, uid, gid); err != nil {
			_ = unix.Close(fd)
			return -1, err
		}
	}
	return fd, nil
}

func temporaryName() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return ".millennium-" + hex.EncodeToString(b[:]), nil
}

// Symlink atomically replaces one entry in an already opened directory.
func Symlink(parent int, name, target string) error {
	if !component(name) {
		return fmt.Errorf("invalid link name")
	}
	tmp, err := temporaryName()
	if err != nil {
		return err
	}
	if err = unix.Symlinkat(target, parent, tmp); err != nil {
		return err
	}
	defer func() { _ = unix.Unlinkat(parent, tmp, 0) }()
	return unix.Renameat(parent, tmp, parent, name)
}

// WriteFile assigns ownership before making a replacement visible to readers.
func WriteFile(parent int, name string, data []byte, mode os.FileMode, uid, gid int) error {
	if !component(name) {
		return fmt.Errorf("invalid file name")
	}
	tmp, err := temporaryName()
	if err != nil {
		return err
	}
	fd, err := unix.Openat(parent, tmp, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_CLOEXEC|unix.O_NOFOLLOW, uint32(mode.Perm()))
	if err != nil {
		return err
	}
	f := os.NewFile(uintptr(fd), tmp)
	defer func() { _ = f.Close() }()
	defer func() { _ = unix.Unlinkat(parent, tmp, 0) }()
	if err = f.Chown(uid, gid); err != nil {
		return err
	}
	if err = f.Chmod(mode); err != nil {
		return err
	}
	if _, err = f.Write(data); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return unix.Renameat(parent, tmp, parent, name)
}
