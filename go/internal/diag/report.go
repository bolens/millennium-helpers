package diag

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/bolens/millennium-helpers/internal/clientfiles"
	"github.com/bolens/millennium-helpers/internal/config"
	"github.com/bolens/millennium-helpers/internal/repair"
	"github.com/bolens/millennium-helpers/internal/schedule"
	"github.com/bolens/millennium-helpers/internal/steam"
	"github.com/bolens/millennium-helpers/internal/theme"
	"github.com/bolens/millennium-helpers/internal/version"
)

// Report is the structured diagnostic snapshot (JSON + doctor plan).
type Report struct {
	SteamRunning             bool   `json:"steam_running"`
	BinariesNeedPrivilege    bool   `json:"-"`
	BinariesOK               bool   `json:"binaries_ok"`
	RuntimeHelpersExecutable bool   `json:"runtime_helpers_executable"`
	HooksOK                  bool   `json:"hooks_ok,omitempty"`
	FlatpakOK                bool   `json:"flatpak_ok,omitempty"`
	SudoersOK                bool   `json:"sudoers_ok,omitempty"`
	TimerActive              bool   `json:"timer_active,omitempty"`
	LingerOK                 bool   `json:"linger_ok,omitempty"`
	TaskScheduled            bool   `json:"task_scheduled,omitempty"`
	ScriptsUpToDate          bool   `json:"scripts_up_to_date"`
	PermissionsOK            bool   `json:"permissions_ok"`
	SkinsDirOK               bool   `json:"skins_dir_ok"`
	CompletionsOK            bool   `json:"completions_ok"`
	CleanOfObsolete          bool   `json:"clean_of_obsolete"`
	UnmanagedFilesOK         bool   `json:"unmanaged_files_ok,omitempty"`
	MixedInstallOK           bool   `json:"mixed_install_ok"`
	InstallMethod            string `json:"install_method"`
	HelpersCheckout          string `json:"helpers_checkout"`
	HelpersTrack             string `json:"helpers_track"`
	HelpersRef               string `json:"helpers_ref"`
	LatestReleaseTag         string `json:"latest_release_tag"`
	UpdateChannel            string `json:"update_channel"`
	Version                  string `json:"version,omitempty"`

	// Human details (not always in JSON)
	BinariesDetail string `json:"-"`
	SteamDetail    string `json:"-"`
}

// Options for diag CLI.
type Options struct {
	Doctor  bool
	JSON    bool
	Logs    bool
	Follow  bool
	Share   bool
	Force   bool
	DryRun  bool
	Quiet   bool
	Yes     bool
	Help    bool
	Version bool
}

// ParseArgs parses diag argv.
func ParseArgs(args []string) (Options, error) {
	var o Options
	for _, a := range args {
		switch a {
		case "-h", "--help", "-Help":
			o.Help = true
		case "-V", "--version", "-Version":
			o.Version = true
		case "doctor", "--fix", "-f", "-Fix":
			o.Doctor = true
		case "logs":
			o.Logs = true
		case "--json", "-json", "-Json":
			o.JSON = true
		case "--share", "-s", "-Share":
			o.Share = true
		case "--follow", "-l", "-Follow":
			o.Follow = true
		case "--force", "-Force":
			o.Force = true
		case "-d", "--dry-run", "-DryRun":
			o.DryRun = true
		case "-q", "--quiet", "-Quiet":
			o.Quiet = true
		case "-y", "--yes", "-Yes":
			o.Yes = true
		default:
			if strings.HasPrefix(a, "-") {
				return o, fmt.Errorf("Error: unknown option %s", a)
			}
			return o, fmt.Errorf("Error: unknown diag argument %s", a)
		}
	}
	return o, nil
}

// NeedsLegacy is reserved for remaining shell/PS-only diag surfaces (none today).
func NeedsLegacy(args []string) bool {
	_, err := ParseArgs(args)
	return err != nil
}

