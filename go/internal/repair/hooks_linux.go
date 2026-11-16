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
	plans := HooksAt(steamRoot, libRoot)
	// Preflight every library before creating or replacing any hook.
	for _, p := range plans {
		st, err := os.Lstat(HostHookTarget(p.Target))
		if err != nil || !st.Mode().IsRegular() {
			return fmt.Errorf("bootstrap library missing or not regular: %s (upgrade Millennium first)", filepath.Base(p.Target))
		}
	}
	for _, p := range plans {
		folder := filepath.Base(filepath.Dir(p.Hook))
		fd, err := safefs.OpenChildDir(root, folder, true, 0o755, -1, -1)
		if err != nil {
			return fmt.Errorf("cannot safely open Steam architecture directory: %w", err)
		}
		dirs = append(dirs, fd)
	}
	for i, fd := range dirs {
		if err := safefs.Symlink(fd, "libXtst.so.6", plans[i].Target); err != nil {
			return err
		}
	}

	return nil
}
