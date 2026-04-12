package validator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/johnfox/claudectx/internal/config"
)

func TestValidateJSONFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid JSON",
			content: `{"model": "opus"}`,
			wantErr: false,
		},
		{
			name:    "valid empty object",
			content: `{}`,
			wantErr: false,
		},
		{
			name:    "valid nested JSON",
			content: `{"env": {"KEY": "value"}, "model": "sonnet"}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON - missing quote",
			content: `{"model: "opus"}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON - trailing comma",
			content: `{"model": "opus",}`,
			wantErr: true,
		},
		{
			name:    "invalid JSON - not JSON",
			content: `this is not JSON`,
			wantErr: true,
		},
		{
			name:    "empty file",
			content: ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tmpDir, "test.json")
			os.WriteFile(testFile, []byte(tt.content), 0644)

			err := ValidateJSONFile(testFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJSONFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Clean up
			os.Remove(testFile)
		})
	}
}

func TestValidateJSONFileNonExistent(t *testing.T) {
	err := ValidateJSONFile("/tmp/nonexistent-file.json")
	if err == nil {
		t.Error("ValidateJSONFile() should fail for non-existent file")
	}
}

func TestValidateSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings *config.Settings
		wantErr  bool
	}{
		{
			name:     "valid settings",
			settings: &config.Settings{Model: "opus"},
			wantErr:  false,
		},
		{
			name:     "valid empty settings",
			settings: &config.Settings{},
			wantErr:  false,
		},
		{
			name:     "valid with env",
			settings: &config.Settings{Env: map[string]string{"KEY": "value"}},
			wantErr:  false,
		},
		{
			name:     "valid with permissions",
			settings: &config.Settings{Permissions: &config.Permissions{Allow: []string{"tool1"}}},
			wantErr:  false,
		},
		{
			name:     "nil settings",
			settings: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSettings(tt.settings)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSettings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateModel(t *testing.T) {
	tests := []struct {
		name    string
		model   string
		wantErr bool
	}{
		{"opus", "opus", false},
		{"sonnet", "sonnet", false},
		{"haiku", "haiku", false},
		{"empty is valid", "", false}, // Empty is OK
		{"custom model", "custom-model-123", false},
		{"with version", "claude-3-opus-20240229", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateModel(tt.model)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateModel(%q) error = %v, wantErr %v", tt.model, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePermissions(t *testing.T) {
	tests := []struct {
		name    string
		perms   *config.Permissions
		wantErr bool
	}{
		{
			name:    "nil permissions",
			perms:   nil,
			wantErr: false, // nil is valid
		},
		{
			name:    "empty permissions",
			perms:   &config.Permissions{},
			wantErr: false,
		},
		{
			name:    "with allow list",
			perms:   &config.Permissions{Allow: []string{"WebSearch", "Bash"}},
			wantErr: false,
		},
		{
			name:    "with deny list",
			perms:   &config.Permissions{Deny: []string{"WebFetch"}},
			wantErr: false,
		},
		{
			name:    "with both lists",
			perms:   &config.Permissions{Allow: []string{"Read"}, Deny: []string{"Write"}},
			wantErr: false,
		},
		{
			name:    "wildcard",
			perms:   &config.Permissions{Allow: []string{"*"}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePermissions(tt.perms)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePermissions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEnv(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr bool
	}{
		{
			name:    "nil env",
			env:     nil,
			wantErr: false,
		},
		{
			name:    "empty env",
			env:     map[string]string{},
			wantErr: false,
		},
		{
			name:    "valid env",
			env:     map[string]string{"API_KEY": "sk-123", "BASE_URL": "https://example.com"},
			wantErr: false,
		},
		{
			name:    "with empty value",
			env:     map[string]string{"KEY": ""},
			wantErr: false, // Empty values are OK
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnv(tt.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSettingsFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid settings file
	validPath := filepath.Join(tmpDir, "valid.json")
	settings := &config.Settings{Model: "opus"}
	config.SaveSettings(validPath, settings)

	// Should pass validation
	err := ValidateSettingsFile(validPath)
	if err != nil {
		t.Errorf("ValidateSettingsFile() failed for valid file: %v", err)
	}

	// Create an invalid JSON file
	invalidPath := filepath.Join(tmpDir, "invalid.json")
	os.WriteFile(invalidPath, []byte("{ invalid json }"), 0644)

	// Should fail validation
	err = ValidateSettingsFile(invalidPath)
	if err == nil {
		t.Error("ValidateSettingsFile() should fail for invalid JSON")
	}
}

func TestValidateClaudeMD(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid markdown",
			content: "# Instructions\n\nUse Python.",
			wantErr: false,
		},
		{
			name:    "empty content",
			content: "",
			wantErr: false, // Empty is valid
		},
		{
			name:    "just whitespace",
			content: "   \n\t  ",
			wantErr: false, // Whitespace is OK
		},
		{
			name:    "long content",
			content: string(make([]byte, 100000)),
			wantErr: false, // Large files are OK
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateClaudeMD(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateClaudeMD() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
