package tui

import (
	"fmt"
	"strings"

	"github.com/alexinslc/rekap/internal/config"
	"github.com/alexinslc/rekap/internal/theme"
	"github.com/alexinslc/rekap/internal/ui"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ThemePreviewModel struct {
	themes       []string
	currentIndex int
	width        int
	height       int
	viewport     viewport.Model
}

func NewThemePreview() ThemePreviewModel {
	themes := theme.ListBuiltIn()
	vp := viewport.New(0, 0)
	
	return ThemePreviewModel{
		themes:       themes,
		currentIndex: 0,
		viewport:     vp,
	}
}

func (m ThemePreviewModel) Init() tea.Cmd {
	return nil
}

func (m ThemePreviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
			
		case "left", "h":
			if m.currentIndex > 0 {
				m.currentIndex--
				m.updateViewportContent()
			}
			
		case "right", "l":
			if m.currentIndex < len(m.themes)-1 {
				m.currentIndex++
				m.updateViewportContent()
			}
			
		case "enter":
			// Apply theme and exit
			themeName := m.themes[m.currentIndex]
			t, err := theme.Load(themeName)
			if err != nil {
				return m, tea.Quit
			}
			
			// Save to config
			cfg, _ := config.Load()
			cfg.ApplyTheme(t)
			if err := cfg.Save(); err != nil {
				return m, tea.Quit
			}
			return m, tea.Quit
		}
		
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4  // Account for borders
		m.viewport.Height = msg.Height - 6 // Account for header and footer
		m.updateViewportContent()
	}
	
	return m, nil
}

func (m *ThemePreviewModel) updateViewportContent() {
	if len(m.themes) == 0 {
		m.viewport.SetContent("No themes available")
		return
	}

	themeName := m.themes[m.currentIndex]
	t, err := theme.Load(themeName)
	if err != nil {
		m.viewport.SetContent(fmt.Sprintf("Error loading theme: %v", err))
		return
	}

	// Create a temporary config with the theme applied
	tempCfg := config.Default()
	tempCfg.ApplyTheme(t)

	// Create sample content
	var preview strings.Builder

	// Title
	preview.WriteString(ui.RenderTitle("📊 Today's rekap", false))
	preview.WriteString("\n\n")

	// Sample summary line
	summaryLines := []string{
		"11h 0m screen-on",
		"Top apps: VS Code (2h22m), Safari (1h29m)",
	}
	preview.WriteString(ui.RenderSummaryLine(summaryLines))
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

	m.viewport.SetContent(preview.String())
}

func (m ThemePreviewModel) View() string {
	if len(m.themes) == 0 {
		return "No themes available"
	}

	themeName := m.themes[m.currentIndex]
	t, err := theme.Load(themeName)
	if err != nil {
		return fmt.Sprintf("Error loading theme: %v", err)
	}

	// Create a temporary config with the theme applied
	tempCfg := config.Default()
	tempCfg.ApplyTheme(t)

	// Apply colors to UI
	ui.ApplyColors(tempCfg)

	// Create header with theme info and navigation
	themeNav := fmt.Sprintf("Theme: %s (%d/%d)", themeName, m.currentIndex+1, len(m.themes))
	help := "← → Navigate • Enter: Apply • q: Cancel"

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(ui.GetColor("primary")).
		PaddingBottom(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(ui.GetColor("muted")).
		PaddingBottom(1)

	header := lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render(themeNav),
		helpStyle.Render(help),
	)

	// Wrap content in a box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.GetColor("muted")).
		Padding(1, 2).
		Width(m.width - 4).
		MaxWidth(m.width - 4)

	content := boxStyle.Render(m.viewport.View())

	// Combine everything
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Center,
			header,
			content,
		),
	)
}

// RunThemePreview starts the interactive theme preview
func RunThemePreview() error {
	p := tea.NewProgram(
		NewThemePreview(),
		tea.WithAltScreen(),
	)

	_, err := p.Run()
	return err
}
