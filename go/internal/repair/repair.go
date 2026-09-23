package repair

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bolens/millennium-helpers/internal/clientfiles"
	"github.com/bolens/millennium-helpers/internal/config"
	"github.com/bolens/millennium-helpers/internal/steam"
	"github.com/bolens/millennium-helpers/internal/theme"
	"github.com/bolens/millennium-helpers/internal/usercontext"
)

// RuntimeHelpersExecutable reports shared-install helper access on Linux.
func RuntimeHelpersExecutable() bool {
	return runtime.GOOS != "linux" || clientfiles.HelpersExecutable(MillenniumLibRoot())
}

// EnsureRuntimeHelpersExecutable restores Linux helper permissions.
func EnsureRuntimeHelpersExecutable() error {
	if runtime.GOOS != "linux" {
		return nil
	}
	return clientfiles.NormalizeHelpers(MillenniumLibRoot())
}

// Target is a path that repair would chown / touch.
type Target struct {
	Path string
	Kind string
}

// Plan lists ownership/cache repair targets for the current user (read-only).
func Plan() ([]Target, error) {
	ctx, err := usercontext.Resolve()
	if err != nil {
		return nil, err
	}
	return planFor(ctx), nil
}

func planFor(ctx usercontext.Context) []Target {
	home := ctx.Home
	xdgConfig, xdgData := ctx.ConfigHome, ctx.DataHome
	steam := theme.FindSteamDir()
	candidates := []string{
		filepath.Join(xdgData, "millennium"),
		filepath.Join(home, ".local", "share", "millennium"),
		filepath.Join(xdgConfig, "millennium"),
		filepath.Join(xdgConfig, "millennium-helpers"),
		filepath.Join(home, ".config", "millennium"),
		filepath.Join(home, ".var", "app", "com.valvesoftware.Steam", "config", "millennium"),
		filepath.Join(home, ".var", "app", "com.valvesoftware.Steam", ".config", "millennium"),
	}
	if steam != "" {
		candidates = append([]string{filepath.Join(steam, "millennium")}, candidates...)
	}
	if runtime.GOOS == "windows" {
		if local := os.Getenv("LOCALAPPDATA"); local != "" {
			candidates = append(candidates, filepath.Join(local, "millennium"))
		}
	}
	var out []Target
	seen := map[string]bool{}
	for _, p := range candidates {
		if seen[p] {
			continue
		}
		if st, err := os.Stat(p); os.IsPermission(err) || (err == nil && (st.IsDir() || st.Mode().IsRegular())) {
			seen[p] = true
			out = append(out, Target{Path: p, Kind: "chown"})
		}
	}
	if steam != "" {
		cache := filepath.Join(steam, "config", "htmlcache")
		if st, err := os.Stat(cache); err == nil && st.IsDir() {
			out = append(out, Target{Path: cache, Kind: "htmlcache"})
		}
	}
	return out
}

// FormatPlan prints repair dry-run text.
func FormatPlan(targets []Target, skipTheme bool) string {
	var b strings.Builder
	b.WriteString("=== DRY RUN MODE: No changes will be made ===\n")
	b.WriteString("[DRY RUN] Would capture Steam's environment and close it if running.\n")
	if runtime.GOOS == "windows" {
		fmt.Fprintf(&b, "[DRY RUN] Would run: millennium upgrade --force --channel %s\n", updateChannel())
	} else if runtime.GOOS == "linux" {
		b.WriteString("[DRY RUN] Would restore executable modes on Millennium runtime helpers.\n")
		hooks := PlanHooks()
		if len(hooks) == 0 {
			b.WriteString("[DRY RUN] Would restore bootstrap hooks (no Steam tree found yet).\n")
		} else {
			for _, h := range hooks {
				fmt.Fprintf(&b, "[DRY RUN] Would link hook: %s -> %s\n", h.Hook, h.Target)
			}
		}
	}
	if len(targets) == 0 {
		b.WriteString("[DRY RUN] No Millennium user/Steam paths found to repair.\n")
	}
	for _, t := range targets {
		fmt.Fprintf(&b, "[DRY RUN] Would fix (%s): %s\n", t.Kind, t.Path)
	}
	if skipTheme {
		b.WriteString("[DRY RUN] Skipping theme refresh (--skip-theme).\n")
	} else {
		b.WriteString("[DRY RUN] Would refresh active theme assets if present.\n")
	}
	b.WriteString("Dry run completed successfully!\n")
	return b.String()
}

