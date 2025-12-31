package exporter

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/johnfox/claudectx/internal/config"
	"github.com/johnfox/claudectx/internal/profile"
	"github.com/johnfox/claudectx/internal/store"
)

func setupTestEnv(t *testing.T) (*store.Store, string) {
	tmpHome := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})
	os.Setenv("HOME", tmpHome)

	claudeDir := filepath.Join(tmpHome, ".claude")
	os.MkdirAll(claudeDir, 0755)

	s, err := store.NewStore()
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	return s, tmpHome
}

func TestExportProfile(t *testing.T) {
	s, _ := setupTestEnv(t)

	// Create a test profile
	prof := profile.NewProfile("test-export")
	prof.Settings.Model = "opus"
	prof.Settings.Env = map[string]string{"KEY": "value"}
	prof.ClaudeMD = "# Test instructions"

	err := s.Save(prof)
	if err != nil {
		t.Fatalf("Failed to save profile: %v", err)
	}

	// Export it
	var buf bytes.Buffer
	err = ExportProfile(s, "test-export", &buf)
	if err != nil {
		t.Fatalf("ExportProfile() failed: %v", err)
	}

	// Verify output is valid JSON
	var exported ExportedProfile
	err = json.Unmarshal(buf.Bytes(), &exported)
	if err != nil {
		t.Fatalf("Exported data is not valid JSON: %v", err)
	}

	// Verify content
	if exported.Name != "test-export" {
		t.Errorf("Name = %q, want %q", exported.Name, "test-export")
	}

	if exported.Settings.Model != "opus" {
		t.Errorf("Model = %q, want %q", exported.Settings.Model, "opus")
	}

	if exported.ClaudeMD != "# Test instructions" {
		t.Errorf("ClaudeMD = %q, want %q", exported.ClaudeMD, "# Test instructions")
	}

	if exported.Version != ExportVersion {
		t.Errorf("Version = %q, want %q", exported.Version, ExportVersion)
	}
}

func TestExportProfileNonExistent(t *testing.T) {
	s, _ := setupTestEnv(t)

	var buf bytes.Buffer
	err := ExportProfile(s, "nonexistent", &buf)
	if err == nil {
		t.Error("ExportProfile() should fail for non-existent profile")
	}
}

func TestImportProfile(t *testing.T) {
	s, _ := setupTestEnv(t)

	// Create export data
	exported := ExportedProfile{
		Version: ExportVersion,
		Name:    "imported",
		Settings: &config.Settings{
			Model: "sonnet",
			Env:   map[string]string{"IMPORTED": "yes"},
		},
		ClaudeMD: "# Imported instructions",
	}

	data, _ := json.Marshal(exported)
	buf := bytes.NewBuffer(data)

	// Import it
	err := ImportProfile(s, buf, "")
	if err != nil {
		t.Fatalf("ImportProfile() failed: %v", err)
	}

	// Verify profile was created
	if !s.Exists("imported") {
		t.Error("Profile was not created")
	}

	// Load and verify
	prof, err := s.Load("imported")
	if err != nil {
		t.Fatalf("Failed to load imported profile: %v", err)
	}

	if prof.Settings.Model != "sonnet" {
		t.Errorf("Model = %q, want %q", prof.Settings.Model, "sonnet")
	}

	if prof.ClaudeMD != "# Imported instructions" {
		t.Errorf("ClaudeMD not imported correctly")
	}
}

func TestImportProfileWithRename(t *testing.T) {
	s, _ := setupTestEnv(t)

	exported := ExportedProfile{
		Version: ExportVersion,
		Name:    "original-name",
		Settings: &config.Settings{
			Model: "haiku",
		},
		ClaudeMD: "",
	}

	data, _ := json.Marshal(exported)
	buf := bytes.NewBuffer(data)

	// Import with different name
	err := ImportProfile(s, buf, "new-name")
	if err != nil {
		t.Fatalf("ImportProfile() failed: %v", err)
	}

	// Should exist with new name
	if !s.Exists("new-name") {
		t.Error("Profile not created with new name")
	}

	// Should not exist with original name
	if s.Exists("original-name") {
		t.Error("Profile should not exist with original name")
	}
}

func TestImportProfileAlreadyExists(t *testing.T) {
	s, _ := setupTestEnv(t)

	// Create existing profile
	existing := profile.NewProfile("existing")
	s.Save(existing)

	// Try to import with same name
	exported := ExportedProfile{
		Version:  ExportVersion,
		Name:     "existing",
		Settings: &config.Settings{},
		ClaudeMD: "",
	}

	data, _ := json.Marshal(exported)
	buf := bytes.NewBuffer(data)

	err := ImportProfile(s, buf, "")
	if err == nil {
		t.Error("ImportProfile() should fail when profile already exists")
	}
}

