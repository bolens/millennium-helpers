//go:build unix

package upgrade

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSudoChildFailureIsNotRetried(t *testing.T) {
	dir := t.TempDir()
	calls := filepath.Join(dir, "calls")
	t.Setenv("MILLENNIUM_TEST_SUDO_LOG", calls)
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$MILLENNIUM_TEST_SUDO_LOG\"\nexit 17\n"
	if err := os.WriteFile(filepath.Join(dir, "sudo"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	input, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdin
	os.Stdin = input
	t.Cleanup(func() { os.Stdin = old; _ = input.Close() })
	code := runSudo([]string{"/bin/millennium", "upgrade", "--yes"})
	if code != 17 {
		t.Fatalf("exit code=%d, want child status 17", code)
	}
	b, err := os.ReadFile(calls)
	if err != nil || string(b) != "-n /bin/millennium upgrade --yes\n" {
		t.Fatalf("mutation was retried or argv changed: %q %v", b, err)
	}
}
