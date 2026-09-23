package usercontext

import (
	"errors"
	"os/user"
	"path/filepath"
	"testing"
)

func TestResolveCallerDirectories(t *testing.T) {
	callerHome, processHome := t.TempDir(), t.TempDir()
	caller := &user.User{Username: "caller", HomeDir: callerHome, Uid: "1000", Gid: "1000"}
	process := &user.User{Username: "process", HomeDir: processHome, Uid: "0", Gid: "0"}
	envs := map[string]string{"SUDO_USER": "caller", "XDG_CONFIG_HOME": filepath.Join(processHome, "config"), "XDG_DATA_HOME": filepath.Join(processHome, "data")}
	env := func(k string) string { return envs[k] }
	current := func() (*user.User, error) { return process, nil }
	home := func() (string, error) { return processHome, nil }
	lookup := func(name string) (*user.User, error) {
		if name != "caller" {
			t.Fatal(name)
		}
		return caller, nil
	}
	c, err := resolve(0, env, home, current, lookup)
	if err != nil {
		t.Fatal(err)
	}
	if c.Name != "caller" || c.Home != callerHome || c.ConfigHome != filepath.Join(callerHome, ".config") || c.DataHome != filepath.Join(callerHome, ".local", "share") || c.UID != "1000" {
		t.Fatalf("wrong sudo context: %+v", c)
	}
	c, err = resolve(1000, env, home, current, lookup)
	if err != nil || c.Home != processHome || c.ConfigHome != envs["XDG_CONFIG_HOME"] {
		t.Fatalf("normal overrides lost: %+v %v", c, err)
	}
	_, err = resolve(0, env, home, current, func(string) (*user.User, error) { return nil, errors.New("unknown") })
	if err == nil {
		t.Fatal("invalid sudo caller silently fell back")
	}
}
