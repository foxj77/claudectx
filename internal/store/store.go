package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/johnfox/claudectx/internal/config"
	"github.com/johnfox/claudectx/internal/paths"
	"github.com/johnfox/claudectx/internal/profile"
)

// Store manages profile persistence on the filesystem
type Store struct {
	profilesDir string
}

// NewStore creates a new Store and ensures the profiles directory exists
func NewStore() (*Store, error) {
	profilesDir, err := paths.ProfilesDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles directory: %w", err)
	}

	// Ensure the profiles directory exists
	err = paths.EnsureProfilesDir()
	if err != nil {
		return nil, fmt.Errorf("failed to create profiles directory: %w", err)
	}

	return &Store{
		profilesDir: profilesDir,
	}, nil
}

// Save saves a profile to disk
func (s *Store) Save(prof *profile.Profile) error {
	if err := prof.Validate(); err != nil {
		return fmt.Errorf("invalid profile: %w", err)
	}

	// Ensure profile directory exists
	err := paths.EnsureProfileDir(prof.Name)
	if err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	// Save settings.json
	settingsPath, err := paths.ProfileFile(prof.Name, "settings.json")
	if err != nil {
		return err
	}

	err = config.SaveSettings(settingsPath, prof.Settings)
	if err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	// Save CLAUDE.md if present
	if strings.TrimSpace(prof.ClaudeMD) != "" {
		claudeMDPath, err := paths.ProfileFile(prof.Name, "CLAUDE.md")
		if err != nil {
			return err
		}

		err = os.WriteFile(claudeMDPath, []byte(prof.ClaudeMD), 0644)
		if err != nil {
			return fmt.Errorf("failed to save CLAUDE.md: %w", err)
		}
	}

	return nil
}

// Load loads a profile from disk
func (s *Store) Load(name string) (*profile.Profile, error) {
	if !s.Exists(name) {
		return nil, fmt.Errorf("profile %q does not exist", name)
	}

	prof := profile.NewProfile(name)

	// Load settings.json
	settingsPath, err := paths.ProfileFile(name, "settings.json")
	if err != nil {
		return nil, err
	}

	settings, err := config.LoadSettings(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}
	prof.Settings = settings

	// Load CLAUDE.md if it exists
	claudeMDPath, err := paths.ProfileFile(name, "CLAUDE.md")
	if err != nil {
		return nil, err
	}

	if config.FileExists(claudeMDPath) {
		content, err := os.ReadFile(claudeMDPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CLAUDE.md: %w", err)
		}
		prof.ClaudeMD = string(content)
	}

	return prof, nil
}

// List returns the names of all profiles
func (s *Store) List() ([]string, error) {
	entries, err := os.ReadDir(s.profilesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read profiles directory: %w", err)
	}

	var profiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Verify it has a settings.json file to be considered a valid profile
			settingsPath := filepath.Join(s.profilesDir, entry.Name(), "settings.json")
			if config.FileExists(settingsPath) {
				profiles = append(profiles, entry.Name())
			}
		}
	}

	return profiles, nil
}

// Exists checks if a profile exists
func (s *Store) Exists(name string) bool {
	profileDir, err := paths.ProfileDir(name)
	if err != nil {
		return false
	}

	settingsPath := filepath.Join(profileDir, "settings.json")
	return config.FileExists(settingsPath)
}

// Delete removes a profile from disk
func (s *Store) Delete(name string) error {
	if !s.Exists(name) {
		return fmt.Errorf("profile %q does not exist", name)
	}

	profileDir, err := paths.ProfileDir(name)
	if err != nil {
		return err
	}

	err = os.RemoveAll(profileDir)
	if err != nil {
		return fmt.Errorf("failed to delete profile directory: %w", err)
	}

	return nil
}

// GetCurrent returns the name of the current profile
func (s *Store) GetCurrent() (string, error) {
	currentFile, err := paths.CurrentProfileFile()
	if err != nil {
		return "", err
	}

	if !config.FileExists(currentFile) {
		return "", nil
	}

	content, err := os.ReadFile(currentFile)
	if err != nil {
		return "", fmt.Errorf("failed to read current profile file: %w", err)
	}

	return strings.TrimSpace(string(content)), nil
}

// SetCurrent sets the current profile name
func (s *Store) SetCurrent(name string) error {
	currentFile, err := paths.CurrentProfileFile()
	if err != nil {
		return err
	}

	if name == "" {
		// Remove the current profile file
		if config.FileExists(currentFile) {
			return os.Remove(currentFile)
		}
		return nil
	}

	err = os.WriteFile(currentFile, []byte(name), 0644)
	if err != nil {
		return fmt.Errorf("failed to write current profile file: %w", err)
	}

	return nil
}

// GetPrevious returns the name of the previous profile
func (s *Store) GetPrevious() (string, error) {
	prevFile, err := paths.PreviousProfileFile()
	if err != nil {
		return "", err
	}

	if !config.FileExists(prevFile) {
		return "", nil
	}

	content, err := os.ReadFile(prevFile)
	if err != nil {
		return "", fmt.Errorf("failed to read previous profile file: %w", err)
	}

	return strings.TrimSpace(string(content)), nil
}

// SetPrevious sets the previous profile name
func (s *Store) SetPrevious(name string) error {
	prevFile, err := paths.PreviousProfileFile()
	if err != nil {
		return err
	}

	if name == "" {
		// Remove the previous profile file
		if config.FileExists(prevFile) {
			return os.Remove(prevFile)
		}
		return nil
	}

	err = os.WriteFile(prevFile, []byte(name), 0644)
	if err != nil {
		return fmt.Errorf("failed to write previous profile file: %w", err)
	}

	return nil
}
