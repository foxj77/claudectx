package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Settings represents the structure of settings.json.
// Known fields (Env, Model, Permissions) are typed for convenient access.
// Any additional fields written by Claude Code are preserved transparently
// in extras so that a LoadSettings→SaveSettings roundtrip never destroys data.
type Settings struct {
	Env         map[string]string `json:"env,omitempty"`
	Model       string            `json:"model,omitempty"`
	Permissions *Permissions      `json:"permissions,omitempty"`

	// extras holds any JSON keys not explicitly modelled above.
	// This prevents claudectx from silently destroying fields such as
	// effortLevel, autoDreamEnabled, skipDangerousModePermissionPrompt,
	// or future plugin-tracking fields that Claude Code may add.
	extras map[string]json.RawMessage
}

// UnmarshalJSON implements json.Unmarshaler for Settings.
// It populates the typed fields and stores all unrecognised keys in extras.
func (s *Settings) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	known := map[string]bool{"env": true, "model": true, "permissions": true}
	s.extras = make(map[string]json.RawMessage)
	for k, v := range raw {
		if !known[k] {
			s.extras[k] = v
		}
	}

	if v, ok := raw["model"]; ok {
		if err := json.Unmarshal(v, &s.Model); err != nil {
			return fmt.Errorf("parsing model: %w", err)
		}
	}
	if v, ok := raw["env"]; ok {
		if err := json.Unmarshal(v, &s.Env); err != nil {
			return fmt.Errorf("parsing env: %w", err)
		}
	}
	if v, ok := raw["permissions"]; ok {
		s.Permissions = &Permissions{}
		if err := json.Unmarshal(v, s.Permissions); err != nil {
			return fmt.Errorf("parsing permissions: %w", err)
		}
	}
	return nil
}

// MarshalJSON implements json.Marshaler for Settings.
// It emits all extras first (lower priority), then the typed fields which take
// precedence, ensuring known fields are always up-to-date on disk.
func (s *Settings) MarshalJSON() ([]byte, error) {
	out := make(map[string]json.RawMessage, len(s.extras)+3)
	for k, v := range s.extras {
		out[k] = v
	}
	if s.Model != "" {
		b, err := json.Marshal(s.Model)
		if err != nil {
			return nil, err
		}
		out["model"] = b
	}
	if len(s.Env) > 0 {
		b, err := json.Marshal(s.Env)
		if err != nil {
			return nil, err
		}
		out["env"] = b
	}
	if s.Permissions != nil {
		b, err := json.Marshal(s.Permissions)
		if err != nil {
			return nil, err
		}
		out["permissions"] = b
	}
	return json.Marshal(out)
}

// Permissions represents the permissions section of settings.
// Unknown sub-fields (e.g. defaultMode) are preserved via the same extras
// mechanism used by Settings.
type Permissions struct {
	Allow []string `json:"allow,omitempty"`
	Deny  []string `json:"deny,omitempty"`

	extras map[string]json.RawMessage
}

// UnmarshalJSON implements json.Unmarshaler for Permissions.
func (p *Permissions) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	known := map[string]bool{"allow": true, "deny": true}
	p.extras = make(map[string]json.RawMessage)
	for k, v := range raw {
		if !known[k] {
			p.extras[k] = v
		}
	}

	if v, ok := raw["allow"]; ok {
		if err := json.Unmarshal(v, &p.Allow); err != nil {
			return fmt.Errorf("parsing allow: %w", err)
		}
	}
	if v, ok := raw["deny"]; ok {
		if err := json.Unmarshal(v, &p.Deny); err != nil {
			return fmt.Errorf("parsing deny: %w", err)
		}
	}
	return nil
}

// MarshalJSON implements json.Marshaler for Permissions.
func (p *Permissions) MarshalJSON() ([]byte, error) {
	out := make(map[string]json.RawMessage, len(p.extras)+2)
	for k, v := range p.extras {
		out[k] = v
	}
	if len(p.Allow) > 0 {
		b, err := json.Marshal(p.Allow)
		if err != nil {
			return nil, err
		}
		out["allow"] = b
	}
	if len(p.Deny) > 0 {
		b, err := json.Marshal(p.Deny)
		if err != nil {
			return nil, err
		}
		out["deny"] = b
	}
	return json.Marshal(out)
}

// LoadSettings reads and parses a settings.json file
func LoadSettings(path string) (*Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings Settings
	err = json.Unmarshal(data, &settings)
	if err != nil {
		return nil, fmt.Errorf("failed to parse settings JSON: %w", err)
	}

	return &settings, nil
}

// LoadSettingsOrEmpty loads settings from a file, or returns empty settings if file doesn't exist
func LoadSettingsOrEmpty(path string) *Settings {
	settings, err := LoadSettings(path)
	if err != nil {
		// Return empty settings if file doesn't exist or can't be read
		return &Settings{
			Env: make(map[string]string),
		}
	}
	return settings
}

// SaveSettings writes settings to a JSON file with formatting
func SaveSettings(path string, settings *Settings) error {
	// Marshal with indentation for readability
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Add newline at end of file (common convention)
	data = append(data, '\n')

	// Write to file
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Sync to ensure data is written to disk
	err = destFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
