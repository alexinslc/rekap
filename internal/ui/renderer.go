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
	// Color palette matching fang's aesthetic
	primaryColor   = lipgloss.Color("13")  // Bright magenta/pink
	secondaryColor = lipgloss.Color("14")  // Cyan
	accentColor    = lipgloss.Color("11")  // Bright yellow
	successColor   = lipgloss.Color("10")  // Bright green
	warningColor   = lipgloss.Color("9")   // Bright red
	mutedColor     = lipgloss.Color("240") // Darker gray
	textColor      = lipgloss.Color("255") // White

	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginTop(1).
			MarginBottom(1)

	// subtitleStyle is currently unused but reserved for future use
	_ = lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true)

	dataStyle = lipgloss.NewStyle().
			Foreground(textColor)

	labelStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	// valueStyle is currently unused but reserved for future use
	_ = lipgloss.NewStyle().
		Foreground(textColor)

	highlightStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	hintStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	dividerStyle = lipgloss.NewStyle().
			Foreground(mutedColor)
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
			// Print character without newline
			fmt.Print(string(r))
		}
		fmt.Println()
		return ""
	}
	return titleStyle.Render(text)
}

// RenderHeader renders a section header
func RenderHeader(text string) string {
	return headerStyle.Render(text)
}

// RenderDivider renders a visual divider
func RenderDivider() string {
	return dividerStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// RenderSummaryLine formats the main summary line with enhanced styling
func RenderSummaryLine(parts []string) string {
	if len(parts) == 0 {
		return ""
	}

	// Create a clean summary line with subtle styling
	content := strings.Join(parts, " • ")
	return labelStyle.Render(content)
}

// RenderDataPoint formats a single data point with icon and enhanced styling
func RenderDataPoint(icon, text string) string {
	return fmt.Sprintf("  %s  %s", icon, dataStyle.Render(text))
}

// RenderHighlight formats highlighted text with extra emphasis
func RenderHighlight(icon, text string) string {
	styledText := highlightStyle.Render(text)
	return fmt.Sprintf("  %s  %s", icon, styledText)
}

// RenderSuccess formats a success message
func RenderSuccess(text string) string {
	return successStyle.Render("✓ " + text)
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
