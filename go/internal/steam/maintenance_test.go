package steam

import (
	"errors"
	"strings"
	"testing"
)

func TestMaintenanceRecovery(t *testing.T) {
	for _, fail := range []string{"", "capture", "close", "work", "resume", "both"} {
		t.Run(fail, func(t *testing.T) {
			var calls []string
			workErr, resumeErr := errors.New("work"), errors.New("resume")
			step := func(name string) func() error {
				return func() error {
					calls = append(calls, name)
					if fail == name || fail == "both" && (name == "work" || name == "resume") {
						if name == "resume" {
							return resumeErr
						}
						return workErr
					}
					return nil
				}
			}
			err := withClosedClient(step("capture"), step("close"), step("resume"), step("work"))
			want := "capture,close,work,resume"
			if fail == "capture" {
				want = "capture"
			}
			if fail == "close" {
				want = "capture,close"
			}
			if strings.Join(calls, ",") != want {
				t.Fatalf("calls=%v", calls)
			}
			if (err == nil) != (fail == "") {
				t.Fatalf("error=%v", err)
			}
			if fail == "both" && (!errors.Is(err, workErr) || !errors.Is(err, resumeErr)) {
				t.Fatal("lost failure during recovery")
			}
		})
	}
}
