package cmd

import (
	"reflect"
	"testing"
)

// TestParseRunArgs_ProfileOnly verifies the simplest valid form.
func TestParseRunArgs_ProfileOnly(t *testing.T) {
	opts, err := ParseRunArgs([]string{"work"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.ProfileName != "work" {
		t.Errorf("ProfileName = %q, want %q", opts.ProfileName, "work")
	}
	if opts.DryRun {
		t.Error("DryRun should be false")
	}
	if len(opts.ClaudeArgs) != 0 {
		t.Errorf("ClaudeArgs should be empty, got %v", opts.ClaudeArgs)
	}
}

// TestParseRunArgs_DryRunBeforeProfile verifies --dry-run can precede the profile name.
func TestParseRunArgs_DryRunBeforeProfile(t *testing.T) {
	opts, err := ParseRunArgs([]string{"--dry-run", "work"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.DryRun {
		t.Error("DryRun should be true")
	}
	if opts.ProfileName != "work" {
		t.Errorf("ProfileName = %q, want %q", opts.ProfileName, "work")
	}
}

// TestParseRunArgs_DryRunAfterProfile verifies --dry-run can follow the profile name.
func TestParseRunArgs_DryRunAfterProfile(t *testing.T) {
	opts, err := ParseRunArgs([]string{"work", "--dry-run"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.DryRun {
		t.Error("DryRun should be true")
	}
	if opts.ProfileName != "work" {
		t.Errorf("ProfileName = %q, want %q", opts.ProfileName, "work")
	}
}

// TestParseRunArgs_PassThroughArgs verifies args after -- are captured verbatim.
func TestParseRunArgs_PassThroughArgs(t *testing.T) {
	opts, err := ParseRunArgs([]string{"work", "--", "--model", "opus"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.ProfileName != "work" {
		t.Errorf("ProfileName = %q, want %q", opts.ProfileName, "work")
	}
	want := []string{"--model", "opus"}
	if !reflect.DeepEqual(opts.ClaudeArgs, want) {
		t.Errorf("ClaudeArgs = %v, want %v", opts.ClaudeArgs, want)
	}
}

// TestParseRunArgs_DryRunWithPassThrough verifies all three parts together.
func TestParseRunArgs_DryRunWithPassThrough(t *testing.T) {
	opts, err := ParseRunArgs([]string{"--dry-run", "work", "--", "--model", "opus", "--permission-mode", "plan"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.DryRun {
		t.Error("DryRun should be true")
	}
	if opts.ProfileName != "work" {
		t.Errorf("ProfileName = %q, want %q", opts.ProfileName, "work")
	}
	want := []string{"--model", "opus", "--permission-mode", "plan"}
	if !reflect.DeepEqual(opts.ClaudeArgs, want) {
		t.Errorf("ClaudeArgs = %v, want %v", opts.ClaudeArgs, want)
	}
}

// TestParseRunArgs_MissingProfileName verifies an error when no profile is given.
func TestParseRunArgs_MissingProfileName(t *testing.T) {
	_, err := ParseRunArgs([]string{})
	if err == nil {
		t.Fatal("expected error for missing profile name")
	}
}

// TestParseRunArgs_OnlyDryRunNoProfile verifies an error when --dry-run is given but no profile.
func TestParseRunArgs_OnlyDryRunNoProfile(t *testing.T) {
	_, err := ParseRunArgs([]string{"--dry-run"})
	if err == nil {
		t.Fatal("expected error for missing profile name after --dry-run")
	}
}

// TestParseRunArgs_UnknownFlagBeforeSeparatorIsError verifies that unknown flags
// before -- produce a clear error with a hint to use --.
func TestParseRunArgs_UnknownFlagBeforeSeparatorIsError(t *testing.T) {
	_, err := ParseRunArgs([]string{"work", "--model", "opus"})
	if err == nil {
		t.Fatal("expected error for unknown flag before -- separator")
	}
}

// TestParseRunArgs_EmptyPassThroughAfterSeparator verifies -- with nothing after is valid.
func TestParseRunArgs_EmptyPassThroughAfterSeparator(t *testing.T) {
	opts, err := ParseRunArgs([]string{"work", "--"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.ProfileName != "work" {
		t.Errorf("ProfileName = %q, want %q", opts.ProfileName, "work")
	}
	if len(opts.ClaudeArgs) != 0 {
		t.Errorf("ClaudeArgs should be empty after bare --, got %v", opts.ClaudeArgs)
	}
}

// TestParseRunArgs_PassThroughPreservesOrder verifies pass-through arg order is maintained.
func TestParseRunArgs_PassThroughPreservesOrder(t *testing.T) {
	opts, err := ParseRunArgs([]string{"personal", "--", "-p", "say hello", "--output-format", "json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"-p", "say hello", "--output-format", "json"}
	if !reflect.DeepEqual(opts.ClaudeArgs, want) {
		t.Errorf("ClaudeArgs = %v, want %v", opts.ClaudeArgs, want)
	}
}
