package tui

import "github.com/charmbracelet/lipgloss"

const sidebarWidth = 22
const minTermWidth = 80

var (
	// Colors -- these get overridden by ApplyThemeColors
	primaryColor   = lipgloss.Color("13")
	secondaryColor = lipgloss.Color("14")
	accentColor    = lipgloss.Color("11")
	successColor   = lipgloss.Color("10")
	warningColor   = lipgloss.Color("9")
	mutedColor     = lipgloss.Color("240")
	textColor      = lipgloss.Color("255")
)

func buildStyles() tuiStyles {
	return tuiStyles{
		titleBar: lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			PaddingLeft(1),

		sidebarContainer: lipgloss.NewStyle().
			Width(sidebarWidth).
			BorderStyle(lipgloss.NormalBorder()).
			BorderRight(true).
			BorderForeground(mutedColor).
			PaddingLeft(1).
			PaddingRight(1),

		sidebarItem: lipgloss.NewStyle().
			Foreground(textColor).
			PaddingLeft(1),

		sidebarActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			PaddingLeft(0),

		sidebarUnavailable: lipgloss.NewStyle().
			Foreground(mutedColor).
			PaddingLeft(1),

		detailPane: lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(1),

		sectionHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1),

		dataLabel: lipgloss.NewStyle().
			Foreground(secondaryColor),

		dataValue: lipgloss.NewStyle().
			Foreground(textColor),

		highlight: lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor),

		success: lipgloss.NewStyle().
			Foreground(successColor),

		warning: lipgloss.NewStyle().
			Foreground(warningColor),

		muted: lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true),

		footerBar: lipgloss.NewStyle().
			Foreground(mutedColor).
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
