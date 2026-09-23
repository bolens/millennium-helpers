package diag

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/bolens/millennium-helpers/internal/clientfiles"
	"github.com/bolens/millennium-helpers/internal/repair"
)

func TestRunReadOnly(t *testing.T) {
	t.Setenv("MILLENNIUM_CONFIG_DIR", t.TempDir())
	t.Setenv("MILLENNIUM_CONFIG_FILE", filepath.Join(t.TempDir(), "missing.json"))
	t.Setenv("DIAG_TEST_BYPASS_CHECKS", "1")
	results := RunReadOnly()
	if len(results) < 2 {
		t.Fatalf("expected checks, got %d", len(results))
	}
	out := FormatReport(results)
	if !strings.Contains(out, "Helpers Version") {
		t.Fatalf("%s", out)
	}
}

func TestNeedsLegacy(t *testing.T) {
	if NeedsLegacy(nil) {
		t.Fatal("bare diag should be native")
	}
	if NeedsLegacy([]string{"--json"}) {
		t.Fatal("json should be native")
	}
	if NeedsLegacy([]string{"--share"}) {
		t.Fatal("share should be native")
	}
	if NeedsLegacy([]string{"logs"}) {
		t.Fatal("logs should be native")
	}
	if NeedsLegacy([]string{"doctor", "--dry-run"}) {
		t.Fatal("doctor dry-run should be native")
	}
	if NeedsLegacy([]string{"doctor"}) {
		t.Fatal("live doctor should be native")
	}
	if NeedsLegacy([]string{"logs", "--follow"}) {
		t.Fatal("follow should be native")
	}
}

func TestDoctorPlan(t *testing.T) {
	r := Report{
		BinariesOK: false, HooksOK: false, SkinsDirOK: false,
		FlatpakOK: true, TimerActive: true, SudoersOK: true,
		LingerOK: true, PermissionsOK: true, TaskScheduled: true,
		RuntimeHelpersExecutable: false,
	}
	steps := DoctorPlan(r, false)
	ids := map[string]bool{}
	for _, s := range steps {
		ids[s.ID] = true
	}
	if ids["upgrade_force"] != (runtime.GOOS != "darwin") || !ids["skins_dir"] {
		t.Fatalf("%#v", steps)
	}
	if runtime.GOOS == "linux" && ids["runtime_helpers"] {
		t.Fatalf("runtime helper repair should wait for healthy binaries: %#v", steps)
	}
	r.BinariesOK = true
	steps = DoctorPlan(r, false)
	ids = map[string]bool{}
	for _, s := range steps {
		ids[s.ID] = true
	}
	if runtime.GOOS == "linux" && !ids["runtime_helpers"] {
		t.Fatalf("missing runtime helper repair: %#v", steps)
	}
	if runtime.GOOS != "linux" && ids["runtime_helpers"] {
		t.Fatalf("unexpected Windows runtime helper repair: %#v", steps)
	}
}

