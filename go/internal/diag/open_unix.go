//go:build unix

package diag

import "os"

func openFollowFile(path string) (*os.File, error) { return os.Open(path) }
