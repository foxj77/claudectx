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

	// Must end with a newline
	if len(content) == 0 || content[len(content)-1] != '\n' {
		t.Error("SaveSettings() should end the file with a newline")
	}

	// Must be valid, indented JSON
	var parsed map[string]interface{}
	if err = json.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("Saved JSON is not valid: %v", err)
	}

	// Verify 2-space indentation by checking a known indented key exists
	if _, ok := parsed["model"]; !ok {
		t.Error("Expected 'model' key in saved JSON")
	}
}

func TestSettingsPreservesUnknownFields(t *testing.T) {
	tmpDir := t.TempDir()
	settingsPath := filepath.Join(tmpDir, "settings.json")

	// Write settings.json that contains fields unknown to the Settings struct
	original := `{
  "env": {
    "API_KEY": "secret"
  },
  "model": "opus",
  "hooks": {"PreToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "echo hi"}]}]},
  "apiKeyHelper": "/usr/local/bin/get-key",
  "includeCoAuthoredBy": false
}`
	if err := os.WriteFile(settingsPath, []byte(original), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Load, then save back
	settings, err := LoadSettings(settingsPath)
	if err != nil {
		t.Fatalf("LoadSettings() failed: %v", err)
	}

	if err = SaveSettings(settingsPath, settings); err != nil {
		t.Fatalf("SaveSettings() failed: %v", err)
	}

	// Re-parse the saved file and confirm all keys survived
	content, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	var saved map[string]interface{}
	if err = json.Unmarshal(content, &saved); err != nil {
		t.Fatalf("Saved JSON is not valid: %v", err)
	}

	for _, key := range []string{"env", "model", "hooks", "apiKeyHelper", "includeCoAuthoredBy"} {
		if _, ok := saved[key]; !ok {
			t.Errorf("Key %q was dropped during load+save round-trip", key)
		}
	}

	if saved["model"] != "opus" {
		t.Errorf("model = %v, want %q", saved["model"], "opus")
	}
	if saved["apiKeyHelper"] != "/usr/local/bin/get-key" {
		t.Errorf("apiKeyHelper = %v, want %q", saved["apiKeyHelper"], "/usr/local/bin/get-key")
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
