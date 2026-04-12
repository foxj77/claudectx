package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Settings represents the structure of settings.json.
// Every known Claude Code settings key is listed explicitly so that a
// load/save round-trip never silently drops data. Fields with complex or
// polymorphic schemas are stored as json.RawMessage and passed through
// verbatim. See https://code.claude.com/docs/en/settings for the full
// reference.
type Settings struct {
	Schema string `json:"$schema,omitempty"`

	// Scalar string settings
	Agent                       string `json:"agent,omitempty"`
	ApiKeyHelper                string `json:"apiKeyHelper,omitempty"`
	AutoMemoryDirectory         string `json:"autoMemoryDirectory,omitempty"`
	AutoUpdatesChannel          string `json:"autoUpdatesChannel,omitempty"`
	AwsAuthRefresh              string `json:"awsAuthRefresh,omitempty"`
	AwsCredentialExport         string `json:"awsCredentialExport,omitempty"`
	DefaultShell                string `json:"defaultShell,omitempty"`
	DisableAutoMode             string `json:"disableAutoMode,omitempty"`
	DisableDeepLinkRegistration string `json:"disableDeepLinkRegistration,omitempty"`
	EffortLevel                 string `json:"effortLevel,omitempty"`
	ForceLoginMethod            string `json:"forceLoginMethod,omitempty"`
	Language                    string `json:"language,omitempty"`
	Model                       string `json:"model,omitempty"`
	OtelHeadersHelper           string `json:"otelHeadersHelper,omitempty"`
	OutputStyle                 string `json:"outputStyle,omitempty"`
	PlansDirectory              string `json:"plansDirectory,omitempty"`
	PluginTrustMessage          string `json:"pluginTrustMessage,omitempty"`

	// Numeric settings
	CleanupPeriodDays  int      `json:"cleanupPeriodDays,omitempty"`
	FeedbackSurveyRate *float64 `json:"feedbackSurveyRate,omitempty"` // 0 means suppress; needs *float64

	// Boolean settings — default false; true enables the feature
	AllowManagedHooksOnly           bool `json:"allowManagedHooksOnly,omitempty"`
	AllowManagedMcpServersOnly      bool `json:"allowManagedMcpServersOnly,omitempty"`
	AllowManagedPermissionRulesOnly bool `json:"allowManagedPermissionRulesOnly,omitempty"`
	AlwaysThinkingEnabled           bool `json:"alwaysThinkingEnabled,omitempty"`
	ChannelsEnabled                 bool `json:"channelsEnabled,omitempty"`
	DisableAllHooks                 bool `json:"disableAllHooks,omitempty"`
	DisableSkillShellExecution      bool `json:"disableSkillShellExecution,omitempty"`
	EnableAllProjectMcpServers      bool `json:"enableAllProjectMcpServers,omitempty"`
	FastModePerSessionOptIn         bool `json:"fastModePerSessionOptIn,omitempty"`
	ForceRemoteSettingsRefresh      bool `json:"forceRemoteSettingsRefresh,omitempty"`
	PrefersReducedMotion            bool `json:"prefersReducedMotion,omitempty"`
	ShowClearContextOnPlanAccept    bool `json:"showClearContextOnPlanAccept,omitempty"`
	ShowThinkingSummaries           bool `json:"showThinkingSummaries,omitempty"`
	VoiceEnabled                    bool `json:"voiceEnabled,omitempty"`

	// Boolean settings — default true; *bool preserves an explicit false value
	IncludeCoAuthoredBy    *bool `json:"includeCoAuthoredBy,omitempty"`  // deprecated; default true
	IncludeGitInstructions *bool `json:"includeGitInstructions,omitempty"` // default true
	RespectGitignore       *bool `json:"respectGitignore,omitempty"`       // default true
	SpinnerTipsEnabled     *bool `json:"spinnerTipsEnabled,omitempty"`     // default true
	UseAutoModeDuringPlan  *bool `json:"useAutoModeDuringPlan,omitempty"`  // default true

	// Array settings
	AllowedHttpHookUrls    []string `json:"allowedHttpHookUrls,omitempty"`
	AvailableModels        []string `json:"availableModels,omitempty"`
	CompanyAnnouncements   []string `json:"companyAnnouncements,omitempty"`
	DisabledMcpjsonServers []string `json:"disabledMcpjsonServers,omitempty"`
	EnabledMcpjsonServers  []string `json:"enabledMcpjsonServers,omitempty"`
	HttpHookAllowedEnvVars []string `json:"httpHookAllowedEnvVars,omitempty"`

	// Map settings
	Env map[string]string `json:"env,omitempty"`

	// Structured settings with their own types
	Permissions *Permissions `json:"permissions,omitempty"`

	// Complex settings stored verbatim as raw JSON.
	// Using json.RawMessage avoids defining bespoke types for every nested
	// schema while still round-tripping the exact bytes from the source file.
	AllowedChannelPlugins   json.RawMessage `json:"allowedChannelPlugins,omitempty"`
	AllowedMcpServers       json.RawMessage `json:"allowedMcpServers,omitempty"`
	Attribution             json.RawMessage `json:"attribution,omitempty"`
	AutoMode                json.RawMessage `json:"autoMode,omitempty"`
	BlockedMarketplaces     json.RawMessage `json:"blockedMarketplaces,omitempty"`
	DeniedMcpServers        json.RawMessage `json:"deniedMcpServers,omitempty"`
	EnabledPlugins          json.RawMessage `json:"enabledPlugins,omitempty"`
	ExtraKnownMarketplaces  json.RawMessage `json:"extraKnownMarketplaces,omitempty"`
	FileSuggestion          json.RawMessage `json:"fileSuggestion,omitempty"`
	ForceLoginOrgUUID       json.RawMessage `json:"forceLoginOrgUUID,omitempty"` // string or []string
	Hooks                   json.RawMessage `json:"hooks,omitempty"`
	ModelOverrides          json.RawMessage `json:"modelOverrides,omitempty"`
	Sandbox                 json.RawMessage `json:"sandbox,omitempty"`
	SpinnerTipsOverride     json.RawMessage `json:"spinnerTipsOverride,omitempty"`
	SpinnerVerbs            json.RawMessage `json:"spinnerVerbs,omitempty"`
	StatusLine              json.RawMessage `json:"statusLine,omitempty"`
	StrictKnownMarketplaces json.RawMessage `json:"strictKnownMarketplaces,omitempty"`
	Worktree                json.RawMessage `json:"worktree,omitempty"`
}

