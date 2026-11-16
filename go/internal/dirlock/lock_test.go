package dirlock

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestLockAcrossProcesses(t *testing.T) {
	if dir := os.Getenv("MILLENNIUM_TEST_LOCK_DIR"); dir != "" {
		lock, err := Acquire(dir)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("locked")
		var b [1]byte
		_, _ = os.Stdin.Read(b[:])
		// Exit without closing the lock to verify kernel cleanup.
		_ = lock
		os.Exit(0)
	}
	dir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestLockAcrossProcesses$")
	cmd.Env = append(os.Environ(), "MILLENNIUM_TEST_LOCK_DIR="+dir)
	in, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = in.Close()
		if cmd.ProcessState == nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	})
	line, err := bufio.NewReader(out).ReadString('\n')
	if err != nil || line != "locked\n" {
		t.Fatalf("child readiness=%q err=%v", line, err)
	}
	if lock, err := Acquire(dir); !errors.Is(err, ErrBusy) {
		if lock != nil {
			_ = lock.Close()
		}
		t.Fatalf("competing lock=%v", err)
	}
	_ = in.Close()
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}
	lock, err := Acquire(dir)
	if err != nil {
		t.Fatalf("lock not released: %v", err)
	}
	if err := lock.Close(); err != nil {
		t.Fatal(err)
	}
}
