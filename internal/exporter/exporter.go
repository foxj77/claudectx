package exporter

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/johnfox/claudectx/internal/config"
	"github.com/johnfox/claudectx/internal/profile"
	"github.com/johnfox/claudectx/internal/store"
	"github.com/johnfox/claudectx/internal/validator"
)

// ExportVersion is the current export format version
const ExportVersion = "1.0.0"

// ExportedProfile represents a profile in export format
type ExportedProfile struct {
	Version    string           `json:"version"`
	Name       string           `json:"name"`
	Settings   *config.Settings `json:"settings"`
	ClaudeMD   string           `json:"claude_md,omitempty"`
	ExportedAt string           `json:"exported_at"`
}

// ExportProfile exports a profile to JSON format
func ExportProfile(s *store.Store, profileName string, w io.Writer) error {
	// Check if profile exists
	if !s.Exists(profileName) {
		return fmt.Errorf("profile %q does not exist", profileName)
	}

	// Load the profile
	prof, err := s.Load(profileName)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	// Create export structure
	exported := ExportedProfile{
		Version:    ExportVersion,
		Name:       profileName,
		Settings:   prof.Settings,
		ClaudeMD:   prof.ClaudeMD,
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Marshal to pretty-printed JSON
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(exported)
	if err != nil {
		return fmt.Errorf("failed to encode profile: %w", err)
	}

	return nil
}

// ImportProfile imports a profile from JSON format
func ImportProfile(s *store.Store, r io.Reader, newName string) error {
	// Decode JSON
	var exported ExportedProfile
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&exported)
	if err != nil {
		return fmt.Errorf("failed to decode profile: %w", err)
	}

	// Validate version compatibility
	if exported.Version != ExportVersion {
		return fmt.Errorf("incompatible export version %q (expected %q)", exported.Version, ExportVersion)
	}

	// Use new name if provided, otherwise use exported name
	profileName := exported.Name
	if newName != "" {
		profileName = newName
	}

	// Check if profile already exists
	if s.Exists(profileName) {
		return fmt.Errorf("profile %q already exists", profileName)
	}

	// Validate settings
	if err := validator.ValidateSettings(exported.Settings); err != nil {
		return fmt.Errorf("imported settings are invalid: %w", err)
	}

	// Validate CLAUDE.md
	if err := validator.ValidateClaudeMD(exported.ClaudeMD); err != nil {
		return fmt.Errorf("imported CLAUDE.md is invalid: %w", err)
	}

	// Create profile from imported data
	prof := profile.ProfileFromCurrent(profileName, exported.Settings, exported.ClaudeMD)

	// Save the profile
	err = s.Save(prof)
	if err != nil {
		return fmt.Errorf("failed to save imported profile: %w", err)
	}

	return nil
}
