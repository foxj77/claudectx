package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/johnfox/claudectx/internal/paths"
	"github.com/johnfox/claudectx/internal/profile"
)

func setupTestEnv(t *testing.T) string {
	// Create a temporary directory for testing
	tmpHome := t.TempDir()

	// Set HOME to temp directory
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})
	os.Setenv("HOME", tmpHome)

	// Create .claude directory
	claudeDir := filepath.Join(tmpHome, ".claude")
	err := os.MkdirAll(claudeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test claude dir: %v", err)
	}

	return tmpHome
}

func TestNewStore(t *testing.T) {
	setupTestEnv(t)

	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() failed: %v", err)
	}

	if store == nil {
		t.Fatal("NewStore() returned nil")
	}

	// Verify profiles directory was created
	profilesDir, _ := paths.ProfilesDir()
	info, err := os.Stat(profilesDir)
	if err != nil {
		t.Fatalf("Profiles directory not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Profiles path is not a directory")
	}
}

func TestSaveAndLoadProfile(t *testing.T) {
	setupTestEnv(t)
	store, _ := NewStore()

	// Create a test profile
	prof := profile.NewProfile("test-save")
	prof.Settings.Model = "opus"
	prof.Settings.Env = map[string]string{"KEY": "value"}
	prof.ClaudeMD = "# Test instructions"

	// Save it
	err := store.Save(prof)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Load it back
	loaded, err := store.Load("test-save")
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if loaded.Name != "test-save" {
		t.Errorf("Name = %q, want %q", loaded.Name, "test-save")
	}

	if loaded.Settings.Model != "opus" {
		t.Errorf("Model = %q, want %q", loaded.Settings.Model, "opus")
	}

	if loaded.Settings.Env["KEY"] != "value" {
		t.Errorf("Env[KEY] = %q, want %q", loaded.Settings.Env["KEY"], "value")
	}

	if loaded.ClaudeMD != "# Test instructions" {
		t.Errorf("ClaudeMD = %q, want %q", loaded.ClaudeMD, "# Test instructions")
	}
}

func TestLoadNonExistent(t *testing.T) {
	setupTestEnv(t)
	store, _ := NewStore()

	_, err := store.Load("nonexistent")
	if err == nil {
		t.Error("Load() should fail for non-existent profile")
	}
}

func TestList(t *testing.T) {
	setupTestEnv(t)
	store, _ := NewStore()

	// Initially should be empty
	profiles, err := store.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(profiles) != 0 {
		t.Errorf("List() = %d profiles, want 0", len(profiles))
	}

	// Create some profiles
	store.Save(profile.NewProfile("work"))
	store.Save(profile.NewProfile("personal"))
	store.Save(profile.NewProfile("test"))

	// List should return 3
	profiles, err = store.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(profiles) != 3 {
		t.Errorf("List() = %d profiles, want 3", len(profiles))
	}

	// Check that all names are present
	names := make(map[string]bool)
	for _, p := range profiles {
		names[p] = true
	}

	expectedNames := []string{"work", "personal", "test"}
	for _, name := range expectedNames {
		if !names[name] {
			t.Errorf("List() missing profile %q", name)
		}
	}
}

func TestExists(t *testing.T) {
	setupTestEnv(t)
	store, _ := NewStore()

	// Should not exist initially
	if store.Exists("test") {
		t.Error("Exists() should return false for non-existent profile")
	}

	// Create it
	store.Save(profile.NewProfile("test"))

	// Now should exist
	if !store.Exists("test") {
		t.Error("Exists() should return true for existing profile")
	}
}

func TestDelete(t *testing.T) {
	setupTestEnv(t)
	store, _ := NewStore()

	// Create a profile
	store.Save(profile.NewProfile("to-delete"))

	if !store.Exists("to-delete") {
		t.Fatal("Profile was not created")
	}

	// Delete it
	err := store.Delete("to-delete")
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Should no longer exist
	if store.Exists("to-delete") {
		t.Error("Profile still exists after Delete()")
	}
}

func TestDeleteNonExistent(t *testing.T) {
	setupTestEnv(t)
	store, _ := NewStore()

	err := store.Delete("nonexistent")
	if err == nil {
		t.Error("Delete() should fail for non-existent profile")
	}
}