func TestFormatJSON(t *testing.T) {
	t.Setenv("DIAG_TEST_BYPASS_CHECKS", "1")
	t.Setenv("MILLENNIUM_CONFIG_DIR", t.TempDir())
	out := FormatJSON(Collect())
	if !strings.HasPrefix(strings.TrimSpace(out), "{") {
		t.Fatalf("%s", out)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["steam_running"]; !ok {
		t.Fatalf("%v", m)
	}
	if _, ok := m["update_channel"]; !ok {
		t.Fatalf("%v", m)
	}
	_, hasRuntimeHelpers := m["runtime_helpers_executable"]
	if (runtime.GOOS == "linux") != hasRuntimeHelpers {
		t.Fatalf("runtime helper field mismatch on %s: %v", runtime.GOOS, m)
	}
}

func TestDoctorDryRun(t *testing.T) {
	r := Report{BinariesOK: false, HooksOK: false, SkinsDirOK: true}
	out := FormatDoctorDryRun(r, false)
	if !strings.Contains(out, "DRY RUN") || (runtime.GOOS != "darwin" && !strings.Contains(out, "upgrade")) {
		t.Fatalf("%s", out)
	}
}

func TestPrintLogsNoSteam(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("STEAM", filepath.Join(t.TempDir(), "nosteam"))
	t.Setenv("MILLENNIUM_STATE_DIR", t.TempDir())
	_ = PrintLogs()
}

func TestRedactAndUpload(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USER", "alice")
	t.Setenv("GITHUB_TOKEN", "ghp_secrettokenvalue1234567890abcd")
	body := home + "/.config has ghp_secrettokenvalue1234567890abcd for alice"
	got := RedactReport(body)
	if strings.Contains(got, home) || strings.Contains(got, "alice") || strings.Contains(got, "ghp_secret") {
		t.Fatalf("not redacted: %s", got)
	}
	if !strings.Contains(got, "[REDACTED]") || !strings.Contains(got, "~") {
		t.Fatalf("%s", got)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method %s", r.Method)
		}
		_, _ = w.Write([]byte("https://paste.rs/abc123"))
	}))
	t.Cleanup(srv.Close)
	prev := pasteEndpoint
	prevDo := httpDo
	t.Cleanup(func() {
		pasteEndpoint = prev
		httpDo = prevDo
	})
	pasteEndpoint = srv.URL
	httpDo = srv.Client().Do

	url, err := UploadPasteRS("hello")
	if err != nil || url != "https://paste.rs/abc123" {
		t.Fatalf("%s %v", url, err)
	}
}

func TestKeepFailedShareDoesNotOverwriteEarlierReport(t *testing.T) {
	stateDir := t.TempDir()
	t.Setenv("MILLENNIUM_STATE_DIR", stateDir)

	first := keepFailedShare("first report")
	second := keepFailedShare("second report")
	if first == "" || second == "" {
		t.Fatalf("failed paths: first=%q second=%q", first, second)
	}
	if first == second {
		t.Fatalf("second report overwrote %q", first)
	}
	got, err := os.ReadFile(first)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "first report" {
		t.Fatalf("first report = %q", got)
	}
}

