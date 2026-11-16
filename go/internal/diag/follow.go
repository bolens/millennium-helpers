package diag

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bolens/millennium-helpers/internal/schedule"
)

// Test seams for FollowLogs (avoid hanging the suite).
var (
	followPollInterval = 200 * time.Millisecond
	followMaxCycles    = 0 // 0 = until process exit / Ctrl+C
	followCycleHook    func()
	followOutput       io.Writer = os.Stdout
)

// logFilterParts returns lowercase substrings for Millennium-related lines.
func logFilterParts() []string {
	parts := []string{"millennium", "bootstrap", "update-check", "plugin_loader", "steamwebhelper"}
	if runtime.GOOS == "windows" {
		parts = append(parts, "wsock32")
	} else {
		parts = append(parts, "pressure-vessel")
	}
	return parts
}

func lineMatchesFilter(line string, parts []string) bool {
	low := strings.ToLower(line)
	for _, p := range parts {
		if strings.Contains(low, p) {
			return true
		}
	}
	return false
}

// FollowLogs prints updater tail once, then filter-tails the newest Steam log.
func FollowLogs() int {
	maxCycles := followMaxCycles
	if v := strings.TrimSpace(os.Getenv("MILLENNIUM_FOLLOW_MAX_CYCLES")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxCycles = n
		}
	}
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
	fmt.Println("Tailing log file (Ctrl+C to exit)...")
	return followWithOptions(logFile, logFilterParts(), 100, followOptions{maxCycles, followPollInterval, followCycleHook, followOutput})
}

type followOptions struct {
	maxCycles    int
	pollInterval time.Duration
	cycleHook    func()
	output       io.Writer
}

func followFiltered(path string, parts []string, initialTail int) int {
	return followWithOptions(path, parts, initialTail, followOptions{followMaxCycles, followPollInterval, followCycleHook, followOutput})
}

func followWithOptions(path string, parts []string, initialTail int, options followOptions) int {
	f, err := openFollowFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	offset := info.Size()
	lines, partial, err := readTail(io.NewSectionReader(f, 0, offset), initialTail)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	emit := func(line string) error {
		if lineMatchesFilter(line, parts) {
			_, err := fmt.Fprintln(options.output, line)
			return err
		}
		return nil
	}
	for _, line := range lines {
		if err := emit(line); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	for cycle := 0; options.maxCycles <= 0 || cycle < options.maxCycles; cycle++ {
		if options.cycleHook != nil {
			options.cycleHook()
		}
		current, err := os.Stat(path)
		if err != nil {
			time.Sleep(options.pollInterval)
			continue
		}
		openInfo, err := f.Stat()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		if !os.SameFile(openInfo, current) || current.Size() < offset {
			replacement, err := openFollowFile(path)
			if err != nil {
				time.Sleep(options.pollInterval)
				continue
			}
			_ = f.Close()
			f = replacement
			offset = 0
			partial = ""
			current, err = f.Stat()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 1
			}
		}
		if current.Size() == offset {
			time.Sleep(options.pollInterval)
			continue
		}
		reader := bufio.NewReader(io.NewSectionReader(f, offset, current.Size()-offset))
		for {
			text, readErr := reader.ReadString('\n')
			offset += int64(len(text))
			partial += text
			if strings.HasSuffix(partial, "\n") {
				line := strings.TrimSuffix(strings.TrimSuffix(partial, "\n"), "\r")
				if err := emit(line); err != nil {
					fmt.Fprintln(os.Stderr, err)
					return 1
				}
				partial = ""
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				fmt.Fprintln(os.Stderr, readErr)
				return 1
			}
		}
	}
	return 0
}
