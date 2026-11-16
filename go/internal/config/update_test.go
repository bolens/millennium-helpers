package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bolens/millennium-helpers/internal/dirlock"
)

func TestUpdateContentionAndPreservation(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("MILLENNIUM_CONFIG_FILE", filepath.Join(dir, "config.json"))
	entered, release := make(chan struct{}), make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- Update(func(data Data) error { close(entered); <-release; return Set(data, "backup_limit", "9") })
	}()
	<-entered
	ran := false
	err := Update(func(Data) error { ran = true; return nil })
	close(release)
	firstErr := <-done
	if !errors.Is(err, dirlock.ErrBusy) || ran || firstErr != nil {
		t.Fatalf("contention=%v ran=%v first=%v", err, ran, firstErr)
	}
	if err := Update(func(data Data) error { return Set(data, "update_channel", "beta") }); err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil || Get(got, "backup_limit") != "9" || Get(got, "update_channel") != "beta" {
		t.Fatalf("data=%v err=%v", got, err)
	}
	before, err := os.ReadFile(Path())
	if err != nil {
		t.Fatal(err)
	}
	failure := errors.New("rejected change")
	if err := Update(func(data Data) error { data["backup_limit"] = 99; return failure }); !errors.Is(err, failure) {
		t.Fatal(err)
	}
	after, err := os.ReadFile(Path())
	if err != nil || string(after) != string(before) {
		t.Fatal("failed update altered config")
	}
}

func captureCLI(t *testing.T, args []string) (int, string) {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "output")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Error(err)
		}
	}()
	stdout, stderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = file, file
	defer func() { os.Stdout, os.Stderr = stdout, stderr }()
	code := RunCLI(args)
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	return code, string(b)
}

func TestConfigMutationArityAndSecretOutput(t *testing.T) {
	t.Setenv("MILLENNIUM_CONFIG_FILE", filepath.Join(t.TempDir(), "config.json"))
	for _, args := range [][]string{{"set", "github_token"}, {"set", "github_token", "x", "extra"}, {"get", "backup_limit", "extra"}, {"list", "extra"}} {
		if code, _ := captureCLI(t, args); code == 0 {
			t.Fatalf("accepted %q", args)
		}
	}
	token := "fixture-credential-do-not-echo"
	for _, args := range [][]string{{"set", "github_token", token, "--dry-run"}, {"set", "github_token", token}} {
		code, out := captureCLI(t, args)
		if code != 0 || strings.Contains(out, token) || !strings.Contains(out, "[REDACTED]") {
			t.Fatalf("code=%d secret redacted=%v", code, !strings.Contains(out, token))
		}
	}
	data, err := Load()
	if err != nil || Get(data, "github_token") != token {
		t.Fatal("redaction altered saved credential")
	}
	if code, _ := captureCLI(t, []string{"set", "github_token", ""}); code != 0 {
		t.Fatal("explicit clearing no longer supported")
	}
}
