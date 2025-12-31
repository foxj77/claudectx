package cmd

import (
	"fmt"

	"github.com/johnfox/claudectx/internal/store"
)

// ShowCurrent shows the current active profile
func ShowCurrent(s *store.Store) error {
	current, err := s.GetCurrent()
	if err != nil {
		return fmt.Errorf("failed to get current profile: %w", err)
	}

	if current == "" {
		fmt.Println("No profile is currently active")
		return nil
	}

	fmt.Println(current)
	return nil
}
