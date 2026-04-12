package health

import (
	"fmt"
	"strings"

	"github.com/johnfox/claudectx/internal/config"
)

// HealthError represents a health check error
type HealthError struct {
	Message string
	Details string
}

func (e *HealthError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// HealthResult represents the result of a health check
type HealthResult struct {
	IsValid  bool
	Warnings []string
	Error    *HealthError
}

// IsHealthy returns true if there are no errors
func (r *HealthResult) IsHealthy() bool {
	return r.Error == nil && r.IsValid
}

// HasWarnings returns true if there are warnings
func (r *HealthResult) HasWarnings() bool {
	return len(r.Warnings) > 0
}

// ProfileHealthReport contains the complete health check results for a profile
type ProfileHealthReport struct {
	Profile     string
	Overall     HealthResult
	Settings    HealthResult
	Model       HealthResult
	Permissions HealthResult
	EnvVars     HealthResult
}

// IsHealthy returns true if the overall health is good
func (r *ProfileHealthReport) IsHealthy() bool {
	return r.Overall.IsHealthy()
}

// TotalWarnings returns the total number of warnings across all checks
func (r *ProfileHealthReport) TotalWarnings() int {
	total := 0
	total += len(r.Overall.Warnings)
	total += len(r.Settings.Warnings)
	total += len(r.Model.Warnings)
	total += len(r.Permissions.Warnings)
	total += len(r.EnvVars.Warnings)
	return total
}

// Summary returns a brief summary of the health status
func (r *ProfileHealthReport) Summary() string {
	if !r.IsHealthy() {
		return "Unhealthy"
	}
	if r.TotalWarnings() > 0 {
		return "Healthy (with warnings)"
	}
	return "Healthy"
}

// CheckProfile performs a comprehensive health check on a profile
func CheckProfile(profileName string, settings *config.Settings, claudeMD string) *ProfileHealthReport {
	report := &ProfileHealthReport{
		Profile: profileName,
	}

	// Check settings
	report.Settings = CheckSettings(settings)

	// If settings check failed, overall health fails
	if !report.Settings.IsHealthy() {
		report.Overall = HealthResult{
			IsValid: false,
			Error:   report.Settings.Error,
		}
		return report
	}

	// Check model
	report.Model = CheckModel(settings.Model)

	// Check permissions
	report.Permissions = CheckPermissions(settings.Permissions)

	// Check environment variables
	report.EnvVars = CheckEnvVars(settings.Env)

	// Determine overall health
	allHealthy := report.Settings.IsHealthy() &&
		report.Model.IsHealthy() &&
		report.Permissions.IsHealthy() &&
		report.EnvVars.IsHealthy()

	report.Overall = HealthResult{
		IsValid: allHealthy,
	}

	// Collect all warnings
	var allWarnings []string
	allWarnings = append(allWarnings, report.Settings.Warnings...)
	allWarnings = append(allWarnings, report.Model.Warnings...)
	allWarnings = append(allWarnings, report.Permissions.Warnings...)
	allWarnings = append(allWarnings, report.EnvVars.Warnings...)

	report.Overall.Warnings = allWarnings

	return report
}

// CheckSettings validates the settings structure
func CheckSettings(settings *config.Settings) HealthResult {
	if settings == nil {
		return HealthResult{
			IsValid: false,
			Error: &HealthError{
				Message: "Settings cannot be nil",
			},
		}
	}

	warnings := []string{}

	// Warn if no model is set
	if settings.Model == "" {
		warnings = append(warnings, "No model specified (will use Claude Code default)")
	}

	// Warn if no environment variables
	if len(settings.Env) == 0 {
		warnings = append(warnings, "No environment variables set")
	}

	return HealthResult{
		IsValid:  true,
		Warnings: warnings,
	}
}

// Known Claude models
var knownModels = []string{
	"opus",
	"sonnet",
	"haiku",
	"claude-3-opus",
	"claude-3-sonnet",
	"claude-3-haiku",
	"claude-3-opus-20240229",
	"claude-3-sonnet-20240229",
	"claude-3-haiku-20240307",
}

// isKnownModel checks if a model name is in the known list
func isKnownModel(model string) bool {
	modelLower := strings.ToLower(model)
	for _, known := range knownModels {
		if strings.Contains(modelLower, known) {
			return true
		}
	}
	return false
}

// CheckModel validates the model configuration
func CheckModel(model string) HealthResult {
	warnings := []string{}

	if model == "" {
		warnings = append(warnings, "No model specified")
		return HealthResult{
			IsValid:  true,
			Warnings: warnings,
		}
	}

	// Check if it's a known model
	if !isKnownModel(model) {
		warnings = append(warnings, fmt.Sprintf("Unknown model %q (custom models are allowed but may not work)", model))
	}

	return HealthResult{
		IsValid:  true,
		Warnings: warnings,
	}
}

// CheckPermissions validates the permissions configuration
func CheckPermissions(perms *config.Permissions) HealthResult {
	if perms == nil {
		return HealthResult{
			IsValid: true,
		}
	}

	warnings := []string{}

	// Check for wildcard
	for _, allow := range perms.Allow {
		if allow == "*" {
			warnings = append(warnings, "Wildcard (*) in allow list grants access to all tools")
			break
		}
	}

	// Warn if both allow and deny are used
	if len(perms.Allow) > 0 && len(perms.Deny) > 0 {
		warnings = append(warnings, "Both allow and deny lists are specified (deny takes precedence)")
	}

	return HealthResult{
		IsValid:  true,
		Warnings: warnings,
	}
}

// CheckEnvVars validates environment variables
func CheckEnvVars(env map[string]string) HealthResult {
	warnings := []string{}

	// Check for empty values
	for key, value := range env {
		if value == "" {
			warnings = append(warnings, fmt.Sprintf("Environment variable %q has empty value", key))
		}
	}

	return HealthResult{
		IsValid:  true,
		Warnings: warnings,
	}
}
