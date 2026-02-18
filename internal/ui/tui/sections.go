package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alexinslc/rekap/internal/collectors"
	"github.com/alexinslc/rekap/internal/config"
	"github.com/alexinslc/rekap/internal/ui"
)

// SummaryData mirrors cmd/rekap.SummaryData to avoid circular imports.
// The caller passes this from the cmd package.
type SummaryData struct {
	Uptime        collectors.UptimeResult
	Battery       collectors.BatteryResult
	Screen        collectors.ScreenResult
	Apps          collectors.AppsResult
	Focus         collectors.FocusResult
	Media         collectors.MediaResult
	Network       collectors.NetworkResult
	Browsers      collectors.BrowsersResult
	Notifications collectors.NotificationsResult
	Issues        collectors.IssuesResult
	Fragmentation collectors.FragmentationResult
	Burnout       collectors.BurnoutResult
}

func BuildSections(data *SummaryData, cfg *config.Config) []Section {
	s := &sectionBuilder{data: data, cfg: cfg}
	return []Section{
		s.system(),
		s.productivity(),
		s.browser(),
		s.network(),
		s.wellness(),
		s.media(),
		s.notifications(),
		s.issues(),
	}
}

type sectionBuilder struct {
	data *SummaryData
	cfg  *config.Config
}

func (s *sectionBuilder) system() Section {
	available := s.data.Uptime.Available || s.data.Battery.Available || s.data.Screen.Available
	if !available {
		return Section{Name: "System", Available: false, HintText: "System data unavailable"}
	}

	var summary, expanded strings.Builder

	if s.data.Uptime.Available {
		summary.WriteString(fmt.Sprintf("Uptime:    %s\n", s.data.Uptime.FormattedTime))
		expanded.WriteString(fmt.Sprintf("Uptime:    %s\n", s.data.Uptime.FormattedTime))
		expanded.WriteString(fmt.Sprintf("Boot time: %s\n",
			ui.FormatTime(s.data.Uptime.BootTime, s.cfg.Display.TimeFormat)))
	}

	if s.data.Battery.Available && s.cfg.ShouldShowBattery() {
		status := "discharging"
		if s.data.Battery.IsPlugged {
			status = "plugged in"
		}
		if s.data.Battery.StartPct != s.data.Battery.CurrentPct {
			summary.WriteString(fmt.Sprintf("Battery:   %d%% -> %d%% (%s)\n",
				s.data.Battery.StartPct, s.data.Battery.CurrentPct, status))
		} else {
			summary.WriteString(fmt.Sprintf("Battery:   %d%% (%s)\n",
				s.data.Battery.CurrentPct, status))
		}
		expanded.WriteString(fmt.Sprintf("Battery:   %d%% -> %d%% (%s)\n",
			s.data.Battery.StartPct, s.data.Battery.CurrentPct, status))
		if s.data.Battery.PlugCount > 0 {
			expanded.WriteString(fmt.Sprintf("Plug events: %d today\n", s.data.Battery.PlugCount))
		}
	}

	if s.data.Screen.Available {
		summary.WriteString(fmt.Sprintf("Screen:    %s on\n", ui.FormatDuration(s.data.Screen.ScreenOnMinutes)))
		expanded.WriteString(fmt.Sprintf("Screen:    %s on\n", ui.FormatDuration(s.data.Screen.ScreenOnMinutes)))
		if s.data.Screen.LockCount > 0 {
			expanded.WriteString(fmt.Sprintf("Locks:     %d", s.data.Screen.LockCount))
			if s.data.Screen.AvgMinsBetweenLock > 0 {
				expanded.WriteString(fmt.Sprintf(" (avg %s between)", ui.FormatDuration(s.data.Screen.AvgMinsBetweenLock)))
			}
			expanded.WriteString("\n")
		}
	}

	return Section{
		Name:      "System",
		Available: true,
		Summary:   strings.TrimRight(summary.String(), "\n"),
		Expanded:  strings.TrimRight(expanded.String(), "\n"),
	}
}

