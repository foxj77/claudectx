package main

import (
	"fmt"
	"os"

	"github.com/johnfox/claudectx/cmd"
	"github.com/johnfox/claudectx/internal/store"
)

const version = "0.1.0"

func main() {
	// Initialize store
	s, err := store.NewStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to initialize claudectx: %v\n", err)
		os.Exit(1)
	}

	// Parse arguments
	if len(os.Args) < 2 {
		// Default action: list profiles
		if err := cmd.ListProfiles(s); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	arg := os.Args[1]

	switch arg {
	case "-h", "--help":
		printHelp()

	case "-v", "--version":
		fmt.Printf("claudectx version %s\n", version)

	case "-c", "--current":
		if err := cmd.ShowCurrent(s); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "-":
		if err := cmd.TogglePrevious(s); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "-n":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: profile name required")
			fmt.Fprintln(os.Stderr, "Usage: claudectx -n <name>")
			os.Exit(1)
		}
		if err := cmd.CreateProfile(s, os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "-d":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: profile name required")
			fmt.Fprintln(os.Stderr, "Usage: claudectx -d <name>")
			os.Exit(1)
		}
		if err := cmd.DeleteProfile(s, os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	default:
		// Assume it's a profile name to switch to
		if err := cmd.SwitchProfile(s, arg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func printHelp() {
	help := `claudectx - Fast way to switch between Claude Code configuration profiles

USAGE:
  claudectx                List all profiles
  claudectx <NAME>         Switch to profile
  claudectx -              Switch to previous profile
  claudectx -c, --current  Show current profile
  claudectx -n <NAME>      Create new profile from current config
  claudectx -d <NAME>      Delete profile
  claudectx -h, --help     Show this help
  claudectx -v, --version  Show version

EXAMPLES:
  claudectx work           Switch to 'work' profile
  claudectx -              Toggle between current and previous profile
  claudectx -n personal    Create 'personal' profile from current settings
  claudectx -d old-work    Delete 'old-work' profile

WHAT CLAUDECTX MANAGES:
  - ~/.claude/settings.json    User-level settings
  - ~/.claude/CLAUDE.md        Global instructions

Profiles are stored in ~/.claude/profiles/

Inspired by kubectx - https://github.com/ahmetb/kubectx
`
	fmt.Print(help)
}
