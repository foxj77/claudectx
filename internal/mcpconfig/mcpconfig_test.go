package mcpconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadMCPServers_FileNotExists(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "claude.json")

	servers, err := LoadMCPServers(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(servers) != 0 {
		t.Fatalf("expected empty servers map, got %+v", servers)
	}
}

func TestLoadMCPServers_InvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "claude.json")
	if err := os.WriteFile(path, []byte("not-json"), 0644); err != nil {
		t.Fatalf("failed to write invalid json file: %v", err)
	}

	_, err := LoadMCPServers(path)
	if err == nil {
		t.Fatal("expected error for invalid json, got nil")
	}
}

func TestLoadAndSaveMCPServers_Roundtrip(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "claude.json")

	in := MCPServers{
		"one": MCPServer{Type: "exec", Command: "echo"},
	}

	if err := SaveMCPServers(path, in); err != nil {
		t.Fatalf("SaveMCPServers failed: %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	// File should be valid JSON and contain mcpServers key
	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("written file is not json: %v", err)
	}

	if _, ok := parsed["mcpServers"]; !ok {
		t.Fatalf("mcpServers key not written")
	}

	out, err := LoadMCPServers(path)
	if err != nil {
		t.Fatalf("LoadMCPServers failed: %v", err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("roundtrip mismatch: in=%+v out=%+v", in, out)
	}
}

func TestSaveMCPServers_PreserveOtherFields(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "claude.json")

	// initial file with other fields
	orig := `{"foo":"bar","mcpServers":{"one":{"type":"exec","command":"echo"}}}`
	_ = os.WriteFile(path, []byte(orig), 0644)

	// update with empty servers (should remove mcpServers)
	if err := SaveMCPServers(path, make(MCPServers)); err != nil {
		t.Fatalf("SaveMCPServers failed: %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("written file is not json: %v", err)
	}

	if _, ok := parsed["mcpServers"]; ok {
		t.Fatalf("mcpServers key should have been removed")
	}

	// Ensure other fields remain
	var foo string
	if err := json.Unmarshal(parsed["foo"], &foo); err != nil {
		t.Fatalf("failed to read foo: %v", err)
	}
	if foo != "bar" {
		t.Fatalf("expected foo=bar, got %q", foo)
	}
}

func TestSaveToFileAndLoadFromFile_Roundtrip(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "mcp.json")

	in := MCPServers{"a": {Type: "exec", Command: "true"}}
	if err := SaveToFile(path, in); err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	out, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("roundtrip mismatch in/out: %v %v", in, out)
	}
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "mcp.json")
	_ = os.WriteFile(path, []byte("x"), 0644)

	_, err := LoadFromFile(path)
	if err == nil {
		t.Fatal("expected error for invalid json, got nil")
	}
}

func TestFileExists(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "f.txt")
	if FileExists(path) {
		t.Fatal("expected FileExists false for non-existent file")
	}
	_ = os.WriteFile(path, []byte("ok"), 0644)
	if !FileExists(path) {
		t.Fatal("expected FileExists true for existing file")
	}
}
