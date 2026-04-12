package printer

import (
	"fmt"
	"os"
)

// Color represents ANSI color codes
type Color string

const (
	// Color codes
	Red    Color = "\033[31m"
	Green  Color = "\033[32m"
	Yellow Color = "\033[33m"
	Blue   Color = "\033[34m"
	Cyan   Color = "\033[36m"
	Reset  Color = "\033[0m"

	// Style codes (not used directly, use BoldStyle/DimStyle functions)
)

// ColorEnabled checks if color output is enabled
func ColorEnabled() bool {
	// Respect NO_COLOR environment variable
	// https://no-color.org/
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	return true
}

// Colorize wraps text in color codes if colors are enabled
func Colorize(text string, color Color) string {
	if !ColorEnabled() {
		return text
	}

	return string(color) + text + string(Reset)
}

// BoldStyle applies bold styling to text
func BoldStyle(text string) string {
	if !ColorEnabled() {
		return text
	}

	return "\033[1m" + text + string(Reset)
}

// DimStyle applies dim styling to text
func DimStyle(text string) string {
	if !ColorEnabled() {
		return text
	}

	return "\033[2m" + text + string(Reset)
}

// Success prints a success message in green
func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(Colorize(msg, Green))
}

// Error prints an error message in red to stderr
func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, Colorize(msg, Red))
}

// Warning prints a warning message in yellow
func Warning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(Colorize(msg, Yellow))
}

// Info prints an info message in blue
func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(Colorize(msg, Blue))
}

// Bold is a convenience wrapper for BoldStyle
func Bold(text string) string {
	return BoldStyle(text)
}

// Dim is a convenience wrapper for DimStyle
func Dim(text string) string {
	return DimStyle(text)
}

// PrintList prints a list of items with optional highlighting
func PrintList(items []string, highlightItems map[string]string) {
	for _, item := range items {
		if label, highlighted := highlightItems[item]; highlighted {
			// Print highlighted item with label
			fmt.Printf("%s %s\n", Colorize(item, Cyan), DimStyle(label))
		} else {
			// Print regular item
			fmt.Println(item)
		}
	}
}
