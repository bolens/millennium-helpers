// Package usercontext resolves the account whose Steam installation is managed.
package usercontext

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

// Context separates the target user's directories from an elevated process's environment.
type Context struct {
	Name, Home, ConfigHome, DataHome string
	UID, GID                         string
}

// Resolve uses the sudo caller when elevated. HOME and XDG overrides apply only
// to the unelevated account; inherited root directories must never become targets.
func Resolve() (Context, error) {
	return resolve(os.Geteuid(), os.Getenv, os.UserHomeDir, user.Current, user.Lookup)
}

func resolve(euid int, env func(string) string, homeDir func() (string, error), current func() (*user.User, error), lookup func(string) (*user.User, error)) (Context, error) {
	var c Context
	var u *user.User
	var err error
	sudo := euid == 0 && env("SUDO_USER") != ""
	if sudo {
		u, err = lookup(env("SUDO_USER"))
	} else {
		u, err = current()
	}
	if err != nil || u == nil {
		return c, fmt.Errorf("cannot resolve invoking user")
	}
	c.Name, c.UID, c.GID = u.Username, u.Uid, u.Gid
	if sudo {
		c.Home = u.HomeDir
	} else {
		c.Home, err = homeDir()
		if err != nil {
			return Context{}, fmt.Errorf("cannot resolve invoking user's home")
		}
		c.ConfigHome, c.DataHome = env("XDG_CONFIG_HOME"), env("XDG_DATA_HOME")
	}
	if !filepath.IsAbs(c.Home) {
		return Context{}, fmt.Errorf("invoking user's home must be absolute")
	}
	if !filepath.IsAbs(c.ConfigHome) {
		c.ConfigHome = filepath.Join(c.Home, ".config")
	}
	if !filepath.IsAbs(c.DataHome) {
		c.DataHome = filepath.Join(c.Home, ".local", "share")
	}
	return c, nil
}
