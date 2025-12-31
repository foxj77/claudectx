package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Settings represents the structure of settings.json
type Settings struct {
	Env         map[string]string `json:"env,omitempty"`
	Model       string            `json:"model,omitempty"`
	Permissions *Permissions      `json:"permissions,omitempty"`
	// Add other fields as needed, using omitempty to preserve minimal configs
}

// Permissions represents the permissions section of settings
type Permissions struct {
	Allow []string `json:"allow,omitempty"`
	Deny  []string `json:"deny,omitempty"`
}

// LoadSettings reads and parses a settings.json file
func LoadSettings(path string) (*Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings Settings
	err = json.Unmarshal(data, &settings)
	if err != nil {
		return nil, fmt.Errorf("failed to parse settings JSON: %w", err)
	}

	return &settings, nil
}

// LoadSettingsOrEmpty loads settings from a file, or returns empty settings if file doesn't exist
func LoadSettingsOrEmpty(path string) *Settings {
	settings, err := LoadSettings(path)
	if err != nil {
		// Return empty settings if file doesn't exist or can't be read
		return &Settings{
			Env: make(map[string]string),
		}
	}
	return settings
}

// SaveSettings writes settings to a JSON file with formatting
func SaveSettings(path string, settings *Settings) error {
	// Marshal with indentation for readability
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Add newline at end of file (common convention)
	data = append(data, '\n')

	// Write to file
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Sync to ensure data is written to disk
	err = destFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