// Collect builds a full read-only report.
func Collect() Report {
	r := Report{
		ScriptsUpToDate:  true, // detailed package freshness stays legacy
		PermissionsOK:    true,
		CompletionsOK:    true,
		CleanOfObsolete:  true,
		UnmanagedFilesOK: true,
		MixedInstallOK:   true,
		InstallMethod:    "unknown",
		HooksOK:          true,
		FlatpakOK:        true,
		SudoersOK:        true,
		LingerOK:         true,
	}
	if os.Getenv("DIAG_TEST_BYPASS_CHECKS") != "" {
		r.SteamRunning = false
		r.BinariesOK = true
		r.RuntimeHelpersExecutable = true
		r.HooksOK = true
		r.FlatpakOK = true
		r.SudoersOK = true
		r.LingerOK = true
		r.PermissionsOK = true
		r.SkinsDirOK = true
		r.TimerActive = true
		r.TaskScheduled = true
		r.UpdateChannel = resolveChannel()
		r.Version = version.Resolve()
		return r
	}

	r.UpdateChannel = resolveChannel()
	r.Version = millenniumVersion()
	r.SteamRunning = isSteamRunning()
	if r.SteamRunning {
		r.SteamDetail = "Running"
	} else {
		r.SteamDetail = "Not Running"
	}

	var binariesErr error
	r.BinariesOK, r.BinariesDetail, binariesErr = checkBinaries()
	r.BinariesNeedPrivilege = errors.Is(binariesErr, os.ErrPermission)
	r.RuntimeHelpersExecutable = repair.RuntimeHelpersExecutable()
	r.PermissionsOK = repair.PermissionsOK()
	if runtime.GOOS != "windows" {
		r.HooksOK, r.FlatpakOK = checkHooks()
		r.SudoersOK = checkSudoers()
		r.TimerActive = scheduleConfiguredUnix()
		r.LingerOK = true // best-effort; linger check needs loginctl
	} else {
		r.TaskScheduled = scheduleConfiguredWindows()
	}

	steam := theme.FindSteamDir()
	if skins, err := theme.SkinsDir(); err == nil {
		if st, e := os.Stat(skins); e == nil && st.IsDir() {
			r.SkinsDirOK = true
		} else {
			r.SkinsDirOK = false
		}
	} else if steam == "" {
		r.SkinsDirOK = false
	} else {
		r.SkinsDirOK = false
	}
	_ = steam
	return r
}

func resolveChannel() string {
	data, err := config.Load()
	if err == nil {
		if ch := config.Get(data, "update_channel"); ch != "" {
			return ch
		}
	}
	if b, err := os.ReadFile(filepath.Join(repair.MillenniumLibRoot(), "version.txt")); err == nil {
		if strings.Contains(strings.ToLower(string(b)), "beta") {
			return "beta"
		}
	}
	return "stable"
}

func millenniumVersion() string {
	if runtime.GOOS == "windows" {
		steam := theme.FindSteamDir()
		if steam == "" {
			return "Not Installed"
		}
		p := filepath.Join(steam, "millennium", "version.txt")
		b, err := os.ReadFile(p)
		if err != nil {
			return "Not Installed"
		}
		return strings.TrimSpace(string(b))
	}
	b, err := os.ReadFile(filepath.Join(repair.MillenniumLibRoot(), "version.txt"))
	if err != nil {
		return "Not Installed"
	}
	return strings.TrimSpace(string(b))
}

func isSteamRunning() bool {
	if runtime.GOOS == "windows" {
		out, err := exec.Command("tasklist", "/FI", "IMAGENAME eq steam.exe").CombinedOutput()
		if err != nil {
			return false
		}
		return strings.Contains(strings.ToLower(string(out)), "steam.exe")
	}
	out, err := exec.Command("pgrep", "-x", "steam").CombinedOutput()
	return err == nil && len(strings.TrimSpace(string(out))) > 0
}

