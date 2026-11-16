package repair

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bolens/millennium-helpers/internal/theme"
	"github.com/bolens/millennium-helpers/internal/usercontext"
)

// HookPlan is one planned bootstrap hook symlink.
type HookPlan struct {
	Hook   string
	Target string
	Steam  string
}

// MillenniumLibRoot returns the install root that holds bootstrap .so files.
func MillenniumLibRoot() string {
	if d := os.Getenv("MOCK_LIB_DIR"); d != "" {
		return filepath.Join(d, "millennium")
	}
	if d := os.Getenv("MILLENNIUM_LIB_DIR"); d != "" {
		return filepath.Join(d, "millennium")
	}
	return "/usr/lib/millennium"
}

// SteamRT3 detects an installed 64-bit client or the explicit opt-in marker.
func SteamRT3(root string) bool {
	for _, name := range []string{"steamrt64/steam", ".steam-enable-steamrt64-client"} {
		if st, err := os.Stat(filepath.Join(root, name)); err == nil && st.Mode().IsRegular() {
			return true
		}
	}
	return false
}

// HooksAt is shared by diagnostics, injection toggles, and the safe Linux writer.
func HooksAt(steam, libRoot string) []HookPlan {
	out := []HookPlan{
		{filepath.Join(steam, "ubuntu12_32/libXtst.so.6"), filepath.Join(libRoot, "libmillennium_bootstrap_x86.so"), steam},
		{filepath.Join(steam, "ubuntu12_64/libXtst.so.6"), filepath.Join(libRoot, "libmillennium_bootstrap_hhx64.so"), steam},
	}
	if SteamRT3(steam) {
		out = append(out, HookPlan{filepath.Join(steam, "steamrt64/libXtst.so.6"),
			filepath.Join("/run/host", libRoot, "libmillennium_bootstrap_x64.so"), steam})
	}
	return out
}

// HostHookTarget maps a pressure-vessel link target back to its host file.
func HostHookTarget(target string) string {
	return strings.TrimPrefix(target, "/run/host")
}

// PlanHooks lists libXtst → bootstrap symlink pairs for existing Steam trees.
// Empty on Windows/Darwin (hooks are not used there).
func PlanHooks() []HookPlan {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return nil
	}
	root := MillenniumLibRoot()
	ctx, err := usercontext.Resolve()
	if err != nil {
		return nil
	}
	home := ctx.Home
	cands := []string{
		filepath.Join(home, ".local/share/Steam"),
		filepath.Join(home, ".steam/steam"),
		filepath.Join(home, ".steam/root"),
		filepath.Join(home, ".var/app/com.valvesoftware.Steam/.local/share/Steam"),
	}
	if steam := theme.FindSteamDir(); steam != "" {
		cands = append([]string{steam}, cands...)
	}
	seen := map[string]bool{}
	var out []HookPlan
	for _, steam := range cands {
		if seen[steam] {
			continue
		}
		seen[steam] = true
		if st, err := os.Stat(steam); err != nil || !st.IsDir() {
			continue
		}
		out = append(out, HooksAt(steam, root)...)
	}
	return out
}

// InstallBootstrapHooks restores Millennium Steam bootstrap symlinks (Unix).
// No-op on Windows/Darwin. Shared by `millennium repair` and `diag doctor`.
func InstallBootstrapHooks() error {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return nil
	}
	plans := PlanHooks()
	if len(plans) == 0 {
		return fmt.Errorf("no Steam directories found to install hooks")
	}
	seen := map[string]bool{}
	for _, p := range plans {
		key, err := filepath.EvalSymlinks(p.Steam)
		if err != nil {
			return fmt.Errorf("cannot resolve Steam hook directory")
		}
		if seen[key] {
			continue
		}
		if err := InstallHooksAt(p.Steam, MillenniumLibRoot()); err != nil {
			return err
		}
		seen[key] = true
		fmt.Println("Restored Steam bootstrap hooks.")
	}
	return nil
}
