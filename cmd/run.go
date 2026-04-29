package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/johnfox/claudectx/internal/mcpconfig"
	"github.com/johnfox/claudectx/internal/paths"
	"github.com/johnfox/claudectx/internal/printer"
	"github.com/johnfox/claudectx/internal/profile"
	"github.com/johnfox/claudectx/internal/store"
	"github.com/johnfox/claudectx/internal/validator"
)

// RunOptions holds the parsed arguments for the run command.
type RunOptions struct {
	ProfileName string
	ClaudeArgs  []string
	DryRun      bool
}

// RunResult holds output from a RunProfile call.
type RunResult struct {
	ProfileName   string
	GeneratedArgs []string
	TempDir       string
	ExitCode      int
}

// ParseRunArgs parses the slice of arguments following "claudectx run".
// Valid forms:
//
//	run <profile>
//	run --dry-run <profile>
//	run <profile> -- <claude args...>
//	run --dry-run <profile> -- <claude args...>
func ParseRunArgs(args []string) (RunOptions, error) {
	var opts RunOptions
	remaining := make([]string, 0, len(args))

	// First pass: extract --dry-run and find the -- separator.
	separatorIdx := -1
	for i, a := range args {
		if a == "--dry-run" {
			opts.DryRun = true
			continue
		}
		if a == "--" {
			separatorIdx = i
			break
		}
		remaining = append(remaining, a)
	}

	// Everything after -- goes to ClaudeArgs.
	if separatorIdx >= 0 {
		opts.ClaudeArgs = args[separatorIdx+1:]
	}

	// The first remaining non-flag token is the profile name.
	for _, a := range remaining {
		if strings.HasPrefix(a, "-") {
			return RunOptions{}, fmt.Errorf(
				"unknown claudectx run flag %q; use: claudectx run <profile> -- %s ...",
				a, a,
			)
		}
		if opts.ProfileName == "" {
			opts.ProfileName = a
		}
	}

	if opts.ProfileName == "" {
		return RunOptions{}, errors.New("profile name required\nUsage: claudectx run <name> [-- <claude args...>]")
	}

	return opts, nil
}

// RunProfile launches claude with the named profile's settings without
// modifying global claudectx state (settings.json, CLAUDE.md, claude.json,
// current/previous profile trackers).
//
// When opts.DryRun is true the function returns the generated command args
// without executing claude and without creating any temp files.
func RunProfile(s *store.Store, opts RunOptions) (RunResult, error) {
	result := RunResult{ProfileName: opts.ProfileName}

	if err := profile.ValidateProfileName(opts.ProfileName); err != nil {
		return result, fmt.Errorf("invalid profile name: %w", err)
	}

	if !s.Exists(opts.ProfileName) {
		return result, fmt.Errorf("profile %q does not exist", opts.ProfileName)
	}

	prof, err := s.Load(opts.ProfileName)
	if err != nil {
		return result, fmt.Errorf("failed to load profile: %w", err)
	}

	if err := prof.Validate(); err != nil {
		return result, fmt.Errorf("profile validation failed: %w", err)
	}

	if prof.Settings != nil {
		if err := validator.ValidateSettings(prof.Settings); err != nil {
			return result, fmt.Errorf("profile settings invalid: %w", err)
		}
	}

	// Build the argument list for claude.
	var claudeArgs []string

	// --settings always included; points directly at the profile's settings.json.
	settingsPath, err := paths.ProfileFile(opts.ProfileName, "settings.json")
	if err != nil {
		return result, fmt.Errorf("failed to resolve settings path: %w", err)
	}
	claudeArgs = append(claudeArgs, "--settings", settingsPath)

	// --append-system-prompt-file only when CLAUDE.md is non-empty.
	if strings.TrimSpace(prof.ClaudeMD) != "" {
		claudeMDPath, err := paths.ProfileFile(opts.ProfileName, "CLAUDE.md")
		if err != nil {
			return result, fmt.Errorf("failed to resolve CLAUDE.md path: %w", err)
		}
		claudeArgs = append(claudeArgs, "--append-system-prompt-file", claudeMDPath)
	}

	// --mcp-config + --strict-mcp-config only when profile has MCP servers.
	// In dry-run mode we compute the would-be path but do not create any files.
	var tempDir string
	if len(prof.MCPServers) > 0 {
		base, pathErr := paths.RunTempDir()
		if pathErr != nil {
			return result, fmt.Errorf("failed to resolve run temp dir: %w", pathErr)
		}
		runName := fmt.Sprintf("run-%d-%d", time.Now().UnixNano(), os.Getpid())
		mcpPath := filepath.Join(base, runName, "mcp.json")

		if !opts.DryRun {
			tempDir = filepath.Dir(mcpPath)
			if err := os.MkdirAll(tempDir, 0700); err != nil {
				return result, fmt.Errorf("failed to create temp dir for MCP config: %w", err)
			}
			result.TempDir = tempDir
			if err := mcpconfig.SaveClaudeMCPConfig(mcpPath, prof.MCPServers); err != nil {
				_ = os.RemoveAll(tempDir)
				return result, fmt.Errorf("failed to write MCP config: %w", err)
			}
		}

		claudeArgs = append(claudeArgs, "--mcp-config", mcpPath, "--strict-mcp-config")
	}

	// User pass-through args come last so they can override profile values.
	claudeArgs = append(claudeArgs, opts.ClaudeArgs...)

	result.GeneratedArgs = claudeArgs

	if opts.DryRun {
		printer.Info("claude %s", strings.Join(claudeArgs, " "))
		return result, nil
	}

	printer.Info("Running Claude with profile %q for this session only", opts.ProfileName)

	exitCode, err := execClaude(claudeArgs)
	result.ExitCode = exitCode

	// Best-effort cleanup of temp MCP config after claude exits.
	// MCP config is read at startup only, so deletion after exit is safe.
	if tempDir != "" {
		if cleanErr := os.RemoveAll(tempDir); cleanErr != nil {
			printer.Warning("failed to clean up temp run dir %q: %v", tempDir, cleanErr)
		}
	}

	if err != nil {
		return result, err
	}
	return result, nil
}

// execClaude runs the claude binary with the given args, inheriting stdio.
// Returns the child exit code and any exec-level error (e.g. binary not found).
func execClaude(args []string) (int, error) {
	cmd := exec.Command("claude", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), nil
		}
		if errors.Is(err, exec.ErrNotFound) {
			return 1, fmt.Errorf(
				"\"claude\" not found in PATH\n" +
					"Install Claude Code: https://claude.ai/code",
			)
		}
		return 1, fmt.Errorf("failed to run claude: %w", err)
	}
	return 0, nil
}

// createRunTempDir creates a unique temp directory under ~/.claude/.claudectx-run/.
func createRunTempDir() (string, error) {
	base, err := paths.RunTempDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(base, 0700); err != nil {
		return "", fmt.Errorf("failed to create run temp base dir: %w", err)
	}
	name := fmt.Sprintf("run-%d-%d", time.Now().UnixNano(), os.Getpid())
	dir := filepath.Join(base, name)
	if err := os.Mkdir(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create run temp dir: %w", err)
	}
	return dir, nil
}
