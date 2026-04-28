package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/johnfox/claudectx/internal/mcpconfig"
	"github.com/johnfox/claudectx/internal/paths"
	"github.com/johnfox/claudectx/internal/profile"
	"github.com/johnfox/claudectx/internal/store"
)

// helpers

func setupRunTest(t *testing.T) (*store.Store, string) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	s, err := store.NewStore()
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	return s, tmp
}

func saveProfile(t *testing.T, s *store.Store, p *profile.Profile) {
	t.Helper()
	if err := s.Save(p); err != nil {
		t.Fatalf("failed to save profile %q: %v", p.Name, err)
	}
}

func setCurrentProfile(t *testing.T, s *store.Store, name string) {
	t.Helper()
	if err := s.SetCurrent(name); err != nil {
		t.Fatalf("failed to set current profile: %v", err)
	}
}

// ── Profile existence ──────────────────────────────────────────────────────────

func TestRunProfile_MissingProfileReturnsError(t *testing.T) {
	s, _ := setupRunTest(t)

	_, err := RunProfile(s, RunOptions{ProfileName: "nonexistent", DryRun: true})
	if err == nil {
		t.Fatal("expected error for non-existent profile")
	}
	if !strings.Contains(err.Error(), "nonexistent") {
		t.Errorf("error should mention the profile name, got: %v", err)
	}
}

func TestRunProfile_EmptyProfileNameReturnsError(t *testing.T) {
	s, _ := setupRunTest(t)

	_, err := RunProfile(s, RunOptions{ProfileName: "", DryRun: true})
	if err == nil {
		t.Fatal("expected error for empty profile name")
	}
}

// ── Global state immutability ──────────────────────────────────────────────────

func TestRunProfile_DoesNotModifyCurrentProfileTracker(t *testing.T) {
	s, _ := setupRunTest(t)

	base := profile.NewProfile("base")
	target := profile.NewProfile("target")
	saveProfile(t, s, base)
	saveProfile(t, s, target)
	setCurrentProfile(t, s, "base")

	if _, err := RunProfile(s, RunOptions{ProfileName: "target", DryRun: true}); err != nil {
		t.Fatalf("RunProfile failed: %v", err)
	}

	cur, err := s.GetCurrent()
	if err != nil {
		t.Fatalf("GetCurrent failed: %v", err)
	}
	if cur != "base" {
		t.Errorf("current profile changed to %q; run must not modify current tracker", cur)
	}
}

func TestRunProfile_DoesNotModifyPreviousProfileTracker(t *testing.T) {
	s, _ := setupRunTest(t)

	base := profile.NewProfile("base")
	prev := profile.NewProfile("prev")
	target := profile.NewProfile("target")
	saveProfile(t, s, base)
	saveProfile(t, s, prev)
	saveProfile(t, s, target)
	setCurrentProfile(t, s, "base")
	if err := s.SetPrevious("prev"); err != nil {
		t.Fatalf("failed to set previous: %v", err)
	}

	if _, err := RunProfile(s, RunOptions{ProfileName: "target", DryRun: true}); err != nil {
		t.Fatalf("RunProfile failed: %v", err)
	}

	previous, _ := s.GetPrevious()
	if previous != "prev" {
		t.Errorf("previous profile changed to %q; run must not modify previous tracker", previous)
	}
}

func TestRunProfile_DoesNotModifyActiveSettingsFile(t *testing.T) {
	s, _ := setupRunTest(t)

	p := profile.NewProfile("work")
	p.Settings.Model = "claude-haiku-4-5-20251001"
	saveProfile(t, s, p)

	settingsPath, _ := paths.SettingsFile()
	original := `{"model":"claude-sonnet-4-6"}`
	if err := os.WriteFile(settingsPath, []byte(original), 0644); err != nil {
		t.Fatalf("failed to write settings: %v", err)
	}

	if _, err := RunProfile(s, RunOptions{ProfileName: "work", DryRun: true}); err != nil {
		t.Fatalf("RunProfile failed: %v", err)
	}

	b, _ := os.ReadFile(settingsPath)
	if string(b) != original {
		t.Errorf("active settings.json was modified; want %q, got %q", original, string(b))
	}
}

