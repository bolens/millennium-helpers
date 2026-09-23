package repair

import "testing"

func TestRepairWithoutLinuxRuntimeHelpers(t *testing.T) {
	t.Setenv("MOCK_LIB_DIR", t.TempDir())
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("STEAM", t.TempDir())
	if contains(FormatPlan(nil, true), "runtime helpers") {
		t.Fatal("Darwin preview includes Linux runtime helpers")
	}
	if code := RunCLI(false, true, true, true); code != 0 {
		t.Fatalf("Darwin repair without Linux runtime helpers returned %d", code)
	}
}
