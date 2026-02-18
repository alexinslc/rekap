package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/alexinslc/rekap/internal/config"
)

// SummaryData is provided by the caller -- we use an interface to avoid circular imports.
// The caller passes section builders that close over the actual data.
type Section struct {
	Name      string
	Available bool
	HintText  string
	Summary   string
	Expanded  string
}

type Model struct {
	sections  []Section
	cursor    int
	drillDown bool
	viewport  viewport.Model
	width     int
	height    int
	ready     bool
	tooSmall  bool
	styles    tuiStyles
	date      string
}

func New(sections []Section, cfg *config.Config) Model {
	applyThemeColors(cfg)
	return Model{
		sections: sections,
		styles:   buildStyles(),
		date:     time.Now().Format("Mon, Jan 2 2006"),
	}
}

func applyThemeColors(cfg *config.Config) {
	if cfg == nil {
		return
	}
	if cfg.Accessibility.Enabled && cfg.Accessibility.HighContrast {
		primaryColor = lipgloss.Color("15")
		secondaryColor = lipgloss.Color("15")
		accentColor = lipgloss.Color("15")
		successColor = lipgloss.Color("15")
		warningColor = lipgloss.Color("15")
		mutedColor = lipgloss.Color("250")
		textColor = lipgloss.Color("15")
	} else {
		primaryColor = lipgloss.Color(cfg.Colors.Primary)
		secondaryColor = lipgloss.Color(cfg.Colors.Secondary)
		accentColor = lipgloss.Color(cfg.Colors.Accent)
		successColor = lipgloss.Color(cfg.Colors.Success)
		warningColor = lipgloss.Color(cfg.Colors.Warning)
		mutedColor = lipgloss.Color(cfg.Colors.Muted)
		textColor = lipgloss.Color(cfg.Colors.Text)
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.tooSmall = msg.Width < minTermWidth

		contentWidth := msg.Width - sidebarWidth - 3 // border + padding
		contentHeight := msg.Height - 3               // title + footer

		if !m.ready {
			m.viewport = viewport.New(contentWidth, contentHeight)
			m.viewport.SetContent(m.detailContent())
			m.ready = true
		} else {
			m.viewport.Width = contentWidth
			m.viewport.Height = contentHeight
			m.viewport.SetContent(m.detailContent())
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc":
			if m.drillDown {
				m.drillDown = false
				m.viewport.SetContent(m.detailContent())
				m.viewport.GotoTop()
			} else {
				return m, tea.Quit
			}

		case "up", "k":
			if !m.drillDown {
				if m.cursor > 0 {
					m.cursor--
					m.viewport.SetContent(m.detailContent())
					m.viewport.GotoTop()
				}
			} else {
				var cmd tea.Cmd
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}

		case "down", "j":
			if !m.drillDown {
				if m.cursor < len(m.sections)-1 {
					m.cursor++
					m.viewport.SetContent(m.detailContent())
					m.viewport.GotoTop()
				}
			} else {
				var cmd tea.Cmd
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}

		case "enter":
			if !m.drillDown {
				m.drillDown = true
				m.viewport.SetContent(m.detailContent())
				m.viewport.GotoTop()
			}

		case "pgup", "ctrl+u":
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd

		case "pgdown", "ctrl+d":
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	if m.tooSmall {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.styles.muted.Render("Terminal too small (need 80+ columns)"))
	}

	// Title bar
	title := m.styles.titleBar.Render(fmt.Sprintf("rekap - %s", m.date))
	titleBar := lipgloss.NewStyle().
		Width(m.width).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(mutedColor).
		Render(title)

	// Sidebar
	sidebar := m.renderSidebar()

	// Detail pane
	detail := m.styles.detailPane.
		Width(m.width - sidebarWidth - 3).
		Render(m.viewport.View())

	// Join sidebar and detail
	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, detail)

	// Footer
	var footerText string
	if m.drillDown {
		footerText = "Esc back  j/k scroll  q quit"
	} else {
		footerText = "j/k navigate  Enter detail  Esc/q quit"
	}
	footer := m.styles.footerBar.Render(footerText)

	return lipgloss.JoinVertical(lipgloss.Left, titleBar, body, footer)
}

func (m Model) renderSidebar() string {
	var rows []string
	for i, section := range m.sections {
		var row string
		if !section.Available {
			row = m.styles.sidebarUnavailable.Render(section.Name + " (n/a)")
		} else if i == m.cursor {
			row = m.styles.sidebarActive.Render("> " + section.Name)
		} else {
			row = m.styles.sidebarItem.Render(section.Name)
		}
		rows = append(rows, row)
	}

	content := strings.Join(rows, "\n")

	return m.styles.sidebarContainer.
		Height(m.height - 3). // title + footer
		Render(content)
}

func (m Model) detailContent() string {
	if m.cursor >= len(m.sections) {
		return ""
	}

	section := m.sections[m.cursor]

	if !section.Available {
		return m.styles.muted.Render(section.HintText)
	}

	header := m.styles.sectionHeader.Render(section.Name)

	var content string
	if m.drillDown {
		content = section.Expanded
	} else {
		content = section.Summary
	}

	return header + "\n" + content
}
