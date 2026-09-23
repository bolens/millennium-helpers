//go:build unix

package upgrade

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRollbackMetadataAndBackupPreservation(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux manifest validation")
	}
	for _, scenario := range []string{"malformed-active", "missing-candidate", "malformed-candidate", "existing-backup"} {
		t.Run(scenario, func(t *testing.T) {
			base := t.TempDir()
			lib := filepath.Join(base, "lib")
			active := filepath.Join(lib, "millennium")
			candidate := filepath.Join(lib, "millennium.bak_1.0.0")
			for _, p := range []string{active, candidate, filepath.Join(base, "outside")} {
				if err := os.MkdirAll(p, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			sentinel := filepath.Join(base, "outside", "sentinel")
			if err := os.WriteFile(sentinel, []byte("keep"), 0o644); err != nil {
				t.Fatal(err)
			}
			version := "2.0.0"
			if scenario == "malformed-active" {
				version = "x/../../outside"
			}
			if err := os.WriteFile(filepath.Join(active, "version.txt"), []byte(version), 0o644); err != nil {
				t.Fatal(err)
			}
			writeRollbackClient(t, candidate)
			switch scenario {
			case "missing-candidate":
				if err := os.Remove(filepath.Join(candidate, "version.txt")); err != nil {
					t.Fatal(err)
				}
			case "malformed-candidate":
				if err := os.WriteFile(filepath.Join(candidate, "version.txt"), []byte("../../bad"), 0o644); err != nil {
					t.Fatal(err)
				}
			case "existing-backup":
				old := filepath.Join(lib, "millennium.bak_2.0.0")
				if err := os.Mkdir(old, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(old, "sentinel"), []byte("older backup"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("MOCK_LIB_DIR", lib)
			err := rollbackPlatform("millennium.bak_1.0.0", Options{Quiet: true})
			reject := scenario == "missing-candidate" || scenario == "malformed-candidate"
			if (err != nil) != reject {
				t.Fatalf("rollback error=%v", err)
			}
			if b, err := os.ReadFile(sentinel); err != nil || string(b) != "keep" {
				t.Fatal("outside directory changed")
			}
			if reject {
				if b, err := os.ReadFile(filepath.Join(active, "version.txt")); err != nil || string(b) != version {
					t.Fatal("invalid candidate changed active install")
				}
				if st, err := os.Stat(filepath.Join(candidate, "libmillennium_pvs64")); err != nil || st.Mode().Perm() != 0o644 {
					t.Fatal("invalid candidate changed backup permissions")
				}
			}
			if scenario == "existing-backup" {
				if b, err := os.ReadFile(filepath.Join(lib, "millennium.bak_2.0.0", "sentinel")); err != nil || string(b) != "older backup" {
					t.Fatal("overwrote older backup")
				}
			}
		})
	}
}
