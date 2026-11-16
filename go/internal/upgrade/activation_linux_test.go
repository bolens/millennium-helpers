//go:build linux

package upgrade

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bolens/millennium-helpers/internal/clientfiles"
)

func TestSteamRT3RejectsLegacyActivationBeforeMutation(t *testing.T) {
	for _, action := range []string{"install", "rollback"} {
		t.Run(action, func(t *testing.T) {
			home, lib := t.TempDir(), t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("SUDO_USER", "")
			t.Setenv("STEAM", "")
			t.Setenv("STEAM_PATH", "")
			t.Setenv("MOCK_LIB_DIR", lib)
			steam := filepath.Join(home, ".local/share/Steam")
			if err := os.MkdirAll(filepath.Join(steam, "steamrt64"), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(steam, "steamrt64/steam"), []byte("client"), 0755); err != nil {
				t.Fatal(err)
			}
			active := filepath.Join(lib, "millennium")
			backup := filepath.Join(lib, "millennium.bak_legacy")
			for _, root := range []string{active, backup} {
				if err := os.Mkdir(root, 0755); err != nil {
					t.Fatal(err)
				}
			}
			marker := filepath.Join(active, "marker")
			if err := os.WriteFile(marker, []byte("working x64 install"), 0644); err != nil {
				t.Fatal(err)
			}
			writeRollbackClient(t, backup)
			hook := filepath.Join(steam, "steamrt64/libXtst.so.6")
			if err := os.Symlink("existing-hook", hook); err != nil {
				t.Fatal(err)
			}
			var err error
			if action == "install" {
				archive := filepath.Join(t.TempDir(), "legacy.tar.gz")
				if err = writeTestTarGz(archive); err != nil {
					t.Fatal(err)
				}
				err = installPlatform(archive, "3.6.3-beta.1", Options{})
			} else {
				err = rollbackPlatform("millennium.bak_legacy", Options{})
			}
			if err == nil || !strings.Contains(err.Error(), "client activation refused") {
				t.Fatalf("error=%v", err)
			}
			if b, err := os.ReadFile(marker); err != nil || string(b) != "working x64 install" {
				t.Fatal("active client changed", err)
			}
			if target, err := os.Readlink(hook); err != nil || target != "existing-hook" {
				t.Fatal("hook changed", err)
			}
			if st, err := os.Stat(filepath.Join(backup, "libmillennium_pvs64")); err != nil || st.Mode().Perm() != 0644 {
				t.Fatal("backup modified before validation", err)
			}
			if _, err := os.Lstat(active + ".tmp"); !os.IsNotExist(err) {
				t.Fatal("failed install left staging behind", err)
			}
			entries, err := os.ReadDir(lib)
			if err != nil {
				t.Fatal(err)
			}
			for _, e := range entries {
				if strings.HasPrefix(e.Name(), "millennium.bak_") && e.Name() != "millennium.bak_legacy" {
					t.Fatal("created backup before validation")
				}
			}
		})
	}
}

func TestSteamRT3ActivationAcceptsCompleteClient(t *testing.T) {
	home, root := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("SUDO_USER", "")
	t.Setenv("STEAM", home)
	t.Setenv("STEAM_PATH", "")
	if err := os.WriteFile(filepath.Join(home, ".steam-enable-steamrt64-client"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	for _, f := range clientfiles.Linux64() {
		if err := os.WriteFile(filepath.Join(root, f.Name), []byte(f.Name), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := validateClientActivation(root, false); err != nil {
		t.Fatal(err)
	}
}

func TestSteamRT3ActivationChecksAllTargetUsers(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("SUDO_USER", "")
	t.Setenv("STEAM", "")
	t.Setenv("STEAM_PATH", "")
	home := t.TempDir()
	steam := filepath.Join(home, ".local/share/Steam/steamrt64")
	if err := os.MkdirAll(steam, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(steam, "steam"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	old := passwdPath
	passwdPath = filepath.Join(t.TempDir(), "passwd")
	t.Cleanup(func() { passwdPath = old })
	if err := os.WriteFile(passwdPath, []byte("fixture:x:1001:1001::"+home+":/bin/sh\n"), 0644); err != nil {
		t.Fatal(err)
	}
	candidate := t.TempDir()
	if err := validateClientActivation(candidate, false); err != nil {
		t.Fatal("checked an untargeted account", err)
	}
	if err := validateClientActivation(candidate, true); err == nil {
		t.Fatal("ignored a targeted SteamRT3 account")
	}
}
