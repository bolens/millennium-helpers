package diag

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/bolens/millennium-helpers/internal/repair"
)

func TestSteamRT3DoctorPlan(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux compatibility")
	}
	r := Report{BinariesOK: true, SteamRT3: true, HooksOK: false}
	steps := DoctorPlan(r, false)
	if len(steps) < 2 || steps[0].ID != "upgrade_force" {
		t.Fatalf("incompatible client not upgraded first: %#v", steps)
	}
	r.SteamRT3Compatible = true
	for _, s := range DoctorPlan(r, false) {
		if s.ID == "upgrade_force" {
			t.Fatal("compatible client forced upgrade")
		}
	}
	r.SteamRT3Compatible = false
	out := FormatReportFromCollect(r)
	if !strings.Contains(out, "SteamRT3 Compatibility") || !strings.Contains(out, "Bootstrap Hooks") {
		t.Fatal(out)
	}
	if !strings.Contains(FormatJSON(r), `"steamrt3_compatible": false`) {
		t.Fatal("incompatible client omitted from JSON")
	}
	if !strings.Contains(FormatJSON(r), `"hooks_ok": false`) {
		t.Fatal("unhealthy hook omitted from JSON")
	}
}

func TestSteamRT3HookDiagnostics(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux hooks")
	}
	home, lib := t.TempDir(), t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("SUDO_USER", "")
	t.Setenv("MOCK_LIB_DIR", lib)
	root := filepath.Join(home, ".local/share/Steam")
	if err := os.MkdirAll(filepath.Join(root, "steamrt64"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "steamrt64/steam"), nil, 0755); err != nil {
		t.Fatal(err)
	}
	lib = repair.MillenniumLibRoot()
	if err := os.MkdirAll(lib, 0755); err != nil {
		t.Fatal(err)
	}
	for _, p := range repair.HooksAt(root, lib) {
		if err := os.WriteFile(repair.HostHookTarget(p.Target), nil, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := repair.InstallHooksAt(root, lib); err != nil {
		t.Fatal(err)
	}
	if ok, _ := checkHooks(); !ok {
		t.Fatal("container link rejected on host")
	}
	hook := filepath.Join(root, "steamrt64/libXtst.so.6")
	if err := os.Remove(hook); err != nil {
		t.Fatal(err)
	}
	if ok, _ := checkHooks(); ok {
		t.Fatal("missing SteamRT3 hook reported healthy")
	}
	if err := os.WriteFile(hook, nil, 0644); err != nil {
		t.Fatal(err)
	}
	if ok, _ := checkHooks(); ok {
		t.Fatal("regular file reported as Millennium hook")
	}
	if err := os.Remove(hook); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("/foreign/bootstrap.so", hook); err != nil {
		t.Fatal(err)
	}
	if ok, _ := checkHooks(); ok {
		t.Fatal("foreign hook reported healthy")
	}
}
