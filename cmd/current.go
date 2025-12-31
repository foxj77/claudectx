package cmd

import (
	"fmt"

	"github.com/johnfox/claudectx/internal/printer"
	"github.com/johnfox/claudectx/internal/store"
)

// ShowCurrent shows the current active profile
func ShowCurrent(s *store.Store) error {
	current, err := s.GetCurrent()
	if err != nil {
		return fmt.Errorf("failed to get current profile: %w", err)
	}

	if current == "" {
		printer.Info("No profile is currently active")
		return nil
	}

	// Use colored output for current profile
	fmt.Println(printer.Colorize(current, printer.Cyan))
	return nil
}