func TestCurrentProfile(t *testing.T) {
	setupTestEnv(t)
	store, _ := NewStore()

	// Initially should be empty
	current, err := store.GetCurrent()
	if err != nil {
		t.Fatalf("GetCurrent() failed: %v", err)
	}

	if current != "" {
		t.Errorf("GetCurrent() = %q, want empty string", current)
	}

	// Set current
	err = store.SetCurrent("work")
	if err != nil {
		t.Fatalf("SetCurrent() failed: %v", err)
	}

	// Get it back
	current, err = store.GetCurrent()
	if err != nil {
		t.Fatalf("GetCurrent() failed: %v", err)
	}

	if current != "work" {
		t.Errorf("GetCurrent() = %q, want %q", current, "work")
	}

	// Clear it
	err = store.SetCurrent("")
	if err != nil {
		t.Fatalf("SetCurrent(\"\") failed: %v", err)
	}

	current, err = store.GetCurrent()
	if err != nil {
		t.Fatalf("GetCurrent() failed: %v", err)
	}

	if current != "" {
		t.Errorf("GetCurrent() = %q after clear, want empty string", current)
	}
}

func TestPreviousProfile(t *testing.T) {
	setupTestEnv(t)
	store, _ := NewStore()

	// Initially should be empty
	prev, err := store.GetPrevious()
	if err != nil {
		t.Fatalf("GetPrevious() failed: %v", err)
	}

	if prev != "" {
		t.Errorf("GetPrevious() = %q, want empty string", prev)
	}

	// Set previous
	err = store.SetPrevious("personal")
	if err != nil {
		t.Fatalf("SetPrevious() failed: %v", err)
	}

	// Get it back
	prev, err = store.GetPrevious()
	if err != nil {
		t.Fatalf("GetPrevious() failed: %v", err)
	}

	if prev != "personal" {
		t.Errorf("GetPrevious() = %q, want %q", prev, "personal")
	}
}

func TestTogglePattern(t *testing.T) {
	setupTestEnv(t)
	store, _ := NewStore()

	// Simulate switching from work to personal
	store.SetCurrent("work")
	store.SetPrevious("")

	// Switch to personal
	oldCurrent, _ := store.GetCurrent()
	store.SetPrevious(oldCurrent) // Save "work" as previous
	store.SetCurrent("personal")

	// Verify state
	current, _ := store.GetCurrent()
	prev, _ := store.GetPrevious()

	if current != "personal" {
		t.Errorf("Current = %q, want %q", current, "personal")
	}

	if prev != "work" {
		t.Errorf("Previous = %q, want %q", prev, "work")
	}

	// Toggle back
	oldCurrent, _ = store.GetCurrent()
	oldPrev, _ := store.GetPrevious()
	store.SetCurrent(oldPrev)   // Switch to "work"
	store.SetPrevious(oldCurrent) // Save "personal" as previous

	current, _ = store.GetCurrent()
	prev, _ = store.GetPrevious()

	if current != "work" {
		t.Errorf("After toggle, current = %q, want %q", current, "work")
	}

	if prev != "personal" {
		t.Errorf("After toggle, previous = %q, want %q", prev, "personal")
	}
}

func TestSaveWithClaudeMD(t *testing.T) {
	setupTestEnv(t)
	store, _ := NewStore()

	prof := profile.NewProfile("with-md")
	prof.ClaudeMD = "# Claude Instructions\n\nUse Python for everything."

	err := store.Save(prof)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Load and verify
	loaded, err := store.Load("with-md")
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if loaded.ClaudeMD != prof.ClaudeMD {
		t.Errorf("ClaudeMD not preserved, got %q", loaded.ClaudeMD)
	}

	// Verify file exists on disk
	profileDir, _ := paths.ProfileDir("with-md")
	claudeMDPath := filepath.Join(profileDir, "CLAUDE.md")
	if _, err := os.Stat(claudeMDPath); err != nil {
		t.Errorf("CLAUDE.md file not created: %v", err)
	}
}

func TestSaveWithoutClaudeMD(t *testing.T) {
	setupTestEnv(t)
	store, _ := NewStore()

	prof := profile.NewProfile("no-md")
	prof.ClaudeMD = "" // Explicitly empty

	err := store.Save(prof)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// CLAUDE.md file should not be created
	profileDir, _ := paths.ProfileDir("no-md")
	claudeMDPath := filepath.Join(profileDir, "CLAUDE.md")
	if _, err := os.Stat(claudeMDPath); err == nil {
		t.Error("CLAUDE.md file should not be created when ClaudeMD is empty")
	}

	// Load and verify
	loaded, err := store.Load("no-md")
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if loaded.ClaudeMD != "" {
		t.Errorf("ClaudeMD should be empty, got %q", loaded.ClaudeMD)
	}
}
