package tui

import (
	"fmt"
	"strings"

	"github.com/alexinslc/rekap/internal/config"
	"github.com/alexinslc/rekap/internal/theme"
	"github.com/alexinslc/rekap/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ThemePreviewModel represents the state of the theme preview TUI
type ThemePreviewModel struct {
	themes       []string
	currentIndex int
	width        int
	height       int
	applied      bool
	quitting     bool
}

// NewThemePreview creates a new theme preview model
func NewThemePreview() ThemePreviewModel {
	return ThemePreviewModel{
		themes:       theme.ListBuiltIn(),
		currentIndex: 0,
		applied:      false,
		quitting:     false,
	}
}

// Init initializes the model
func (m ThemePreviewModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m ThemePreviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "left", "h":
			if m.currentIndex > 0 {
				m.currentIndex--
			}

		case "right", "l":
			if m.currentIndex < len(m.themes)-1 {
				m.currentIndex++
			}

		case "enter":
			// Apply theme and exit
			themeName := m.themes[m.currentIndex]

			// Load config
			cfg, err := config.Load()
			if err != nil {
				cfg = config.Default()
			}

			// Load and apply theme
			t, err := theme.Load(themeName)
			if err == nil {
				cfg.ApplyTheme(t)
				_ = cfg.Save() // Ignore error for now
				m.applied = true
			}

			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the model
func (m ThemePreviewModel) View() string {
	if m.width == 0 {
		// Initial render before we get window size
		return "Loading..."
	}

	// Load current theme
	themeName := m.themes[m.currentIndex]
	t, err := theme.Load(themeName)
	if err != nil {
		return fmt.Sprintf("Error loading theme: %v", err)
	}

	// Apply theme temporarily for preview
	cfg := config.Default()
	cfg.ApplyTheme(t)
	ui.ApplyColors(cfg)

	// Create sample content
	var preview strings.Builder

	// Title
	preview.WriteString(ui.RenderTitle("📊 Today's rekap", false))
	preview.WriteString("\n\n")

	// Sample summary line
	preview.WriteString(ui.RenderSummaryLine([]string{
		"11h 0m screen-on",
		"Top apps: VS Code (2h22m), Safari (1h29m)",
	}))
	preview.WriteString("\n\n")

	// Sample system section
	preview.WriteString(ui.RenderHeader("SYSTEM"))
	preview.WriteString(ui.RenderDataPoint("⏰", "Active since 9:58 AM • 4h 47m awake"))
	preview.WriteString("\n")
	preview.WriteString(ui.RenderDataPoint("🔋", "Battery: 92% → 68% • discharging"))
	preview.WriteString("\n\n")

	// Sample productivity section
	preview.WriteString(ui.RenderHeader("PRODUCTIVITY"))
	preview.WriteString(ui.RenderHighlight("⏱️ ", "Best focus: 1h 27m in VS Code"))
	preview.WriteString("\n")
	preview.WriteString(ui.RenderDataPoint("📱", "VS Code • 2h 22m"))
	preview.WriteString("\n")

	// Wrap in box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(t.Colors.Muted)).
		Padding(1, 2).
		MaxWidth(m.width - 4)

	content := boxStyle.Render(preview.String())

	// Add header with theme info
	themeInfo := fmt.Sprintf("Theme: %s (%d/%d)", t.Name, m.currentIndex+1, len(m.themes))
	if t.Description != "" {
		themeInfo += " - " + t.Description
	}

	themeNavStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(t.Colors.Primary))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Colors.Muted))

	header := lipgloss.JoinVertical(
		lipgloss.Left,
		themeNavStyle.Render(themeInfo),
		helpStyle.Render("← → Navigate • Enter: Apply • q: Cancel"),
		"",
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
	)
}

// GetAppliedTheme returns the name of the applied theme if any
func (m ThemePreviewModel) GetAppliedTheme() string {
	if m.applied {
		return m.themes[m.currentIndex]
	}
	return ""
}
