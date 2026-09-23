//go:build !linux

package repair

func InstallHooksAt(steamRoot, libRoot string) error { return nil }
