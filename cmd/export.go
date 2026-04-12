package cmd

import (
	"fmt"
	"os"

	"github.com/johnfox/claudectx/internal/exporter"
	"github.com/johnfox/claudectx/internal/printer"
	"github.com/johnfox/claudectx/internal/store"
)

// ExportProfile exports a profile to JSON format
func ExportProfile(s *store.Store, profileName string, outputPath string) error {
	// Verify profile exists
	if !s.Exists(profileName) {
		return fmt.Errorf("profile %q does not exist", profileName)
	}

	// Determine output destination
	var output *os.File
	var err error

	if outputPath == "" || outputPath == "-" {
		// Output to stdout
		output = os.Stdout
	} else {
		// Output to file
		output, err = os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer output.Close()
	}

	// Export the profile
	err = exporter.ExportProfile(s, profileName, output)
	if err != nil {
		return fmt.Errorf("failed to export profile: %w", err)
	}

	// Show success message (but not to stdout if that's where we exported)
	if outputPath != "" && outputPath != "-" {
		printer.Success("Exported profile %q to %s", profileName, outputPath)
	}

	return nil
}
