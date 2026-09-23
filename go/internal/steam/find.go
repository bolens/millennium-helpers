package steam

import (
	"os"
	"path/filepath"

	"github.com/bolens/millennium-helpers/internal/usercontext"
)

// FindDir returns the first existing Steam install root, or "".
func FindDir() string {
	for _, c := range dirCandidates() {
		if c == "" {
			continue
		}
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			return c
		}
	}
	return ""
}

// DirCandidates returns likely Steam roots (may not exist).
func DirCandidates() []string {
	return dirCandidates()
}

func dirCandidatesUnix() []string {
	ctx, err := usercontext.Resolve()
	if err != nil {
		return nil
	}
	return dirCandidatesForHome(ctx.Home)
}

func dirCandidatesForHome(home string) []string {
	return []string{
		filepath.Join(home, ".local/share/Steam"),
		filepath.Join(home, ".steam/steam"),
		filepath.Join(home, ".steam/root"),
		filepath.Join(home, ".var/app/com.valvesoftware.Steam/.local/share/Steam"),
		filepath.Join(home, "Library/Application Support/Steam"),
	}
}

func dirCandidatesCommon() []string {
	var out []string
	for _, env := range []string{"STEAM", "STEAM_PATH"} {
		if v := os.Getenv(env); v != "" {
			out = append(out, v)
		}
	}
	return out
}
