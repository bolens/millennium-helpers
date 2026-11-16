package diag

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUnknownChecksAreNotReportedAsHealthy(t *testing.T) {
	report := Report{BinariesOK: true, ScriptsUpToDate: true, UnverifiedChecks: []string{"scripts_up_to_date"}}
	var got map[string]any
	if err := json.Unmarshal([]byte(FormatJSON(report)), &got); err != nil {
		t.Fatal(err)
	}
	if value, present := got["scripts_up_to_date"]; !present || value != false {
		t.Fatalf("unmeasured freshness=%v present=%v", value, present)
	}
	if got["binaries_ok"] != true {
		t.Fatal("measured check lost")
	}
	if !strings.Contains(FormatReportFromCollect(report), "Not verified") {
		t.Fatal("human report hides unknown checks")
	}
}
