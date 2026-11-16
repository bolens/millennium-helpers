package upgrade

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/bolens/millennium-helpers/internal/dirlock"
)

func TestCompetingActivationIsRejectedBeforeMutation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix install root fixture")
	}
	lib := t.TempDir()
	t.Setenv("MOCK_LIB_DIR", lib)
	staged := filepath.Join(lib, "millennium.tmp")
	if err := os.Mkdir(staged, 0755); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(staged, "in-flight")
	if err := os.WriteFile(sentinel, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	lock, err := dirlock.Acquire(lib)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := lock.Close(); err != nil {
			t.Error(err)
		}
	}()
	for _, err := range []error{installPlatform("/missing/archive", "1.0.0", Options{}), rollbackPlatform("missing", Options{})} {
		if !errors.Is(err, dirlock.ErrBusy) {
			t.Fatalf("err=%v", err)
		}
	}
	if got, err := os.ReadFile(sentinel); err != nil || string(got) != "keep" {
		t.Fatal("competing installer deleted staged work")
	}
}

func TestVersionInferencePreservesPrereleasesAndIsRepeatable(t *testing.T) {
	for _, tc := range []struct{ path, fallback, want string }{
		{"millennium-v3.6.3-beta.1-linux-x86_64.tar.gz", "", "3.6.3-beta.1"},
		{"custom.zip", "v3.6.3", "3.6.3"},
		{"custom.zip", "", "local"},
	} {
		for i := 0; i < 2; i++ {
			if got := InferVersion(tc.path, tc.fallback); got != tc.want {
				t.Fatalf("%s: %q", tc.path, got)
			}
		}
	}
}

func TestTimestampedBackupsDoNotCollideOrEscape(t *testing.T) {
	root := t.TempDir()
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	first, err := reserveTimestampedBackup(root, "3.6.3", now)
	if err != nil {
		t.Fatal(err)
	}
	second, err := reserveTimestampedBackup(root, "3.6.3", now)
	if err != nil {
		t.Fatal(err)
	}
	if first == second || filepath.Base(second) != "3.6.3-2_20261116120000" {
		t.Fatalf("backups=%s,%s", first, second)
	}
	invalid, err := reserveTimestampedBackup(root, "../../outside", now)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(invalid) != root || !isWindowsBackupName(filepath.Base(invalid)) {
		t.Fatalf("unsafe or unrecognized backup=%s", invalid)
	}
}
