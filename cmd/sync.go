package cmd

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"

	"github.com/johnfox/claudectx/internal/config"
	"github.com/johnfox/claudectx/internal/paths"
	"github.com/johnfox/claudectx/internal/printer"
	"github.com/johnfox/claudectx/internal/store"
)

// hasConfigChanged compares active configuration with stored profile
// Returns true if they differ, false otherwise
func hasConfigChanged(s *store.Store, profileName string) (bool, error) {
	// Load the stored profile
	stored, err := s.Load(profileName)
	if err != nil {
		return false, fmt.Errorf("failed to load profile: %w", err)
	}

	// Read active settings
	settingsPath, err := paths.SettingsFile()
	if err != nil {
		return false, err
	}

	activeSettings, err := config.LoadSettings(settingsPath)
	if err != nil {
		// If file doesn't exist or can't be read, consider it changed
		return true, nil
	}

	// Read active CLAUDE.md
	claudeMDPath, err := paths.ClaudeMDFile()
	if err != nil {
		return false, err
	}

	var activeClaudeMD string
	if config.FileExists(claudeMDPath) {
		content, err := os.ReadFile(claudeMDPath)
		if err != nil {
			return true, nil
		}
		activeClaudeMD = string(content)
	}

	// Compare by hashing (more efficient than deep comparison)
	return !profilesEqual(activeSettings, activeClaudeMD, stored.Settings, stored.ClaudeMD), nil
}

// profilesEqual compares two profile configurations by content hash
func profilesEqual(settings1 *config.Settings, claudeMD1 string, settings2 *config.Settings, claudeMD2 string) bool {
	// Hash settings
	hash1 := hashSettings(settings1)
	hash2 := hashSettings(settings2)

	if hash1 != hash2 {
		return false
	}

	// Hash CLAUDE.md
	hashMD1 := hashString(claudeMD1)
	hashMD2 := hashString(claudeMD2)

	return hashMD1 == hashMD2
}

// hashSettings creates a hash of settings for comparison
func hashSettings(settings *config.Settings) string {
	if settings == nil {
		return "nil"
	}

	// Marshal to JSON for consistent hashing
	data, err := json.Marshal(settings)
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", md5.Sum(data))
}

// hashString creates a hash of a string for comparison
func hashString(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// syncCurrentProfile saves the active configuration back to the current profile
func syncCurrentProfile(s *store.Store, profileName string) error {
	// Read active settings
	settingsPath, err := paths.SettingsFile()
	if err != nil {
		return fmt.Errorf("failed to get settings path: %w", err)
	}

	activeSettings, err := config.LoadSettings(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to load active settings: %w", err)
	}

	// Read active CLAUDE.md
	claudeMDPath, err := paths.ClaudeMDFile()
	if err != nil {
		return fmt.Errorf("failed to get CLAUDE.md path: %w", err)
	}

	var activeClaudeMD string
	if config.FileExists(claudeMDPath) {
		content, err := os.ReadFile(claudeMDPath)
		if err != nil {
			return fmt.Errorf("failed to read CLAUDE.md: %w", err)
		}
		activeClaudeMD = string(content)
	}

	// Load the existing profile to update it
	prof, err := s.Load(profileName)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	// Update the profile with active configuration
	prof.Settings = activeSettings
	prof.ClaudeMD = activeClaudeMD
	prof.Touch()

	// Save the updated profile
	if err := s.Save(prof); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	return nil
}

// SyncProfile saves the current active configuration to the specified profile
func SyncProfile(s *store.Store, profileName string) error {
	// Validate profile exists
	if !s.Exists(profileName) {
		return fmt.Errorf("profile %q does not exist", profileName)
	}

	// Sync the configuration
	if err := syncCurrentProfile(s, profileName); err != nil {
		return err
	}

	printer.Success("Synced active configuration to profile %q", profileName)
	return nil
}

// SyncCurrentProfile saves the current active configuration to the current profile
func SyncCurrentProfile(s *store.Store) error {
	// Get current profile name
	currentName, err := s.GetCurrent()
	if err != nil {
		return fmt.Errorf("failed to get current profile: %w", err)
	}

	if currentName == "" {
		return fmt.Errorf("no current profile to sync to")
	}

	return SyncProfile(s, currentName)
}
