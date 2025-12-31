package cmd

import (
	"fmt"
	"sort"

	"github.com/johnfox/claudectx/internal/printer"
	"github.com/johnfox/claudectx/internal/store"
)

// ListProfiles lists all available profiles, highlighting the current one
func ListProfiles(s *store.Store) error {
	profiles, err := s.List()
	if err != nil {
		return fmt.Errorf("failed to list profiles: %w", err)
	}

	if len(profiles) == 0 {
		printer.Info("No profiles found. Create one with: claudectx -n <name>")
		return nil
	}

	// Get current profile
	current, err := s.GetCurrent()
	if err != nil {
		return fmt.Errorf("failed to get current profile: %w", err)
	}

	// Sort profiles alphabetically
	sort.Strings(profiles)

	// Build highlight map
	highlightMap := make(map[string]string)
	if current != "" {
		highlightMap[current] = "(current)"
	}

	// Print using the printer package
	printer.PrintList(profiles, highlightMap)

	return nil
}