// Permissions represents the permissions section of settings.json.
type Permissions struct {
	Allow                             []string `json:"allow,omitempty"`
	Ask                               []string `json:"ask,omitempty"`
	Deny                              []string `json:"deny,omitempty"`
	AdditionalDirectories             []string `json:"additionalDirectories,omitempty"`
	DefaultMode                       string   `json:"defaultMode,omitempty"`
	DisableBypassPermissionsMode      string   `json:"disableBypassPermissionsMode,omitempty"`
	SkipDangerousModePermissionPrompt bool     `json:"skipDangerousModePermissionPrompt,omitempty"`
}

// LoadSettings reads and parses a settings.json file.
func LoadSettings(path string) (*Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings Settings
	if err = json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings JSON: %w", err)
	}

	return &settings, nil
}

// LoadSettingsOrEmpty loads settings from a file, or returns empty settings if
// the file does not exist or cannot be read.
func LoadSettingsOrEmpty(path string) *Settings {
	settings, err := LoadSettings(path)
	if err != nil {
		return &Settings{
			Env: make(map[string]string),
		}
	}
	return settings
}

// SaveSettings writes settings to a JSON file with formatting.
func SaveSettings(path string, settings *Settings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	data = append(data, '\n')

	if err = os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

// CopyFile copies a file from src to dst.
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

	if _, err = io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	if err = destFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}

// FileExists checks if a file exists.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
