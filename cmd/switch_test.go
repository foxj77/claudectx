package cmd

import (
	"os"
	"testing"

	"github.com/johnfox/claudectx/internal/mcpconfig"
	"github.com/johnfox/claudectx/internal/paths"
	"github.com/johnfox/claudectx/internal/profile"
	"github.com/johnfox/claudectx/internal/store"
)

func TestSwitchProfile_InvalidName(t *testing.T) {
	tmp := t.TempDir()
	// isolate HOME
	_ = os.Setenv("HOME", tmp)

	s, err := store.NewStore()
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	if err := SwitchProfile(s, ""); err == nil {
		t.Fatal("expected error for empty profile name")
	}
}

func TestSwitchProfile_ProfileDoesNotExist(t *testing.T) {
	tmp := t.TempDir()
	_ = os.Setenv("HOME", tmp)

	s, err := store.NewStore()
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	if err := SwitchProfile(s, "nope"); err == nil {
		t.Fatal("expected error for non-existent profile")
	}
}

func TestSwitchProfile_Success(t *testing.T) {
	tmp := t.TempDir()
	_ = os.Setenv("HOME", tmp)

	s, err := store.NewStore()
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Create an "old" profile and set it as current
	old := profile.NewProfile("old")
	old.Settings.Model = "old-model"
	old.ClaudeMD = "old"
	if err := s.Save(old); err != nil {
		t.Fatalf("failed to save old profile: %v", err)
	}
	if err := s.SetCurrent("old"); err != nil {
		t.Fatalf("failed to set current: %v", err)
	}

	// Create target profile
	work := profile.NewProfile("work")
	work.Settings.Model = "work-model"
	work.ClaudeMD = "hello"
	work.MCPServers = mcpconfig.MCPServers{"one": {Type: "exec", Command: "echo"}}
	if err := s.Save(work); err != nil {
		t.Fatalf("failed to save work profile: %v", err)
	}

	// Ensure there is an existing active settings file to be backed up
	settingsPath, err := paths.SettingsFile()
	if err != nil {
		t.Fatalf("failed to determine settings file path: %v", err)
	}
	if err := os.WriteFile(settingsPath, []byte(`{"model":"existing"}`), 0644); err != nil {
		t.Fatalf("failed to write initial settings file: %v", err)
	}

	// Run switch
	if err := SwitchProfile(s, "work"); err != nil {
		t.Fatalf("SwitchProfile failed: %v", err)
	}

	// Verify current and previous
	cur, err := s.GetCurrent()
	if err != nil {
		t.Fatalf("GetCurrent failed: %v", err)
	}
	if cur != "work" {
		t.Fatalf("expected current=work, got %q", cur)
	}

	prev, err := s.GetPrevious()
	if err != nil {
		t.Fatalf("GetPrevious failed: %v", err)
	}
	if prev != "old" {
		t.Fatalf("expected previous=old, got %q", prev)
	}

	// Verify active CLAUDE.md
	claudeMDPath, _ := paths.ClaudeMDFile()
	b, _ := os.ReadFile(claudeMDPath)
	if string(b) != "hello" {
		t.Fatalf("CLAUDE.md content mismatch: %q", string(b))
	}

	// Verify claude.json contains the MCP server
	claudeJSONPath, _ := paths.ClaudeJSONFile()
	servers, err := mcpconfig.LoadMCPServers(claudeJSONPath)
	if err != nil {
		t.Fatalf("failed to load claude.json: %v", err)
	}
	if _, ok := servers["one"]; !ok {
		t.Fatalf("expected mcp server 'one' present")
	}
}

func TestSwitchProfile_SaveSettingsFails_Rollback(t *testing.T) {
	tmp := t.TempDir()
	_ = os.Setenv("HOME", tmp)

	s, err := store.NewStore()
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Create current profile "old"
	old := profile.NewProfile("old")
	old.Settings.Model = "old-model"
	if err := s.Save(old); err != nil {
		t.Fatalf("failed to save old profile: %v", err)
	}
	if err := s.SetCurrent("old"); err != nil {
		t.Fatalf("failed to set current: %v", err)
	}

	// Create target profile
	work := profile.NewProfile("work")
	work.Settings.Model = "work-model"
	if err := s.Save(work); err != nil {
		t.Fatalf("failed to save work profile: %v", err)
	}

	// Create active settings file and make it read-only to force write error
	settingsPath, _ := paths.SettingsFile()
	_ = os.WriteFile(settingsPath, []byte(`{"model":"existing"}`), 0444)
	t.Cleanup(func() {
		_ = os.Chmod(settingsPath, 0644)
	})

	if err := SwitchProfile(s, "work"); err == nil {
		t.Fatal("expected error when settings file cannot be written")
	}

	// Ensure current profile was not changed
	cur, _ := s.GetCurrent()
	if cur != "old" {
		t.Fatalf("expected current still old after failed switch, got %q", cur)
	}
}