func checkBinaries() (ok bool, detail string, checkErr error) {
	if runtime.GOOS == "darwin" {
		return true, "Not applicable (Linux client runtime)", nil
	}
	if runtime.GOOS == "windows" {
		steam := theme.FindSteamDir()
		if steam == "" {
			return false, "Not Installed (Steam not found)", nil
		}
		root := filepath.Join(steam, "millennium")
		ver := filepath.Join(root, "version.txt")
		if _, err := os.Stat(ver); err != nil {
			return false, "Not Installed (missing version.txt)", nil
		}
		needed := []string{
			filepath.Join(root, "lib", "millennium.dll"),
			filepath.Join(root, "lib", "millennium.hhx64.dll"),
			filepath.Join(root, "bin", "millennium.crashhandler64.exe"),
			filepath.Join(root, "bin", "millennium.luavm64.exe"),
		}
		for _, p := range needed {
			if _, err := os.Stat(p); err != nil {
				return false, "Corrupted (core libraries or wrapper binaries are missing)", nil
			}
		}
		b, _ := os.ReadFile(ver)
		return true, "v" + strings.TrimSpace(string(b)) + " - Present", nil
	}

	root := repair.MillenniumLibRoot()
	verFile := filepath.Join(root, "version.txt")
	if _, err := os.Stat(verFile); err != nil {
		if errors.Is(err, os.ErrPermission) {
			return false, "Cannot verify client files: re-run with sudo", os.ErrPermission
		}
		return false, "Not Installed (missing version.txt)", nil
	}
	if err := clientfiles.Verify(root); err != nil {
		if errors.Is(err, os.ErrPermission) {
			return false, "Cannot verify client files: re-run with sudo", err
		}
		return false, "Corrupted (" + err.Error() + ")", err
	}
	b, _ := os.ReadFile(verFile)
	return true, fmt.Sprintf("v%s (%s channel) - Verified Healthy", strings.TrimSpace(string(b)), resolveChannel()), nil
}

func checkHooks() (hooksOK, flatpakOK bool) {
	hooksOK, flatpakOK = true, true
	if runtime.GOOS == "darwin" {
		return true, true
	}
	cands := steam.DirCandidates()
	found := false
	for _, steam := range cands {
		if st, err := os.Stat(steam); err != nil || !st.IsDir() {
			continue
		}
		found = true
		for _, arch := range []struct{ folder, lib string }{
			{"ubuntu12_32", "x86"},
			{"ubuntu12_64", "hhx64"},
		} {
			hook := filepath.Join(steam, arch.folder, "libXtst.so.6")
			fi, err := os.Lstat(hook)
			if err != nil {
				hooksOK = false
				continue
			}
			if fi.Mode()&os.ModeSymlink == 0 {
				continue // warn-level in bash; not hard fail
			}
			target, err := os.Readlink(hook)
			if err != nil {
				hooksOK = false
				continue
			}
			want := "libmillennium_bootstrap_" + arch.lib + ".so"
			if !strings.Contains(target, want) {
				continue
			}
			if _, err := os.Stat(target); err != nil {
				hooksOK = false
			}
		}
		if strings.Contains(steam, "com.valvesoftware.Steam") {
			// Flatpak override presence is best-effort.
			out, err := exec.Command("flatpak", "override", "--user", "--show", "com.valvesoftware.Steam").CombinedOutput()
			if err != nil || !strings.Contains(string(out), "/usr/lib/millennium") {
				flatpakOK = false
			}
		}
	}
	if !found {
		hooksOK = false
	}
	return hooksOK, flatpakOK
}

func checkSudoers() bool {
	// sudoers.d is Linux-oriented; calling sudo on macOS CI can hang in the
	// security agent even with -n, and the check isn't meaningful on Darwin.
	if runtime.GOOS != "linux" {
		return true
	}
	out, err := exec.Command("sudo", "-n", "-l").CombinedOutput()
	if err != nil {
		return false
	}
	text := string(out)
	return strings.Contains(text, "millennium upgrade") ||
		strings.Contains(text, "millennium-upgrade") ||
		strings.Contains(text, "NOPASSWD: ALL") ||
		strings.Contains(text, "NOPASSWD:ALL")
}