func (s *sectionBuilder) productivity() Section {
	available := s.data.Apps.Available || s.data.Focus.Available
	if !available {
		return Section{
			Name:      "Productivity",
			Available: false,
			HintText:  "Grant Full Disk Access to enable app tracking.\nRun 'rekap init' for setup.",
		}
	}

	var summary, expanded strings.Builder

	if s.data.Focus.Available && s.data.Focus.StreakMinutes > 0 {
		summary.WriteString(fmt.Sprintf("Focus:     %s in %s\n",
			ui.FormatDuration(s.data.Focus.StreakMinutes), s.data.Focus.AppName))
		expanded.WriteString(fmt.Sprintf("Focus:     %s in %s\n",
			ui.FormatDuration(s.data.Focus.StreakMinutes), s.data.Focus.AppName))
	}

	if s.data.Apps.Available && len(s.data.Apps.TopApps) > 0 {
		summary.WriteString("\nTop Apps:\n")
		for i, app := range s.data.Apps.TopApps {
			if i >= 3 {
				break
			}
			summary.WriteString(fmt.Sprintf("  %d. %-16s %s\n", i+1, app.Name, ui.FormatDuration(app.Minutes)))
		}

		expanded.WriteString("\nAll Apps:\n")
		for i, app := range s.data.Apps.TopApps {
			if i >= 10 {
				break
			}
			expanded.WriteString(fmt.Sprintf("  %d. %-16s %s  (%s)\n",
				i+1, app.Name, ui.FormatDuration(app.Minutes), app.BundleID))
		}

		if s.data.Apps.SwitchingAvailable {
			expanded.WriteString(fmt.Sprintf("\nSwitches:  %d total (%.1f/hr)\n",
				s.data.Apps.TotalSwitches, s.data.Apps.SwitchesPerHour))
			if s.data.Apps.AvgMinsBetween > 0 {
				expanded.WriteString(fmt.Sprintf("Avg between: %.1f min\n", s.data.Apps.AvgMinsBetween))
			}
		}
	}

	return Section{
		Name:      "Productivity",
		Available: true,
		Summary:   strings.TrimRight(summary.String(), "\n"),
		Expanded:  strings.TrimRight(expanded.String(), "\n"),
	}
}

