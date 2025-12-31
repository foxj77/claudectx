package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClaudeDir(t *testing.T) {
	dir, err := ClaudeDir()
	if err != nil {
		t.Fatalf("ClaudeDir() failed: %v", err)
	}

	if dir == "" {
		t.Fatal("ClaudeDir() returned empty string")
	}

	// Should end with .claude
	if filepath.Base(dir) != ".claude" {
		t.Errorf("ClaudeDir() = %q, want path ending with .claude", dir)
	}
}

func TestProfilesDir(t *testing.T) {
	dir, err := ProfilesDir()
	if err != nil {
		t.Fatalf("ProfilesDir() failed: %v", err)
	}

	if dir == "" {
		t.Fatal("ProfilesDir() returned empty string")
	}

	// Should end with profiles
	if filepath.Base(dir) != "profiles" {
		t.Errorf("ProfilesDir() = %q, want path ending with profiles", dir)
	}
}

func TestProfileDir(t *testing.T) {
	dir, err := ProfileDir("work")
	if err != nil {
		t.Fatalf("ProfileDir() failed: %v", err)
	}

	if filepath.Base(dir) != "work" {
		t.Errorf("ProfileDir(\"work\") = %q, want path ending with work", dir)
	}

	// Test with empty name
	_, err = ProfileDir("")
	if err == nil {
		t.Error("ProfileDir(\"\") should return error for empty name")
	}
}

func TestCurrentProfileFile(t *testing.T) {
	file, err := CurrentProfileFile()
	if err != nil {
		t.Fatalf("CurrentProfileFile() failed: %v", err)
	}

	if filepath.Base(file) != ".claudectx-current" {
		t.Errorf("CurrentProfileFile() = %q, want .claudectx-current", file)
	}
}

func TestPreviousProfileFile(t *testing.T) {
	file, err := PreviousProfileFile()
	if err != nil {
		t.Fatalf("PreviousProfileFile() failed: %v", err)
	}

	if filepath.Base(file) != ".claudectx-previous" {
		t.Errorf("PreviousProfileFile() = %q, want .claudectx-previous", file)
	}
}

func TestSettingsFile(t *testing.T) {
	file, err := SettingsFile()
	if err != nil {
		t.Fatalf("SettingsFile() failed: %v", err)
	}

	if filepath.Base(file) != "settings.json" {
		t.Errorf("SettingsFile() = %q, want settings.json", file)
	}
}

func TestClaudeMDFile(t *testing.T) {
	file, err := ClaudeMDFile()
	if err != nil {
		t.Fatalf("ClaudeMDFile() failed: %v", err)
	}

	if filepath.Base(file) != "CLAUDE.md" {
		t.Errorf("ClaudeMDFile() = %q, want CLAUDE.md", file)
	}
}

func TestProfileFiles(t *testing.T) {
	tests := []struct {
		name     string
		profile  string
		filename string
		expected string
	}{
		{"settings", "work", "settings.json", "settings.json"},
		{"claude md", "work", "CLAUDE.md", "CLAUDE.md"},
		{"auth", "work", "auth.json", "auth.json"},
		{"mcp", "work", "mcp-servers.json", "mcp-servers.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := ProfileFile(tt.profile, tt.filename)
			if err != nil {
				t.Fatalf("ProfileFile(%q, %q) failed: %v", tt.profile, tt.filename, err)
			}

			if filepath.Base(path) != tt.expected {
				t.Errorf("ProfileFile(%q, %q) = %q, want file ending with %q",
					tt.profile, tt.filename, path, tt.expected)
			}
		})
	}
}

func TestEnsureProfilesDir(t *testing.T) {
	// Create a temporary directory for testing
	tmpHome := t.TempDir()

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set temporary HOME
	os.Setenv("HOME", tmpHome)

	// Profiles dir shouldn't exist yet
	profilesDir := filepath.Join(tmpHome, ".claude", "profiles")
	if _, err := os.Stat(profilesDir); err == nil {
		t.Fatal("Profiles dir should not exist yet")
	}

	// Ensure it gets created
	err := EnsureProfilesDir()
	if err != nil {
		t.Fatalf("EnsureProfilesDir() failed: %v", err)
	}

	// Now it should exist
	info, err := os.Stat(profilesDir)
	if err != nil {
		t.Fatalf("Profiles dir was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Profiles path is not a directory")
	}

	// Calling again should not error
	err = EnsureProfilesDir()
	if err != nil {
		t.Errorf("EnsureProfilesDir() second call failed: %v", err)
	}
}

func TestEnsureProfileDir(t *testing.T) {
	// Create a temporary directory for testing
	tmpHome := t.TempDir()

	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set temporary HOME
	os.Setenv("HOME", tmpHome)

	// Ensure profiles dir exists first
	err := EnsureProfilesDir()
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Profile dir shouldn't exist yet
	profileDir := filepath.Join(tmpHome, ".claude", "profiles", "testprofile")
	if _, err := os.Stat(profileDir); err == nil {
		t.Fatal("Profile dir should not exist yet")
	}

	// Ensure it gets created
	err = EnsureProfileDir("testprofile")
	if err != nil {
		t.Fatalf("EnsureProfileDir() failed: %v", err)
	}

	// Now it should exist
	info, err := os.Stat(profileDir)
	if err != nil {
		t.Fatalf("Profile dir was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Profile path is not a directory")
	}

	// Calling again should not error
	err = EnsureProfileDir("testprofile")
	if err != nil {
		t.Errorf("EnsureProfileDir() second call failed: %v", err)
	}
}
