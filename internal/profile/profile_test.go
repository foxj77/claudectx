package profile

import (
	"testing"
	"time"

	"github.com/johnfox/claudectx/internal/config"
)

func TestNewProfile(t *testing.T) {
	name := "test-profile"
	profile := NewProfile(name)

	if profile.Name != name {
		t.Errorf("Name = %q, want %q", profile.Name, name)
	}

	if profile.Settings == nil {
		t.Error("Settings should be initialized")
	}

	if profile.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}

	if profile.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}

	// CreatedAt and UpdatedAt should be approximately equal for new profile
	diff := profile.UpdatedAt.Sub(profile.CreatedAt)
	if diff < 0 || diff > time.Second {
		t.Errorf("CreatedAt and UpdatedAt should be very close, got diff %v", diff)
	}
}

func TestProfileFromCurrent(t *testing.T) {
	settings := &config.Settings{
		Model: "opus",
		Env: map[string]string{
			"KEY": "value",
		},
	}

	profile := ProfileFromCurrent("my-profile", settings, "# Claude Instructions")

	if profile.Name != "my-profile" {
		t.Errorf("Name = %q, want %q", profile.Name, "my-profile")
	}

	if profile.Settings.Model != "opus" {
		t.Errorf("Settings.Model = %q, want %q", profile.Settings.Model, "opus")
	}

	if profile.ClaudeMD != "# Claude Instructions" {
		t.Errorf("ClaudeMD = %q, want %q", profile.ClaudeMD, "# Claude Instructions")
	}

	if profile.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestProfileTouch(t *testing.T) {
	profile := NewProfile("test")

	// Wait a tiny bit to ensure time difference
	time.Sleep(10 * time.Millisecond)

	originalUpdated := profile.UpdatedAt

	profile.Touch()

	if !profile.UpdatedAt.After(originalUpdated) {
		t.Error("Touch() should update UpdatedAt to a later time")
	}
}

func TestProfileValidate(t *testing.T) {
	tests := []struct {
		name    string
		profile *Profile
		wantErr bool
	}{
		{
			name: "valid profile",
			profile: &Profile{
				Name:     "valid",
				Settings: &config.Settings{Model: "opus"},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			profile: &Profile{
				Name:     "",
				Settings: &config.Settings{},
			},
			wantErr: true,
		},
		{
			name: "nil settings",
			profile: &Profile{
				Name:     "test",
				Settings: nil,
			},
			wantErr: true,
		},
		{
			name: "minimal valid profile",
			profile: &Profile{
				Name:     "minimal",
				Settings: &config.Settings{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.profile.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProfileIsEmpty(t *testing.T) {
	tests := []struct {
		name    string
		profile *Profile
		want    bool
	}{
		{
			name: "empty profile",
			profile: &Profile{
				Name:     "empty",
				Settings: &config.Settings{},
			},
			want: true,
		},
		{
			name: "profile with model",
			profile: &Profile{
				Name:     "with-model",
				Settings: &config.Settings{Model: "opus"},
			},
			want: false,
		},
		{
			name: "profile with env",
			profile: &Profile{
				Name: "with-env",
				Settings: &config.Settings{
					Env: map[string]string{"KEY": "value"},
				},
			},
			want: false,
		},
		{
			name: "profile with claude md",
			profile: &Profile{
				Name:     "with-md",
				Settings: &config.Settings{},
				ClaudeMD: "# Instructions",
			},
			want: false,
		},
		{
			name: "profile with permissions",
			profile: &Profile{
				Name: "with-perms",
				Settings: &config.Settings{
					Permissions: &config.Permissions{
						Allow: []string{"WebSearch"},
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.profile.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateProfileName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid simple", "work", false},
		{"valid with dash", "work-2025", false},
		{"valid with underscore", "work_prod", false},
		{"valid with dot", "work.prod", false},
		{"empty string", "", true},
		{"with slash", "work/prod", true},
		{"with backslash", "work\\prod", true},
		{"with space", "work prod", true},
		{"just dots", "..", true},
		{"just dot", ".", true},
		{"starts with dash", "-work", false}, // Actually valid
		{"ends with dash", "work-", false},   // Actually valid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProfileName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateProfileName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