func (s *sectionBuilder) browser() Section {
	if !s.data.Browsers.Available || (s.data.Browsers.TotalTabs == 0 && s.data.Browsers.TotalURLsVisited == 0) {
		return Section{Name: "Browser", Available: false, HintText: "No browser data available"}
	}

	var summary, expanded strings.Builder

	if s.data.Browsers.TotalTabs > 0 {
		summary.WriteString(fmt.Sprintf("Tabs:      %d open\n", s.data.Browsers.TotalTabs))
	}
	if s.data.Browsers.TotalURLsVisited > 0 {
		summary.WriteString(fmt.Sprintf("Visited:   %d URLs today\n", s.data.Browsers.TotalURLsVisited))
	}
	if s.data.Browsers.TopHistoryDomain != "" {
		summary.WriteString(fmt.Sprintf("Top site:  %s (%d visits)\n",
			s.data.Browsers.TopHistoryDomain, s.data.Browsers.TopDomainVisits))
	}

	// Expanded: per-browser breakdown
	if s.data.Browsers.Chrome.Available {
		expanded.WriteString(fmt.Sprintf("Chrome:    %d tabs\n", s.data.Browsers.Chrome.TabCount))
	}
	if s.data.Browsers.Safari.Available {
		expanded.WriteString(fmt.Sprintf("Safari:    %d tabs\n", s.data.Browsers.Safari.TabCount))
	}
	if s.data.Browsers.Edge.Available {
		expanded.WriteString(fmt.Sprintf("Edge:      %d tabs\n", s.data.Browsers.Edge.TabCount))
	}

	if s.data.Browsers.TotalURLsVisited > 0 {
		expanded.WriteString(fmt.Sprintf("\nURLs visited: %d\n", s.data.Browsers.TotalURLsVisited))
		if s.data.Browsers.TopHistoryDomain != "" {
			expanded.WriteString(fmt.Sprintf("Top domain:   %s (%d visits)\n",
				s.data.Browsers.TopHistoryDomain, s.data.Browsers.TopDomainVisits))
		}
	}

	// Top tab domains
	if len(s.data.Browsers.TopDomains) > 0 {
		type dc struct {
			domain string
			count  int
		}
		var domains []dc
		for domain, count := range s.data.Browsers.TopDomains {
			domains = append(domains, dc{domain, count})
		}
		sort.Slice(domains, func(i, j int) bool {
			return domains[i].count > domains[j].count
		})
		expanded.WriteString("\nTop tab domains:\n")
		for i, d := range domains {
			if i >= 5 {
				break
			}
			expanded.WriteString(fmt.Sprintf("  %s (%d)\n", d.domain, d.count))
		}
	}

	// Work/distraction breakdown
	total := s.data.Browsers.WorkVisits + s.data.Browsers.DistractionVisits + s.data.Browsers.NeutralVisits
	if total > 0 {
		expanded.WriteString("\nDomain breakdown:\n")
		expanded.WriteString(fmt.Sprintf("  Work:        %d visits (%d%%)\n",
			s.data.Browsers.WorkVisits, pct(s.data.Browsers.WorkVisits, total)))
		expanded.WriteString(fmt.Sprintf("  Distraction: %d visits (%d%%)\n",
			s.data.Browsers.DistractionVisits, pct(s.data.Browsers.DistractionVisits, total)))
		expanded.WriteString(fmt.Sprintf("  Neutral:     %d visits (%d%%)\n",
			s.data.Browsers.NeutralVisits, pct(s.data.Browsers.NeutralVisits, total)))
	}

	return Section{
		Name:      "Browser",
		Available: true,
		Summary:   strings.TrimRight(summary.String(), "\n"),
		Expanded:  strings.TrimRight(expanded.String(), "\n"),
	}
}

func (s *sectionBuilder) network() Section {
	if !s.data.Network.Available {
		return Section{Name: "Network", Available: false, HintText: "No network data available"}
	}

	qualifier := ""
	if s.data.Network.SinceBoot {
		qualifier = " (since boot)"
	}

	summary := fmt.Sprintf("%s: %s down / %s up%s",
		s.data.Network.InterfaceName,
		collectors.FormatBytes(s.data.Network.BytesReceived),
		collectors.FormatBytes(s.data.Network.BytesSent),
		qualifier)

	expanded := fmt.Sprintf("Interface: %s\nNetwork:   %s\nReceived:  %s\nSent:      %s%s",
		s.data.Network.InterfaceName,
		s.data.Network.NetworkName,
		collectors.FormatBytes(s.data.Network.BytesReceived),
		collectors.FormatBytes(s.data.Network.BytesSent),
		qualifier)

	return Section{
		Name:      "Network",
		Available: true,
		Summary:   summary,
		Expanded:  expanded,
	}
}

