package selector

import (
	"testing"
)

func TestOption(t *testing.T) {
	tests := []struct {
		name      string
		opt       Option
		wantLabel string
		wantCur   bool
	}{
		{
			name:      "simple option",
			opt:       Option{Label: "test", IsCurrent: false},
			wantLabel: "test",
			wantCur:   false,
		},
		{
			name:      "current option",
			opt:       Option{Label: "work", IsCurrent: true},
			wantLabel: "work",
			wantCur:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.opt.Label != tt.wantLabel {
				t.Errorf("Label = %q, want %q", tt.opt.Label, tt.wantLabel)
			}
			if tt.opt.IsCurrent != tt.wantCur {
				t.Errorf("IsCurrent = %v, want %v", tt.opt.IsCurrent, tt.wantCur)
			}
		})
	}
}

func TestSelectNonInteractive(t *testing.T) {
	// Test that Select returns error when not in a terminal
	// This test is primarily for documentation - actual interactive
	// testing must be done manually
	options := []Option{
		{Label: "option1", IsCurrent: false},
		{Label: "option2", IsCurrent: true},
	}

	// In test environment (non-TTY), Select should return an error
	_, err := Select("Test", options)
	if err == nil {
		t.Skip("Test is running in a TTY, skipping non-interactive test")
	}

	// We expect an error about requiring a terminal
	if err.Error() != "interactive mode requires a terminal" {
		t.Errorf("Expected terminal error, got: %v", err)
	}
}
