package mcp

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

type protocolResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *struct {
		Code int `json:"code"`
	} `json:"error"`
}

func responses(t *testing.T, input string) []protocolResponse {
	t.Helper()
	var out bytes.Buffer
	if err := ServeStdio(strings.NewReader(input), &out); err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(&out)
	var result []protocolResponse
	for {
		var response protocolResponse
		err := decoder.Decode(&response)
		if errors.Is(err, io.EOF) {
			return result
		}
		if err != nil {
			t.Fatal(err)
		}
		if response.JSONRPC != "2.0" {
			t.Fatalf("invalid protocol=%q", response.JSONRPC)
		}
		result = append(result, response)
	}
}

func TestInitializeAndToolCatalog(t *testing.T) {
	got := responses(t, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`+"\n"+`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`+"\n")
	if len(got) != 2 || string(got[0].ID) != "1" || string(got[1].ID) != "2" {
		t.Fatalf("responses=%+v", got)
	}
	var init struct {
		ProtocolVersion string `json:"protocolVersion"`
		ServerInfo      struct {
			Name string `json:"name"`
		} `json:"serverInfo"`
	}
	if err := json.Unmarshal(got[0].Result, &init); err != nil {
		t.Fatal(err)
	}
	if init.ProtocolVersion == "" || init.ServerInfo.Name != "millennium-helpers-mcp" {
		t.Fatalf("init=%+v", init)
	}
	var catalog struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(got[1].Result, &catalog); err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, tool := range catalog.Tools {
		if names[tool.Name] {
			t.Fatalf("duplicate tool=%s", tool.Name)
		}
		names[tool.Name] = true
	}
	for _, name := range []string{"millennium_diag", "millennium_theme", "millennium_config", "millennium_upgrade", "millennium_schedule", "millennium_repair", "millennium_purge"} {
		if !names[name] {
			t.Fatalf("missing tool=%s", name)
		}
	}
}

func TestProtocolFailuresAndExactIDs(t *testing.T) {
	cases := []struct {
		input, id string
		code      int
	}{
		{`{"jsonrpc":"2.0","id":9007199254740993,"method":"unknown"}`, "9007199254740993", -32601},
		{`{"jsonrpc":"2.0","id":"literal:a,b","method":"unknown"}`, `"literal:a,b"`, -32601},
		{`{"jsonrpc":"2.0","id":null,"method":"unknown"}`, "null", -32601},
		{`{`, "null", -32700}, {`null`, "null", -32600}, {`[]`, "null", -32600},
		{`{"jsonrpc":"2.0","id":true,"method":"tools/list"}`, "null", -32600},
		{`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"millennium_diag","arguments":[]}}`, "4", -32602},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got := responses(t, tc.input+"\n")
			if len(got) != 1 || got[0].Error == nil || got[0].Error.Code != tc.code || string(got[0].ID) != tc.id || len(got[0].Result) != 0 {
				t.Fatalf("responses=%+v", got)
			}
		})
	}
}

func TestNotificationsAndRecovery(t *testing.T) {
	t.Setenv("TEST_SUITE_RUN", "1")
	t.Setenv("MOCK_BIN", t.TempDir())
	input := `{"jsonrpc":"2.0","method":"tools/call","params":{"name":"millennium_purge","arguments":{"confirm":true}}}` + "\n" + `bad json` + "\n" + `{"jsonrpc":"2.0","method":"notifications/initialized"}` + "\n" + `{"jsonrpc":"2.0","id":7,"method":"tools/list"}` + "\n"
	got := responses(t, input)
	if len(got) != 2 || got[0].Error == nil || got[0].Error.Code != -32700 || string(got[1].ID) != "7" {
		t.Fatalf("responses=%+v", got)
	}
}

func TestServerWriterFailure(t *testing.T) {
	err := ServeStdio(strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`+"\n"), failedWriter{})
	if !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("err=%v", err)
	}
}

type failedWriter struct{}

func (failedWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }
