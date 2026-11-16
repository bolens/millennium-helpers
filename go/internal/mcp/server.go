package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/bolens/millennium-helpers/internal/version"
)

// ServeStdio processes one JSON-RPC message per line. IDs use json.Number so
// integers beyond float64's exact range round-trip without changing identity.
func ServeStdio(in io.Reader, out io.Writer) error {
	logf("Millennium Helpers MCP server started.")
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	encoder := json.NewEncoder(out)
	respond := func(id any, result any, code int, message string) error {
		response := map[string]any{"jsonrpc": "2.0", "id": id}
		if code != 0 {
			response["error"] = map[string]any{"code": code, "message": message}
		} else {
			response["result"] = result
		}
		return encoder.Encode(response)
	}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if !json.Valid([]byte(line)) {
			if err := respond(nil, nil, -32700, "Parse error"); err != nil {
				return err
			}
			continue
		}
		var req map[string]any
		decoder := json.NewDecoder(strings.NewReader(line))
		decoder.UseNumber()
		if err := decoder.Decode(&req); err != nil || req == nil {
			if err := respond(nil, nil, -32600, "Invalid Request"); err != nil {
				return err
			}
			continue
		}
		method, ok := req["method"].(string)
		id, hasID := req["id"]
		validID := true
		switch id.(type) {
		case nil, string, json.Number:
		default:
			validID = false
		}
		if req["jsonrpc"] != "2.0" || !ok || method == "" || !validID {
			if err := respond(nil, nil, -32600, "Invalid Request"); err != nil {
				return err
			}
			continue
		}
		// MCP tool invocation is a request. Notifications must never execute a tool
		// or produce a response, including notifications with an unknown method.
		if !hasID {
			continue
		}
		var result any
		switch method {
		case "initialize":
			result = map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{"tools": map[string]any{}}, "serverInfo": map[string]any{"name": "millennium-helpers-mcp", "version": version.Resolve()}}
		case "tools/list":
			result = map[string]any{"tools": ToolsList()}
		case "tools/call":
			params, valid := req["params"].(map[string]any)
			name, nameOK := params["name"].(string)
			args, argsOK := params["arguments"].(map[string]any)
			if !valid || !nameOK || name == "" || (params["arguments"] != nil && !argsOK) {
				if err := respond(id, nil, -32602, "Invalid params"); err != nil {
					return err
				}
				continue
			}
			result = HandleToolCall(name, args)
		default:
			if err := respond(id, nil, -32601, fmt.Sprintf("Method not found: %s", method)); err != nil {
				return err
			}
			continue
		}
		if err := respond(id, result, 0, ""); err != nil {
			return err
		}
	}
	return scanner.Err()
}