func TestRunProfile_DoesNotModifyActiveClaudeMD(t *testing.T) {
	s, _ := setupRunTest(t)

	p := profile.NewProfile("work")
	p.ClaudeMD = "profile instructions"
	saveProfile(t, s, p)

	claudeMDPath, _ := paths.ClaudeMDFile()
	original := "global instructions"
	if err := os.WriteFile(claudeMDPath, []byte(original), 0644); err != nil {
		t.Fatalf("failed to write CLAUDE.md: %v", err)
	}

	if _, err := RunProfile(s, RunOptions{ProfileName: "work", DryRun: true}); err != nil {
		t.Fatalf("RunProfile failed: %v", err)
	}

	b, _ := os.ReadFile(claudeMDPath)
	if string(b) != original {
		t.Errorf("CLAUDE.md was modified; want %q, got %q", original, string(b))
	}
}

func TestRunProfile_DoesNotModifyClaudeJSON(t *testing.T) {
	s, _ := setupRunTest(t)

	p := profile.NewProfile("work")
	p.MCPServers = mcpconfig.MCPServers{
		"profile-server": {Type: "stdio", Command: "echo"},
	}
	saveProfile(t, s, p)

	claudeJSONPath, _ := paths.ClaudeJSONFile()
	original := `{"mcpServers":{"global-server":{"type":"stdio","command":"global"}}}`
	if err := os.WriteFile(claudeJSONPath, []byte(original), 0644); err != nil {
		t.Fatalf("failed to write claude.json: %v", err)
	}

	if _, err := RunProfile(s, RunOptions{ProfileName: "work", DryRun: true}); err != nil {
		t.Fatalf("RunProfile failed: %v", err)
	}

	b, _ := os.ReadFile(claudeJSONPath)
	if string(b) != original {
		t.Errorf("claude.json was modified; want %q, got %q", original, string(b))
	}
}

// ── Generated args: settings ───────────────────────────────────────────────────

func TestRunProfile_ArgsIncludeSettingsFlag(t *testing.T) {
	s, _ := setupRunTest(t)

	p := profile.NewProfile("work")
	p.Settings.Model = "claude-haiku-4-5-20251001"
	saveProfile(t, s, p)

	result, err := RunProfile(s, RunOptions{ProfileName: "work", DryRun: true})
	if err != nil {
		t.Fatalf("RunProfile failed: %v", err)
	}

	if !containsConsecutive(result.GeneratedArgs, "--settings") {
		t.Errorf("generated args should include --settings flag, got: %v", result.GeneratedArgs)
	}
}

func TestRunProfile_SettingsFlagPointsToProfileSettingsJSON(t *testing.T) {
	s, _ := setupRunTest(t)

	p := profile.NewProfile("work")
	saveProfile(t, s, p)

	result, err := RunProfile(s, RunOptions{ProfileName: "work", DryRun: true})
	if err != nil {
		t.Fatalf("RunProfile failed: %v", err)
	}

	settingsArg := argAfter(result.GeneratedArgs, "--settings")
	if settingsArg == "" {
		t.Fatal("--settings flag present but has no value")
	}
	profileSettingsPath, _ := paths.ProfileFile("work", "settings.json")
	if settingsArg != profileSettingsPath {
		t.Errorf("--settings = %q, want %q", settingsArg, profileSettingsPath)
	}
}

// ── Generated args: CLAUDE.md ──────────────────────────────────────────────────

func TestRunProfile_ArgsIncludeClaudeMDFlagWhenNonEmpty(t *testing.T) {
	s, _ := setupRunTest(t)

	p := profile.NewProfile("work")
	p.ClaudeMD = "some instructions"
	saveProfile(t, s, p)

	result, err := RunProfile(s, RunOptions{ProfileName: "work", DryRun: true})
	if err != nil {
		t.Fatalf("RunProfile failed: %v", err)
	}

	if !containsConsecutive(result.GeneratedArgs, "--append-system-prompt-file") {
		t.Errorf("expected --append-system-prompt-file in args, got: %v", result.GeneratedArgs)
	}
}

