package diag

import "testing"

func TestAmbiguousDiagnosticOptionsFailBeforeExecution(t *testing.T) {
	for _, args := range [][]string{{"doctor", "logs"}, {"doctor", "--share"}, {"logs", "--json"}, {"doctor", "--json"}, {"--force"}, {"--dry-run"}, {"--yes"}, {"--follow"}} {
		if _, err := ParseArgs(args); err == nil {
			t.Fatalf("accepted %q", args)
		}
	}
	for _, args := range [][]string{{"doctor", "--dry-run", "--yes"}, {"logs", "--follow"}, {"--share", "--json"}, {"doctor", "logs", "--help"}} {
		if _, err := ParseArgs(args); err != nil {
			t.Fatalf("rejected %q: %v", args, err)
		}
	}
}
