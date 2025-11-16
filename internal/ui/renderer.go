package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alexinslc/rekap/internal/config"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var (
	// Color palette matching fang's aesthetic (defaults)
	primaryColor   = lipgloss.Color("13")  // Bright magenta/pink
	secondaryColor = lipgloss.Color("14")  // Cyan
	accentColor    = lipgloss.Color("11")  // Bright yellow
	successColor   = lipgloss.Color("10")  // Bright green
	warningColor   = lipgloss.Color("9")   // Bright red
	mutedColor     = lipgloss.Color("240") // Darker gray
	textColor      = lipgloss.Color("255") // White

	// Accessibility settings
	accessibilityEnabled  = false
	accessibilityNoEmoji  = false

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

// ApplyColors updates the color scheme based on config
func ApplyColors(cfg *config.Config) {
	if cfg == nil {
		return
	}

	// Apply accessibility settings
	accessibilityEnabled = cfg.Accessibility.Enabled
	accessibilityNoEmoji = cfg.Accessibility.NoEmoji

	// Update color palette
	// In high contrast mode (when both enabled and high_contrast are true), use black and white
	if cfg.Accessibility.Enabled && cfg.Accessibility.HighContrast {
		primaryColor = lipgloss.Color("15")   // White
		secondaryColor = lipgloss.Color("15") // White
		accentColor = lipgloss.Color("15")    // White
		successColor = lipgloss.Color("15")   // White
		warningColor = lipgloss.Color("15")   // White
		mutedColor = lipgloss.Color("250")    // Light gray
		textColor = lipgloss.Color("15")      // White
	} else {
		primaryColor = lipgloss.Color(cfg.Colors.Primary)
		secondaryColor = lipgloss.Color(cfg.Colors.Secondary)
		accentColor = lipgloss.Color(cfg.Colors.Accent)
		successColor = lipgloss.Color(cfg.Colors.Success)
		warningColor = lipgloss.Color(cfg.Colors.Warning)
		mutedColor = lipgloss.Color(cfg.Colors.Muted)
		textColor = lipgloss.Color(cfg.Colors.Text)
	}

	// Rebuild styles with new colors
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		MarginBottom(1)

	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		MarginTop(1).
		MarginBottom(1)

	dataStyle = lipgloss.NewStyle().
		Foreground(textColor)

	labelStyle = lipgloss.NewStyle().
		Foreground(secondaryColor)

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
}

// IsTTY returns true if stdout is a terminal
func IsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// RenderTitle renders the main title with optional animation
func RenderTitle(text string, animate bool) string {
	// Remove emoji if accessibility mode is enabled and no-emoji is set
	if accessibilityEnabled && accessibilityNoEmoji {
		text = removeEmoji(text)
	}
	
	// Add visual markers in accessibility mode
	if accessibilityEnabled {
		text = "=== " + text + " ==="
	}
	
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
	if accessibilityEnabled {
		// Add visual separators for sections
		return headerStyle.Render(">> " + text + " <<")
	}
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
	if accessibilityNoEmoji {
		icon = getAccessibleIcon(icon)
	}
	if accessibilityEnabled {
		// Add bullet point for better distinction
		return fmt.Sprintf("  • %s  %s", icon, dataStyle.Render(text))
	}
	return fmt.Sprintf("  %s  %s", icon, dataStyle.Render(text))
}

// RenderHighlight formats highlighted text with extra emphasis
func RenderHighlight(icon, text string) string {
	if accessibilityNoEmoji {
		icon = getAccessibleIcon(icon)
	}
	styledText := highlightStyle.Render(text)
	if accessibilityEnabled {
		// Add visual emphasis with markers
		return fmt.Sprintf("  ** %s  %s **", icon, styledText)
	}
	return fmt.Sprintf("  %s  %s", icon, styledText)
}

// RenderSubItem formats a sub-item with indentation
func RenderSubItem(text string) string {
	return fmt.Sprintf("      %s", hintStyle.Render(text))
}

// RenderSuccess formats a success message
func RenderSuccess(text string) string {
	if accessibilityEnabled {
		return successStyle.Render("[OK] " + text)
	}
	return successStyle.Render("✓ " + text)
}

// RenderHint formats a hint message
func RenderHint(text string) string {
	if accessibilityEnabled {
		return hintStyle.Render("[INFO] " + text)
	}
	return hintStyle.Render("💡 " + text)
}

// RenderError formats an error message
func RenderError(text string) string {
	if accessibilityEnabled {
		return errorStyle.Render("[ERROR] " + text)
	}
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

// FormatTime formats a time according to the config's preference
func FormatTime(t time.Time, timeFormat string) string {
	if timeFormat == "24h" {
		return t.Format("15:04")
	}
	return t.Format("3:04 PM")
}

// ClearScreen clears the terminal screen (if TTY)
func ClearScreen() {
	if IsTTY() {
		fmt.Print("\033[H\033[2J")
	}
}

// removeEmoji strips emoji characters from text
func removeEmoji(text string) string {
	// Simple approach: keep only ASCII printable characters and spaces
	var result strings.Builder
	lastWasSpace := false
	for _, r := range text {
		if (r >= 32 && r <= 126) || r == '\n' || r == '\t' {
			if r == ' ' || r == '\t' {
				// Only add space if last char wasn't a space
				if !lastWasSpace && result.Len() > 0 {
					result.WriteRune(' ')
					lastWasSpace = true
				}
			} else {
				result.WriteRune(r)
				lastWasSpace = false
			}
		} else {
			// Replace emoji with space but avoid double spaces
			if !lastWasSpace && result.Len() > 0 {
				result.WriteRune(' ')
				lastWasSpace = true
			}
		}
	}
	return strings.TrimSpace(result.String())
}

// getAccessibleIcon returns a text-based alternative to emoji icons
var accessibleIconMap = map[string]string{
	"⏰": "[TIME]",
	"🔋": "[BAT]",
	"🔌": "[PWR]",
	"📱": "[APP]",
	"⏱️": "[FOCUS]",
	"🎵": "[MUSIC]",
	"🌐": "[NET]",
	"📊": "[DATA]",
	"💡": "[INFO]",
	"✓":  "[OK]",
	"✗":  "[ERR]",
}

func getAccessibleIcon(emoji string) string {
	if alt, ok := accessibleIconMap[emoji]; ok {
		return alt
	}
	return "[*]"
}