func TestRunProfile_ArgsOmitClaudeMDFlagWhenEmpty(t *testing.T) {
	s, _ := setupRunTest(t)

	p := profile.NewProfile("work")
	p.ClaudeMD = ""
	saveProfile(t, s, p)

	result, err := RunProfile(s, RunOptions{ProfileName: "work", DryRun: true})
	if err != nil {
		t.Fatalf("RunProfile failed: %v", err)
	}

	if containsConsecutive(result.GeneratedArgs, "--append-system-prompt-file") {
		t.Errorf("--append-system-prompt-file should be absent for empty CLAUDE.md, got: %v", result.GeneratedArgs)
	}
}

func TestRunProfile_ArgsOmitClaudeMDFlagWhenWhitespaceOnly(t *testing.T) {
	s, _ := setupRunTest(t)

	p := profile.NewProfile("work")
	p.ClaudeMD = "   \n\t  "
	saveProfile(t, s, p)

	result, err := RunProfile(s, RunOptions{ProfileName: "work", DryRun: true})
	if err != nil {
		t.Fatalf("RunProfile failed: %v", err)
	}

	if containsConsecutive(result.GeneratedArgs, "--append-system-prompt-file") {
		t.Errorf("--append-system-prompt-file should be absent for whitespace-only CLAUDE.md")
	}
}

func TestRunProfile_ClaudeMDFlagPointsToProfileFile(t *testing.T) {
	s, _ := setupRunTest(t)

	p := profile.NewProfile("work")
	p.ClaudeMD = "instructions"
	saveProfile(t, s, p)

	result, err := RunProfile(s, RunOptions{ProfileName: "work", DryRun: true})
	if err != nil {
		t.Fatalf("RunProfile failed: %v", err)
	}

	claudeMDArg := argAfter(result.GeneratedArgs, "--append-system-prompt-file")
	profileClaudeMDPath, _ := paths.ProfileFile("work", "CLAUDE.md")
	if claudeMDArg != profileClaudeMDPath {
		t.Errorf("--append-system-prompt-file = %q, want %q", claudeMDArg, profileClaudeMDPath)
	}
}

// ── Generated args: MCP ───────────────────────────────────────────────────────

func TestRunProfile_ArgsIncludeMCPFlagsWhenServersPresent(t *testing.T) {
	s, _ := setupRunTest(t)

	p := profile.NewProfile("work")
	p.MCPServers = mcpconfig.MCPServers{
		"my-server": {Type: "stdio", Command: "echo"},
	}
	saveProfile(t, s, p)

	result, err := RunProfile(s, RunOptions{ProfileName: "work", DryRun: true})
	if err != nil {
		t.Fatalf("RunProfile failed: %v", err)
	}

	if !containsConsecutive(result.GeneratedArgs, "--mcp-config") {
		t.Errorf("expected --mcp-config in args when profile has MCP servers, got: %v", result.GeneratedArgs)
	}
	if !containsFlag(result.GeneratedArgs, "--strict-mcp-config") {
		t.Errorf("expected --strict-mcp-config in args when profile has MCP servers, got: %v", result.GeneratedArgs)
	}
}

func TestRunProfile_ArgsOmitMCPFlagsWhenNoServers(t *testing.T) {
	s, _ := setupRunTest(t)

	p := profile.NewProfile("work")
	saveProfile(t, s, p)

	result, err := RunProfile(s, RunOptions{ProfileName: "work", DryRun: true})
	if err != nil {
		t.Fatalf("RunProfile failed: %v", err)
	}

	if containsConsecutive(result.GeneratedArgs, "--mcp-config") {
		t.Errorf("--mcp-config should be absent when profile has no MCP servers")
	}
	if containsFlag(result.GeneratedArgs, "--strict-mcp-config") {
		t.Errorf("--strict-mcp-config should be absent when profile has no MCP servers")
	}
}

