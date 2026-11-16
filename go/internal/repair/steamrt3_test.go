package repair

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSteamRT3HookPlan(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux container hook targets")
	}
	root := t.TempDir()
	if SteamRT3(root) {
		t.Fatal("empty tree detected as SteamRT3")
	}
	if err := os.MkdirAll(filepath.Join(root, "steamrt64"), 0755); err != nil {
		t.Fatal(err)
	}
	if SteamRT3(root) {
		t.Fatal("directory alone detected as SteamRT3")
	}
	if err := os.WriteFile(filepath.Join(root, "steamrt64/steam"), []byte("client"), 0755); err != nil {
		t.Fatal(err)
	}
	plans := HooksAt(root, "/usr/lib/millennium")
	if len(plans) != 3 || plans[2].Target != "/run/host/usr/lib/millennium/libmillennium_bootstrap_x64.so" {
		t.Fatalf("%#v", plans)
	}
	if HostHookTarget(plans[2].Target) != "/usr/lib/millennium/libmillennium_bootstrap_x64.so" {
		t.Fatal("wrong host target")
	}
}

func TestSteamRT3HookRepair(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux safe hook writer")
	}
	root, lib := t.TempDir(), t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "steamrt64"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "steamrt64/steam"), nil, 0755); err != nil {
		t.Fatal(err)
	}
	plans := HooksAt(root, lib)
	for _, p := range plans[:2] {
		if err := os.WriteFile(p.Target, []byte("lib"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := InstallHooksAt(root, lib); err == nil {
		t.Fatal("accepted missing x64 bootstrap")
	}
	if _, err := os.Lstat(plans[0].Hook); !os.IsNotExist(err) {
		t.Fatal("changed legacy hooks before source validation")
	}
	if err := os.WriteFile(HostHookTarget(plans[2].Target), []byte("lib"), 0644); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := InstallHooksAt(root, lib); err != nil {
			t.Fatal(err)
		}
	}
	for _, p := range plans {
		got, err := os.Readlink(p.Hook)
		if err != nil || got != p.Target {
			t.Fatalf("hook=%s err=%v", got, err)
		}
	}
	// A symlinked SteamRT3 architecture parent must never be followed.
	outside := t.TempDir()
	if err := os.Remove(plans[2].Hook); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "steamrt64/steam")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "steamrt64")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".steam-enable-steamrt64-client"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "steamrt64")); err != nil {
		t.Fatal(err)
	}
	if err := InstallHooksAt(root, lib); err == nil {
		t.Fatal("followed SteamRT3 parent symlink")
	}
	if _, err := os.Lstat(filepath.Join(outside, "libXtst.so.6")); !os.IsNotExist(err) {
		t.Fatal("modified outside tree")
	}
}
