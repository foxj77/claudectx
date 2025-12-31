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

	case "export":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: profile name required")
			fmt.Fprintln(os.Stderr, "Usage: claudectx export <name> [output-file]")
			os.Exit(1)
		}
		profileName := os.Args[2]
		outputPath := ""
		if len(os.Args) > 3 {
			outputPath = os.Args[3]
		}
		if err := cmd.ExportProfile(s, profileName, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "import":
		inputPath := ""
		newName := ""
		if len(os.Args) > 2 {
			inputPath = os.Args[2]
		}
		if len(os.Args) > 3 {
			newName = os.Args[3]
		}
		if err := cmd.ImportProfile(s, inputPath, newName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "health":
		args := []string{}
		if len(os.Args) > 2 {
			args = os.Args[2:]
		}
		if err := cmd.Health(args); err != nil {
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
  claudectx                        List all profiles
  claudectx <NAME>                 Switch to profile
  claudectx -                      Switch to previous profile
  claudectx -c, --current          Show current profile
  claudectx -n <NAME>              Create new profile from current config
  claudectx -d <NAME>              Delete profile
  claudectx export <NAME> [FILE]   Export profile to JSON (stdout if no file)
  claudectx import [FILE] [NAME]   Import profile from JSON (stdin if no file)
  claudectx health [NAME]          Check profile health (current if no name given)
  claudectx -h, --help             Show this help
  claudectx -v, --version          Show version

EXAMPLES:
  claudectx work                   Switch to 'work' profile
  claudectx -                      Toggle between current and previous profile
  claudectx -n personal            Create 'personal' profile from current settings
  claudectx -d old-work            Delete 'old-work' profile
  claudectx export work work.json  Export 'work' profile to file
  claudectx export work            Export to stdout (for piping)
  claudectx import work.json       Import profile from file
  claudectx import work.json new   Import and rename to 'new'
  cat work.json | claudectx import Import from stdin
  claudectx health                 Check current profile health
  claudectx health work            Check 'work' profile health

WHAT CLAUDECTX MANAGES:
  - ~/.claude/settings.json    User-level settings
  - ~/.claude/CLAUDE.md        Global instructions
  - Automatic backups in ~/.claude/backups/

Profiles are stored in ~/.claude/profiles/

Inspired by kubectx - https://github.com/ahmetb/kubectx
`
	fmt.Print(help)
}
