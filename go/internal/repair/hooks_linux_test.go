package repair

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHooksRefuseSymlinkedArchitecture(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "Steam")
	lib := filepath.Join(base, "lib")
	outside := filepath.Join(base, "outside")
	for _, p := range []string{root, lib, outside} {
		if err := os.Mkdir(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"libmillennium_bootstrap_x86.so", "libmillennium_bootstrap_hhx64.so"} {
		if err := os.WriteFile(filepath.Join(lib, name), []byte("library"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	sentinel := filepath.Join(outside, "libXtst.so.6")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	arch := filepath.Join(root, "ubuntu12_32")
	if err := os.Symlink(outside, arch); err != nil {
		t.Fatal(err)
	}
	if err := InstallHooksAt(root, lib); err == nil {
		t.Fatal("accepted symlinked hook parent")
	}
	if b, err := os.ReadFile(sentinel); err != nil || string(b) != "keep" {
		t.Fatal("outside hook changed")
	}
	if err := os.Remove(arch); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := InstallHooksAt(root, lib); err != nil {
			t.Fatal(err)
		}
	}
	target, err := os.Readlink(filepath.Join(arch, "libXtst.so.6"))
	if err != nil || target != filepath.Join(lib, "libmillennium_bootstrap_x86.so") {
		t.Fatalf("hook=%s %v", target, err)
	}
}
