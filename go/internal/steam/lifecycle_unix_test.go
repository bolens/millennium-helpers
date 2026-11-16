//go:build unix

package steam

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestShutdownMustVerifyExit(t *testing.T) {
	if err := finishShutdown(func() bool { return true }, func(time.Duration) {}, func() error { return errors.New("kill failed") }); err == nil {
		t.Fatal("reported running client stopped")
	}
	running := true
	if err := finishShutdown(func() bool { return running }, func(time.Duration) {}, func() error { running = false; return nil }); err != nil {
		t.Fatal(err)
	}
}

func TestRelaunchRetainsCapturedEnvironment(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux shell launcher")
	}
	t.Setenv("TEST_SUITE_RUN", "")
	t.Setenv("PSTESTS", "")
	dir := t.TempDir()
	t.Setenv("MILLENNIUM_STATE_DIR", dir)
	result := filepath.Join(dir, "result")
	sentinel := filepath.Join(dir, "must-not-exist")
	argv := []string{"-srt-logger-opened", "space arg", "quote'arg", "$(touch " + sentinel + ")", "; touch " + sentinel + " #", ""}
	var quoted []string
	for _, arg := range argv {
		quoted = append(quoted, shellQuote(arg))
	}
	display := ":42's display"
	exe := filepath.Join(dir, "steam")
	if err := os.WriteFile(exe, []byte("#!/bin/sh\nprintf '%s\\000' \"$DISPLAY\" \"$@\" > "+shellQuote(result)+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("DISPLAY", "outer")
	name, home, err := TargetUser()
	if err != nil {
		t.Fatal(err)
	}
	state := RelaunchStateFile(name, home)
	if err := writeRelaunchEnv(state, map[string]string{"DISPLAY": display}, strings.Join(quoted, " "), false); err != nil {
		t.Fatal(err)
	}
	attempted, err := relaunchFromState(name, home, runSteamCmdWithEnv)
	if !attempted || err != nil {
		t.Fatalf("relaunch %v %v", attempted, err)
	}
	expected := strings.Join(append([]string{display}, argv[1:]...), "\x00") + "\x00"
	if b, err := os.ReadFile(result); err != nil || string(b) != expected {
		t.Fatalf("captured environment/arguments lost: %q %v", b, err)
	}
	if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
		t.Fatal("captured arguments executed as shell commands", err)
	}
	if _, err := os.Stat(state); !os.IsNotExist(err) {
		t.Fatal("successful relaunch retained state")
	}
	if err := writeRelaunchEnv(state, map[string]string{}, "", false); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir) // Steam exists, but no shell/runuser: dispatch fails safely.
	if _, err := relaunchFromState(name, home, runSteamCmdWithEnv); err == nil {
		t.Fatal("missing launcher reported success")
	}
	if _, err := os.Stat(state); err != nil {
		t.Fatal("failed relaunch lost recovery state")
	}
}

func TestLauncherFailureAndReadiness(t *testing.T) {
	failure := errors.New("launcher rejected executable")
	done := make(chan error, 1)
	done <- failure
	if err := waitForSteamStart(done, func() bool { return false }, time.Second); !errors.Is(err, failure) {
		t.Fatalf("lost launch error: %v", err)
	}
	exited := make(chan error, 1)
	exited <- nil
	if err := waitForSteamStart(exited, func() bool { return false }, time.Millisecond); err == nil {
		t.Fatal("launcher exit without Steam reported success")
	}
	if err := waitForSteamStart(make(chan error), func() bool { return true }, time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestFailedExecutableRetainsRelaunchState(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux launcher")
	}
	t.Setenv("TEST_SUITE_RUN", "")
	dir := t.TempDir()
	t.Setenv("MILLENNIUM_STATE_DIR", dir)
	if err := os.WriteFile(filepath.Join(dir, "steam"), []byte("#!/nonexistent/interpreter\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	name, home, err := TargetUser()
	if err != nil {
		t.Fatal(err)
	}
	state := RelaunchStateFile(name, home)
	if err := writeRelaunchEnv(state, nil, "", false); err != nil {
		t.Fatal(err)
	}
	// Execute the same command synchronously to make the dispatch failure
	// deterministic even when an unrelated host Steam process is running.
	if _, err := relaunchFromState(name, home, runSteamCmdWithEnv); err == nil {
		t.Fatal("invalid executable reported successful relaunch")
	}
	if _, err := os.Stat(state); err != nil {
		t.Fatal("lost failed relaunch state")
	}
}

func TestCaptureEnvPreservesEmptyArguments(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux proc capture")
	}
	root := t.TempDir()
	t.Setenv("MOCK_PROC", root)
	t.Setenv("MILLENNIUM_STATE_DIR", t.TempDir())
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "pgrep"), []byte("#!/bin/sh\nprintf '4242\\n'\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	process := filepath.Join(root, "4242")
	if err := os.Mkdir(process, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(process, "cmdline"), []byte("steam\x00-srt-logger-opened\x00\x00"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(process, "environ"), []byte("DISPLAY=:fixture\x00"), 0644); err != nil {
		t.Fatal(err)
	}
	name, home, err := TargetUser()
	if err != nil {
		t.Fatal(err)
	}
	if err := CaptureEnv(name, home); err != nil {
		t.Fatal(err)
	}
	vals, err := ParseRelaunchEnv(RelaunchStateFile(name, home))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := vals["STEAM_ARGS"], "'-srt-logger-opened' '' "; got != want {
		t.Fatalf("args=%q, want %q", got, want)
	}
}

func TestMalformedRelaunchArgumentsPreserveState(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux relaunch arguments")
	}
	t.Setenv("MILLENNIUM_STATE_DIR", t.TempDir())
	name, home, err := TargetUser()
	if err != nil {
		t.Fatal(err)
	}
	state := RelaunchStateFile(name, home)
	if err := writeRelaunchEnv(state, nil, "'unterminated", false); err != nil {
		t.Fatal(err)
	}
	called := false
	attempted, err := relaunchFromState(name, home, func(string, string, map[string]string) error { called = true; return nil })
	if err == nil || attempted || called {
		t.Fatalf("malformed state launched: %v %v %v", attempted, called, err)
	}
	if _, err := os.Stat(state); err != nil {
		t.Fatal("lost recovery state", err)
	}
}
