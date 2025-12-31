package cmd

import (
	"fmt"

	"github.com/johnfox/claudectx/internal/store"
)

// DeleteProfile deletes a profile
func DeleteProfile(s *store.Store, name string) error {
	// Check if profile exists
	if !s.Exists(name) {
		return fmt.Errorf("profile %q does not exist", name)
	}

	// Check if it's the current profile
	current, err := s.GetCurrent()
	if err != nil {
		return fmt.Errorf("failed to get current profile: %w", err)
	}

	if current == name {
		return fmt.Errorf("cannot delete current profile %q - switch to another profile first", name)
	}

	// Delete the profile
	err = s.Delete(name)
	if err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	// Clear previous profile if it was this one
	prev, err := s.GetPrevious()
	if err == nil && prev == name {
		s.SetPrevious("")
	}

	fmt.Printf("Deleted profile %q\n", name)
	return nil
}
