package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/johnfox/claudectx/internal/printer"
	"github.com/johnfox/claudectx/internal/selector"
	"github.com/johnfox/claudectx/internal/store"
	"golang.org/x/term"
)

// ListProfilesInteractive displays an interactive profile selector
func ListProfilesInteractive(s *store.Store) error {
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
		// Not a fatal error - just means no current profile
		current = ""
	}

	// Sort profiles alphabetically
	sort.Strings(profiles)

	// Check if we're in a TTY
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		// Not a TTY, fall back to simple list
		return ListProfiles(s)
	}

	// Build options for selector
	options := make([]selector.Option, len(profiles))
	for i, profile := range profiles {
		options[i] = selector.Option{
			Label:     profile,
			IsCurrent: profile == current,
		}
	}

	// Show interactive selector
	selected, err := selector.Select("Select a profile:", options)
	if err != nil {
		// User cancelled or error occurred
		if err.Error() == "cancelled" {
			return nil // Exit gracefully
		}
		return err
	}

	// Get selected profile
	selectedProfile := profiles[selected]

	// If it's already the current profile, no need to switch
	if selectedProfile == current {
		printer.Info("Already using profile %q", selectedProfile)
		return nil
	}

	// Switch to the selected profile
	return SwitchProfile(s, selectedProfile)
}
