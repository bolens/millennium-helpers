package mcp

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/bolens/millennium-helpers/internal/dirlock"
)

func TestRegisterWritesCursorAndClaude(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", filepath.Join(home, "AppData"))

	claudeDir := filepath.Join(home, ".config", "Claude")
	cursorDir := filepath.Join(home, ".cursor")
	if err := os.MkdirAll(claudeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cursorDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "claude_desktop_config.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res := Register()
	if !res.RegisteredAny {
		t.Fatalf("expected registration, lines=%v", res.Lines)
	}

	raw, err := os.ReadFile(filepath.Join(cursorDir, "mcp.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cursor map[string]any
	if err := json.Unmarshal(raw, &cursor); err != nil {
		t.Fatal(err)
	}
	servers, _ := cursor["mcpServers"].(map[string]any)
	entry, _ := servers["millennium-helpers"].(map[string]any)
	args, _ := entry["args"].([]any)
	if entry["command"] != "millennium" || len(args) != 1 || args[0] != "mcp" {
		t.Fatalf("cursor command: %#v", entry)
	}
}

func TestRegisterPreservesInvalidConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", filepath.Join(home, "AppData"))

	cursorDir := filepath.Join(home, ".cursor")
	if err := os.MkdirAll(cursorDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(cursorDir, "mcp.json")
	const invalid = "{not-json\n"
	if err := os.WriteFile(path, []byte(invalid), 0o644); err != nil {
		t.Fatal(err)
	}

	res := Register()
	if res.RegisteredAny {
		t.Fatalf("invalid-only config should not register: %v", res.Lines)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != invalid {
		t.Fatalf("invalid config was overwritten: %q", raw)
	}
}

func TestRegisterPreservesDocumentAndRejectsInvalidNamespace(t *testing.T) {
	config := mcpHostConfig{Label: "fixture", Path: filepath.Join(t.TempDir(), "config.json"), Key: "mcpServers"}
	for _, body := range []string{"null", `{"mcpServers":[]}`, `{"mcpServers":"keep"}`, `{broken`} {
		if err := os.WriteFile(config.Path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := registerHost(config, map[string]any{"command": mcpCommand, "args": []string{"mcp"}}); err == nil {
			t.Fatalf("accepted %s", body)
		}
		after, err := os.ReadFile(config.Path)
		if err != nil || string(after) != body {
			t.Fatal("invalid document overwritten")
		}
	}
	if err := os.WriteFile(config.Path, []byte(`{"unrelated":9007199254740993}`), 0600); err != nil {
		t.Fatal(err)
	}
	if ok, err := registerHost(config, map[string]any{"command": mcpCommand, "args": []string{"mcp"}}); err != nil || !ok {
		t.Fatalf("registered=%v err=%v", ok, err)
	}
	after, err := os.ReadFile(config.Path)
	if err != nil || !strings.Contains(string(after), "9007199254740993") {
		t.Fatal("unrelated integer changed")
	}
}

func TestRegisterRejectsCompetingWriter(t *testing.T) {
	config := mcpHostConfig{Label: "fixture", Path: filepath.Join(t.TempDir(), "config.json"), Key: "mcpServers"}
	before := []byte(`{"keep":true}`)
	if err := os.WriteFile(config.Path, before, 0600); err != nil {
		t.Fatal(err)
	}
	lock, err := dirlock.Acquire(filepath.Dir(config.Path))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := lock.Close(); err != nil {
			t.Error(err)
		}
	}()
	if _, err := registerHost(config, map[string]any{}); !errors.Is(err, dirlock.ErrBusy) {
		t.Fatalf("err=%v", err)
	}
	after, err := os.ReadFile(config.Path)
	if err != nil || string(after) != string(before) {
		t.Fatal("competing writer changed host config")
	}
}

func TestRegisterDoesNotOverwriteUnreadableDocument(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("Unix discretionary read permissions")
	}
	config := mcpHostConfig{Label: "fixture", Path: filepath.Join(t.TempDir(), "config.json"), Key: "mcpServers"}
	before := []byte(`{"keep":"private"}`)
	if err := os.WriteFile(config.Path, before, 0000); err != nil {
		t.Fatal(err)
	}
	if _, err := registerHost(config, map[string]any{}); !errors.Is(err, os.ErrPermission) {
		t.Fatalf("err=%v", err)
	}
	if err := os.Chmod(config.Path, 0600); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(config.Path)
	if err != nil || string(after) != string(before) {
		t.Fatal("unreadable host config overwritten")
	}
}
