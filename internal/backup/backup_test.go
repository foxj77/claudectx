package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/johnfox/claudectx/internal/config"
	"github.com/johnfox/claudectx/internal/paths"
)

func setupTestEnv(t *testing.T) string {
	tmpHome := t.TempDir()
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

func TestNewManager(t *testing.T) {
	setupTestEnv(t)

	mgr, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	if mgr == nil {
		t.Fatal("NewManager() returned nil")
	}

	// Verify backup directory was created
	backupDir, _ := paths.ClaudeDir()
	backupPath := filepath.Join(backupDir, "backups")
	info, err := os.Stat(backupPath)
	if err != nil {
		t.Fatalf("Backup directory not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Backup path is not a directory")
	}
}

func TestCreateBackup(t *testing.T) {
	setupTestEnv(t)
	mgr, _ := NewManager()

	// Create a test settings file
	settingsPath, _ := paths.SettingsFile()
	settings := &config.Settings{Model: "opus"}
	config.SaveSettings(settingsPath, settings)

	// Create backup
	backupID, err := mgr.Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if backupID == "" {
		t.Error("Create() returned empty backup ID")
	}

	// Verify backup exists
	backups, err := mgr.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(backups) != 1 {
		t.Errorf("Expected 1 backup, got %d", len(backups))
	}

	if backups[0].ID != backupID {
		t.Errorf("Backup ID mismatch: got %s, want %s", backups[0].ID, backupID)
	}
}

func TestCreateBackupWithClaudeMD(t *testing.T) {
	setupTestEnv(t)
	mgr, _ := NewManager()

	// Create settings and CLAUDE.md
	settingsPath, _ := paths.SettingsFile()
	config.SaveSettings(settingsPath, &config.Settings{Model: "sonnet"})

	claudeMDPath, _ := paths.ClaudeMDFile()
	os.WriteFile(claudeMDPath, []byte("# Instructions"), 0644)

	// Create backup
	backupID, err := mgr.Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Verify both files were backed up
	backupDir := filepath.Join(mgr.backupDir, backupID)
	settingsBackup := filepath.Join(backupDir, "settings.json")
	claudeMDBackup := filepath.Join(backupDir, "CLAUDE.md")

	if !config.FileExists(settingsBackup) {
		t.Error("settings.json not backed up")
	}

	if !config.FileExists(claudeMDBackup) {
		t.Error("CLAUDE.md not backed up")
	}
}

func TestRestoreBackup(t *testing.T) {
	setupTestEnv(t)
	mgr, _ := NewManager()

	// Create original settings
	settingsPath, _ := paths.SettingsFile()
	originalSettings := &config.Settings{Model: "opus"}
	config.SaveSettings(settingsPath, originalSettings)

	// Create backup
	backupID, err := mgr.Create()
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Modify settings
	modifiedSettings := &config.Settings{Model: "haiku"}
	config.SaveSettings(settingsPath, modifiedSettings)

	// Restore backup
	err = mgr.Restore(backupID)
	if err != nil {
		t.Fatalf("Restore() failed: %v", err)
	}

	// Verify settings were restored
	restored, err := config.LoadSettings(settingsPath)
	if err != nil {
		t.Fatalf("Failed to load restored settings: %v", err)
	}

	if restored.Model != "opus" {
		t.Errorf("Model = %s, want opus", restored.Model)
	}
}

func TestRestoreWithClaudeMD(t *testing.T) {
	setupTestEnv(t)
	mgr, _ := NewManager()

	// Create original files
	settingsPath, _ := paths.SettingsFile()
	claudeMDPath, _ := paths.ClaudeMDFile()

	config.SaveSettings(settingsPath, &config.Settings{Model: "opus"})
	os.WriteFile(claudeMDPath, []byte("# Original"), 0644)

	// Create backup
	backupID, _ := mgr.Create()

	// Modify files
	config.SaveSettings(settingsPath, &config.Settings{Model: "haiku"})
	os.WriteFile(claudeMDPath, []byte("# Modified"), 0644)

	// Restore
	err := mgr.Restore(backupID)
	if err != nil {
		t.Fatalf("Restore() failed: %v", err)
	}

	// Verify CLAUDE.md was restored
	content, _ := os.ReadFile(claudeMDPath)
	if string(content) != "# Original" {
		t.Errorf("CLAUDE.md = %s, want '# Original'", string(content))
	}
}

func TestRestoreNonExistentBackup(t *testing.T) {
	setupTestEnv(t)
	mgr, _ := NewManager()

	err := mgr.Restore("nonexistent-backup")
	if err == nil {
		t.Error("Restore() should fail for non-existent backup")
	}
}

func TestListBackups(t *testing.T) {
	setupTestEnv(t)
	mgr, _ := NewManager()

	// Initially empty
	backups, err := mgr.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(backups) != 0 {
		t.Errorf("Expected 0 backups, got %d", len(backups))
	}

	// Create a settings file first
	settingsPath, _ := paths.SettingsFile()
	config.SaveSettings(settingsPath, &config.Settings{Model: "opus"})

	// Create some backups
	mgr.Create()
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	mgr.Create()
	time.Sleep(10 * time.Millisecond)
	mgr.Create()

	backups, err = mgr.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(backups) != 3 {
		t.Errorf("Expected 3 backups, got %d", len(backups))
	}

	// Verify backups are sorted by time (newest first)
	for i := 0; i < len(backups)-1; i++ {
		if backups[i].CreatedAt.Before(backups[i+1].CreatedAt) {
			t.Error("Backups not sorted by time (newest first)")
		}
	}
}

func TestPruneOldBackups(t *testing.T) {
	setupTestEnv(t)
	mgr, _ := NewManager()

	// Create settings file
	settingsPath, _ := paths.SettingsFile()
	config.SaveSettings(settingsPath, &config.Settings{Model: "opus"})

	// Create more backups than the limit (default 10)
	for i := 0; i < 12; i++ {
		mgr.Create()
		time.Sleep(5 * time.Millisecond)
	}

	// Prune to keep only 5
	err := mgr.Prune(5)
	if err != nil {
		t.Fatalf("Prune() failed: %v", err)
	}

	// Verify only 5 remain
	backups, _ := mgr.List()
	if len(backups) != 5 {
		t.Errorf("Expected 5 backups after pruning, got %d", len(backups))
	}
}

func TestDeleteBackup(t *testing.T) {
	setupTestEnv(t)
	mgr, _ := NewManager()

	// Create settings file
	settingsPath, _ := paths.SettingsFile()
	config.SaveSettings(settingsPath, &config.Settings{Model: "opus"})

	// Create backup
	backupID, _ := mgr.Create()

	// Delete it
	err := mgr.Delete(backupID)
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify it's gone
	backups, _ := mgr.List()
	if len(backups) != 0 {
		t.Errorf("Expected 0 backups after delete, got %d", len(backups))
	}
}

func TestGetLatestBackup(t *testing.T) {
	setupTestEnv(t)
	mgr, _ := NewManager()

	// No backups initially
	latest := mgr.GetLatest()
	if latest != "" {
		t.Errorf("GetLatest() = %s, want empty string when no backups", latest)
	}

	// Create settings file
	settingsPath, _ := paths.SettingsFile()
	config.SaveSettings(settingsPath, &config.Settings{Model: "opus"})

	// Create backups
	mgr.Create()
	time.Sleep(10 * time.Millisecond)
	mgr.Create()
	time.Sleep(10 * time.Millisecond)
	latestID, _ := mgr.Create()

	// Get latest
	latest = mgr.GetLatest()
	if latest != latestID {
		t.Errorf("GetLatest() = %s, want %s", latest, latestID)
	}
}

func TestRestoreLatest(t *testing.T) {
	setupTestEnv(t)
	mgr, _ := NewManager()

	settingsPath, _ := paths.SettingsFile()

	// Create original
	config.SaveSettings(settingsPath, &config.Settings{Model: "opus"})
	mgr.Create()

	// Modify
	config.SaveSettings(settingsPath, &config.Settings{Model: "haiku"})

	// Restore latest
	err := mgr.RestoreLatest()
	if err != nil {
		t.Fatalf("RestoreLatest() failed: %v", err)
	}

	// Verify
	restored, _ := config.LoadSettings(settingsPath)
	if restored.Model != "opus" {
		t.Errorf("Model = %s, want opus", restored.Model)
	}
}

func TestRestoreLatestWhenNoBackups(t *testing.T) {
	setupTestEnv(t)
	mgr, _ := NewManager()

	err := mgr.RestoreLatest()
	if err == nil {
		t.Error("RestoreLatest() should fail when no backups exist")
	}
}
