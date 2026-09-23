package diag

import "testing"

func TestDarwinDoctorExcludesLinuxRepairs(t *testing.T) {
	for _, force := range []bool{false, true} {
		for _, step := range DoctorPlan(Report{}, force) {
			switch step.ID {
			case "upgrade_force", "runtime_helpers", "repair_hooks", "flatpak", "sudoers_hint", "linger":
				t.Fatalf("Linux repair on Darwin with force=%v: %s", force, step.ID)
			}
		}
	}
	if ok, detail, _ := checkBinaries(); !ok {
		t.Fatalf("requires Linux binaries: %s", detail)
	}
}
