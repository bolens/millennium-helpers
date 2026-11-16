package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/bolens/millennium-helpers/internal/atomicfile"
	"github.com/bolens/millennium-helpers/internal/dirlock"
	"github.com/bolens/millennium-helpers/internal/jsonobject"
)

const (
	mcpServerName = "millennium-helpers"
	mcpCommand    = "millennium"
)

// RegisterResult summarizes --register outcome.
type RegisterResult struct {
	RegisteredAny bool
	Lines         []string
}

type mcpHostConfig struct {
	Label string
	Path  string
	Key   string
}

func mcpHostConfigs() []mcpHostConfig {
	home, _ := os.UserHomeDir()
	claude := filepath.Join(home, ".config", "Claude", "claude_desktop_config.json")
	if runtime.GOOS == "windows" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			claude = filepath.Join(appdata, "Claude", "claude_desktop_config.json")
		}
	}
	return []mcpHostConfig{
		{Label: "Claude Desktop", Path: claude, Key: "mcpServers"},
		{Label: "Windsurf", Path: filepath.Join(home, ".codeium", "windsurf", "mcp_config.json"), Key: "mcpServers"},
		{Label: "Cursor", Path: filepath.Join(home, ".cursor", "mcp.json"), Key: "mcpServers"},
	}
}

// Register writes millennium-helpers into detected AI client MCP configs.
func Register() RegisterResult {
	var out RegisterResult
	serverConfig := map[string]any{"command": mcpCommand, "args": []string{"mcp"}}

	for _, cfg := range mcpHostConfigs() {
		dir := filepath.Dir(cfg.Path)
		if st, err := os.Stat(dir); err != nil || !st.IsDir() {
			continue
		}

		out.Lines = append(out.Lines, fmt.Sprintf("Registering Millennium Helpers MCP server in %s...", cfg.Label))

		registered, err := registerHost(cfg, serverConfig)
		if err != nil {
			out.Lines = append(out.Lines, fmt.Sprintf("  Error: %s config left unchanged: %v", cfg.Label, err))
			continue
		}
		if registered {
			out.Lines = append(out.Lines, fmt.Sprintf("  Registered in %s config: %s", cfg.Label, cfg.Path))
			out.RegisteredAny = true
		}
	}

	snippet, _ := json.MarshalIndent(map[string]any{
		"mcpServers": map[string]any{mcpServerName: serverConfig},
	}, "", "  ")
	out.Lines = append(out.Lines, "", "Manual config snippet (any MCP host):", string(snippet))

	if !out.RegisteredAny {
		out.Lines = append(out.Lines, "",
			"No active config directories found (Claude Desktop, Windsurf, or Cursor).",
			"Paste the snippet above into your MCP client's config file.")
	} else {
		out.Lines = append(out.Lines, "",
			"Registration check completed successfully.",
			"Restart Cursor / Claude Desktop / Windsurf so the MCP server appears.",
			"See docs/mcp.md for tool details and troubleshooting.")
	}
	return out
}

// RunRegister prints Register() output and returns the process exit code.
func RunRegister() int {
	res := Register()
	for _, line := range res.Lines {
		fmt.Println(line)
	}
	if !res.RegisteredAny {
		return 1
	}
	return 0
}

func registerHost(cfg mcpHostConfig, serverConfig map[string]any) (bool, error) {
	lock, err := dirlock.Acquire(filepath.Dir(cfg.Path))
	if err != nil {
		return false, err
	}
	defer func() { _ = lock.Close() }()
	data := map[string]any{}
	raw, err := os.ReadFile(cfg.Path)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	if err == nil {
		data, err = jsonobject.Decode(raw)
		if err != nil {
			return false, fmt.Errorf("invalid JSON: %w", err)
		}
	}
	servers := map[string]any{}
	if value, present := data[cfg.Key]; present {
		var ok bool
		servers, ok = value.(map[string]any)
		if !ok {
			return false, fmt.Errorf("%s must be a JSON object", cfg.Key)
		}
	} else {
		data[cfg.Key] = servers
	}
	if existing, ok := servers[mcpServerName].(map[string]any); ok {
		args, _ := existing["args"].([]any)
		if command, _ := existing["command"].(string); command == mcpCommand && len(args) == 1 && args[0] == "mcp" {
			return true, nil
		}
	}
	servers[mcpServerName] = serverConfig
	raw, err = json.MarshalIndent(data, "", "  ")
	if err != nil {
		return false, err
	}
	mode := os.FileMode(0644)
	if info, err := os.Stat(cfg.Path); err == nil {
		mode = info.Mode().Perm()
	}
	if err := atomicfile.WriteFile(cfg.Path, append(raw, '\n'), mode); err != nil {
		return false, err
	}
	return true, nil
}
