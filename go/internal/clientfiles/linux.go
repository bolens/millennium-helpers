// Package clientfiles defines the Linux client files shared by maintenance commands.
package clientfiles

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// File describes a required Linux client component.
type File struct {
	Name       string
	Executable bool
}

// Linux returns a fresh manifest, including both runtime executables.
func Linux() []File {
	return []File{
		{"libmillennium_bootstrap_x86.so", false},
		{"libmillennium_bootstrap_hhx64.so", false},
		{"libmillennium_x86.so", false},
		{"libmillennium_hhx64.so", false},
		{"libmillennium_pvs64", true},
		{"libmillennium_luavm_x86", true},
	}
}

// ValidateFiles rejects incomplete installs and non-regular components.
func ValidateFiles(root string) error {
	for _, f := range Linux() {
		st, err := os.Lstat(filepath.Join(root, f.Name))
		if errors.Is(err, os.ErrPermission) {
			return fmt.Errorf("cannot inspect client file %s: %w", f.Name, os.ErrPermission)
		}
		if err != nil || !st.Mode().IsRegular() {
			return fmt.Errorf("required client file missing or not regular: %s", f.Name)
		}
	}
	return nil
}

// HelpersExecutable checks shared-install modes, not the privileged caller's access.
func HelpersExecutable(root string) bool {
	for _, f := range Linux() {
		if !f.Executable {
			continue
		}
		st, err := os.Lstat(filepath.Join(root, f.Name))
		if err != nil || !st.Mode().IsRegular() || st.Mode().Perm()&0o555 != 0o555 {
			return false
		}
	}
	return true
}

// NormalizeHelpers validates every helper before changing any modes.
func NormalizeHelpers(root string) error {
	for _, f := range Linux() {
		if !f.Executable {
			continue
		}
		st, err := os.Lstat(filepath.Join(root, f.Name))
		if err != nil || !st.Mode().IsRegular() {
			return fmt.Errorf("runtime helper missing or not regular: %s", f.Name)
		}
	}
	for _, f := range Linux() {
		if f.Executable {
			if err := os.Chmod(filepath.Join(root, f.Name), 0o755); err != nil {
				return fmt.Errorf("cannot restore executable mode: %s", f.Name)
			}
		}
	}
	return nil
}

func digest(root, name string) (string, error) {
	f, err := os.Open(filepath.Join(root, name))
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return "", fmt.Errorf("cannot read client file %s: %w", name, os.ErrPermission)
		}
		return "", fmt.Errorf("cannot read client file: %s", name)
	}
	defer f.Close()
	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", fmt.Errorf("cannot hash client file: %s", name)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// WriteChecksums records every required component or fails the staged installation.
func WriteChecksums(root string) error {
	if err := ValidateFiles(root); err != nil {
		return err
	}
	var b strings.Builder
	for _, f := range Linux() {
		sum, err := digest(root, f.Name)
		if err != nil {
			return err
		}
		fmt.Fprintf(&b, "%s  %s\n", sum, f.Name)
	}
	if err := os.WriteFile(filepath.Join(root, "checksums.txt"), []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("cannot write client integrity manifest")
	}
	return nil
}

// Verify requires complete, unambiguous checksums for the manifest's regular files.
func Verify(root string) error {
	if _, err := Version(root); err != nil {
		return err
	}
	if err := ValidateFiles(root); err != nil {
		return err
	}
	b, err := os.ReadFile(filepath.Join(root, "checksums.txt"))
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return fmt.Errorf("cannot read client integrity manifest: %w", os.ErrPermission)
		}
		return fmt.Errorf("client integrity manifest missing or unreadable")
	}
	expected := map[string]string{}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 || len(fields[0]) != 64 {
			return fmt.Errorf("malformed client integrity manifest")
		}
		name := strings.TrimPrefix(fields[1], "*")
		if filepath.Base(name) != name || strings.ContainsAny(name, `/\`) {
			return fmt.Errorf("invalid integrity manifest filename")
		}
		if _, exists := expected[name]; exists {
			return fmt.Errorf("duplicate integrity manifest entry")
		}
		if _, err := hex.DecodeString(fields[0]); err != nil {
			return fmt.Errorf("invalid integrity manifest checksum")
		}
		expected[name] = strings.ToLower(fields[0])
	}
	for _, f := range Linux() {
		want, ok := expected[f.Name]
		if !ok {
			return fmt.Errorf("client integrity manifest omits %s", f.Name)
		}
		got, err := digest(root, f.Name)
		if err != nil {
			return err
		}
		if got != want {
			return fmt.Errorf("client checksum mismatch: %s", f.Name)
		}
	}
	return nil
}
