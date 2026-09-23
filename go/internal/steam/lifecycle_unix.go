//go:build unix

package steam

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// IsSteamRunning reports whether the Steam client process is present.
func IsSteamRunning() bool {
	if runtime.GOOS == "darwin" {
		return exec.Command("pgrep", "-ix", "Steam").Run() == nil
	}
	return exec.Command("pgrep", "-x", "steam").Run() == nil
}

// IsGameRunning detects an active Steam game (exit 75 path for scheduler).
func IsGameRunning() bool {
	if runtime.GOOS == "darwin" {
		out, err := exec.Command("ps", "-A", "-o", "command").CombinedOutput()
		if err != nil {
			return false
		}
		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "steamapps/common") && !strings.Contains(line, "grep") {
				return true
			}
		}
		return false
	}
	root := procRoot()
	entries, err := os.ReadDir(root)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pid := e.Name()
		if pid == "" || pid[0] < '0' || pid[0] > '9' {
			continue
		}
		comm, _ := os.ReadFile(filepath.Join(root, pid, "comm"))
		c := strings.TrimSpace(string(comm))
		if c == "steam" || c == "steamwebhelper" {
			continue
		}
		environ, err := os.ReadFile(filepath.Join(root, pid, "environ"))
		if err != nil {
			continue
		}
		for _, kv := range bytes.Split(environ, []byte{0}) {
			if bytes.HasPrefix(kv, []byte("SteamAppId=")) {
				id := string(kv[len("SteamAppId="):])
				if id != "" && id != "0" {
					return true
				}
			}
		}
	}
	return false
}

func currentUsername() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.Username
}

func flatpakSteamRunning(username string) bool {
	if _, err := exec.LookPath("flatpak"); err != nil {
		return false
	}
	var cmd *exec.Cmd
	if effectiveUID() == 0 && username != "" && username != currentUsername() {
		cmd = exec.Command("runuser", "-l", username, "-c", "flatpak ps")
	} else {
		cmd = exec.Command("flatpak", "ps")
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "com.valvesoftware.Steam")
}

// CaptureEnv writes relaunch.env for a later post-update relaunch.
func CaptureEnv(username, home string) error {
	state := RelaunchStateFile(username, home)
	if err := os.MkdirAll(filepath.Dir(state), 0o700); err != nil {
		return err
	}
	if effectiveUID() == 0 {
		chownUser(filepath.Dir(state), username)
	}
	if runtime.GOOS == "darwin" {
		if err := os.WriteFile(state, []byte("export WAS_MACOS='true'\n"), 0o600); err != nil {
			return err
		}
		if effectiveUID() == 0 {
			chownUser(state, username)
		}
		return nil
	}

	wasFlatpak := flatpakSteamRunning(username)
	env := map[string]string{}
	steamArgs := ""
	pid := firstSteamPID()
	if pid != "" {
		root := procRoot()
		raw, _ := os.ReadFile(filepath.Join(root, pid, "environ"))
		for _, kv := range bytes.Split(raw, []byte{0}) {
			if len(kv) == 0 {
				continue
			}
			parts := bytes.SplitN(kv, []byte{'='}, 2)
			if len(parts) != 2 {
				continue
			}
			key := string(parts[0])
			for _, want := range EnvKeys {
				if key == want {
					env[key] = string(parts[1])
				}
			}
		}
		cmdRaw, _ := os.ReadFile(filepath.Join(root, pid, "cmdline"))
		args := bytes.Split(cmdRaw, []byte{0})
		var quoted []string
		for i, a := range args {
			if i == 0 || len(a) == 0 {
				continue
			}
			quoted = append(quoted, shellQuote(string(a)))
		}
		steamArgs = strings.Join(quoted, " ")
		if steamArgs != "" {
			steamArgs += " "
		}
	}
	if err := writeRelaunchEnv(state, env, steamArgs, wasFlatpak); err != nil {
		return err
	}
	if effectiveUID() == 0 {
		chownUser(state, username)
	}
	return nil
}

