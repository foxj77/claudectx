package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/johnfox/claudectx/internal/paths"
	"github.com/johnfox/claudectx/internal/profile"
	"github.com/johnfox/claudectx/internal/store"
)

// TestSyncCurrentProfile_PreservesUnknownSettingsFields verifies that an explicit
// sync captures unknown fields from the active settings.json into the stored
// profile. This is the write-side of the issue #16 data-loss bug.
func TestSyncCurrentProfile_PreservesUnknownSettingsFields(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	s, err := store.NewStore()
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Create a profile and mark it as current.
	prof := profile.NewProfile("myprofile")
	prof.Settings.Model = "opus"
	if err := s.Save(prof); err != nil {
		t.Fatalf("failed to save profile: %v", err)
	}
	if err := s.SetCurrent("myprofile"); err != nil {
		t.Fatalf("failed to set current: %v", err)
	}

	// Write active settings.json with unknown fields (simulating Claude Code writing them).
	activeSettingsPath, err := paths.SettingsFile()
	if err != nil {
		t.Fatalf("failed to get settings path: %v", err)
	}
	activeSettings := `{
  "model": "opus",
  "effortLevel": "medium",
  "skipDangerousModePermissionPrompt": true
}`
	if err := os.WriteFile(activeSettingsPath, []byte(activeSettings), 0644); err != nil {
		t.Fatalf("failed to write active settings: %v", err)
	}

	// Sync the active config into the profile.
	if err := SyncCurrentProfile(s); err != nil {
		t.Fatalf("SyncCurrentProfile failed: %v", err)
	}

	// Read the stored profile's settings.json and assert unknown fields survived.
	profileDir, err := paths.ProfileDir("myprofile")
	if err != nil {
		t.Fatalf("failed to get profile dir: %v", err)
	}
	raw := make(map[string]json.RawMessage)
	data, err := os.ReadFile(filepath.Join(profileDir, "settings.json"))
	if err != nil {
		t.Fatalf("failed to read stored profile settings: %v", err)
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to parse stored profile settings: %v", err)
	}

	if _, ok := raw["effortLevel"]; !ok {
		t.Error("effortLevel was stripped during SyncCurrentProfile — issue #16")
	}
	if _, ok := raw["skipDangerousModePermissionPrompt"]; !ok {
		t.Error("skipDangerousModePermissionPrompt was stripped during SyncCurrentProfile — issue #16")
	}
	modelRaw, ok := raw["model"]
	if !ok {
		t.Fatal("model key missing from stored profile settings")
	}
	var model string
	if err := json.Unmarshal(modelRaw, &model); err != nil {
		t.Fatalf("failed to parse stored profile model: %v", err)
	}
	if model != "opus" {
		t.Errorf("model = %q, want %q", model, "opus")
	}
}

// TestHashSettings_StableAcrossUnknownFields verifies that hashSettings produces
// the same hash before and after a LoadSettings→SaveSettings roundtrip when the
// settings file contains unknown fields. Without this, hasConfigChanged would
// incorrectly report a change every time after a switch, triggering an unnecessary
// auto-sync on the next switch.
func TestHashSettings_StableAcrossUnknownFields(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	s, err := store.NewStore()
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	// Create a profile with settings that include unknown fields.
	prof := profile.NewProfile("stable")
	prof.Settings.Model = "haiku"
	if err := s.Save(prof); err != nil {
		t.Fatalf("failed to save profile: %v", err)
	}
	if err := s.SetCurrent("stable"); err != nil {
		t.Fatalf("failed to set current: %v", err)
	}

	// Write active settings.json with unknown fields.
	activeSettingsPath, err := paths.SettingsFile()
	if err != nil {
		t.Fatalf("failed to get settings path: %v", err)
	}
	activeSettings := `{"model":"haiku","effortLevel":"low","autoDreamEnabled":false}`
	if err := os.WriteFile(activeSettingsPath, []byte(activeSettings), 0644); err != nil {
		t.Fatalf("failed to write active settings: %v", err)
	}

	// Sync once so the profile matches the active config.
	if err := SyncCurrentProfile(s); err != nil {
		t.Fatalf("first SyncCurrentProfile failed: %v", err)
	}

	// Check that hasConfigChanged reports no change (the profile now matches active).
	changed, err := hasConfigChanged(s, "stable")
	if err != nil {
		t.Fatalf("hasConfigChanged failed: %v", err)
	}
	if changed {
		t.Error("hasConfigChanged returned true immediately after sync — unknown fields are causing hash instability (issue #16)")
	}
}
