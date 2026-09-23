package repair

import (
	"github.com/bolens/millennium-helpers/internal/usercontext"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestPlanAndFormat(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, "cfg"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(home, "data"))
	steam := filepath.Join(home, ".local", "share", "Steam")
	mill := filepath.Join(steam, "millennium")
	if err := os.MkdirAll(mill, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("STEAM", steam)

	targets, err := Plan()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, tg := range targets {
		if tg.Path == mill {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected millennium path in %#v", targets)
	}
	out := FormatPlan(targets, true)
	if !contains(out, "skip-theme") && !contains(out, "Skipping theme") {
		t.Fatalf("%s", out)
	}
	if !contains(out, "DRY RUN") {
		t.Fatalf("missing dry-run header: %s", out)
	}
}

func TestPlanHooksMentionsBootstrap(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	lib := filepath.Join(home, "lib")
	t.Setenv("MOCK_LIB_DIR", lib)
	steam := filepath.Join(home, ".local", "share", "Steam")
	if err := os.MkdirAll(filepath.Join(steam, "ubuntu12_32"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(lib, "millennium"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("STEAM", steam)
	plans := PlanHooks()
	out := FormatPlan(nil, true)
	if len(plans) > 0 && !contains(out, "Would link hook") {
		t.Fatalf("expected hook dry-run lines: %s", out)
	}
}

func TestApplyHtmlcache(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	cache := filepath.Join(root, "htmlcache")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cache, "blob"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	mill := filepath.Join(root, "millennium")
	if err := os.MkdirAll(mill, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Apply([]Target{
		{Path: mill, Kind: "chown"},
		{Path: cache, Kind: "htmlcache"},
	}, true); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(cache)
	if err != nil || len(entries) != 0 {
		t.Fatalf("cache %#v err=%v", entries, err)
	}
}

func TestRuntimeHelpersExecutableAndRepair(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Unix executable modes are not meaningful on Windows")
	}
	lib := filepath.Join(t.TempDir(), "lib")
	root := filepath.Join(lib, "millennium")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOCK_LIB_DIR", lib)
	for _, name := range []string{"libmillennium_pvs64", "libmillennium_luavm_x86"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("helper"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if RuntimeHelpersExecutable() {
		t.Fatal("non-executable runtime helpers reported healthy")
	}
	if err := EnsureRuntimeHelpersExecutable(); err != nil {
		t.Fatal(err)
	}
	if !RuntimeHelpersExecutable() {
		t.Fatal("repaired runtime helpers still reported unhealthy")
	}
	for _, name := range []string{"libmillennium_pvs64", "libmillennium_luavm_x86"} {
		st, err := os.Stat(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		if st.Mode().Perm() != 0o755 {
			t.Fatalf("%s mode = %o, want 755", name, st.Mode().Perm())
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestRepairCLIRestoresRuntimeModes(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux runtime helpers")
	}
	lib := t.TempDir()
	t.Setenv("MOCK_LIB_DIR", lib)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	t.Setenv("STEAM", t.TempDir())
	root := filepath.Join(lib, "millennium")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"libmillennium_pvs64", "libmillennium_luavm_x86", "libmillennium_bootstrap_x86.so", "libmillennium_bootstrap_hhx64.so"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("helper"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if !contains(FormatPlan(nil, true), "runtime helpers") {
		t.Fatal("dry run omits runtime permission repair")
	}
	if code := RunCLI(true, true, true, true); code != 0 {
		t.Fatalf("dry run: %d", code)
	}
	if RuntimeHelpersExecutable() {
		t.Fatal("dry run changed permissions")
	}
	if code := RunCLI(false, true, true, true); code != 0 {
		t.Fatalf("repair: %d", code)
	}
	if !RuntimeHelpersExecutable() {
		t.Fatal("repair left runtime helpers non-executable")
	}
}

func TestPlanUsesResolvedCallerDirectories(t *testing.T) {
	home, processHome := t.TempDir(), t.TempDir()
	t.Setenv("HOME", processHome)
	t.Setenv("STEAM", t.TempDir())
	t.Setenv("STEAM_PATH", "")
	c := usercontext.Context{Home: home, ConfigHome: filepath.Join(home, ".config"), DataHome: filepath.Join(home, ".local", "share")}
	for _, base := range []string{c.ConfigHome, c.DataHome, filepath.Join(processHome, ".config")} {
		if err := os.MkdirAll(filepath.Join(base, "millennium"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	targets := planFor(c)
	if len(targets) != 2 {
		t.Fatalf("wrong targets: %#v", targets)
	}
	for _, target := range targets {
		if target.Path != filepath.Join(c.ConfigHome, "millennium") && target.Path != filepath.Join(c.DataHome, "millennium") {
			t.Fatalf("process path selected: %s", target.Path)
		}
	}
}
