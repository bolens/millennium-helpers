//go:build unix

package repair

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/sys/unix"
)

func TestOwnershipRepairDoesNotFollowSymlinks(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	// Following either link would fail. Neither target belongs to the repair.
	for _, name := range []string{"file-link", "directory-link"} {
		if err := os.Symlink(filepath.Join(t.TempDir(), "absent"), filepath.Join(root, name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "regular"), []byte("fixture"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := chownTree(root); err != nil {
		t.Fatal(err)
	}
	if err := chownTree(filepath.Join(root, "directory-link")); err != nil {
		t.Fatal(err)
	}
}

func TestOwnershipRepairRejectsSymlinkedAncestor(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(base, "outside")
	if err := os.MkdirAll(filepath.Join(out, "millennium"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(base, "share")
	if err := os.Symlink(out, link); err != nil {
		t.Fatal(err)
	}
	if err := chownTree(filepath.Join(link, "millennium")); err == nil {
		t.Fatal("followed symlinked ancestor outside repair tree")
	}
}

func TestOwnershipTraversalUsesOpenDirectory(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "original")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "child"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := os.Rename(root, filepath.Join(base, "moved")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(base, "absent"), root); err != nil {
		t.Fatal(err)
	}
	if err := chownEntry(int(f.Fd()), "child", os.Getuid(), os.Getgid()); err != nil {
		t.Fatal(err)
	}
	// A replaced leaf must not be opened through its link either.
	if err := os.Symlink(filepath.Join(base, "absent"), filepath.Join(base, "leaf")); err != nil {
		t.Fatal(err)
	}
	if err := chownEntry(unix.AT_FDCWD, filepath.Join(base, "leaf"), os.Getuid(), os.Getgid()); err != nil {
		t.Fatal(err)
	}
}

func TestCacheCleanupRejectsEscapes(t *testing.T) {
	for _, ancestor := range []bool{false, true} {
		t.Run(map[bool]string{false: "root", true: "ancestor"}[ancestor], func(t *testing.T) {
			base, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			outside := filepath.Join(base, "outside")
			if err := os.MkdirAll(filepath.Join(outside, "htmlcache"), 0o755); err != nil {
				t.Fatal(err)
			}
			victim := filepath.Join(outside, "htmlcache", "sentinel")
			if err := os.WriteFile(victim, []byte("keep"), 0o644); err != nil {
				t.Fatal(err)
			}
			link := filepath.Join(base, "link")
			target, cache := filepath.Dir(victim), link
			if ancestor {
				target = outside
				cache = filepath.Join(link, "htmlcache")
			}
			if err := os.Symlink(target, link); err != nil {
				t.Fatal(err)
			}
			if err := Apply([]Target{{Path: cache, Kind: "htmlcache"}}, true); err == nil {
				t.Fatal("accepted symlinked cache")
			}
			if b, err := os.ReadFile(victim); err != nil || string(b) != "keep" {
				t.Fatal("deleted outside cache")
			}
		})
	}
}

func TestCacheCleanupStaysWithOpenRoot(t *testing.T) {
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cache, outside := filepath.Join(base, "cache"), filepath.Join(base, "outside")
	if err := os.MkdirAll(filepath.Join(cache, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outside, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cache, "nested", "file"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(cache, "child-link")); err != nil {
		t.Fatal(err)
	}
	fd, err := unix.Open(cache, unix.O_RDONLY|unix.O_DIRECTORY, 0)
	if err != nil {
		t.Fatal(err)
	}
	// The opened tree is renamed and its old path replaced before deletion.
	moved := filepath.Join(base, "moved")
	if err := os.Rename(cache, moved); err != nil {
		unix.Close(fd)
		t.Fatal(err)
	}
	if err := os.Symlink(base, cache); err != nil {
		unix.Close(fd)
		t.Fatal(err)
	}
	if err := removeCacheChildren(fd); err != nil {
		t.Fatal(err)
	}
	if entries, err := os.ReadDir(moved); err != nil || len(entries) != 0 {
		t.Fatalf("cache not cleared: %v %v", entries, err)
	}
	if b, err := os.ReadFile(outside); err != nil || string(b) != "keep" {
		t.Fatal("followed link outside cache")
	}
}

func TestOwnershipHealthAndFailurePropagation(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	home := filepath.Join(root, "home")
	cfg := filepath.Join(root, "cfg")
	owned := filepath.Join(cfg, "millennium-helpers")
	if err := os.MkdirAll(owned, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", cfg)
	t.Setenv("XDG_DATA_HOME", filepath.Join(root, "data"))
	t.Setenv("STEAM", filepath.Join(root, "Steam"))
	t.Setenv("STEAM_PATH", "")
	if !PermissionsOK() {
		t.Fatal("caller-owned files unhealthy")
	}
	fd, name, err := openRepairParent(owned)
	if err != nil {
		t.Fatal(err)
	}
	if err := walkOwnership(fd, name, os.Getuid()+1, os.Getgid(), false); err == nil {
		t.Fatal("ownership mismatch not detected")
	}
	unix.Close(fd)
	link := filepath.Join(root, "alias")
	if err := os.Symlink(cfg, link); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", link)
	if PermissionsOK() {
		t.Fatal("unsafe ownership target reported healthy")
	}
	targets, err := Plan()
	if err != nil {
		t.Fatal(err)
	}
	if err := Apply(targets, true); err == nil {
		t.Fatal("failed ownership repair reported success")
	}
}
