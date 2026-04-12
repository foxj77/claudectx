package cmd

import (
	"fmt"
	"os"

	"github.com/johnfox/claudectx/internal/backup"
	"github.com/johnfox/claudectx/internal/config"
	"github.com/johnfox/claudectx/internal/mcpconfig"
	"github.com/johnfox/claudectx/internal/paths"
	"github.com/johnfox/claudectx/internal/printer"
	"github.com/johnfox/claudectx/internal/profile"
	"github.com/johnfox/claudectx/internal/store"
	"github.com/johnfox/claudectx/internal/validator"
)

// SwitchProfile switches to a different profile with backup and validation
func SwitchProfile(s *store.Store, name string) error {
	// Validate profile name
	if err := profile.ValidateProfileName(name); err != nil {
		return fmt.Errorf("invalid profile name: %w", err)
	}

	// Check if profile exists
	if !s.Exists(name) {
		return fmt.Errorf("profile %q does not exist", name)
	}

	// Load the target profile
	prof, err := s.Load(name)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	// Validate the profile before switching
	if err := prof.Validate(); err != nil {
		return fmt.Errorf("profile validation failed: %w", err)
	}

	// Validate settings
	if err := validator.ValidateSettings(prof.Settings); err != nil {
		return fmt.Errorf("profile settings are invalid: %w", err)
	}

	// Validate CLAUDE.md content
	if err := validator.ValidateClaudeMD(prof.ClaudeMD); err != nil {
		return fmt.Errorf("profile CLAUDE.md is invalid: %w", err)
	}

	// Create backup manager
	backupMgr, err := backup.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize backup manager: %w", err)
	}

	// Create backup before switching
	backupID, err := backupMgr.Create()
	if err != nil {
		printer.Warning("Warning: Failed to create backup: %v", err)
		printer.Warning("Continuing without backup...")
	} else {
		printer.Info("Created backup: %s", backupID)
	}

	// Get current profile name (for previous tracking and auto-sync)
	currentName, err := s.GetCurrent()
	if err != nil {
		return fmt.Errorf("failed to get current profile: %w", err)
	}

	// Auto-sync: Save current changes back to current profile before switching
	if currentName != "" && currentName != name {
		changed, err := hasConfigChanged(s, currentName)
		if err != nil {
			printer.Warning("Warning: Could not detect config changes: %v", err)
		} else if changed {
			printer.Info("Auto-syncing changes to profile %q...", currentName)
			err := syncCurrentProfile(s, currentName)
			if err != nil {
				printer.Warning("Warning: Failed to auto-sync profile: %v", err)
				printer.Warning("Continuing with switch anyway...")
			} else {
				printer.Success("Auto-synced %d changes to profile %q", 1, currentName)
			}
		}
	}

	// Save settings to active location
	settingsPath, err := paths.SettingsFile()
	if err != nil {
		rollback(backupMgr, backupID)
		return fmt.Errorf("failed to get settings path: %w", err)
	}

	err = config.SaveSettings(settingsPath, prof.Settings)
	if err != nil {
		rollback(backupMgr, backupID)
		return fmt.Errorf("failed to save settings: %w", err)
	}

	// Handle CLAUDE.md
	claudeMDPath, err := paths.ClaudeMDFile()
	if err != nil {
		rollback(backupMgr, backupID)
		return fmt.Errorf("failed to get CLAUDE.md path: %w", err)
	}

	if prof.ClaudeMD != "" {
		// Write CLAUDE.md
		err = os.WriteFile(claudeMDPath, []byte(prof.ClaudeMD), 0644)
		if err != nil {
			rollback(backupMgr, backupID)
			return fmt.Errorf("failed to write CLAUDE.md: %w", err)
		}
	} else {
		// Remove CLAUDE.md if profile doesn't have one
		if config.FileExists(claudeMDPath) {
			os.Remove(claudeMDPath)
		}
	}

	// Handle MCP servers in ~/.claude.json
	claudeJSONPath, err := paths.ClaudeJSONFile()
	if err != nil {
		rollback(backupMgr, backupID)
		return fmt.Errorf("failed to get claude.json path: %w", err)
	}

	err = mcpconfig.SaveMCPServers(claudeJSONPath, prof.MCPServers)
	if err != nil {
		rollback(backupMgr, backupID)
		return fmt.Errorf("failed to save MCP servers: %w", err)
	}

	// Update previous profile (if there was a current one)
	if currentName != "" {
		err = s.SetPrevious(currentName)
		if err != nil {
			rollback(backupMgr, backupID)
			return fmt.Errorf("failed to set previous profile: %w", err)
		}
	}

	// Update current profile
	err = s.SetCurrent(name)
	if err != nil {
		rollback(backupMgr, backupID)
		return fmt.Errorf("failed to set current profile: %w", err)
	}

	// Prune old backups (keep last 10)
	if err := backupMgr.Prune(10); err != nil {
		printer.Warning("Warning: Failed to prune old backups: %v", err)
	}

	printer.Success("Switched to profile %q", name)
	return nil
}

// rollback attempts to restore from backup
func rollback(backupMgr *backup.Manager, backupID string) {
	if backupID == "" {
		return // No backup to restore from
	}

	printer.Warning("Rolling back due to error...")
	err := backupMgr.Restore(backupID)
	if err != nil {
		printer.Error("Rollback failed: %v", err)
		printer.Error("You may need to manually restore your configuration")
		printer.Error("Backup ID: %s", backupID)
	} else {
		printer.Info("Successfully rolled back to previous configuration")
	}
}
