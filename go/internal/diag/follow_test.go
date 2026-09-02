package diag

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLineMatchesFilter(t *testing.T) {
	parts := []string{"millennium", "bootstrap"}
	if !lineMatchesFilter("MILLENNIUM ready", parts) {
		t.Fatal("expected match")
	}
	if lineMatchesFilter("unrelated noise", parts) {
		t.Fatal("expected no match")
	}
}

func TestFollowFilteredAppend(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "webhelper.txt")
	if err := os.WriteFile(p, []byte("noise\nMILLENNIUM start\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	followMaxCycles = 1
	followOutput = &output
	followCycleHook = func() {
		f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.WriteString("still noise\nBOOTSTRAP ok\n"); err != nil {
			_ = f.Close()
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		followMaxCycles = 0
		followCycleHook = nil
		followOutput = os.Stdout
	})

	if code := followFiltered(p, logFilterParts(), 100); code != 0 {
		t.Fatalf("code=%d", code)
	}
	if got := output.String(); !strings.Contains(got, "MILLENNIUM start") || !strings.Contains(got, "BOOTSTRAP ok") {
		t.Fatalf("filtered output missing expected lines: %q", got)
	}
}

func TestFollowLogsMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("STEAM", filepath.Join(home, "nosteam"))
	t.Setenv("MILLENNIUM_STATE_DIR", filepath.Join(home, "state"))
	// Isolate candidates that may resolve via default Steam paths on the host.
	t.Setenv("MILLENNIUM_SKINS_DIR", filepath.Join(home, "skins"))
	code := FollowLogs()
	if code != 1 && code != 0 {
		t.Fatalf("unexpected code %d", code)
	}
	_ = strings.Contains
}
