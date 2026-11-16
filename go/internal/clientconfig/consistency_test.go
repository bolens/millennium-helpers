package clientconfig

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/bolens/millennium-helpers/internal/dirlock"
)

func TestMutationsRejectConcurrentWriter(t *testing.T) {
	path, _, _ := fixture(t)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := dirlock.Acquire(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := lock.Close(); err != nil {
			t.Error(err)
		}
	}()
	for _, change := range []func() error{
		func() error { return SetPlugins([]string{"beta"}, true, false) },
		func() error { return SetTheme("Demo", false) },
		func() error { return DisableTheme("", false) },
	} {
		if err := change(); !errors.Is(err, dirlock.ErrBusy) {
			t.Fatalf("err=%v", err)
		}
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(before) {
		t.Fatal("contender overwrote configuration")
	}
}

func TestNullConfigAndBuiltInTheme(t *testing.T) {
	path, _, themes := fixture(t)
	if err := os.Remove(filepath.Join(themes, "Steam")); err != nil {
		t.Fatal(err)
	}
	if err := SetTheme("Demo", false); err != nil {
		t.Fatal(err)
	}
	if err := DisableTheme("Demo", false); err != nil {
		t.Fatal(err)
	}
	got, err := Show()
	if err != nil || got.ActiveTheme != "Steam" {
		t.Fatalf("summary=%+v err=%v", got, err)
	}
	if err := os.WriteFile(path, []byte("null"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Show(); err == nil {
		t.Fatal("null config accepted")
	}
	if err := SetPlugins([]string{"alpha"}, true, false); err == nil {
		t.Fatal("null config mutation accepted")
	}
}

func TestDisableErrorsDoesNotDisableChangedTheme(t *testing.T) {
	path, _, _ := fixture(t)
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = DisableFlagged([]Finding{{Kind: "theme", Name: "Demo", Active: true}}, "themes", false)
	if err == nil {
		t.Fatal("stale error report disabled a different active theme")
	}
	after, readErr := os.ReadFile(path)
	if readErr != nil || string(after) != string(before) {
		t.Fatal("stale report altered configuration")
	}
}
