//go:build unix

package safefs

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

func testParent(t *testing.T) (string, int) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	fd, err := OpenDir(root, false, 0700, -1, -1)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := unix.Close(fd); err != nil {
			t.Error(err)
		}
	})
	return root, fd
}

func TestOpenRefusesSymlinkedAncestors(t *testing.T) {
	root, fd := testParent(t)
	outside, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "redirect")); err != nil {
		t.Fatal(err)
	}
	if child, err := OpenChildDir(fd, "redirect", true, 0700, -1, -1); err == nil {
		_ = unix.Close(child)
		t.Fatal("followed child symlink")
	}
	if child, err := OpenDir(filepath.Join(root, "redirect", "created"), true, 0700, -1, -1); err == nil {
		_ = unix.Close(child)
		t.Fatal("followed ancestor symlink")
	}
	entries, err := os.ReadDir(outside)
	if err != nil || len(entries) != 0 {
		t.Fatalf("outside directory changed: %v %v", entries, err)
	}
	if child, err := OpenDir("relative/path", true, 0700, -1, -1); err == nil {
		_ = unix.Close(child)
		t.Fatal("accepted relative privileged path")
	}
}

func TestInvalidComponentsCannotEscapeParent(t *testing.T) {
	root, fd := testParent(t)
	for _, name := range []string{"", ".", "..", "nested/name"} {
		if child, err := OpenChildDir(fd, name, true, 0700, -1, -1); err == nil {
			_ = unix.Close(child)
			t.Fatalf("created invalid child %q", name)
		}
		if err := Symlink(fd, name, "/foreign"); err == nil {
			t.Fatalf("accepted invalid link %q", name)
		}
		if err := WriteFile(fd, name, []byte("data"), 0600, -1, -1); err == nil {
			t.Fatalf("accepted invalid file %q", name)
		}
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("invalid input changed directory: %v %v", entries, err)
	}
}

func TestAtomicWriteReplacesLinkWithoutFollowingIt(t *testing.T) {
	root, fd := testParent(t)
	victim := filepath.Join(t.TempDir(), "victim")
	if err := os.WriteFile(victim, []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "config")
	if err := os.Symlink(victim, target); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(fd, "config", []byte("new configuration"), 0640, os.Getuid(), os.Getgid()); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(victim)
	if err != nil || string(got) != "private" {
		t.Fatal("write followed final symlink")
	}
	got, err = os.ReadFile(target)
	if err != nil || string(got) != "new configuration" {
		t.Fatalf("replacement=%q err=%v", got, err)
	}
	info, err := os.Lstat(target)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0640 {
		t.Fatalf("replacement metadata=%v err=%v", info, err)
	}
	var st unix.Stat_t
	if err := unix.Fstatat(fd, "config", &st, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		t.Fatal(err)
	}
	if int(st.Uid) != os.Getuid() || int(st.Gid) != os.Getgid() {
		t.Fatalf("wrong owner=%d:%d", st.Uid, st.Gid)
	}
}

func TestFailedReplacementPreservesDestinationAndCleansTemporaryEntries(t *testing.T) {
	root, fd := testParent(t)
	dest := filepath.Join(root, "destination")
	if err := os.Mkdir(dest, 0700); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(dest, "keep")
	if err := os.WriteFile(sentinel, []byte("untouched"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := WriteFile(fd, "destination", []byte("replace"), 0600, -1, -1); err == nil {
		t.Fatal("replaced a directory with a file")
	}
	if err := Symlink(fd, "destination", "/foreign"); err == nil {
		t.Fatal("replaced a directory with a symlink")
	}
	got, err := os.ReadFile(sentinel)
	if err != nil || string(got) != "untouched" {
		t.Fatal("failed replacement damaged destination")
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 || entries[0].Name() != "destination" {
		t.Fatalf("temporary entries leaked: %v err=%v", entries, err)
	}
}

func TestSymlinkReplacementChangesOnlyTheNamedEntry(t *testing.T) {
	root, fd := testParent(t)
	if err := os.Symlink("/old", filepath.Join(root, "hook")); err != nil {
		t.Fatal(err)
	}
	if err := Symlink(fd, "hook", "/new"); err != nil {
		t.Fatal(err)
	}
	got, err := os.Readlink(filepath.Join(root, "hook"))
	if err != nil || got != "/new" {
		t.Fatalf("hook=%q err=%v", got, err)
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 1 {
		t.Fatalf("unexpected entries=%v err=%v", entries, err)
	}
}