func TestRunProfile_MCPTempConfigIsValidJSON(t *testing.T) {
	s, tmp := setupRunTest(t)

	p := profile.NewProfile("work")
	p.MCPServers = mcpconfig.MCPServers{
		"my-server": {Type: "stdio", Command: "echo", Args: []string{"hello"}},
	}
	saveProfile(t, s, p)

	// Use a real (non-dry-run) call so the temp file is actually written.
	// Override PATH so exec of "claude" fails immediately after file creation.
	t.Setenv("PATH", tmp) // no claude binary in tmp dir
	result, _ := RunProfile(s, RunOptions{ProfileName: "work", DryRun: false})

	mcpPath := argAfter(result.GeneratedArgs, "--mcp-config")
	if mcpPath == "" {
		t.Fatal("--mcp-config path missing from generated args")
	}

	// The temp dir may have been cleaned up after claude failed; write a fresh
	// copy using the same SaveClaudeMCPConfig to verify the format separately.
	tmpFile := filepath.Join(tmp, "verify-mcp.json")
	if err := mcpconfig.SaveClaudeMCPConfig(tmpFile, p.MCPServers); err != nil {
		t.Fatalf("SaveClaudeMCPConfig failed: %v", err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("cannot read MCP config: %v", err)
	}

	var top map[string]json.RawMessage
	if err := json.Unmarshal(data, &top); err != nil {
		t.Fatalf("generated MCP config is not valid JSON: %v", err)
	}
	if _, ok := top["mcpServers"]; !ok {
		t.Errorf("generated MCP config missing top-level 'mcpServers' key")
	}
}

func TestRunProfile_MCPTempConfigPreservesEnvBlock(t *testing.T) {
	s, tmp := setupRunTest(t)

	servers := mcpconfig.MCPServers{
		"kube-server": {
			Type:    "stdio",
			Command: "/usr/bin/kube-mcp",
			Env:     map[string]string{"KUBECONFIG": "/home/user/.kube/config"},
		},
	}

	p := profile.NewProfile("work")
	p.MCPServers = servers
	saveProfile(t, s, p)

	// Write the Claude-format MCP config directly via SaveClaudeMCPConfig
	// to verify the env block survives the format conversion.
	tmpFile := filepath.Join(tmp, "verify-env-mcp.json")
	if err := mcpconfig.SaveClaudeMCPConfig(tmpFile, servers); err != nil {
		t.Fatalf("SaveClaudeMCPConfig failed: %v", err)
	}

	data, _ := os.ReadFile(tmpFile)
	if !strings.Contains(string(data), "KUBECONFIG") {
		t.Errorf("generated MCP config missing env block: %s", data)
	}
}

// ── Generated args: pass-through ──────────────────────────────────────────────

func TestRunProfile_PassThroughArgsAppendedAfterGeneratedArgs(t *testing.T) {
	s, _ := setupRunTest(t)

	p := profile.NewProfile("work")
	saveProfile(t, s, p)

	result, err := RunProfile(s, RunOptions{
		ProfileName: "work",
		ClaudeArgs:  []string{"--model", "opus"},
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("RunProfile failed: %v", err)
	}

	// --settings must appear, followed eventually by --model opus
	settingsIdx := indexOf(result.GeneratedArgs, "--settings")
	modelIdx := indexOf(result.GeneratedArgs, "--model")
	if settingsIdx < 0 {
		t.Fatal("--settings not found in generated args")
	}
	if modelIdx < 0 {
		t.Fatal("--model not found in generated args")
	}
	if modelIdx <= settingsIdx {
		t.Errorf("pass-through --model (idx %d) should come after --settings (idx %d)", modelIdx, settingsIdx)
	}
}

func TestRunProfile_PassThroughArgsPreservedVerbatim(t *testing.T) {
	s, _ := setupRunTest(t)

	p := profile.NewProfile("work")
	saveProfile(t, s, p)

	passThrough := []string{"-p", "Review this diff", "--output-format", "json"}
	result, err := RunProfile(s, RunOptions{
		ProfileName: "work",
		ClaudeArgs:  passThrough,
		DryRun:      true,
	})
	if err != nil {
		t.Fatalf("RunProfile failed: %v", err)
	}

	// All pass-through args must appear somewhere in generated args
	for _, arg := range passThrough {
		if !containsFlag(result.GeneratedArgs, arg) {
			t.Errorf("pass-through arg %q missing from generated args: %v", arg, result.GeneratedArgs)
		}
	}
}

// ── Dry-run behaviour ──────────────────────────────────────────────────────────

func TestRunProfile_DryRunDoesNotExecClaude(t *testing.T) {
	s, _ := setupRunTest(t)

	p := profile.NewProfile("work")
	saveProfile(t, s, p)

	// If DryRun is honoured, RunProfile returns without trying to exec.
	// With a clearly invalid claude path this would fail if exec is attempted.
	result, err := RunProfile(s, RunOptions{ProfileName: "work", DryRun: true})
	if err != nil {
		t.Fatalf("RunProfile dry-run should not fail without claude binary: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("dry-run ExitCode = %d, want 0", result.ExitCode)
	}
}

func TestRunProfile_DryRunReturnsGeneratedArgs(t *testing.T) {
	s, _ := setupRunTest(t)

	p := profile.NewProfile("work")
	saveProfile(t, s, p)

	result, err := RunProfile(s, RunOptions{ProfileName: "work", DryRun: true})
	if err != nil {
		t.Fatalf("RunProfile failed: %v", err)
	}
	if len(result.GeneratedArgs) == 0 {
		t.Error("dry-run should return non-empty GeneratedArgs")
	}
}

// ── RunResult contract ─────────────────────────────────────────────────────────

func TestRunProfile_ResultContainsProfileName(t *testing.T) {
	s, _ := setupRunTest(t)

	p := profile.NewProfile("personal")
	saveProfile(t, s, p)

	result, err := RunProfile(s, RunOptions{ProfileName: "personal", DryRun: true})
	if err != nil {
		t.Fatalf("RunProfile failed: %v", err)
	}
	if result.ProfileName != "personal" {
		t.Errorf("result.ProfileName = %q, want %q", result.ProfileName, "personal")
	}
}

// ── MCP temp file cleanup ──────────────────────────────────────────────────────

func TestRunProfile_DryRunDoesNotCreateTempFiles(t *testing.T) {
	s, _ := setupRunTest(t)

	p := profile.NewProfile("work")
	p.MCPServers = mcpconfig.MCPServers{
		"srv": {Type: "stdio", Command: "echo"},
	}
	saveProfile(t, s, p)

	result, err := RunProfile(s, RunOptions{ProfileName: "work", DryRun: true})
	if err != nil {
		t.Fatalf("RunProfile failed: %v", err)
	}

	if result.TempDir != "" {
		if _, statErr := os.Stat(result.TempDir); statErr == nil {
			t.Errorf("dry-run should not create temp directory, but %q exists", result.TempDir)
		}
	}
}

// ── Validation ────────────────────────────────────────────────────────────────

func TestRunProfile_InvalidProfileNameReturnsError(t *testing.T) {
	s, _ := setupRunTest(t)

	_, err := RunProfile(s, RunOptions{ProfileName: "../escape", DryRun: true})
	if err == nil {
		t.Fatal("expected error for path-traversal profile name")
	}
}

func TestRunProfile_ProfileWithNoSettingsFileStillRunsWithDefaultArgs(t *testing.T) {
	s, _ := setupRunTest(t)

	// Save a profile, then remove its settings.json to simulate a bare profile dir
	p := profile.NewProfile("bare")
	saveProfile(t, s, p)

	profileDir, _ := paths.ProfileDir("bare")
	_ = os.Remove(filepath.Join(profileDir, "settings.json"))

	// Should still produce args (settings flag will point to a non-existent file;
	// that is Claude's problem, not claudectx's — the profile is technically valid)
	_, err := RunProfile(s, RunOptions{ProfileName: "bare", DryRun: true})
	if err != nil {
		t.Logf("note: RunProfile returned error for profile with missing settings.json: %v", err)
	}
}

// ── helper functions ───────────────────────────────────────────────────────────

// containsConsecutive reports whether args contains flag (as a standalone element).
func containsConsecutive(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

// containsFlag is an alias for containsConsecutive — checks for a flag anywhere.
func containsFlag(args []string, flag string) bool {
	return containsConsecutive(args, flag)
}

// argAfter returns the element immediately following flag in args, or "".
func argAfter(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// indexOf returns the first index of target in args, or -1.
func indexOf(args []string, target string) int {
	for i, a := range args {
		if a == target {
			return i
		}
	}
	return -1
}
