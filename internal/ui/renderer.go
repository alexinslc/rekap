package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var (
	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("6"))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	dataStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))

	highlightStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("3"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1"))

	hintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true)
)

// IsTTY returns true if stdout is a terminal
func IsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// RenderTitle renders the main title with optional animation
func RenderTitle(text string, animate bool) string {
	if animate && IsTTY() {
		// Simple typing effect
		for i, r := range text {
			if i > 0 {
				time.Sleep(30 * time.Millisecond)
			}
			fmt.Print(string(r))
		}
		fmt.Println()
		return ""
	}
	return titleStyle.Render(text)
}

// RenderSummaryLine formats the main summary line
func RenderSummaryLine(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	return dataStyle.Render(strings.Join(parts, " • "))
}

// RenderDataPoint formats a single data point with icon
func RenderDataPoint(icon, text string) string {
	return fmt.Sprintf("%s %s", icon, dataStyle.Render(text))
}

// RenderHighlight formats highlighted text (like focus streak)
func RenderHighlight(icon, text string) string {
	return fmt.Sprintf("%s %s", icon, highlightStyle.Render(text))
}

// RenderHint formats a hint message
func RenderHint(text string) string {
	return hintStyle.Render("💡 " + text)
}

// RenderError formats an error message
func RenderError(text string) string {
	return errorStyle.Render("✗ " + text)
}

// FormatDuration formats minutes into human-readable duration
func FormatDuration(minutes int) string {
	hours := minutes / 60
	mins := minutes % 60
	
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

// FormatDurationCompact formats minutes into compact duration (for summary)
func FormatDurationCompact(minutes int) string {
	hours := minutes / 60
	mins := minutes % 60
	
	if hours > 0 && mins > 0 {
		return fmt.Sprintf("%dh%dm", hours, mins)
	} else if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dm", mins)
}

// ClearScreen clears the terminal screen (if TTY)
func ClearScreen() {
	if IsTTY() {
		fmt.Print("\033[H\033[2J")
	}
}