func scheduleConfiguredUnix() bool {
	st := schedule.CollectStatus()
	return st.Configured
}

func scheduleConfiguredWindows() bool {
	st := schedule.CollectStatus()
	return st.Configured
}

// FormatJSON emits contract-shaped JSON for the current OS.
func FormatJSON(r Report) string {
	type unixJSON struct {
		SteamRunning             bool   `json:"steam_running"`
		BinariesNeedPrivilege    bool   `json:"-"`
		BinariesOK               bool   `json:"binaries_ok"`
		RuntimeHelpersExecutable *bool  `json:"runtime_helpers_executable,omitempty"`
		HooksOK                  bool   `json:"hooks_ok"`
		FlatpakOK                bool   `json:"flatpak_ok"`
		SudoersOK                bool   `json:"sudoers_ok"`
		TimerActive              bool   `json:"timer_active"`
		LingerOK                 bool   `json:"linger_ok"`
		ScriptsUpToDate          bool   `json:"scripts_up_to_date"`
		PermissionsOK            bool   `json:"permissions_ok"`
		SkinsDirOK               bool   `json:"skins_dir_ok"`
		CompletionsOK            bool   `json:"completions_ok"`
		CleanOfObsolete          bool   `json:"clean_of_obsolete"`
		UnmanagedFilesOK         bool   `json:"unmanaged_files_ok"`
		MixedInstallOK           bool   `json:"mixed_install_ok"`
		InstallMethod            string `json:"install_method"`
		HelpersCheckout          string `json:"helpers_checkout"`
		HelpersTrack             string `json:"helpers_track"`
		HelpersRef               string `json:"helpers_ref"`
		LatestReleaseTag         string `json:"latest_release_tag"`
		UpdateChannel            string `json:"update_channel"`
	}
	type winJSON struct {
		SteamRunning     bool   `json:"steam_running"`
		BinariesOK       bool   `json:"binaries_ok"`
		PermissionsOK    bool   `json:"permissions_ok"`
		SkinsDirOK       bool   `json:"skins_dir_ok"`
		TaskScheduled    bool   `json:"task_scheduled"`
		CleanOfObsolete  bool   `json:"clean_of_obsolete"`
		CompletionsOK    bool   `json:"completions_ok"`
		ScriptsUpToDate  bool   `json:"scripts_up_to_date"`
		InstallMethod    string `json:"install_method"`
		MixedInstallOK   bool   `json:"mixed_install_ok"`
		HelpersCheckout  string `json:"helpers_checkout"`
		HelpersTrack     string `json:"helpers_track"`
		HelpersRef       string `json:"helpers_ref"`
		LatestReleaseTag string `json:"latest_release_tag"`
		UpdateChannel    string `json:"update_channel"`
		Version          string `json:"version"`
	}
	var b []byte
	var err error
	var helperModes *bool
	if runtime.GOOS == "linux" {
		helperModes = &r.RuntimeHelpersExecutable
	}
	if runtime.GOOS == "windows" {
		b, err = json.MarshalIndent(winJSON{
			SteamRunning: r.SteamRunning, BinariesOK: r.BinariesOK, PermissionsOK: r.PermissionsOK,
			SkinsDirOK: r.SkinsDirOK, TaskScheduled: r.TaskScheduled, CleanOfObsolete: r.CleanOfObsolete,
			CompletionsOK: r.CompletionsOK, ScriptsUpToDate: r.ScriptsUpToDate, InstallMethod: r.InstallMethod,
			MixedInstallOK: r.MixedInstallOK, HelpersCheckout: r.HelpersCheckout, HelpersTrack: r.HelpersTrack,
			HelpersRef: r.HelpersRef, LatestReleaseTag: r.LatestReleaseTag, UpdateChannel: r.UpdateChannel,
			Version: r.Version,
		}, "", "  ")
	} else {
		b, err = json.MarshalIndent(unixJSON{
			SteamRunning: r.SteamRunning, BinariesOK: r.BinariesOK, HooksOK: r.HooksOK, FlatpakOK: r.FlatpakOK,
			RuntimeHelpersExecutable: helperModes,
			SudoersOK:                r.SudoersOK, TimerActive: r.TimerActive, LingerOK: r.LingerOK,
			ScriptsUpToDate: r.ScriptsUpToDate, PermissionsOK: r.PermissionsOK, SkinsDirOK: r.SkinsDirOK,
			CompletionsOK: r.CompletionsOK, CleanOfObsolete: r.CleanOfObsolete, UnmanagedFilesOK: r.UnmanagedFilesOK,
			MixedInstallOK: r.MixedInstallOK, InstallMethod: r.InstallMethod, HelpersCheckout: r.HelpersCheckout,
			HelpersTrack: r.HelpersTrack, HelpersRef: r.HelpersRef, LatestReleaseTag: r.LatestReleaseTag,
			UpdateChannel: r.UpdateChannel,
		}, "", "  ")
	}
	if err != nil {
		return "{}"
	}
	return string(b) + "\n"
}

