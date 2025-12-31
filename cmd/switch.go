package cmd

import (
	"fmt"
	"os"

	"github.com/johnfox/claudectx/internal/config"
	"github.com/johnfox/claudectx/internal/paths"
	"github.com/johnfox/claudectx/internal/store"
)

// SwitchProfile switches to a different profile
func SwitchProfile(s *store.Store, name string) error {
	// Check if profile exists
	if !s.Exists(name) {
		return fmt.Errorf("profile %q does not exist", name)
	}

	// Load the target profile
	prof, err := s.Load(name)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	// Get current profile name (for previous tracking)
	currentName, err := s.GetCurrent()
	if err != nil {
		return fmt.Errorf("failed to get current profile: %w", err)
	}

	// Save settings to active location
	settingsPath, err := paths.SettingsFile()
	if err != nil {
		return fmt.Errorf("failed to get settings path: %w", err)
	}

	err = config.SaveSettings(settingsPath, prof.Settings)
	if err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	// Handle CLAUDE.md
	claudeMDPath, err := paths.ClaudeMDFile()
	if err != nil {
		return fmt.Errorf("failed to get CLAUDE.md path: %w", err)
	}

	if prof.ClaudeMD != "" {
		// Write CLAUDE.md
		err = os.WriteFile(claudeMDPath, []byte(prof.ClaudeMD), 0644)
		if err != nil {
			return fmt.Errorf("failed to write CLAUDE.md: %w", err)
		}
	} else {
		// Remove CLAUDE.md if profile doesn't have one
		if config.FileExists(claudeMDPath) {
			os.Remove(claudeMDPath)
		}
	}

	// Update previous profile (if there was a current one)
	if currentName != "" {
		err = s.SetPrevious(currentName)
		if err != nil {
			return fmt.Errorf("failed to set previous profile: %w", err)
		}
	}

	// Update current profile
	err = s.SetCurrent(name)
	if err != nil {
		return fmt.Errorf("failed to set current profile: %w", err)
	}

	fmt.Printf("Switched to profile %q\n", name)
	return nil
}
