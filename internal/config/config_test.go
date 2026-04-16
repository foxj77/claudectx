package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSettings(t *testing.T) {
	// Create a temporary settings file
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	settingsContent := `{
  "env": {
    "API_KEY": "test-key",
    "BASE_URL": "https://api.example.com"
  },
  "model": "opus",
  "permissions": {
    "allow": ["WebSearch", "WebFetch"]
  }
}`

	err := os.WriteFile(settingsPath, []byte(settingsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Load the settings
	settings, err := LoadSettings(settingsPath)
	if err != nil {
		t.Fatalf("LoadSettings() failed: %v", err)
	}

	// Verify the loaded data
	if settings.Model != "opus" {
		t.Errorf("Model = %q, want %q", settings.Model, "opus")
	}

	if settings.Env["API_KEY"] != "test-key" {
		t.Errorf("Env[API_KEY] = %q, want %q", settings.Env["API_KEY"], "test-key")
	}

	if len(settings.Permissions.Allow) != 2 {
		t.Errorf("len(Permissions.Allow) = %d, want 2", len(settings.Permissions.Allow))
	}
}

func TestLoadSettingsNonExistent(t *testing.T) {
	// Try to load a non-existent file
	settings, err := LoadSettings("/tmp/nonexistent-claudectx-test.json")
	if err == nil {
		t.Error("LoadSettings() should fail for non-existent file")
	}
	if settings != nil {
		t.Error("LoadSettings() should return nil for non-existent file")
	}
}

func TestLoadSettingsInvalidJSON(t *testing.T) {
	// Create a temporary invalid JSON file
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "invalid.json")

	err := os.WriteFile(settingsPath, []byte("{ invalid json }"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Load should fail
	settings, err := LoadSettings(settingsPath)
	if err == nil {
		t.Error("LoadSettings() should fail for invalid JSON")
	}
	if settings != nil {
		t.Error("LoadSettings() should return nil for invalid JSON")
	}
}

func TestSaveSettings(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	settings := &Settings{
		Model: "sonnet",
		Env: map[string]string{
			"TEST_VAR": "test-value",
		},
		Permissions: &Permissions{
			Allow: []string{"WebSearch"},
		},
	}

	// Save the settings
	err := SaveSettings(settingsPath, settings)
	if err != nil {
		t.Fatalf("SaveSettings() failed: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(settingsPath); err != nil {
		t.Fatalf("Settings file was not created: %v", err)
	}

	// Read back and verify
	loaded, err := LoadSettings(settingsPath)
	if err != nil {
		t.Fatalf("Failed to load saved settings: %v", err)
	}

	if loaded.Model != "sonnet" {
		t.Errorf("Model = %q, want %q", loaded.Model, "sonnet")
	}

	if loaded.Env["TEST_VAR"] != "test-value" {
		t.Errorf("Env[TEST_VAR] = %q, want %q", loaded.Env["TEST_VAR"], "test-value")
	}
}

func TestSaveSettingsPreservesFormatting(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	settings := &Settings{
		Model: "opus",
		Env: map[string]string{
			"KEY1": "value1",
			"KEY2": "value2",
		},
	}

	err := SaveSettings(settingsPath, settings)
	if err != nil {
		t.Fatalf("SaveSettings() failed: %v", err)
	}

	// Read the raw file content
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	// Check that it's properly formatted (indented)
	var parsed map[string]interface{}
	err = json.Unmarshal(content, &parsed)
	if err != nil {
		t.Fatalf("Saved JSON is not valid: %v", err)
	}

	// Should be nicely formatted with indentation
	// We can check by re-marshaling with indent and comparing
	formatted, _ := json.MarshalIndent(parsed, "", "  ")
	if string(content) != string(formatted)+"\n" {
		t.Error("SaveSettings() should preserve formatting with 2-space indentation")
	}
}

func TestLoadSettingsOrEmpty(t *testing.T) {
	// Test with non-existent file
	tmpDir := t.TempDir()
	nonExistentPath := filepath.Join(tmpDir, "nonexistent.json")

	settings := LoadSettingsOrEmpty(nonExistentPath)
	if settings == nil {
		t.Fatal("LoadSettingsOrEmpty() should return empty settings, not nil")
	}

	// Should be an empty/default settings object
	if settings.Model != "" {
		t.Errorf("Empty settings should have empty model, got %q", settings.Model)
	}

	// Test with existing valid file
	existingPath := filepath.Join(tmpDir, "existing.json")
	err := os.WriteFile(existingPath, []byte(`{"model": "haiku"}`), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	settings = LoadSettingsOrEmpty(existingPath)
	if settings.Model != "haiku" {
		t.Errorf("Model = %q, want %q", settings.Model, "haiku")
	}
}

func TestSettingsJSON(t *testing.T) {
	settings := &Settings{
		Model: "opus",
		Env: map[string]string{
			"KEY": "value",
		},
		Permissions: &Permissions{
			Allow: []string{"tool1", "tool2"},
			Deny:  []string{"tool3"},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("Failed to marshal settings: %v", err)
	}

	// Unmarshal back
	var loaded Settings
	err = json.Unmarshal(data, &loaded)
	if err != nil {
		t.Fatalf("Failed to unmarshal settings: %v", err)
	}

	if loaded.Model != settings.Model {
		t.Errorf("Model = %q, want %q", loaded.Model, settings.Model)
	}

	if len(loaded.Env) != len(settings.Env) {
		t.Errorf("len(Env) = %d, want %d", len(loaded.Env), len(settings.Env))
	}

	if len(loaded.Permissions.Allow) != 2 {
		t.Errorf("len(Allow) = %d, want 2", len(loaded.Permissions.Allow))
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "dest.txt")

	content := "test content"
	err := os.WriteFile(srcPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy the file
	err = CopyFile(srcPath, dstPath)
	if err != nil {
		t.Fatalf("CopyFile() failed: %v", err)
	}

	// Verify destination exists
	copiedContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination: %v", err)
	}

	if string(copiedContent) != content {
		t.Errorf("Copied content = %q, want %q", string(copiedContent), content)
	}
}

func TestCopyFileNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "nonexistent.txt")
	dstPath := filepath.Join(tmpDir, "dest.txt")

	err := CopyFile(srcPath, dstPath)
	if err == nil {
		t.Error("CopyFile() should fail for non-existent source")
	}
}

func TestFileExists(t *testing.T) {
	tmpDir := t.TempDir()
	existingPath := filepath.Join(tmpDir, "exists.txt")
	nonExistentPath := filepath.Join(tmpDir, "nonexistent.txt")

	// Create the file
	err := os.WriteFile(existingPath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if !FileExists(existingPath) {
		t.Error("FileExists() should return true for existing file")
	}

	if FileExists(nonExistentPath) {
		t.Error("FileExists() should return false for non-existent file")
	}
}

// --- Issue #16: unknown field preservation tests ---

// TestLoadSaveSettings_PreservesUnknownTopLevelFields verifies that fields in
// settings.json that claudectx does not model (e.g. effortLevel, autoDreamEnabled)
// survive a LoadSettings → SaveSettings roundtrip unchanged.
func TestLoadSaveSettings_PreservesUnknownTopLevelFields(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "settings.json")

	original := `{
  "model": "opus",
  "effortLevel": "medium",
  "autoDreamEnabled": true,
  "skipDangerousModePermissionPrompt": true
}`
	if err := os.WriteFile(path, []byte(original), 0644); err != nil {
		t.Fatalf("failed to write settings: %v", err)
	}

	settings, err := LoadSettings(path)
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}
	if err := SaveSettings(path, settings); err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	raw := make(map[string]json.RawMessage)
	data, _ := os.ReadFile(path)
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("re-reading saved file failed: %v", err)
	}

	for _, key := range []string{"effortLevel", "autoDreamEnabled", "skipDangerousModePermissionPrompt"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("key %q was stripped from settings.json during LoadSettings→SaveSettings roundtrip", key)
		}
	}

	var effortLevel string
	if err := json.Unmarshal(raw["effortLevel"], &effortLevel); err != nil || effortLevel != "medium" {
		t.Errorf("effortLevel = %q, want %q", effortLevel, "medium")
	}
}

// TestLoadSaveSettings_PreservesUnknownPermissionsFields verifies that fields
// inside the permissions object that claudectx does not model (e.g. defaultMode)
// survive a roundtrip.
func TestLoadSaveSettings_PreservesUnknownPermissionsFields(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "settings.json")

	original := `{
  "model": "sonnet",
  "permissions": {
    "allow": ["WebSearch"],
    "deny": [],
    "defaultMode": "bypassPermissions"
  }
}`
	if err := os.WriteFile(path, []byte(original), 0644); err != nil {
		t.Fatalf("failed to write settings: %v", err)
	}

	settings, err := LoadSettings(path)
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}
	if err := SaveSettings(path, settings); err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	raw := make(map[string]json.RawMessage)
	data, _ := os.ReadFile(path)
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("re-reading saved file failed: %v", err)
	}

	permsRaw, ok := raw["permissions"]
	if !ok {
		t.Fatal("permissions key missing after roundtrip")
	}
	perms := make(map[string]json.RawMessage)
	if err := json.Unmarshal(permsRaw, &perms); err != nil {
		t.Fatalf("parsing permissions failed: %v", err)
	}
	if _, ok := perms["defaultMode"]; !ok {
		t.Error("permissions.defaultMode was stripped during LoadSettings→SaveSettings roundtrip")
	}
	var defaultMode string
	if err := json.Unmarshal(perms["defaultMode"], &defaultMode); err != nil || defaultMode != "bypassPermissions" {
		t.Errorf("permissions.defaultMode = %q, want %q", defaultMode, "bypassPermissions")
	}
}

