package cmd

import (
	"fmt"
	"os"

	"github.com/johnfox/claudectx/internal/config"
	"github.com/johnfox/claudectx/internal/paths"
	"github.com/johnfox/claudectx/internal/profile"
	"github.com/johnfox/claudectx/internal/store"
)

// CreateProfile creates a new profile from the current Claude configuration
func CreateProfile(s *store.Store, name string) error {
	// Validate profile name
	if err := profile.ValidateProfileName(name); err != nil {
		return fmt.Errorf("invalid profile name: %w", err)
	}

	// Check if profile already exists
	if s.Exists(name) {
		return fmt.Errorf("profile %q already exists", name)
	}

	// Load current settings
	settingsPath, err := paths.SettingsFile()
	if err != nil {
		return fmt.Errorf("failed to get settings path: %w", err)
	}

	settings := config.LoadSettingsOrEmpty(settingsPath)

	// Load current CLAUDE.md if it exists
	claudeMDPath, err := paths.ClaudeMDFile()
	if err != nil {
		return fmt.Errorf("failed to get CLAUDE.md path: %w", err)
	}

	var claudeMD string
	if config.FileExists(claudeMDPath) {
		content, err := os.ReadFile(claudeMDPath)
		if err != nil {
			return fmt.Errorf("failed to read CLAUDE.md: %w", err)
		}
		claudeMD = string(content)
	}

	// Create the profile
	prof := profile.ProfileFromCurrent(name, settings, claudeMD)

	// Save it
	err = s.Save(prof)
	if err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	fmt.Printf("Created profile %q from current configuration\n", name)
	return nil
}