func firstSteamPID() string {
	out, err := exec.Command("pgrep", "-x", "steam").CombinedOutput()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

// CloseGracefully asks Steam to shut down, then force-kills if needed.
func CloseGracefully(username, home string) error {
	if runtime.GOOS == "darwin" {
		if testSuite() {
			fmt.Println("[TEST] Bypassing macOS Steam quit to protect host")
			fmt.Println("Steam closed successfully.")
			return nil
		}
		if currentUsername() == username {
			_ = exec.Command("osascript", "-e", `quit app "Steam"`).Run()
		} else {
			_ = exec.Command("runuser", "-l", username, "-c", `osascript -e 'quit app "Steam"'`).Run()
		}
		return finishShutdown(IsSteamRunning, waitSteamGone, func() error { return exec.Command("killall", "-9", "Steam").Run() })
	}

	wasFlatpak := flatpakSteamRunning(username)
	state := RelaunchStateFile(username, home)
	var shutdown string
	switch {
	case wasFlatpak:
		shutdown = "flatpak run com.valvesoftware.Steam -shutdown"
	case lookPath("steam"):
		shutdown = "steam -shutdown"
	default:
		local := filepath.Join(home, ".local", "bin", "steam")
		if st, err := os.Stat(local); err == nil && !st.IsDir() && st.Mode()&0o111 != 0 {
			shutdown = shellQuote(local) + " -shutdown"
		}
	}
	if shutdown != "" {
		_ = runSteamCmd(username, shutdown, state)
	}
	return finishShutdown(IsSteamRunning, waitSteamGone, func() error { return exec.Command("killall", "-9", "steam", "steamwebhelper").Run() })
}

func finishShutdown(running func() bool, wait func(time.Duration), force func() error) error {
	wait(30 * time.Second)
	if running() {
		fmt.Fprintln(os.Stderr, "Steam did not close gracefully. Force killing...")
		killErr := force()
		wait(5 * time.Second)
		if running() {
			return fmt.Errorf("Steam is still running after shutdown attempt (force-stop error: %v)", killErr)
		}
	}
	fmt.Println("Steam closed successfully.")
	return nil
}

func waitSteamGone(d time.Duration) {
	deadline := time.Now().Add(d)
	for IsSteamRunning() && time.Now().Before(deadline) {
		time.Sleep(time.Second)
	}
}

func lookPath(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func runSteamCmd(username, cmd, stateFile string) error {
	vals := map[string]string{}
	if stateFile != "" {
		var err error
		vals, err = ParseRelaunchEnv(stateFile)
		if err != nil {
			return err
		}
	}
	return runSteamCmdWithEnv(username, cmd, vals)
}

func steamCommand(cmd string, vals map[string]string) string {
	var parts []string
	for _, k := range EnvKeys {
		if v := vals[k]; v != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", k, shellQuote(v)))
		}
	}
	if len(parts) == 0 {
		return cmd
	}
	return "env " + strings.Join(parts, " ") + " " + cmd
}

func runSteamCmdWithEnv(username, cmd string, vals map[string]string) error {
	full := steamCommand(cmd, vals)
	if testSuite() {
		if mock := os.Getenv("MOCK_BIN"); mock != "" {
			if _, err := os.Stat(filepath.Join(mock, "runuser")); err == nil {
				return exec.Command("runuser", username, "-c", full).Run()
			}
		}
		fmt.Printf("[TEST] Bypassing command execution to protect host: %s\n", full)
		return nil
	}
	if effectiveUID() == 0 {
		return exec.Command("runuser", username, "-c", full).Run()
	}
	return exec.Command("sh", "-c", full).Run()
}

// RelaunchFromState starts Steam using relaunch.env and deletes the file.
func RelaunchFromState(username, home string) (bool, error) {
	return relaunchFromState(username, home, startSteamCommand)
}

func relaunchFromState(username, home string, launch func(string, string, map[string]string) error) (attempted bool, err error) {
	state := RelaunchStateFile(username, home)
	if !IsSafeRelaunchStateFile(username, state) {
		return false, nil
	}
	vals, err := ParseRelaunchEnv(state)
	if err != nil {
		return false, err
	}
	argDisplay := vals["STEAM_ARGS"]
	if argDisplay == "" {
		argDisplay = "none"
	}
	wasFlatpak := vals["WAS_FLATPAK"] == "true"
	fmt.Printf("Relaunching Steam client with arguments: %s (Flatpak: %v)...\n", argDisplay, wasFlatpak)
	defer func() {
		if attempted && err == nil {
			_ = os.Remove(state)
		}
	}()

	if testSuite() {
		fmt.Println("[TEST] Bypassing real Steam relaunch in test suite.")
		return true, nil
	}

	if runtime.GOOS == "darwin" {
		if currentUsername() == username {
			return true, exec.Command("open", "-a", "Steam").Start()
		} else {
			return true, exec.Command("runuser", "-l", username, "-c", "open -a Steam >/dev/null 2>&1 &").Start()
		}
	}

	steamArgs := vals["STEAM_ARGS"]
	var cmd string
	switch {
	case wasFlatpak:
		cmd = "flatpak run com.valvesoftware.Steam " + steamArgs
	case lookPath("steam"):
		cmd = "steam " + steamArgs
	default:
		local := filepath.Join(home, ".local", "bin", "steam")
		if st, err := os.Stat(local); err == nil && !st.IsDir() {
			cmd = shellQuote(local) + " " + steamArgs
		}
	}
	if cmd == "" {
		return false, fmt.Errorf("Steam executable not found for relaunch")
	}
	return true, launch(username, cmd, vals)
}

// startSteamCommand detaches the launcher from the maintenance process, but waits
// for observable startup before allowing relaunch state to be discarded.
func startSteamCommand(username, command string, env map[string]string) error {
	full := "exec " + steamCommand(command, env)
	var cmd *exec.Cmd
	if effectiveUID() == 0 {
		cmd = exec.Command("runuser", username, "-c", full)
	} else {
		cmd = exec.Command("sh", "-c", full)
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	return waitForSteamStart(done, IsSteamRunning, 10*time.Second)
}

func waitForSteamStart(done <-chan error, running func() bool, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case err := <-done:
			if err != nil {
				return fmt.Errorf("Steam launcher failed: %w", err)
			}
			done = nil // A launcher may exit successfully before its child is visible.
		default:
		}
		if running() {
			return nil
		}
		select {
		case err := <-done:
			if err != nil {
				return fmt.Errorf("Steam launcher failed: %w", err)
			}
			done = nil
		case <-tick.C:
		case <-timer.C:
			return fmt.Errorf("Steam did not start before the startup timeout")
		}
	}
}
