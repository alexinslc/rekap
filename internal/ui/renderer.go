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
	// Color palette - vibrant and eye-catching
	primaryColor   = lipgloss.Color("13") // Bright magenta
	secondaryColor = lipgloss.Color("14") // Cyan
	accentColor    = lipgloss.Color("11") // Bright yellow
	successColor   = lipgloss.Color("10") // Bright green
	warningColor   = lipgloss.Color("9")  // Bright red
	mutedColor     = lipgloss.Color("8")  // Gray
	textColor      = lipgloss.Color("7")  // White

	// Styles with enhanced visual appeal
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(secondaryColor)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	dataStyle = lipgloss.NewStyle().
			Foreground(textColor)

	labelStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(accentColor)

	highlightStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor).
			Background(lipgloss.Color("0")).
			Padding(0, 1)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	hintStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(1, 2)

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

	// Create a visually appealing summary box
	content := strings.Join(parts, " • ")
	styledContent := dataStyle.Render(content)

	return boxStyle.Render(styledContent)
}

// RenderDataPoint formats a single data point with icon and enhanced styling
func RenderDataPoint(icon, text string) string {
	// Split text on bullets or colons to add colored labels
	if strings.Contains(text, ":") {
		parts := strings.SplitN(text, ":", 2)
		if len(parts) == 2 {
			label := labelStyle.Render(parts[0] + ":")
			value := dataStyle.Render(parts[1])
			return fmt.Sprintf("%s %s%s", icon, label, value)
		}
	}

	return fmt.Sprintf("%s %s", icon, dataStyle.Render(text))
}

// RenderHighlight formats highlighted text with extra emphasis
func RenderHighlight(icon, text string) string {
	// Add some visual flair to the highlight
	styledText := highlightStyle.Render(" ✨ " + text + " ✨")
	return fmt.Sprintf("%s %s", icon, styledText)
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
