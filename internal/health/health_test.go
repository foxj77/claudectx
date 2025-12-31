package health

import (
	"testing"

	"github.com/johnfox/claudectx/internal/config"
)

func TestCheckSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings *config.Settings
		wantErr  bool
		wantWarnings bool
	}{
		{
			name:     "valid settings with model",
			settings: &config.Settings{Model: "opus"},
			wantErr:  false,
			wantWarnings: false,
		},
		{
			name:     "empty settings",
			settings: &config.Settings{},
			wantErr:  false,
			wantWarnings: true, // Warning about no model set
		},
		{
			name:     "nil settings",
			settings: nil,
			wantErr:  true,
			wantWarnings: false,
		},
		{
			name: "settings with env vars",
			settings: &config.Settings{
				Model: "sonnet",
				Env:   map[string]string{"API_KEY": "test"},
			},
			wantErr:  false,
			wantWarnings: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckSettings(tt.settings)

			if (result.Error != nil) != tt.wantErr {
				t.Errorf("CheckSettings() error = %v, wantErr %v", result.Error, tt.wantErr)
			}

			if tt.wantWarnings && len(result.Warnings) == 0 {
				t.Error("CheckSettings() expected warnings but got none")
			}
		})
	}
}

func TestCheckModel(t *testing.T) {
	tests := []struct {
		name         string
		model        string
		wantValid    bool
		wantWarnings bool
	}{
		{
			name:      "opus",
			model:     "opus",
			wantValid: true,
		},
		{
			name:      "sonnet",
			model:     "sonnet",
			wantValid: true,
		},
		{
			name:      "haiku",
			model:     "haiku",
			wantValid: true,
		},
		{
			name:         "empty model",
			model:        "",
			wantValid:    true,
			wantWarnings: true,
		},
		{
			name:         "custom model",
			model:        "custom-model-123",
			wantValid:    true,
			wantWarnings: true, // Warning that it's not a known model
		},
		{
			name:      "claude-3 opus full name",
			model:     "claude-3-opus-20240229",
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckModel(tt.model)

			if result.IsValid != tt.wantValid {
				t.Errorf("CheckModel() IsValid = %v, want %v", result.IsValid, tt.wantValid)
			}

			if tt.wantWarnings && len(result.Warnings) == 0 {
				t.Error("Expected warnings but got none")
			}
		})
	}
}

func TestCheckPermissions(t *testing.T) {
	tests := []struct {
		name      string
		perms     *config.Permissions
		wantValid bool
		wantWarnings bool
	}{
		{
			name:      "nil permissions",
			perms:     nil,
			wantValid: true,
		},
		{
			name:      "empty permissions",
			perms:     &config.Permissions{},
			wantValid: true,
		},
		{
			name: "valid allow list",
			perms: &config.Permissions{
				Allow: []string{"WebSearch", "Bash"},
			},
			wantValid: true,
		},
		{
			name: "wildcard",
			perms: &config.Permissions{
				Allow: []string{"*"},
			},
			wantValid:    true,
			wantWarnings: true, // Warning about allowing all tools
		},
		{
			name: "both allow and deny",
			perms: &config.Permissions{
				Allow: []string{"Read"},
				Deny:  []string{"Write"},
			},
			wantValid:    true,
			wantWarnings: true, // Warning about having both
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPermissions(tt.perms)

			if result.IsValid != tt.wantValid {
				t.Errorf("CheckPermissions() IsValid = %v, want %v", result.IsValid, tt.wantValid)
			}

			if tt.wantWarnings && len(result.Warnings) == 0 {
				t.Error("Expected warnings but got none")
			}
		})
	}
}

