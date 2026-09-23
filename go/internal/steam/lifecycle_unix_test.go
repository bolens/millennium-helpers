//go:build unix

package steam

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
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
	exe := filepath.Join(dir, "steam")
	if err := os.WriteFile(exe, []byte("#!/bin/sh\nprintf '%s' \"$DISPLAY\" > "+shellQuote(result)+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("DISPLAY", "outer")
	name, home, err := TargetUser()
	if err != nil {
		t.Fatal(err)
	}
	state := RelaunchStateFile(name, home)
	if err := writeRelaunchEnv(state, map[string]string{"DISPLAY": ":42"}, "", false); err != nil {
		t.Fatal(err)
	}
	attempted, err := relaunchFromState(name, home, runSteamCmdWithEnv)
	if !attempted || err != nil {
		t.Fatalf("relaunch %v %v", attempted, err)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if b, err := os.ReadFile(result); err == nil && string(b) == ":42" {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if b, err := os.ReadFile(result); err != nil || string(b) != ":42" {
		t.Fatalf("captured display lost: %q %v", b, err)
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
