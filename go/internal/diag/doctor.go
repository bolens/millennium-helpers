package diag

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/bolens/millennium-helpers/internal/repair"
	"github.com/bolens/millennium-helpers/internal/steam"
	"github.com/bolens/millennium-helpers/internal/theme"
	"github.com/bolens/millennium-helpers/internal/usercontext"
)

// DoctorStep is one live repair action.
type DoctorStep struct {
	ID     string
	Detail string
}

// DoctorPlan lists live steps from a report (same conditions as dry-run).
func DoctorPlan(r Report, force bool) []DoctorStep {
	var out []DoctorStep
	add := func(cond bool, id, detail string) {
		if force || cond {
			out = append(out, DoctorStep{ID: id, Detail: detail})
		}
	}
	if r.BinariesNeedPrivilege {
		out = append(out, DoctorStep{ID: "binaries_access", Detail: "re-run sudo millennium diag doctor to verify unreadable client files"})
	} else if runtime.GOOS != "darwin" {
		add(!r.BinariesOK, "upgrade_force", "millennium upgrade --force")
	}
	if runtime.GOOS != "windows" {
		if runtime.GOOS == "linux" {
			if !r.BinariesNeedPrivilege {
				add(r.BinariesOK && !r.RuntimeHelpersExecutable, "runtime_helpers", "restore executable modes on Millennium runtime helpers")
			}
			add(!r.HooksOK, "repair_hooks", "restore bootstrap libXtst hooks")
			add(!r.FlatpakOK, "flatpak", "flatpak override --user --filesystem=/usr/lib/millennium")
		}
		add(!r.TimerActive, "schedule_enable", "millennium schedule enable")
		if runtime.GOOS == "linux" {
			add(!r.SudoersOK, "sudoers_hint", "re-run installer for passwordless sudoers")
			add(!r.LingerOK, "linger", "loginctl enable-linger")
		}
		add(!r.PermissionsOK, "permissions", "fix ownership on user Millennium paths")
	} else {
		add(!r.TaskScheduled, "schedule_enable", "millennium schedule enable")
	}
	add(!r.SkinsDirOK, "skins_dir", "create Steam steamui/skins directory")
	return out
}

// RunDoctorLive applies DoctorPlan repairs.
func RunDoctorLive(o Options) int {
	if _, err := usercontext.Resolve(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Println("=== Running Millennium Doctor (native) ===")
	r := Collect()
	if o.Force {
		fmt.Println("Force option specified. Applying all planned doctor repairs...")
	}
	steps := DoctorPlan(r, o.Force)
	if len(steps) == 0 {
		fmt.Println("No issues detected. Your Millennium installation looks healthy!")
		return 0
	}

	needSteamClose := false
	for _, s := range steps {
		if s.ID == "upgrade_force" || s.ID == "repair_hooks" || s.ID == "runtime_helpers" {
			needSteamClose = true
			break
		}
	}
	work := func() error {
		var failures []error
		for _, s := range steps {
			fmt.Printf("\n[DOCTOR] %s...\n", s.Detail)
			if err := applyDoctorStep(s, r, o); err != nil {
				failures = append(failures, err)
			}
		}
		return errors.Join(failures...)
	}
	var err error
	if needSteamClose && os.Getenv("MOCK_LIB_DIR") == "" {
		err = steam.WithClosedClient(o.Yes, work)
	} else {
		err = work()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Doctor failed: %v\n", err)
		return 1
	}
	fmt.Printf("\nDoctor repairs applied successfully.\nChannel: %s. Re-run millennium diag to verify, or millennium diag doctor again if issues remain.\n", r.UpdateChannel)
	if !r.CompletionsOK || !r.ScriptsUpToDate || !r.CleanOfObsolete || !r.UnmanagedFilesOK || !r.MixedInstallOK {
		fmt.Println("Note: package/completions/obsolete cleanup may still need a helpers reinstall (install.sh / package manager).")
	}
	return 0
}

func applyDoctorStep(s DoctorStep, r Report, o Options) error {
	switch s.ID {
	case "binaries_access":
		return fmt.Errorf("client integrity could not be checked; re-run sudo millennium diag doctor before choosing a repair")
	case "upgrade_force":
		return runSelf("upgrade", "--channel", r.UpdateChannel, "--force", "--yes")
	case "repair_hooks":
		return repair.InstallBootstrapHooks()
	case "runtime_helpers":
		return repair.EnsureRuntimeHelpersExecutable()
	case "flatpak":
		cmd := exec.Command("flatpak", "override", "--user", "--filesystem=/usr/lib/millennium", "com.valvesoftware.Steam")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("flatpak override: %v (%s)", err, strings.TrimSpace(string(out)))
		}
		return nil
	case "schedule_enable":
		args := []string{"schedule", "enable", r.UpdateChannel}
		if o.Quiet {
			args = append(args, "-q")
		}
		return runSelf(args...)
	case "sudoers_hint":
		fmt.Println("Sudoers drop-in is missing or unauthorized.")
		fmt.Println("Re-run the installer to set up passwordless rules: sudo ./install.sh")
		return nil
	case "linger":
		user := os.Getenv("SUDO_USER")
		if user == "" {
			user = os.Getenv("USER")
		}
		if user == "" {
			user = "root"
		}
		cmd := exec.Command("loginctl", "enable-linger", user)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("enable-linger: %v (%s)", err, strings.TrimSpace(string(out)))
		}
		return nil
	case "permissions":
		targets, err := repair.Plan()
		if err != nil {
			return err
		}
		var ownership []repair.Target
		for _, target := range targets {
			if target.Kind == "chown" {
				ownership = append(ownership, target)
			}
		}
		return repair.Apply(ownership, true)
	case "skins_dir":
		dir, err := theme.SkinsDir()
		if err != nil || dir == "" {
			return fmt.Errorf("skins directory unknown: %v", err)
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		fmt.Printf("Created: %s\n", dir)
		return nil
	default:
		return fmt.Errorf("unknown doctor step %s", s.ID)
	}
}

func runSelf(args ...string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
