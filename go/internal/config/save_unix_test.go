//go:build unix

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOwnedSaveRefusesSymlinkParents(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(root, "outside")
	if err := os.Mkdir(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(outside, "config.json")
	if err := os.WriteFile(path, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	if err := saveOwnedFile(filepath.Join(link, "config.json"), []byte("replace"), os.Getuid(), os.Getgid(), false); err == nil {
		t.Fatal("followed symlink parent")
	}
	if b, err := os.ReadFile(path); err != nil || string(b) != "keep" {
		t.Fatal("outside config changed")
	}
	safe := filepath.Join(root, "new", "nested", "config.json")
	if err := saveOwnedFile(safe, []byte("new"), os.Getuid(), os.Getgid(), false); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(safe); err != nil || string(b) != "new" {
		t.Fatal("config not written")
	}
}

func TestFileOverridePreservesSharedParent(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(root, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "config.json")
	t.Setenv("MILLENNIUM_CONFIG_FILE", path)
	if err := Save(Data{"update_channel": "beta"}); err != nil {
		t.Fatal(err)
	}
	if err := saveOwnedFile(path, []byte("{}"), os.Getuid(), os.Getgid(), false); err != nil {
		t.Fatal(err)
	}
	if st, err := os.Stat(root); err != nil || st.Mode().Perm() != 0o755 {
		t.Fatal("changed shared config parent")
	}
}
