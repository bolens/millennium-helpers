//go:build unix

package steam

import (
	"os"
	"path/filepath"
	"strings"
)

func dirCandidates() []string {
	// Wine commonly exports STEAM or STEAM_PATH with a Windows client path;
	// never let that override or supplement native Unix Steam discovery.
	var out []string
	for _, candidate := range dirCandidatesCommon() {
		if !windowsSteamCandidate(candidate) {
			out = append(out, candidate)
		}
	}
	return append(out, dirCandidatesUnix()...)
}

func windowsSteamCandidate(path string) bool {
	if path == "" {
		return false
	}
	lower := strings.ToLower(filepath.ToSlash(path))
	lower = strings.ReplaceAll(lower, `\`, "/")
	if len(lower) >= 3 && lower[1] == ':' && lower[2] == '/' {
		return true
	}
	for _, marker := range []string{
		"/.wine/", "/wineprefixes/", "/dosdevices/", "/drive_c/",
		"/program files/", "/program files (x86)/",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	_, err := os.Stat(filepath.Join(path, "steam.exe"))
	return err == nil
}
