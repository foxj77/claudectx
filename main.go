package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		// Default action: list profiles
		listProfiles()
		return
	}

	arg := os.Args[1]

	switch arg {
	case "-h", "--help":
		printHelp()
	case "-v", "--version":
		fmt.Printf("claudectx version %s\n", version)
	case "-c", "--current":
		showCurrent()
	case "-":
		togglePrevious()
	case "-n":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: profile name required")
			os.Exit(1)
		}
		createProfile(os.Args[2])
	case "-d":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: profile name required")
			os.Exit(1)
		}
		deleteProfile(os.Args[2])
	default:
		// Check if it's a rename operation (name=oldname)
		// Check if it's a profile switch
		switchProfile(arg)
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

Inspired by kubectx - https://github.com/ahmetb/kubectx
`
	fmt.Print(help)
}

func listProfiles() {
	fmt.Println("claudectx: list profiles - TODO")
}

func showCurrent() {
	fmt.Println("claudectx: show current - TODO")
}

func togglePrevious() {
	fmt.Println("claudectx: toggle previous - TODO")
}

func createProfile(name string) {
	fmt.Printf("claudectx: create profile '%s' - TODO\n", name)
}

func deleteProfile(name string) {
	fmt.Printf("claudectx: delete profile '%s' - TODO\n", name)
}

func switchProfile(name string) {
	fmt.Printf("claudectx: switch to profile '%s' - TODO\n", name)
}
