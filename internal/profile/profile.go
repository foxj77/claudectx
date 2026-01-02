package profile

import (
	"errors"
	"strings"
	"time"

	"github.com/johnfox/claudectx/internal/config"
	"github.com/johnfox/claudectx/internal/mcpconfig"
)

// Profile represents a complete Claude configuration profile
type Profile struct {
	Name       string
	Settings   *config.Settings
	ClaudeMD   string
	MCPServers mcpconfig.MCPServers
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// NewProfile creates a new empty profile with the given name
func NewProfile(name string) *Profile {
	now := time.Now()
	return &Profile{
		Name:       name,
		Settings:   &config.Settings{Env: make(map[string]string)},
		ClaudeMD:   "",
		MCPServers: make(mcpconfig.MCPServers),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// ProfileFromCurrent creates a profile from current configuration
func ProfileFromCurrent(name string, settings *config.Settings, claudeMD string, mcpServers mcpconfig.MCPServers) *Profile {
	now := time.Now()
	return &Profile{
		Name:       name,
		Settings:   settings,
		ClaudeMD:   claudeMD,
		MCPServers: mcpServers,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// Touch updates the UpdatedAt timestamp
func (p *Profile) Touch() {
	p.UpdatedAt = time.Now()
}

// Validate checks if the profile is valid
func (p *Profile) Validate() error {
	if p.Name == "" {
		return errors.New("profile name cannot be empty")
	}

	if p.Settings == nil {
		return errors.New("profile settings cannot be nil")
	}

	return nil
}

// IsEmpty returns true if the profile has no meaningful configuration
func (p *Profile) IsEmpty() bool {
	if p.Settings == nil {
		return true
	}

	// Check if settings has any content
	hasModel := p.Settings.Model != ""
	hasEnv := len(p.Settings.Env) > 0
	hasPermissions := p.Settings.Permissions != nil &&
		(len(p.Settings.Permissions.Allow) > 0 || len(p.Settings.Permissions.Deny) > 0)
	hasClaudeMD := strings.TrimSpace(p.ClaudeMD) != ""
	hasMCPServers := len(p.MCPServers) > 0

	return !hasModel && !hasEnv && !hasPermissions && !hasClaudeMD && !hasMCPServers
}

// ValidateProfileName checks if a profile name is valid
func ValidateProfileName(name string) error {
	if name == "" {
		return errors.New("profile name cannot be empty")
	}

	// Check for invalid characters (path separators, spaces, etc.)
	invalidChars := []string{"/", "\\", " ", "\t", "\n", "\r"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return errors.New("profile name cannot contain spaces or path separators")
		}
	}

	// Prevent . and .. which are special directory names
	if name == "." || name == ".." {
		return errors.New("profile name cannot be '.' or '..'")
	}

	return nil
}
