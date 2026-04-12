package cmd

import (
	"fmt"

	"github.com/johnfox/claudectx/internal/store"
)

// TogglePrevious switches to the previous profile (like `cd -`)
func TogglePrevious(s *store.Store) error {
	// Get previous profile
	prev, err := s.GetPrevious()
	if err != nil {
		return fmt.Errorf("failed to get previous profile: %w", err)
	}

	if prev == "" {
		return fmt.Errorf("no previous profile to switch to")
	}

	// Check if previous profile still exists
	if !s.Exists(prev) {
		return fmt.Errorf("previous profile %q no longer exists", prev)
	}

	// Switch to it (this will handle updating current/previous)
	return SwitchProfile(s, prev)
}
