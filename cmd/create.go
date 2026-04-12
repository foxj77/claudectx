package cmd

import (
	"fmt"
	"os"

	"github.com/johnfox/claudectx/internal/config"
	"github.com/johnfox/claudectx/internal/mcpconfig"
	"github.com/johnfox/claudectx/internal/paths"
	"github.com/johnfox/claudectx/internal/printer"
	"github.com/johnfox/claudectx/internal/profile"
	"github.com/johnfox/claudectx/internal/store"
	"github.com/johnfox/claudectx/internal/validator"
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

	// Validate current settings before creating profile
	if err := validator.ValidateSettings(settings); err != nil {
		printer.Warning("Warning: Current settings may be invalid: %v", err)
		printer.Warning("Creating profile anyway...")
	}

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

		// Validate CLAUDE.md
		if err := validator.ValidateClaudeMD(claudeMD); err != nil {
			printer.Warning("Warning: CLAUDE.md may be invalid: %v", err)
			printer.Warning("Creating profile anyway...")
		}
	}

	// Load current MCP servers from ~/.claude.json
	claudeJSONPath, err := paths.ClaudeJSONFile()
	if err != nil {
		return fmt.Errorf("failed to get claude.json path: %w", err)
	}

	mcpServers, err := mcpconfig.LoadMCPServers(claudeJSONPath)
	if err != nil {
		// If we can't load MCP servers, just use empty
		mcpServers = make(mcpconfig.MCPServers)
	}

	// Create the profile
	prof := profile.ProfileFromCurrent(name, settings, claudeMD, mcpServers)

	// Save it
	err = s.Save(prof)
	if err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	printer.Success("Created profile %q from current configuration", name)

	// Show what was captured
	if settings.Model != "" {
		printer.Info("  Model: %s", settings.Model)
	}
	if len(settings.Env) > 0 {
		printer.Info("  Environment variables: %d", len(settings.Env))
	}
	if settings.Permissions != nil && (len(settings.Permissions.Allow) > 0 || len(settings.Permissions.Deny) > 0) {
		printer.Info("  Permissions configured: yes")
	}
	if claudeMD != "" {
		printer.Info("  CLAUDE.md included: yes")
	}
	if len(mcpServers) > 0 {
		printer.Info("  MCP servers: %d", len(mcpServers))
	}

	return nil
}
