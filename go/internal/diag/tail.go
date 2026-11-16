package diag

import (
	"bufio"
	"io"
	"os"
	"strings"
)

// readTail retains at most n complete lines and returns a trailing partial line.
func readTail(reader io.Reader, n int) ([]string, string, error) {
	if n < 0 {
		n = 0
	}
	lines := make([]string, n)
	count, next := 0, 0
	buffered := bufio.NewReader(reader)
	for {
		line, err := buffered.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return nil, "", err
			}
			ordered := make([]string, 0, count)
			start := 0
			if count == n && n > 0 {
				start = next
			}
			for i := 0; i < count; i++ {
				ordered = append(ordered, lines[(start+i)%n])
			}
			return ordered, line, nil
		}
		if n > 0 {
			lines[next] = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
			next = (next + 1) % n
			if count < n {
				count++
			}
		}
	}
}

func tailLines(path string, n int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	lines, partial, err := readTail(f, n)
	if err != nil {
		return nil, err
	}
	if partial != "" {
		lines = append(lines, partial)
	}
	if n <= 0 {
		return []string{}, nil
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines, nil
}
