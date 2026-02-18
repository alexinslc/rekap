package tui

import "github.com/charmbracelet/lipgloss"

const sidebarWidth = 22
const minTermWidth = 80

func buildStylesFromPalette(p colorPalette) tuiStyles {
	return tuiStyles{
		titleBar: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.primary).
			PaddingLeft(1),

		sidebarContainer: lipgloss.NewStyle().
			Width(sidebarWidth).
			BorderStyle(lipgloss.NormalBorder()).
			BorderRight(true).
			BorderForeground(p.muted).
			PaddingLeft(1).
			PaddingRight(1),

		sidebarItem: lipgloss.NewStyle().
			Foreground(p.text).
			PaddingLeft(1),

		sidebarActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.primary).
			PaddingLeft(0),

		sidebarUnavailable: lipgloss.NewStyle().
			Foreground(p.muted).
			PaddingLeft(1),

		detailPane: lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(1),

		sectionHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.primary).
			MarginBottom(1),

		dataLabel: lipgloss.NewStyle().
			Foreground(p.secondary),

		dataValue: lipgloss.NewStyle().
			Foreground(p.text),

		highlight: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.accent),

		success: lipgloss.NewStyle().
			Foreground(p.success),

		warning: lipgloss.NewStyle().
			Foreground(p.warning),

		muted: lipgloss.NewStyle().
			Foreground(p.muted).
			Italic(true),

		footerBar: lipgloss.NewStyle().
			Foreground(p.muted).
			PaddingLeft(1),
	}
}

type tuiStyles struct {
	titleBar           lipgloss.Style
	sidebarContainer   lipgloss.Style
	sidebarItem        lipgloss.Style
	sidebarActive      lipgloss.Style
	sidebarUnavailable lipgloss.Style
	detailPane         lipgloss.Style
	sectionHeader      lipgloss.Style
	dataLabel          lipgloss.Style
	dataValue          lipgloss.Style
	highlight          lipgloss.Style
	success            lipgloss.Style
	warning            lipgloss.Style
	muted              lipgloss.Style
	footerBar          lipgloss.Style
}