// TestLoadSettings_KnownFieldsStillAccessible is a regression guard: after adding
// unknown-field preservation, known fields must still be correctly populated.
func TestLoadSettings_KnownFieldsStillAccessible(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "settings.json")

	content := `{
  "model": "haiku",
  "env": {"MY_VAR": "hello"},
  "permissions": {"allow": ["Bash"], "deny": ["WebSearch"]},
  "effortLevel": "low",
  "unknownArray": [1, 2, 3]
}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write settings: %v", err)
	}

	settings, err := LoadSettings(path)
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}

	if settings.Model != "haiku" {
		t.Errorf("Model = %q, want %q", settings.Model, "haiku")
	}
	if settings.Env["MY_VAR"] != "hello" {
		t.Errorf("Env[MY_VAR] = %q, want %q", settings.Env["MY_VAR"], "hello")
	}
	if settings.Permissions == nil {
		t.Fatal("Permissions is nil")
	}
	if len(settings.Permissions.Allow) != 1 || settings.Permissions.Allow[0] != "Bash" {
		t.Errorf("Permissions.Allow = %v, want [Bash]", settings.Permissions.Allow)
	}
	if len(settings.Permissions.Deny) != 1 || settings.Permissions.Deny[0] != "WebSearch" {
		t.Errorf("Permissions.Deny = %v, want [WebSearch]", settings.Permissions.Deny)
	}
}

// TestSaveSettings_ModifyKnownField_PreservesUnknownFields verifies that
// changing a known field (Model) in the in-memory struct and saving does not
// erase unknown fields that were loaded from disk.
func TestSaveSettings_ModifyKnownField_PreservesUnknownFields(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "settings.json")

	original := `{"model":"opus","effortLevel":"high","autoDreamEnabled":false}`
	if err := os.WriteFile(path, []byte(original), 0644); err != nil {
		t.Fatalf("failed to write settings: %v", err)
	}

	settings, err := LoadSettings(path)
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}

	// Simulate claudectx updating a known field
	settings.Model = "sonnet"

	if err := SaveSettings(path, settings); err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	raw := make(map[string]json.RawMessage)
	data, _ := os.ReadFile(path)
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("re-reading saved file failed: %v", err)
	}

	var model string
	json.Unmarshal(raw["model"], &model)
	if model != "sonnet" {
		t.Errorf("model = %q after update, want %q", model, "sonnet")
	}

	if _, ok := raw["effortLevel"]; !ok {
		t.Error("effortLevel was stripped when modifying an unrelated known field")
	}
	if _, ok := raw["autoDreamEnabled"]; !ok {
		t.Error("autoDreamEnabled was stripped when modifying an unrelated known field")
	}
}

// TestSaveSettings_NilPermissions_PreservesExistingPermissionsUnknownFields
// verifies that when Permissions is nil in the in-memory struct but the file on
// disk previously had permissions.defaultMode, that sub-field survives.
func TestSaveSettings_NilPermissions_PreservesExistingPermissionsUnknownFields(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "settings.json")

	original := `{"model":"opus","permissions":{"defaultMode":"bypassPermissions"}}`
	if err := os.WriteFile(path, []byte(original), 0644); err != nil {
		t.Fatalf("failed to write settings: %v", err)
	}

	settings, err := LoadSettings(path)
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}

	// Simulate a case where Permissions struct has no Allow/Deny but was loaded
	// from a file that had defaultMode — the loaded Permissions object should
	// still carry that unknown field through to the save.
	if err := SaveSettings(path, settings); err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	raw := make(map[string]json.RawMessage)
	data, _ := os.ReadFile(path)
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("re-reading saved file failed: %v", err)
	}

	permsRaw, ok := raw["permissions"]
	if !ok {
		t.Fatal("permissions key missing after roundtrip")
	}
	perms := make(map[string]json.RawMessage)
	if err := json.Unmarshal(permsRaw, &perms); err != nil {
		t.Fatalf("parsing permissions failed: %v", err)
	}
	if _, ok := perms["defaultMode"]; !ok {
		t.Error("permissions.defaultMode was stripped when Permissions has no Allow/Deny")
	}
}
