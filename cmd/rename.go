package cmd

import (
	"fmt"

	"github.com/johnfox/claudectx/internal/printer"
	"github.com/johnfox/claudectx/internal/profile"
	"github.com/johnfox/claudectx/internal/store"
)

// RenameProfile renames an existing profile
func RenameProfile(s *store.Store, oldName, newName string) error {
	// Validate new profile name
	if err := profile.ValidateProfileName(newName); err != nil {
		return fmt.Errorf("invalid new profile name: %w", err)
	}

	// Check if old profile exists
	profiles, err := s.List()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	found := false
	for _, p := range profiles {
		if p == oldName {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("profile %q does not exist", oldName)
	}

	// Check if new name already exists
	for _, p := range profiles {
		if p == newName {
			return fmt.Errorf("profile %q already exists", newName)
		}
	}

	// Load the profile
	prof, err := s.Load(oldName)
	if err != nil {
		return fmt.Errorf("failed to load profile %q: %w", oldName, err)
	}

	// Update the profile name
	prof.Name = newName

	// Save with new name
	if err := s.Save(prof); err != nil {
		return fmt.Errorf("failed to save renamed profile: %w", err)
	}

	// Delete old profile
	if err := s.Delete(oldName); err != nil {
		// Try to rollback - delete the new one
		s.Delete(newName)
		return fmt.Errorf("failed to delete old profile: %w", err)
	}

	// Update current profile if it was the renamed one
	current, err := s.GetCurrent()
	if err == nil && current == oldName {
		if err := s.SetCurrent(newName); err != nil {
			printer.Warning("Profile renamed but failed to update current profile tracker")
		}
	}

	// Update previous profile if it was the renamed one
	previous, err := s.GetPrevious()
	if err == nil && previous == oldName {
		if err := s.SetPrevious(newName); err != nil {
			printer.Warning("Profile renamed but failed to update previous profile tracker")
		}
	}

	printer.Success("Renamed profile %q to %q", oldName, newName)
	return nil
}
