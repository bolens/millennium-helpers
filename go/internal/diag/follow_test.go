package diag

import (
	"bytes"
	"io"
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

func TestFollowFilteredMissing(t *testing.T) {
	if code := followFiltered(filepath.Join(t.TempDir(), "missing.log"), logFilterParts(), 100); code != 1 {
		t.Fatalf("code=%d", code)
	}
}

func TestFollowRotationPartialAndWriteFailure(t *testing.T) {
	for _, scenario := range []string{"rotate", "truncate", "partial", "writer-error"} {
		t.Run(scenario, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "log")
			if err := os.WriteFile(path, []byte("BOOTSTRAP old\n"), 0600); err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			cycle := 0
			options := followOptions{maxCycles: 2, output: &output, cycleHook: func() {
				cycle++
				switch scenario {
				case "rotate":
					if cycle == 1 {
						if err := os.Rename(path, path+".old"); err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(path, []byte("BOOTSTRAP new replacement log is larger\n"), 0600); err != nil {
							t.Fatal(err)
						}
					}
				case "truncate":
					if cycle == 1 {
						if err := os.WriteFile(path, []byte("BOOTSTRAP n\n"), 0600); err != nil {
							t.Fatal(err)
						}
					}
				case "partial":
					f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
					if err != nil {
						t.Fatal(err)
					}
					text := "BOOTSTRAP part"
					if cycle == 2 {
						text = "ial\n"
					}
					if _, err := f.WriteString(text); err != nil {
						t.Fatal(err)
					}
					_ = f.Close()
				}
			}}
			if scenario == "writer-error" {
				options.output = failingLogWriter{}
			}
			code := followWithOptions(path, logFilterParts(), 100, options)
			if scenario == "writer-error" {
				if code != 1 {
					t.Fatal("writer failure ignored")
				}
				return
			}
			expected := map[string]string{"rotate": "BOOTSTRAP new replacement log is larger\n", "truncate": "BOOTSTRAP n\n", "partial": "BOOTSTRAP partial\n"}[scenario]
			if code != 0 || output.String() != "BOOTSTRAP old\n"+expected {
				t.Fatalf("code=%d output=%q", code, output.String())
			}
		})
	}
}

type failingLogWriter struct{}

func (failingLogWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

type appendOnFirstLine struct {
	bytes.Buffer
	append func()
}

func (w *appendOnFirstLine) Write(b []byte) (int, error) {
	n, err := w.Buffer.Write(b)
	if w.append != nil {
		appendNow := w.append
		w.append = nil
		appendNow()
	}
	return n, err
}

func TestFollowDoesNotLoseAppendDuringInitialOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "log")
	if err := os.WriteFile(path, []byte("BOOTSTRAP initial\n"), 0600); err != nil {
		t.Fatal(err)
	}
	output := &appendOnFirstLine{append: func() {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.WriteString("BOOTSTRAP during initial output\n"); err != nil {
			t.Fatal(err)
		}
		_ = f.Close()
	}}
	if code := followWithOptions(path, logFilterParts(), 100, followOptions{maxCycles: 1, output: output}); code != 0 {
		t.Fatal(code)
	}
	if got := output.String(); got != "BOOTSTRAP initial\nBOOTSTRAP during initial output\n" {
		t.Fatalf("output=%q", got)
	}
}

func TestTailRetainsOnlyRequestedLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "log")
	if err := os.WriteFile(path, []byte("one\r\ntwo\nthree\nfour"), 0600); err != nil {
		t.Fatal(err)
	}
	lines, err := tailLines(path, 2)
	if err != nil || strings.Join(lines, "|") != "three|four" {
		t.Fatalf("lines=%q err=%v", lines, err)
	}
	lines, err = tailLines(path, 0)
	if err != nil || len(lines) != 0 {
		t.Fatalf("zero tail=%q err=%v", lines, err)
	}
}
