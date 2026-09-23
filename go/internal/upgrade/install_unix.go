//go:build unix

package upgrade

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/bolens/millennium-helpers/internal/archive"
	"github.com/bolens/millennium-helpers/internal/clientfiles"
	"github.com/bolens/millennium-helpers/internal/repair"
)

func installPlatform(archivePath, version string, o Options) error {
	if !clientfiles.ValidVersion(version) {
		return fmt.Errorf("invalid client version label")
	}
	tmp, err := os.MkdirTemp("", "millennium-install-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	if err := archive.SafeExtractTarGz(archivePath, tmp); err != nil {
		return err
	}
	src := filepath.Join(tmp, "usr", "lib", "millennium")
	if st, err := os.Stat(src); err != nil || !st.IsDir() {
		src = tmp
	}

	dest := InstallRoot()
	destTmp := dest + ".tmp"
	_ = os.RemoveAll(destTmp)
	if err := os.MkdirAll(destTmp, 0o755); err != nil {
		return err
	}
	if err := copyTreeFiles(src, destTmp); err != nil {
		return err
	}
	if err := normalizeRuntimeHelperModes(destTmp); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(destTmp, "version.txt"), []byte(version+"\n"), 0o644); err != nil {
		return err
	}
	InstallLicense(destTmp)
	if runtime.GOOS == "linux" {
		if err := clientfiles.WriteChecksums(destTmp); err != nil {
			return err
		}
	}

	destBak := ""
	if st, err := os.Lstat(dest); err == nil {
		if !st.IsDir() {
			return fmt.Errorf("active client must be a real directory")
		}
		destBak, err = reserveBackup(filepath.Dir(dest), dest)
		if err != nil {
			return err
		}
		if err = os.Rename(dest, destBak); err != nil {
			_ = os.Remove(destBak)
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.Rename(destTmp, dest); err != nil {
		if st, e := os.Stat(destBak); e == nil && st.IsDir() {
			_ = os.Rename(destBak, dest)
		}
		return err
	}
	if err := PruneBackups(); err != nil {
		return err
	}
	if err := linkHooksCurrentUser(o.AllUsers); err != nil {
		return err
	}
	_ = runtime.GOOS
	return nil
}

func copyTreeFiles(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil || rel == "." {
			if info.IsDir() {
				return nil
			}
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() { _ = in.Close() }()
		mode := info.Mode()
		if mode&0o111 != 0 {
			mode = 0o755
		} else {
			mode = 0o644
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(out, in)
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})
}

func linkHooksCurrentUser(allUsers bool) error {
	if runtime.GOOS == "darwin" {
		return nil
	}
	homes, err := hookHomes(allUsers)
	if err != nil {
		return err
	}
	for _, home := range homes {
		if err := linkHooksForHome(home); err != nil {
			return fmt.Errorf("link hooks for %s: %w", home, err)
		}
	}
	return nil
}

func linkHooksForHome(home string) error {
	var steam string
	for _, cand := range []string{
		filepath.Join(home, ".local/share/Steam"),
		filepath.Join(home, ".steam/steam"),
		filepath.Join(home, ".steam/root"),
		filepath.Join(home, ".var/app/com.valvesoftware.Steam/.local/share/Steam"),
	} {
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			steam = cand
			break
		}
	}
	if steam == "" {
		return nil
	}
	return repair.InstallHooksAt(steam, InstallRoot())
}
