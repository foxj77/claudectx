package paths

import (
	"errors"
	"os"
	"path/filepath"
)

// ClaudeDir returns the path to the Claude configuration directory (~/.claude)
func ClaudeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude"), nil
}

// ProfilesDir returns the path to the profiles directory (~/.claude/profiles)
func ProfilesDir() (string, error) {
	claudeDir, err := ClaudeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(claudeDir, "profiles"), nil
}

// ProfileDir returns the path to a specific profile directory
func ProfileDir(name string) (string, error) {
	if name == "" {
		return "", errors.New("profile name cannot be empty")
	}
	profilesDir, err := ProfilesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(profilesDir, name), nil
}

// CurrentProfileFile returns the path to the current profile tracking file
func CurrentProfileFile() (string, error) {
	claudeDir, err := ClaudeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(claudeDir, ".claudectx-current"), nil
}

// PreviousProfileFile returns the path to the previous profile tracking file
func PreviousProfileFile() (string, error) {
	claudeDir, err := ClaudeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(claudeDir, ".claudectx-previous"), nil
}

// SettingsFile returns the path to the active settings.json file
func SettingsFile() (string, error) {
	claudeDir, err := ClaudeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(claudeDir, "settings.json"), nil
}

// ClaudeMDFile returns the path to the active CLAUDE.md file
func ClaudeMDFile() (string, error) {
	claudeDir, err := ClaudeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(claudeDir, "CLAUDE.md"), nil
}

// ProfileFile returns the path to a specific file within a profile directory
func ProfileFile(profileName, filename string) (string, error) {
	profileDir, err := ProfileDir(profileName)
	if err != nil {
		return "", err
	}
	return filepath.Join(profileDir, filename), nil
}

// EnsureProfilesDir creates the profiles directory if it doesn't exist
func EnsureProfilesDir() error {
	profilesDir, err := ProfilesDir()
	if err != nil {
		return err
	}
	return os.MkdirAll(profilesDir, 0755)
}

// EnsureProfileDir creates a specific profile directory if it doesn't exist
func EnsureProfileDir(name string) error {
	profileDir, err := ProfileDir(name)
	if err != nil {
		return err
	}
	return os.MkdirAll(profileDir, 0755)
}
