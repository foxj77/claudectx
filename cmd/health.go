package cmd

import (
	"fmt"

	"github.com/johnfox/claudectx/internal/health"
	"github.com/johnfox/claudectx/internal/printer"
	"github.com/johnfox/claudectx/internal/store"
)

// Health checks the health of a profile
func Health(args []string) error {
	s, err := store.NewStore()
	if err != nil {
		return err
	}

	// Determine which profile to check
	profileName := ""
	if len(args) > 0 {
		profileName = args[0]
	} else {
		// Check current profile
		current, err := s.GetCurrent()
		if err != nil {
			return fmt.Errorf("no profile specified and no current profile set")
		}
		profileName = current
	}

	// Load the profile
	prof, err := s.Load(profileName)
	if err != nil {
		return fmt.Errorf("failed to load profile %q: %w", profileName, err)
	}

	// Run health checks
	report := health.CheckProfile(prof.Name, prof.Settings, prof.ClaudeMD)

	// Display the report
	displayHealthReport(report)

	// Return error if unhealthy
	if !report.IsHealthy() {
		return fmt.Errorf("profile %q is unhealthy", profileName)
	}

	return nil
}

// displayHealthReport prints the health report with colored output
func displayHealthReport(report *health.ProfileHealthReport) {
	fmt.Printf("Health Check for Profile: %s\n", printer.Colorize(report.Profile, printer.Cyan))
	fmt.Println()

	// Overall status
	if report.IsHealthy() {
		printer.Success("✓ Overall Status: %s", report.Summary())
	} else {
		printer.Error("✗ Overall Status: %s", report.Summary())
	}
	fmt.Println()

	// Settings check
	displayHealthResult("Settings", report.Settings)

	// Model check
	displayHealthResult("Model", report.Model)

	// Permissions check
	displayHealthResult("Permissions", report.Permissions)

	// Environment variables check
	displayHealthResult("Environment Variables", report.EnvVars)

	// Summary
	if report.TotalWarnings() > 0 {
		fmt.Println()
		printer.Warning("Total warnings: %d", report.TotalWarnings())
	}

	if report.Overall.Error != nil {
		fmt.Println()
		printer.Error("Error: %s", report.Overall.Error.Error())
	}
}

// displayHealthResult prints a single health check result
func displayHealthResult(name string, result health.HealthResult) {
	if result.Error != nil {
		printer.Error("✗ %s: %s", name, result.Error.Error())
	} else if result.IsValid {
		if len(result.Warnings) > 0 {
			printer.Warning("⚠ %s: Valid with warnings", name)
			for _, warning := range result.Warnings {
				fmt.Printf("  - %s\n", warning)
			}
		} else {
			printer.Success("✓ %s: Valid", name)
		}
	} else {
		printer.Error("✗ %s: Invalid", name)
	}
}