func TestDoctorRepairsRuntimePermissions(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux runtime helpers")
	}
	lib := t.TempDir()
	t.Setenv("MOCK_LIB_DIR", lib)
	t.Setenv("MILLENNIUM_CONFIG_DIR", t.TempDir())
	root := filepath.Join(lib, "millennium")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	names := []string{"libmillennium_bootstrap_x86.so", "libmillennium_bootstrap_hhx64.so", "libmillennium_x86.so", "libmillennium_hhx64.so", "libmillennium_pvs64", "libmillennium_luavm_x86", "version.txt"}
	var sums strings.Builder
	for _, name := range names {
		data := []byte("fixture")
		if err := os.WriteFile(filepath.Join(root, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&sums, "%x  %s\n", sha256.Sum256(data), name)
	}
	if err := os.WriteFile(filepath.Join(root, "checksums.txt"), []byte(sums.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	ok, detail, _ := checkBinaries()
	if !ok {
		t.Fatalf("fixture integrity failed: %s", detail)
	}
	if !strings.Contains(detail, "vfixture") {
		t.Fatalf("diagnosis inspected the wrong install root: %s", detail)
	}
	r := Report{BinariesOK: ok, RuntimeHelpersExecutable: repair.RuntimeHelpersExecutable(), HooksOK: true, FlatpakOK: true, TimerActive: true, SudoersOK: true, LingerOK: true, PermissionsOK: true, SkinsDirOK: true}
	if r.RuntimeHelpersExecutable {
		t.Fatal("non-executable helpers reported healthy")
	}
	if !strings.Contains(FormatJSON(r), `"runtime_helpers_executable": false`) {
		t.Fatal("JSON omitted failed permission check")
	}
	if !strings.Contains(FormatDoctorDryRun(r, false), "restore executable modes") {
		t.Fatal("dry run omitted permission fix")
	}
	if repair.RuntimeHelpersExecutable() {
		t.Fatal("diagnosis changed modes")
	}
	steps := DoctorPlan(r, false)
	if len(steps) != 1 || steps[0].ID != "runtime_helpers" {
		t.Fatalf("unexpected repairs: %#v", steps)
	}
	if err := applyDoctorStep(steps[0], r, Options{}); err != nil {
		t.Fatal(err)
	}
	r.RuntimeHelpersExecutable = repair.RuntimeHelpersExecutable()
	if !r.RuntimeHelpersExecutable {
		t.Fatal("doctor did not restore executable permissions")
	}
	if steps := DoctorPlan(r, false); len(steps) != 0 {
		t.Fatalf("repair not idempotent: %#v", steps)
	}
	if ok, detail, _ := checkBinaries(); !ok {
		t.Fatalf("repair changed file integrity: %s", detail)
	}
	if err := os.Remove(filepath.Join(root, "libmillennium_luavm_x86")); err != nil {
		t.Fatal(err)
	}
	r.BinariesOK, _, _ = checkBinaries()
	r.RuntimeHelpersExecutable = repair.RuntimeHelpersExecutable()
	steps = DoctorPlan(r, false)
	if len(steps) != 1 || steps[0].ID != "upgrade_force" {
		t.Fatalf("missing Lua must select reinstall: %#v", steps)
	}
}

func TestDoctorRequiresReadableIntegrityBeforeReinstall(t *testing.T) {
	if runtime.GOOS != "linux" || os.Geteuid() == 0 {
		t.Skip("unprivileged Linux access check")
	}
	lib := t.TempDir()
	t.Setenv("MOCK_LIB_DIR", lib)
	root := filepath.Join(lib, "millennium")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"version.txt", "libmillennium_bootstrap_x86.so", "libmillennium_bootstrap_hhx64.so", "libmillennium_x86.so", "libmillennium_hhx64.so", "libmillennium_pvs64", "libmillennium_luavm_x86"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("fixture"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := clientfiles.WriteChecksums(root); err != nil {
		t.Fatal(err)
	}
	lua := filepath.Join(root, "libmillennium_luavm_x86")
	if err := os.Chmod(lua, 0); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(lua, 0o644) }()
	ok, detail, err := checkBinaries()
	if ok || !errors.Is(err, os.ErrPermission) || strings.Contains(detail, "Corrupted") {
		t.Fatalf("incorrect access diagnosis: %v %s %v", ok, detail, err)
	}
	for _, force := range []bool{false, true} {
		steps := DoctorPlan(Report{BinariesNeedPrivilege: true}, force)
		found := false
		for _, step := range steps {
			if step.ID == "upgrade_force" || step.ID == "runtime_helpers" {
				t.Fatalf("unverified integrity selects %s", step.ID)
			}
			if step.ID == "binaries_access" {
				found = true
				if applyDoctorStep(step, Report{}, Options{}) == nil {
					t.Fatal("access error reported success")
				}
			}
		}
		if !found {
			t.Fatal("missing privileged verification instruction")
		}
	}
}

func TestDoctorOwnershipPreservesCache(t *testing.T) {
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))
	steam := filepath.Join(home, "Steam")
	t.Setenv("STEAM", steam)
	t.Setenv("STEAM_PATH", "")
	cache := filepath.Join(steam, "config", "htmlcache")
	for _, path := range []string{cache, filepath.Join(steam, "millennium")} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	sentinel := filepath.Join(cache, "live-cache")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := applyDoctorStep(DoctorStep{ID: "permissions"}, Report{}, Options{}); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(sentinel); err != nil || string(b) != "keep" {
		t.Fatal("ownership step cleared Steam cache")
	}
}
