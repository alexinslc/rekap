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

// Section represents a single summary section shown in the TUI.
// Callers construct and pass populated sections into the UI model.
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
	palette   colorPalette
	date      string
}

func New(sections []Section, cfg *config.Config) Model {
	palette := colorsFromConfig(cfg)
	return Model{
		sections: sections,
		styles:   buildStylesFromPalette(palette),
		palette:  palette,
		date:     time.Now().Format("Mon, Jan 2 2006"),
	}
}

type colorPalette struct {
	primary   lipgloss.Color
	secondary lipgloss.Color
	accent    lipgloss.Color
	success   lipgloss.Color
	warning   lipgloss.Color
	muted     lipgloss.Color
	text      lipgloss.Color
}

func colorsFromConfig(cfg *config.Config) colorPalette {
	if cfg == nil {
		return colorPalette{
			primary: "13", secondary: "14", accent: "11",
			success: "10", warning: "9", muted: "240", text: "255",
		}
	}
	if cfg.Accessibility.Enabled && cfg.Accessibility.HighContrast {
		return colorPalette{
			primary: "15", secondary: "15", accent: "15",
			success: "15", warning: "15", muted: "250", text: "15",
		}
	}
	return colorPalette{
		primary:   lipgloss.Color(cfg.Colors.Primary),
		secondary: lipgloss.Color(cfg.Colors.Secondary),
		accent:    lipgloss.Color(cfg.Colors.Accent),
		success:   lipgloss.Color(cfg.Colors.Success),
		warning:   lipgloss.Color(cfg.Colors.Warning),
		muted:     lipgloss.Color(cfg.Colors.Muted),
		text:      lipgloss.Color(cfg.Colors.Text),
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
		if contentWidth < 0 {
			contentWidth = 0
		}
		contentHeight := msg.Height - 3 // title + footer
		if contentHeight < 0 {
			contentHeight = 0
		}

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
		BorderForeground(m.palette.muted).
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