func TestCheckEnvVars(t *testing.T) {
	tests := []struct {
		name         string
		env          map[string]string
		wantWarnings bool
	}{
		{
			name: "empty env",
			env:  map[string]string{},
		},
		{
			name: "valid env vars",
			env: map[string]string{
				"API_KEY":  "test",
				"BASE_URL": "https://example.com",
			},
		},
		{
			name: "env with empty value",
			env: map[string]string{
				"EMPTY_VAR": "",
			},
			wantWarnings: true,
		},
		{
			name: "api key set",
			env: map[string]string{
				"ANTHROPIC_API_KEY": "sk-test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckEnvVars(tt.env)

			if !result.IsValid {
				t.Error("CheckEnvVars() should always be valid")
			}

			if tt.wantWarnings && len(result.Warnings) == 0 {
				t.Error("Expected warnings but got none")
			}
		})
	}
}

func TestCheckProfile(t *testing.T) {
	// Create a valid settings object
	validSettings := &config.Settings{
		Model: "opus",
		Env: map[string]string{
			"API_KEY": "test",
		},
	}

	result := CheckProfile("test-profile", validSettings, "# Instructions")

	if !result.IsHealthy() {
		t.Error("CheckProfile() should be healthy for valid profile")
	}

	if result.Profile != "test-profile" {
		t.Errorf("Profile = %q, want %q", result.Profile, "test-profile")
	}

	// Check that sub-checks were run
	if result.Settings.IsValid == false {
		t.Error("Settings check should be valid")
	}

	if result.Model.IsValid == false {
		t.Error("Model check should be valid")
	}
}

func TestCheckProfileNilSettings(t *testing.T) {
	result := CheckProfile("test", nil, "")

	if result.IsHealthy() {
		t.Error("CheckProfile() should not be healthy with nil settings")
	}

	if result.Overall.Error == nil {
		t.Error("Expected error for nil settings")
	}
}

func TestHealthResult(t *testing.T) {
	result := HealthResult{
		IsValid: true,
		Warnings: []string{"Warning 1", "Warning 2"},
		Error: nil,
	}

	if !result.IsHealthy() {
		t.Error("IsHealthy() should be true when IsValid and no Error")
	}

	if !result.HasWarnings() {
		t.Error("HasWarnings() should be true")
	}

	// Test with error
	resultWithError := HealthResult{
		IsValid: false,
		Error: &HealthError{Message: "test error"},
	}

	if resultWithError.IsHealthy() {
		t.Error("IsHealthy() should be false with error")
	}
}

func TestProfileHealthReport(t *testing.T) {
	report := ProfileHealthReport{
		Profile: "test",
		Overall: HealthResult{
			IsValid: true,
		},
		Settings: HealthResult{
			IsValid: true,
		},
		Model: HealthResult{
			IsValid:  true,
			Warnings: []string{"Using custom model"},
		},
	}

	if !report.IsHealthy() {
		t.Error("IsHealthy() should be true when overall is healthy")
	}

	if report.TotalWarnings() != 1 {
		t.Errorf("TotalWarnings() = %d, want 1", report.TotalWarnings())
	}

	// Test summary
	summary := report.Summary()
	if summary != "Healthy (with warnings)" {
		t.Errorf("Summary() = %q, want %q", summary, "Healthy (with warnings)")
	}

	// Test with errors
	unhealthyReport := ProfileHealthReport{
		Overall: HealthResult{
			IsValid: false,
			Error:   &HealthError{Message: "error"},
		},
	}

	if unhealthyReport.IsHealthy() {
		t.Error("IsHealthy() should be false with errors")
	}

	summary = unhealthyReport.Summary()
	if summary != "Unhealthy" {
		t.Errorf("Summary() = %q, want %q", summary, "Unhealthy")
	}
}

func TestKnownModels(t *testing.T) {
	knownModels := []string{"opus", "sonnet", "haiku"}

	for _, model := range knownModels {
		if !isKnownModel(model) {
			t.Errorf("isKnownModel(%q) should be true", model)
		}
	}

	if isKnownModel("unknown-model") {
		t.Error("isKnownModel('unknown-model') should be false")
	}
}
