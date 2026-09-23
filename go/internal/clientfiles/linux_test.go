package clientfiles

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, name := range []string{"libmillennium_bootstrap_x86.so", "libmillennium_bootstrap_hhx64.so", "libmillennium_x86.so", "libmillennium_hhx64.so", "libmillennium_pvs64", "libmillennium_luavm_x86"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := WriteChecksums(root); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestIntegrityRejectsIncompleteAndCorruptInstall(t *testing.T) {
	for _, damage := range []string{"none", "missing-lua", "corrupt-lua", "empty", "omitted", "duplicate", "malformed", "traversal"} {
		t.Run(damage, func(t *testing.T) {
			root := fixture(t)
			path := filepath.Join(root, "checksums.txt")
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			lines := strings.SplitAfter(string(b), "\n")
			switch damage {
			case "missing-lua":
				err = os.Remove(filepath.Join(root, "libmillennium_luavm_x86"))
			case "corrupt-lua":
				err = os.WriteFile(filepath.Join(root, "libmillennium_luavm_x86"), []byte("bad"), 0o644)
			case "empty":
				b = nil
			case "omitted":
				b = []byte(strings.Join(lines[:len(lines)-2], ""))
			case "duplicate":
				b = append(b, []byte(lines[0])...)
			case "malformed":
				b = []byte("not a checksum\n")
			case "traversal":
				b = append(b, []byte(strings.Repeat("0", 64)+"  ../private\n")...)
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, b, 0o644); err != nil {
				t.Fatal(err)
			}
			err = Verify(root)
			if (err == nil) != (damage == "none") {
				t.Fatalf("Verify = %v", err)
			}
			if err != nil && strings.Contains(err.Error(), root) {
				t.Fatal("private path in diagnostic")
			}
		})
	}
}

func TestHelperModesAndPreflight(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix modes and symlinks")
	}
	for _, mode := range []os.FileMode{0o644, 0o700, 0o711, 0o755} {
		t.Run(mode.String(), func(t *testing.T) {
			root := fixture(t)
			for _, name := range []string{"libmillennium_pvs64", "libmillennium_luavm_x86"} {
				if err := os.Chmod(filepath.Join(root, name), mode); err != nil {
					t.Fatal(err)
				}
			}
			if HelpersExecutable(root) != (mode == 0o755) {
				t.Fatal("incorrect shared-install access check")
			}
			if err := NormalizeHelpers(root); err != nil {
				t.Fatal(err)
			}
			if !HelpersExecutable(root) {
				t.Fatal("repair did not restore access")
			}
			if err := Verify(root); err != nil {
				t.Fatal(err)
			}
		})
	}
	root := fixture(t)
	lua := filepath.Join(root, "libmillennium_luavm_x86")
	outside := filepath.Join(t.TempDir(), "external")
	if err := os.WriteFile(outside, []byte("external"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(lua); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, lua); err != nil {
		t.Fatal(err)
	}
	if Verify(root) == nil || NormalizeHelpers(root) == nil || HelpersExecutable(root) {
		t.Fatal("accepted symlink helper")
	}
	for _, path := range []string{filepath.Join(root, "libmillennium_pvs64"), outside} {
		st, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if st.Mode().Perm()&0o111 != 0 {
			t.Fatal("failed preflight changed modes")
		}
	}
}
