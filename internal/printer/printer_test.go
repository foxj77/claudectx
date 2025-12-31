package printer

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestColorEnabled(t *testing.T) {
	// Save original
	originalNoColor := os.Getenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", originalNoColor)

	// Test with NO_COLOR unset
	os.Unsetenv("NO_COLOR")
	if !ColorEnabled() {
		t.Error("ColorEnabled() should return true when NO_COLOR is not set")
	}

	// Test with NO_COLOR set
	os.Setenv("NO_COLOR", "1")
	if ColorEnabled() {
		t.Error("ColorEnabled() should return false when NO_COLOR is set")
	}
}

func TestColorize(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		color Color
		want  string
	}{
		{
			name:  "green text",
			text:  "success",
			color: Green,
			want:  "\033[32msuccess\033[0m",
		},
		{
			name:  "red text",
			text:  "error",
			color: Red,
			want:  "\033[31merror\033[0m",
		},
		{
			name:  "yellow text",
			text:  "warning",
			color: Yellow,
			want:  "\033[33mwarning\033[0m",
		},
		{
			name:  "blue text",
			text:  "info",
			color: Blue,
			want:  "\033[34minfo\033[0m",
		},
		{
			name:  "cyan text",
			text:  "current",
			color:  Cyan,
			want:  "\033[36mcurrent\033[0m",
		},
	}

	// Enable colors for testing
	originalNoColor := os.Getenv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", originalNoColor)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Colorize(tt.text, tt.color)
			if got != tt.want {
				t.Errorf("Colorize(%q, %v) = %q, want %q", tt.text, tt.color, got, tt.want)
			}
		})
	}
}

func TestColorizeWithNoColor(t *testing.T) {
	// Disable colors
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	result := Colorize("text", Green)
	if result != "text" {
		t.Errorf("Colorize with NO_COLOR should return plain text, got %q", result)
	}
}

func TestSuccess(t *testing.T) {
	originalNoColor := os.Getenv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", originalNoColor)

	output := captureOutput(func() {
		Success("Operation succeeded")
	})

	if !strings.Contains(output, "Operation succeeded") {
		t.Error("Success() should print the message")
	}

	// Should contain green color code when colors enabled
	if !strings.Contains(output, "\033[32m") {
		t.Error("Success() should use green color")
	}
}

func TestError(t *testing.T) {
	originalNoColor := os.Getenv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", originalNoColor)

	// Capture stderr instead of stdout
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	Error("Something went wrong")

	w.Close()
	os.Stderr = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Something went wrong") {
		t.Error("Error() should print the message")
	}

	// Should contain red color code
	if !strings.Contains(output, "\033[31m") {
		t.Error("Error() should use red color")
	}
}

func TestWarning(t *testing.T) {
	originalNoColor := os.Getenv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", originalNoColor)

	output := captureOutput(func() {
		Warning("This is a warning")
	})

	if !strings.Contains(output, "This is a warning") {
		t.Error("Warning() should print the message")
	}

	// Should contain yellow color code
	if !strings.Contains(output, "\033[33m") {
		t.Error("Warning() should use yellow color")
	}
}

func TestInfo(t *testing.T) {
	originalNoColor := os.Getenv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", originalNoColor)

	output := captureOutput(func() {
		Info("Information message")
	})

	if !strings.Contains(output, "Information message") {
		t.Error("Info() should print the message")
	}

	// Should contain blue color code
	if !strings.Contains(output, "\033[34m") {
		t.Error("Info() should use blue color")
	}
}

func TestBold(t *testing.T) {
	originalNoColor := os.Getenv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", originalNoColor)

	result := Bold("important")
	expected := "\033[1mimportant\033[0m"

	if result != expected {
		t.Errorf("Bold() = %q, want %q", result, expected)
	}

	// Test with NO_COLOR
	os.Setenv("NO_COLOR", "1")
	result = Bold("important")
	if result != "important" {
		t.Errorf("Bold() with NO_COLOR = %q, want %q", result, "important")
	}
}

func TestDim(t *testing.T) {
	originalNoColor := os.Getenv("NO_COLOR")
	os.Unsetenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", originalNoColor)

	result := Dim("subtle")
	expected := "\033[2msubtle\033[0m"

	if result != expected {
		t.Errorf("Dim() = %q, want %q", result, expected)
	}
}
