package repair

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bolens/millennium-helpers/internal/safefs"
	"golang.org/x/sys/unix"
)

// InstallHooksAt is shared by upgrade and repair. Steam's root and architecture
// directories must be real directories; the final hook may be an existing link.
func InstallHooksAt(steamRoot, libRoot string) error {
	root, err := safefs.OpenDir(steamRoot, false, 0, -1, -1)
	if err != nil {
		return fmt.Errorf("cannot safely open Steam hook root: %w", err)
	}
	defer func() { _ = unix.Close(root) }()
	var dirs []int
	defer func() {
		for _, fd := range dirs {
			_ = unix.Close(fd)
		}
	}()
	targets := []string{"libmillennium_bootstrap_x86.so", "libmillennium_bootstrap_hhx64.so"}
	for i, folder := range []string{"ubuntu12_32", "ubuntu12_64"} {
		st, err := os.Lstat(filepath.Join(libRoot, targets[i]))
		if err != nil || !st.Mode().IsRegular() {
			return fmt.Errorf("bootstrap library missing or not regular: %s", targets[i])
		}
		fd, err := safefs.OpenChildDir(root, folder, true, 0o755, -1, -1)
		if err != nil {
			return fmt.Errorf("cannot safely open Steam architecture directory: %w", err)
		}
		dirs = append(dirs, fd)
	}
	for i, fd := range dirs {
		if err := safefs.Symlink(fd, "libXtst.so.6", filepath.Join(libRoot, targets[i])); err != nil {
			return err
		}
	}
	return nil
}