func TestImportProfileInvalidJSON(t *testing.T) {
	s, _ := setupTestEnv(t)

	buf := bytes.NewBufferString("{ invalid json }")

	err := ImportProfile(s, buf, "")
	if err == nil {
		t.Error("ImportProfile() should fail for invalid JSON")
	}
}

func TestImportProfileWrongVersion(t *testing.T) {
	s, _ := setupTestEnv(t)

	exported := ExportedProfile{
		Version:  "999.0.0",
		Name:     "test",
		Settings: &config.Settings{},
		ClaudeMD: "",
	}

	data, _ := json.Marshal(exported)
	buf := bytes.NewBuffer(data)

	err := ImportProfile(s, buf, "")
	if err == nil {
		t.Error("ImportProfile() should fail for incompatible version")
	}
}

func TestExportedProfileJSON(t *testing.T) {
	exported := ExportedProfile{
		Version: "1.0.0",
		Name:    "test",
		Settings: &config.Settings{
			Model: "opus",
		},
		ClaudeMD:  "# Instructions",
		ExportedAt: "2024-01-01T00:00:00Z",
	}

	// Marshal to JSON
	data, err := json.Marshal(exported)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back
	var loaded ExportedProfile
	err = json.Unmarshal(data, &loaded)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if loaded.Name != "test" {
		t.Errorf("Name = %q, want %q", loaded.Name, "test")
	}

	if loaded.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", loaded.Version, "1.0.0")
	}
}

func TestExportToFile(t *testing.T) {
	s, tmpHome := setupTestEnv(t)

	// Create a profile
	prof := profile.NewProfile("file-export")
	prof.Settings.Model = "opus"
	s.Save(prof)

	// Export to file
	exportPath := filepath.Join(tmpHome, "export.json")
	file, err := os.Create(exportPath)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	err = ExportProfile(s, "file-export", file)
	if err != nil {
		t.Fatalf("ExportProfile() failed: %v", err)
	}

	// Verify file was created and contains valid JSON
	data, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("Failed to read export file: %v", err)
	}

	var exported ExportedProfile
	err = json.Unmarshal(data, &exported)
	if err != nil {
		t.Fatalf("Export file contains invalid JSON: %v", err)
	}

	if exported.Name != "file-export" {
		t.Errorf("Exported name = %q, want %q", exported.Name, "file-export")
	}
}

func TestImportFromFile(t *testing.T) {
	s, tmpHome := setupTestEnv(t)

	// Create import file
	exported := ExportedProfile{
		Version:  ExportVersion,
		Name:     "file-import",
		Settings: &config.Settings{Model: "haiku"},
		ClaudeMD: "",
	}

	importPath := filepath.Join(tmpHome, "import.json")
	data, _ := json.Marshal(exported)
	os.WriteFile(importPath, data, 0644)

	// Import from file
	file, err := os.Open(importPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	err = ImportProfile(s, file, "")
	if err != nil {
		t.Fatalf("ImportProfile() failed: %v", err)
	}

	// Verify
	if !s.Exists("file-import") {
		t.Error("Profile was not imported")
	}
}

func TestPrettyPrintExport(t *testing.T) {
	s, _ := setupTestEnv(t)

	prof := profile.NewProfile("pretty")
	prof.Settings.Model = "opus"
	s.Save(prof)

	var buf bytes.Buffer
	err := ExportProfile(s, "pretty", &buf)
	if err != nil {
		t.Fatalf("ExportProfile() failed: %v", err)
	}

	// Check that output is pretty-printed (has newlines and indentation)
	output := buf.String()
	if !strings.Contains(output, "\n") {
		t.Error("Export should be pretty-printed with newlines")
	}

	if !strings.Contains(output, "  ") {
		t.Error("Export should be indented")
	}
}

func TestImportWithInvalidSettings(t *testing.T) {
	s, _ := setupTestEnv(t)

	// Export with nil settings (invalid)
	exported := ExportedProfile{
		Version:  ExportVersion,
		Name:     "invalid",
		Settings: nil, // This should cause validation to fail
		ClaudeMD: "",
	}

	data, _ := json.Marshal(exported)
	buf := bytes.NewBuffer(data)

	err := ImportProfile(s, buf, "")
	if err == nil {
		t.Error("ImportProfile() should fail for nil settings")
	}
}

func TestExportIncludesMetadata(t *testing.T) {
	s, _ := setupTestEnv(t)

	prof := profile.NewProfile("meta-test")
	s.Save(prof)

	var buf bytes.Buffer
	err := ExportProfile(s, "meta-test", &buf)
	if err != nil {
		t.Fatalf("ExportProfile() failed: %v", err)
	}

	var exported ExportedProfile
	json.Unmarshal(buf.Bytes(), &exported)

	// Should include version
	if exported.Version == "" {
		t.Error("Export should include version")
	}

	// Should include export timestamp
	if exported.ExportedAt == "" {
		t.Error("Export should include ExportedAt timestamp")
	}
}
