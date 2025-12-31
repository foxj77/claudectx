package cmd

import (
	"fmt"
	"sort"

	"github.com/johnfox/claudectx/internal/store"
)

// ListProfiles lists all available profiles, highlighting the current one
func ListProfiles(s *store.Store) error {
	profiles, err := s.List()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	if len(profiles) == 0 {
		fmt.Println("No profiles found. Create one with: claudectx -n <name>")
		return nil
	}

	// Get current profile
	current, err := s.GetCurrent()
	if err != nil {
		return fmt.Errorf("failed to get current profile: %w", err)
	}

	// Sort profiles alphabetically
	sort.Strings(profiles)

	// Print each profile
	for _, name := range profiles {
		if name == current {
			// Highlight current profile (using bold/color in future)
			fmt.Printf("%s (current)\n", name)
		} else {
			fmt.Println(name)
		}
	}

	return nil
}
