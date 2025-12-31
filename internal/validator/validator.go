package validator

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/johnfox/claudectx/internal/config"
)

// ValidateJSONFile validates that a file contains valid JSON
func ValidateJSONFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if len(data) == 0 {
		return fmt.Errorf("file is empty")
	}

	// Try to parse as JSON
	var result interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}

// ValidateSettings validates a Settings struct
func ValidateSettings(settings *config.Settings) error {
	if settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}

	// Validate model if present
	if settings.Model != "" {
		if err := ValidateModel(settings.Model); err != nil {
			return err
		}
	}

	// Validate permissions if present
	if err := ValidatePermissions(settings.Permissions); err != nil {
		return err
	}

	// Validate env if present
	if err := ValidateEnv(settings.Env); err != nil {
		return err
	}

	return nil
}

// ValidateModel validates a model name
func ValidateModel(model string) error {
	// Empty is OK
	if model == "" {
		return nil
	}

	// We don't restrict model names since users might use custom models
	// Just ensure it's not too long
	if len(model) > 255 {
		return fmt.Errorf("model name too long (max 255 characters)")
	}

	return nil
}

// ValidatePermissions validates permissions configuration
func ValidatePermissions(perms *config.Permissions) error {
	// nil is valid
	if perms == nil {
		return nil
	}

	// Just ensure lists aren't unreasonably large
	if len(perms.Allow) > 1000 {
		return fmt.Errorf("too many entries in allow list (max 1000)")
	}

	if len(perms.Deny) > 1000 {
		return fmt.Errorf("too many entries in deny list (max 1000)")
	}

	return nil
}

// ValidateEnv validates environment variables
func ValidateEnv(env map[string]string) error {
	// nil and empty are both valid
	if env == nil || len(env) == 0 {
		return nil
	}

	// Just ensure it's not unreasonably large
	if len(env) > 1000 {
		return fmt.Errorf("too many environment variables (max 1000)")
	}

	return nil
}

// ValidateSettingsFile validates a settings.json file
func ValidateSettingsFile(path string) error {
	// First check if it's valid JSON
	if err := ValidateJSONFile(path); err != nil {
		return err
	}

	// Load and validate the settings structure
	settings, err := config.LoadSettings(path)
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	return ValidateSettings(settings)
}

// ValidateClaudeMD validates CLAUDE.md content
func ValidateClaudeMD(content string) error {
	// CLAUDE.md can contain anything, so we just do basic checks
	// Empty content is valid
	// Very large files might be suspicious but we'll allow them
	if len(content) > 10*1024*1024 { // 10MB limit
		return fmt.Errorf("CLAUDE.md file too large (max 10MB)")
	}

	return nil
}