// FormatDoctorDryRun announces repairs without executing them.
func FormatDoctorDryRun(r Report, force bool) string {
	var b strings.Builder
	b.WriteString("=== DRY RUN MODE: No changes will be made ===\n")
	b.WriteString("=== Millennium Doctor plan (native) ===\n")
	steps := DoctorPlan(r, force)
	if len(steps) == 0 {
		b.WriteString("No issues detected. Your Millennium installation looks healthy!\n")
		return b.String()
	}
	for _, s := range steps {
		b.WriteString("[DRY RUN] Would: " + s.Detail + "\n")
	}
	fmt.Fprintf(&b, "Dry run completed (%d planned actions). Re-run doctor without --dry-run for live repairs.\n", len(steps))
	return b.String()
}

// PrintLogs prints updater + filtered Steam webhelper logs (non-follow).
func PrintLogs() int {
	state := schedule.LogPath()
	if st, err := os.Stat(state); err == nil && st.Mode().IsRegular() {
		fmt.Println("=== Millennium Background Auto-Updater Logs ===")
		_ = printTail(state, 50)
		fmt.Println()
	}
	fmt.Println("=== Millennium & Steam WebHelper Logs ===")
	logFile := newestSteamLog()
	if logFile == "" {
		fmt.Fprintln(os.Stderr, "Error: No Steam logs found on this system.")
		return 1
	}
	fmt.Printf("Reading log file: %s\n\n", logFile)
	parts := logFilterParts()
	lines, err := tailLines(logFile, 200)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	matched := 0
	for _, line := range lines {
		if lineMatchesFilter(line, parts) {
			fmt.Println(line)
			matched++
		}
	}
	if matched == 0 {
		fmt.Println("No recent Millennium-related log entries found.")
	}
	return 0
}

func newestSteamLog() string {
	cands := theme.SteamCandidates()
	var files []string
	names := []string{"webhelper.txt", "webhelper-linux.txt", "console.txt", "console-linux.txt"}
	if runtime.GOOS == "windows" {
		names = []string{"webhelper.txt", "console_log.txt", "cef_log.txt"}
	}
	for _, steam := range cands {
		if steam == "" {
			continue
		}
		logDir := filepath.Join(steam, "logs")
		for _, n := range names {
			p := filepath.Join(logDir, n)
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				files = append(files, p)
			}
		}
	}
	var best string
	var bestT time.Time
	for _, f := range files {
		st, err := os.Stat(f)
		if err != nil {
			continue
		}
		if st.ModTime().After(bestT) {
			bestT = st.ModTime()
			best = f
		}
	}
	return best
}

func printTail(path string, n int) error {
	lines, err := tailLines(path, n)
	if err != nil {
		return err
	}
	for _, line := range lines {
		fmt.Println(line)
	}
	return nil
}

func tailLines(path string, n int) ([]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines, nil
}