func (s *sectionBuilder) wellness() Section {
	fragAvail := s.data.Fragmentation.Available
	burnoutAvail := s.data.Burnout.Available && len(s.data.Burnout.Warnings) > 0
	if !fragAvail && !burnoutAvail {
		return Section{Name: "Wellness", Available: false, HintText: "No wellness data available"}
	}

	var summary, expanded strings.Builder

	if fragAvail {
		summary.WriteString(fmt.Sprintf("Fragmentation: %d/100 (%s)\n",
			s.data.Fragmentation.Score, s.data.Fragmentation.Level))

		expanded.WriteString(fmt.Sprintf("Fragmentation: %d/100 (%s)\n\n",
			s.data.Fragmentation.Score, s.data.Fragmentation.Level))
		expanded.WriteString("Score Breakdown:\n")
		b := s.data.Fragmentation.Breakdown
		expanded.WriteString(fmt.Sprintf("  Apps:     %d unique (weight: 30%%)\n", b.UniqueApps))
		expanded.WriteString(fmt.Sprintf("  Tabs:     %d total (weight: 25%%)\n", b.TotalTabs))
		expanded.WriteString(fmt.Sprintf("  Domains:  %d unique (weight: 25%%)\n", b.UniqueDomains))
		expanded.WriteString(fmt.Sprintf("  Switches: %.1f/hr (weight: 20%%)\n", b.AppSwitchesPerHour))
	}

	if burnoutAvail {
		summary.WriteString(fmt.Sprintf("Warnings:      %d\n", len(s.data.Burnout.Warnings)))

		expanded.WriteString("\nBurnout Warnings:\n")
		severityOrder := map[string]int{"high": 0, "medium": 1, "low": 2}
		sorted := make([]collectors.BurnoutWarning, len(s.data.Burnout.Warnings))
		copy(sorted, s.data.Burnout.Warnings)
		sort.Slice(sorted, func(i, j int) bool {
			return severityOrder[sorted[i].Severity] < severityOrder[sorted[j].Severity]
		})
		for _, w := range sorted {
			expanded.WriteString(fmt.Sprintf("  [%s] %s\n", w.Severity, w.Message))
		}
	} else if fragAvail {
		summary.WriteString("Warnings:      none\n")
	}

	return Section{
		Name:      "Wellness",
		Available: true,
		Summary:   strings.TrimRight(summary.String(), "\n"),
		Expanded:  strings.TrimRight(expanded.String(), "\n"),
	}
}

func (s *sectionBuilder) media() Section {
	if !s.data.Media.Available || !s.cfg.ShouldShowMedia() {
		return Section{Name: "Media", Available: false, HintText: "No media playing"}
	}

	content := fmt.Sprintf("\"%s\" in %s", s.data.Media.Track, s.data.Media.App)
	return Section{
		Name:      "Media",
		Available: true,
		Summary:   content,
		Expanded:  content,
	}
}

func (s *sectionBuilder) notifications() Section {
	if !s.data.Notifications.Available || s.data.Notifications.TotalNotifications == 0 {
		return Section{Name: "Notifications", Available: false, HintText: "No notifications today"}
	}

	var summary, expanded strings.Builder

	summary.WriteString(fmt.Sprintf("Total: %d notifications\n", s.data.Notifications.TotalNotifications))
	if len(s.data.Notifications.TopApps) > 0 {
		summary.WriteString(fmt.Sprintf("Top:   %s (%d)\n",
			s.data.Notifications.TopApps[0].Name, s.data.Notifications.TopApps[0].Count))
	}

	expanded.WriteString(fmt.Sprintf("Total: %d notifications\n\nTop Apps:\n", s.data.Notifications.TotalNotifications))
	for i, app := range s.data.Notifications.TopApps {
		if i >= 10 {
			break
		}
		expanded.WriteString(fmt.Sprintf("  %d. %-16s %d\n", i+1, app.Name, app.Count))
	}

	return Section{
		Name:      "Notifications",
		Available: true,
		Summary:   strings.TrimRight(summary.String(), "\n"),
		Expanded:  strings.TrimRight(expanded.String(), "\n"),
	}
}

func (s *sectionBuilder) issues() Section {
	if !s.data.Issues.Available || len(s.data.Issues.Issues) == 0 {
		return Section{Name: "Issues", Available: false, HintText: "No issues/tickets viewed today"}
	}

	var summary, expanded strings.Builder

	summary.WriteString(fmt.Sprintf("%d issues/tickets viewed today", len(s.data.Issues.Issues)))

	expanded.WriteString("Issues/Tickets Viewed:\n")
	for i, issue := range s.data.Issues.Issues {
		if i >= 20 {
			break
		}
		expanded.WriteString(fmt.Sprintf("  %s (%s, %d visits)\n",
			issue.ID, issue.Tracker, issue.VisitCount))
	}

	return Section{
		Name:      "Issues",
		Available: true,
		Summary:   strings.TrimRight(summary.String(), "\n"),
		Expanded:  strings.TrimRight(expanded.String(), "\n"),
	}
}

func pct(part, total int) int {
	if total == 0 {
		return 0
	}
	return int(float64(part) / float64(total) * 100)
}
