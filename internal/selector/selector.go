package selector

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// Option represents a selectable option
type Option struct {
	Label     string
	IsCurrent bool
}

// Select displays an interactive selector and returns the selected index
func Select(title string, options []Option) (int, error) {
	// Check if we're in a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return -1, fmt.Errorf("interactive mode requires a terminal")
	}

	// Put terminal in raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return -1, fmt.Errorf("failed to enter raw mode: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Find the current option as the default selection
	selected := 0
	for i, opt := range options {
		if opt.IsCurrent {
			selected = i
			break
		}
	}

	// Main selection loop
	for {
		// Clear screen and move cursor to top
		fmt.Fprint(os.Stderr, "\033[2J\033[H")

		// Print title
		if title != "" {
			fmt.Fprintf(os.Stderr, "\033[1m%s\033[0m\r\n\r\n", title)
		}

		// Print options
		for i, opt := range options {
			if i == selected {
				// Selected option - highlighted with arrow
				fmt.Fprintf(os.Stderr, "\033[36m❯ %s\033[0m", opt.Label)
				if opt.IsCurrent {
					fmt.Fprint(os.Stderr, " \033[2m(current)\033[0m")
				}
				fmt.Fprint(os.Stderr, "\r\n")
			} else {
				// Unselected option
				fmt.Fprintf(os.Stderr, "  %s", opt.Label)
				if opt.IsCurrent {
					fmt.Fprint(os.Stderr, " \033[2m(current)\033[0m")
				}
				fmt.Fprint(os.Stderr, "\r\n")
			}
		}

		// Print help text
		fmt.Fprint(os.Stderr, "\r\n\033[2mUse ↑/↓ to navigate, Enter to select, Esc/Ctrl+C to cancel\033[0m")

		// Read single byte
		buf := make([]byte, 3)
		n, err := os.Stdin.Read(buf)
		if err != nil {
			return -1, fmt.Errorf("failed to read input: %w", err)
		}

		// Handle input
		if n == 1 {
			switch buf[0] {
			case 13: // Enter
				fmt.Fprint(os.Stderr, "\033[2J\033[H") // Clear screen
				return selected, nil
			case 3: // Ctrl+C
				fmt.Fprint(os.Stderr, "\033[2J\033[H") // Clear screen
				return -1, fmt.Errorf("cancelled")
			case 27: // ESC
				fmt.Fprint(os.Stderr, "\033[2J\033[H") // Clear screen
				return -1, fmt.Errorf("cancelled")
			}
		} else if n == 3 && buf[0] == 27 && buf[1] == 91 {
			// Arrow keys
			switch buf[2] {
			case 65: // Up arrow
				if selected > 0 {
					selected--
				}
			case 66: // Down arrow
				if selected < len(options)-1 {
					selected++
				}
			}
		}
	}
}
