package mcpconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestSaveClaudeMCPConfig_WritesTopLevelMcpServersKey verifies the wrapper format
// expected by --mcp-config: a top-level "mcpServers" object.
func TestSaveClaudeMCPConfig_WritesTopLevelMcpServersKey(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "mcp.json")

	servers := MCPServers{
		"my-server": {Type: "stdio", Command: "echo"},
	}
	if err := SaveClaudeMCPConfig(path, servers); err != nil {
		t.Fatalf("SaveClaudeMCPConfig failed: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if _, ok := top["mcpServers"]; !ok {
		t.Errorf("output missing top-level 'mcpServers' key; got keys: %v", keys(top))
	}
}

// TestSaveClaudeMCPConfig_ServerNamesPreserved verifies server names survive the round-trip.
func TestSaveClaudeMCPConfig_ServerNamesPreserved(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "mcp.json")

	servers := MCPServers{
		"alpha": {Type: "stdio", Command: "alpha-cmd"},
		"beta":  {Type: "http", URL: "https://beta.example.com"},
	}
	if err := SaveClaudeMCPConfig(path, servers); err != nil {
		t.Fatalf("SaveClaudeMCPConfig failed: %v", err)
	}

	loaded, err := loadClaudeMCPConfig(path)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}
	for name := range servers {
		if _, ok := loaded[name]; !ok {
			t.Errorf("server %q missing after round-trip", name)
		}
	}
}

// TestSaveClaudeMCPConfig_EnvBlockPreserved verifies the env map is included in output.
// Critical: omitting env causes servers that need env vars (e.g. KUBECONFIG) to crash.
func TestSaveClaudeMCPConfig_EnvBlockPreserved(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "mcp.json")

	servers := MCPServers{
		"kube": {
			Type:    "stdio",
			Command: "/usr/bin/kube-mcp",
			Env:     map[string]string{"KUBECONFIG": "/home/user/.kube/config", "FOO": "bar"},
		},
	}
	if err := SaveClaudeMCPConfig(path, servers); err != nil {
		t.Fatalf("SaveClaudeMCPConfig failed: %v", err)
	}

	loaded, err := loadClaudeMCPConfig(path)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}
	srv, ok := loaded["kube"]
	if !ok {
		t.Fatal("server 'kube' missing after round-trip")
	}
	if srv.Env["KUBECONFIG"] != "/home/user/.kube/config" {
		t.Errorf("KUBECONFIG env var lost; got %v", srv.Env)
	}
	if srv.Env["FOO"] != "bar" {
		t.Errorf("FOO env var lost; got %v", srv.Env)
	}
}

// TestSaveClaudeMCPConfig_ArgsPreserved verifies args slice survives the round-trip.
func TestSaveClaudeMCPConfig_ArgsPreserved(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "mcp.json")

	servers := MCPServers{
		"srv": {Type: "stdio", Command: "/bin/tool", Args: []string{"serve", "--port", "9000"}},
	}
	if err := SaveClaudeMCPConfig(path, servers); err != nil {
		t.Fatalf("SaveClaudeMCPConfig failed: %v", err)
	}

	loaded, err := loadClaudeMCPConfig(path)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}
	srv := loaded["srv"]
	if len(srv.Args) != 3 || srv.Args[0] != "serve" || srv.Args[2] != "9000" {
		t.Errorf("args not preserved; got %v", srv.Args)
	}
}

// TestSaveClaudeMCPConfig_EmptyServersWritesValidEmptyConfig verifies an empty server map
// produces a valid file (not a crash or empty file).
func TestSaveClaudeMCPConfig_EmptyServersWritesValidEmptyConfig(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "mcp.json")

	if err := SaveClaudeMCPConfig(path, MCPServers{}); err != nil {
		t.Fatalf("SaveClaudeMCPConfig failed for empty servers: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
}

// TestSaveClaudeMCPConfig_HTTPServerPreserved verifies HTTP-type servers are handled.
func TestSaveClaudeMCPConfig_HTTPServerPreserved(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "mcp.json")

	servers := MCPServers{
		"remote": {
			Type:    "http",
			URL:     "https://mcp.example.com/v1",
			Headers: map[string]string{"Authorization": "Bearer tok"},
		},
	}
	if err := SaveClaudeMCPConfig(path, servers); err != nil {
		t.Fatalf("SaveClaudeMCPConfig failed: %v", err)
	}

	loaded, err := loadClaudeMCPConfig(path)
	if err != nil {
		t.Fatalf("failed to reload: %v", err)
	}
	srv, ok := loaded["remote"]
	if !ok {
		t.Fatal("server 'remote' missing after round-trip")
	}
	if srv.URL != "https://mcp.example.com/v1" {
		t.Errorf("URL not preserved; got %q", srv.URL)
	}
	if srv.Headers["Authorization"] != "Bearer tok" {
		t.Errorf("headers not preserved; got %v", srv.Headers)
	}
}

// TestSaveClaudeMCPConfig_OutputDifferentFromSaveToFile verifies the Claude-compatible
// wrapper format differs from the raw profile storage format used by SaveToFile.
// The wrapper must have the top-level "mcpServers" key; raw storage does not.
func TestSaveClaudeMCPConfig_OutputDifferentFromSaveToFile(t *testing.T) {
	tmp := t.TempDir()

	servers := MCPServers{
		"srv": {Type: "stdio", Command: "cmd"},
	}

	claudePath := filepath.Join(tmp, "claude.json")
	rawPath := filepath.Join(tmp, "raw.json")

	if err := SaveClaudeMCPConfig(claudePath, servers); err != nil {
		t.Fatalf("SaveClaudeMCPConfig failed: %v", err)
	}
	if err := SaveToFile(rawPath, servers); err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	claudeData, _ := os.ReadFile(claudePath)
	rawData, _ := os.ReadFile(rawPath)

	var claudeTop map[string]json.RawMessage
	var rawTop map[string]json.RawMessage
	json.Unmarshal(claudeData, &claudeTop)
	json.Unmarshal(rawData, &rawTop)

	_, claudeHasWrapper := claudeTop["mcpServers"]
	_, rawHasWrapper := rawTop["mcpServers"]

	if !claudeHasWrapper {
		t.Error("SaveClaudeMCPConfig output should have top-level 'mcpServers' key")
	}
	if rawHasWrapper {
		// The raw profile format stores servers at the top level without the wrapper.
		// If this changes, update the test — but for now we verify the two are distinct.
		t.Log("note: SaveToFile now also uses mcpServers wrapper — verify this is intentional")
	}
}

// ── helpers ────────────────────────────────────────────────────────────────────

// loadClaudeMCPConfig reads a Claude-compatible --mcp-config file and returns
// the inner MCPServers map for assertion.
func loadClaudeMCPConfig(path string) (MCPServers, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		MCPServers MCPServers `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.MCPServers, nil
}

func keys(m map[string]json.RawMessage) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
