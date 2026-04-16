package cmd

import (
	"encoding/json"
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

// TestSwitchProfile_PreservesProfileStoredUnknownFields verifies that when
// switching to a profile whose stored settings.json contains unknown fields
// (e.g. effortLevel), those fields are written to the active settings.json.
// This is the core scenario from issue #16: if a profile was captured while
// effortLevel was set, switching to it must restore effortLevel.
func TestSwitchProfile_PreservesProfileStoredUnknownFields(t *testing.T) {
	tmp := t.TempDir()
	_ = os.Setenv("HOME", tmp)

	s, err := store.NewStore()
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Create a "work" profile whose settings.json has an unknown field (effortLevel).
	// We write the raw file directly into the profile directory to bypass the
	// struct-based Save (which would strip it before the fix).
	profileDir, err := paths.ProfileDir("work")
	if err != nil {
		t.Fatalf("failed to get profile dir: %v", err)
	}
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		t.Fatalf("failed to create profile dir: %v", err)
	}
	settingsWithUnknown := `{"model":"work-model","effortLevel":"high","autoDreamEnabled":true}`
	if err := os.WriteFile(profileDir+"/settings.json", []byte(settingsWithUnknown), 0644); err != nil {
		t.Fatalf("failed to write profile settings: %v", err)
	}

	// Write a minimal active settings.json so the backup manager has something to back up.
	activeSettingsPath, err := paths.SettingsFile()
	if err != nil {
		t.Fatalf("failed to get settings path: %v", err)
	}
	if err := os.WriteFile(activeSettingsPath, []byte(`{"model":"old"}`), 0644); err != nil {
		t.Fatalf("failed to write active settings: %v", err)
	}

	if err := SwitchProfile(s, "work"); err != nil {
		t.Fatalf("SwitchProfile failed: %v", err)
	}

	// Read the active settings.json and check the unknown fields survived.
	raw := make(map[string]json.RawMessage)
	data, err := os.ReadFile(activeSettingsPath)
	if err != nil {
		t.Fatalf("failed to read active settings: %v", err)
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to parse active settings: %v", err)
	}

	if _, ok := raw["effortLevel"]; !ok {
		t.Error("effortLevel was stripped from active settings.json when switching to profile — issue #16")
	}
	if _, ok := raw["autoDreamEnabled"]; !ok {
		t.Error("autoDreamEnabled was stripped from active settings.json when switching to profile — issue #16")
	}
	var model string
	json.Unmarshal(raw["model"], &model)
	if model != "work-model" {
		t.Errorf("model = %q, want %q", model, "work-model")
	}
}

// TestSwitchProfile_AutoSync_PreservesUnknownFieldsInProfile verifies that when
// claudectx auto-syncs the current profile before switching away, unknown fields
// in the active settings.json (including plugin references) are preserved in the
// stored profile. Without this, switching away destroys the plugin state.
func TestSwitchProfile_AutoSync_PreservesUnknownFieldsInProfile(t *testing.T) {
	tmp := t.TempDir()
	_ = os.Setenv("HOME", tmp)

	s, err := store.NewStore()
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Create and activate profile "current" (simulates user's current profile).
	cur := profile.NewProfile("current")
	cur.Settings.Model = "current-model"
	if err := s.Save(cur); err != nil {
		t.Fatalf("failed to save current profile: %v", err)
	}
	if err := s.SetCurrent("current"); err != nil {
		t.Fatalf("failed to set current: %v", err)
	}

	// Write active settings.json with unknown fields (simulates plugin install).
	activeSettingsPath, err := paths.SettingsFile()
	if err != nil {
		t.Fatalf("failed to get settings path: %v", err)
	}
	activeSettings := `{"model":"current-model","effortLevel":"medium","plugins":["superpowers"]}`
	if err := os.WriteFile(activeSettingsPath, []byte(activeSettings), 0644); err != nil {
		t.Fatalf("failed to write active settings: %v", err)
	}

	// Create profile "other" to switch to.
	other := profile.NewProfile("other")
	other.Settings.Model = "other-model"
	if err := s.Save(other); err != nil {
		t.Fatalf("failed to save other profile: %v", err)
	}

	// Switch away — this triggers auto-sync of "current" profile.
	if err := SwitchProfile(s, "other"); err != nil {
		t.Fatalf("SwitchProfile failed: %v", err)
	}

	// Load the stored "current" profile and check unknown fields were captured.
	profileDir, err := paths.ProfileDir("current")
	if err != nil {
		t.Fatalf("failed to get profile dir: %v", err)
	}
	raw := make(map[string]json.RawMessage)
	data, err := os.ReadFile(profileDir + "/settings.json")
	if err != nil {
		t.Fatalf("failed to read stored profile settings: %v", err)
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to parse stored profile settings: %v", err)
	}

	if _, ok := raw["effortLevel"]; !ok {
		t.Error("effortLevel was stripped from stored profile during auto-sync — issue #16")
	}
	if _, ok := raw["plugins"]; !ok {
		t.Error("plugins key was stripped from stored profile during auto-sync — issue #16")
	}
}