// Apply performs live ownership/cache repairs and theme refresh.
func Apply(targets []Target, skipTheme bool) error {
	var failures []error
	for _, t := range targets {
		switch t.Kind {
		case "htmlcache":
			fmt.Printf("Clearing Steam htmlcache: %s\n", t.Path)
			if err := clearCache(t.Path); err != nil {
				return err
			}

		case "chown":
			fmt.Printf("Fixing ownership: %s\n", t.Path)
			if err := chownTree(t.Path); err != nil {
				failures = append(failures, fmt.Errorf("ownership repair failed: %w", err))
			}
		}
	}
	if skipTheme {
		fmt.Println("Skipping theme refresh (--skip-theme).")
	} else if err := refreshThemes(); err != nil {
		failures = append(failures, fmt.Errorf("theme refresh failed: %w", err))
	}
	return errors.Join(failures...)
}

func refreshThemes() error {
	name := theme.ActiveThemeName()
	if name != "" {
		fmt.Printf("Refreshing theme '%s'...\n", name)
		if err := theme.UpdateOne(name, false); err != nil {
			return err
		}
		return nil
	}
	fmt.Println("No active theme configured; checking for installed themes to update...")
	if err := theme.UpdateAll(false); err != nil {
		// Missing Steam/skins is not a hard repair failure.
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}
	return nil
}

func updateChannel() string {
	data, err := config.Load()
	if err != nil {
		return "stable"
	}
	ch := strings.TrimSpace(config.Get(data, "update_channel"))
	switch ch {
	case "stable", "beta", "main":
		return ch
	default:
		return "stable"
	}
}

// forceReinstallWindows runs native upgrade --force (Windows binary reinstall).
func forceReinstallWindows(yes bool) error {
	ch := updateChannel()
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	args := []string{"upgrade", "--channel", ch, "--force"}
	if yes {
		args = append(args, "--yes")
	}
	fmt.Printf("Force-reinstalling Millennium binaries (channel %s)...\n", ch)
	cmd := exec.Command(exe, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ParseFlags parses repair CLI flags.
func ParseFlags(args []string) (dryRun, yes, quiet, skipTheme, help, version bool, err error) {
	for _, a := range args {
		switch a {
		case "-d", "--dry-run", "-DryRun":
			dryRun = true
		case "-y", "--yes", "-Yes":
			yes = true
		case "-q", "--quiet", "-Quiet":
			quiet = true
		case "-s", "--skip-theme", "-SkipTheme":
			skipTheme = true
		case "-h", "--help", "-Help":
			help = true
		case "-V", "--version", "-Version":
			version = true
		default:
			if strings.HasPrefix(a, "-") {
				return false, false, false, false, false, false, fmt.Errorf("Error: Unknown option %s", a)
			}
		}
	}
	return dryRun, yes, quiet, skipTheme, help, version, nil
}

// RunCLI runs dry-run or live native repair (hooks/binary + ownership/cache/theme).
func RunCLI(dryRun, skipTheme, quiet, yes bool) int {
	targets, err := Plan()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if dryRun {
		fmt.Print(FormatPlan(targets, skipTheme))
		return 0
	}

	fmt.Println("=== Initiating Millennium Repair ===")
	work := func() error {
		if runtime.GOOS == "windows" {
			if err := forceReinstallWindows(yes); err != nil {
				return err
			}
		} else if runtime.GOOS == "linux" {
			if err := EnsureRuntimeHelpersExecutable(); err != nil {
				return err
			}
			if err := InstallBootstrapHooks(); err != nil {
				return err
			}
		}
		return Apply(targets, skipTheme)
	}
	if os.Getenv("MOCK_LIB_DIR") != "" {
		err = work()
	} else {
		err = steam.WithClosedClient(yes, work)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Repair failed: %v\n", err)
		return 1
	}

	if !quiet {
		fmt.Println("Repair completed successfully.")
		if runtime.GOOS == "windows" {
			fmt.Println("Tip: millennium schedule status — re-enable the updater task if it was cleared.")
		} else {
			fmt.Println("Tip: millennium diag — verify hooks and schedule after repair.")
		}
	}
	return 0
}

// RunDryRunCLI prints a native repair plan.
func RunDryRunCLI(skipTheme bool) int {
	return RunCLI(true, skipTheme, false, false)
}

// PermissionsOK checks the ownership of the same paths repair would change.
func PermissionsOK() bool {
	targets, err := Plan()
	if err != nil {
		return false
	}
	for _, target := range targets {
		if target.Kind == "chown" {
			if err := checkOwnership(target.Path); err != nil {
				return false
			}
		}
	}
	return true
}
