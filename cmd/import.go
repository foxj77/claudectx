package cmd

import (
	"fmt"
	"os"

	"github.com/johnfox/claudectx/internal/exporter"
	"github.com/johnfox/claudectx/internal/printer"
	"github.com/johnfox/claudectx/internal/store"
)

// ImportProfile imports a profile from JSON format
func ImportProfile(s *store.Store, inputPath string, newName string) error {
	// Determine input source
	var input *os.File
	var err error

	if inputPath == "" || inputPath == "-" {
		// Read from stdin
		input = os.Stdin
	} else {
		// Read from file
		input, err = os.Open(inputPath)
		if err != nil {
			return fmt.Errorf("failed to open input file: %w", err)
		}
		defer input.Close()
	}

	// Import the profile
	err = exporter.ImportProfile(s, input, newName)
	if err != nil {
		return fmt.Errorf("failed to import profile: %w", err)
	}

	// Determine what name was used
	profileName := newName
	if profileName == "" {
		// Profile was imported with its original name
		// We don't have easy access to it here, so just say "profile"
		printer.Success("Profile imported successfully")
	} else {
		printer.Success("Imported profile as %q", profileName)
	}

	return nil
}
