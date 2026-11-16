//go:build unix

package steam

import (
	"fmt"
	"strings"
)

// relaunchArguments removes the previous launcher's logger transport marker.
// The detached process cannot inherit its logging pipes. Keep the original
// quoting of all other arguments, including saved state from older helpers.
func relaunchArguments(args string) (string, error) {
	var words []string
	start := -1
	var quote byte
	escaped := false
	keep := func(end int) {
		word := args[start:end]
		switch word {
		case "-srt-logger-opened", "'-srt-logger-opened'", `"-srt-logger-opened"`:
		default:
			words = append(words, word)
		}
	}
	for i := 0; i < len(args); i++ {
		c := args[i]
		whitespace := c == ' ' || c == '\t' || c == '\n'
		if start < 0 {
			if whitespace {
				continue
			}
			start = i
		}
		if escaped {
			escaped = false
			continue
		}
		if quote == '\'' {
			if c == '\'' {
				quote = 0
			}
			continue
		}
		if c == '\\' {
			escaped = true
			continue
		}
		if quote == '"' {
			if c == '"' {
				quote = 0
			}
			continue
		}
		if c == '\'' || c == '"' {
			quote = c
			continue
		}
		if whitespace {
			keep(i)
			start = -1
		}
	}
	if escaped || quote != 0 {
		return "", fmt.Errorf("invalid quoting in saved Steam arguments; capture a new Steam session before relaunch")
	}
	if start >= 0 {
		keep(len(args))
	}
	return strings.Join(words, " "), nil
}
