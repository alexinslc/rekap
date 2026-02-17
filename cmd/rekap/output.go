package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alexinslc/rekap/internal/collectors"
	"github.com/alexinslc/rekap/internal/config"
	"github.com/alexinslc/rekap/internal/ui"
)

func printQuiet(cfg *config.Config, data *SummaryData) {
	if data.Uptime.Available {
		fmt.Printf("awake_minutes=%d\n", data.Uptime.AwakeMinutes)
		fmt.Printf("boot_time=%d\n", data.Uptime.BootTime.Unix())
	}

	if data.Battery.Available {
		fmt.Printf("battery_start_pct=%d\n", data.Battery.StartPct)
		fmt.Printf("battery_now_pct=%d\n", data.Battery.CurrentPct)
		fmt.Printf("plug_events=%d\n", data.Battery.PlugCount)
		if data.Battery.IsPlugged {
			fmt.Printf("is_plugged=1\n")
		} else {
			fmt.Printf("is_plugged=0\n")
		}
	}

	if data.Screen.Available {
		fmt.Printf("screen_on_minutes=%d\n", data.Screen.ScreenOnMinutes)
		if data.Screen.LockCount > 0 {
			fmt.Printf("screen_lock_count=%d\n", data.Screen.LockCount)
			fmt.Printf("avg_mins_between_locks=%d\n", data.Screen.AvgMinsBetweenLock)
		}
	}

	if data.Apps.Available {
		for i, app := range data.Apps.TopApps {
			if i >= 3 {
				break
			}
			fmt.Printf("top_app_%d=%s\n", i+1, app.Name)
			fmt.Printf("top_app_%d_minutes=%d\n", i+1, app.Minutes)
		}
	}

	if data.Focus.Available {
		fmt.Printf("focus_streak_minutes=%d\n", data.Focus.StreakMinutes)
		fmt.Printf("focus_streak_app=%s\n", data.Focus.AppName)
	}

	if data.Media.Available {
		fmt.Printf("media_track=%s\n", data.Media.Track)
		fmt.Printf("media_app=%s\n", data.Media.App)
	}

	if data.Network.Available {
		fmt.Printf("network_interface=%s\n", data.Network.InterfaceName)
		fmt.Printf("network_name=%s\n", data.Network.NetworkName)
		fmt.Printf("network_bytes_received=%d\n", data.Network.BytesReceived)
		fmt.Printf("network_bytes_sent=%d\n", data.Network.BytesSent)
	}

	if data.Browsers.Available {
		fmt.Printf("browser_total_tabs=%d\n", data.Browsers.TotalTabs)
		if data.Browsers.Chrome.Available {
			fmt.Printf("browser_chrome_tabs=%d\n", data.Browsers.Chrome.TabCount)
		}
		if data.Browsers.Safari.Available {
			fmt.Printf("browser_safari_tabs=%d\n", data.Browsers.Safari.TabCount)
		}
		if data.Browsers.Edge.Available {
			fmt.Printf("browser_edge_tabs=%d\n", data.Browsers.Edge.TabCount)
		}
		totalCategorized := data.Browsers.WorkVisits + data.Browsers.DistractionVisits + data.Browsers.NeutralVisits
		if totalCategorized > 0 {
			fmt.Printf("browser_work_visits=%d\n", data.Browsers.WorkVisits)
			fmt.Printf("browser_distraction_visits=%d\n", data.Browsers.DistractionVisits)
			fmt.Printf("browser_neutral_visits=%d\n", data.Browsers.NeutralVisits)
		}
		if data.Browsers.TotalURLsVisited > 0 {
			fmt.Printf("browser_urls_visited=%d\n", data.Browsers.TotalURLsVisited)
		}
		if data.Browsers.TopHistoryDomain != "" {
			fmt.Printf("browser_top_domain=%s\n", data.Browsers.TopHistoryDomain)
			fmt.Printf("browser_top_domain_visits=%d\n", data.Browsers.TopDomainVisits)
		}
		if len(data.Browsers.AllIssueURLs) > 0 {
			fmt.Printf("browser_issues_viewed=%d\n", len(data.Browsers.AllIssueURLs))
		}
	}

	if data.Notifications.Available {
		fmt.Printf("notifications_total=%d\n", data.Notifications.TotalNotifications)
		for i, app := range data.Notifications.TopApps {
			if i >= 3 {
				break
			}
			fmt.Printf("notification_app_%d=%s\n", i+1, app.Name)
			fmt.Printf("notification_app_%d_count=%d\n", i+1, app.Count)
		}
	}

	if data.Fragmentation.Available {
		fmt.Printf("fragmentation_score=%d\n", data.Fragmentation.Score)
		fmt.Printf("fragmentation_level=%s\n", data.Fragmentation.Level)
	}

	if data.Issues.Available {
		fmt.Printf("issues_count=%d\n", len(data.Issues.Issues))
		for i, issue := range data.Issues.Issues {
			if i >= 10 {
				break
			}
			fmt.Printf("issue_%d_id=%s\n", i+1, issue.ID)
			fmt.Printf("issue_%d_tracker=%s\n", i+1, issue.Tracker)
			fmt.Printf("issue_%d_visits=%d\n", i+1, issue.VisitCount)
		}
	}

	overload := collectors.CheckContextOverload(data.Apps, data.Browsers)
	if overload.IsOverloaded {
		fmt.Printf("context_overload=1\n")
		fmt.Printf("context_overload_message=%s\n", overload.WarningMessage)
	} else {
		fmt.Printf("context_overload=0\n")
	}
}

func printHuman(cfg *config.Config, data *SummaryData) {
	title := ui.RenderTitle("📊 Today's rekap", ui.IsTTY())
	if title != "" {
		fmt.Println(title)
	}
	fmt.Println()

	// Check for context overload
	overload := collectors.CheckContextOverload(data.Apps, data.Browsers)
	if overload.IsOverloaded {
		fmt.Println(ui.RenderWarning("Context overload: " + overload.WarningMessage))
		fmt.Println()
	}

	// Build summary line
	var summaryParts []string

	if data.Screen.Available {
		summaryParts = append(summaryParts, ui.FormatDuration(data.Screen.ScreenOnMinutes)+" screen-on")
	}

	if data.Apps.Available && len(data.Apps.TopApps) > 0 {
		var appList []string
		for i, app := range data.Apps.TopApps {
			if i >= 3 {
				break
			}
			appList = append(appList, fmt.Sprintf("%s (%s)", app.Name, ui.FormatDurationCompact(app.Minutes)))
		}
		if len(appList) > 0 {
			summaryParts = append(summaryParts, "Top apps: "+strings.Join(appList, ", "))
		}
	}

	if len(summaryParts) > 0 {
		fmt.Println(ui.RenderSummaryLine(summaryParts))
		fmt.Println()
	}

	// System Status Section
	fmt.Println(ui.RenderHeader("SYSTEM"))

	if data.Uptime.Available {
		text := fmt.Sprintf("Active since %s • %s",
			ui.FormatTime(data.Uptime.BootTime, cfg.Display.TimeFormat),
			data.Uptime.FormattedTime)
		fmt.Println(ui.RenderDataPoint("⏰", text))
	}

	if data.Battery.Available && cfg.ShouldShowBattery() {
		status := "discharging"
		if data.Battery.IsPlugged {
			status = "plugged in"
		}
		var text string
		if data.Battery.StartPct != data.Battery.CurrentPct {
			text = fmt.Sprintf("%d%% → %d%% • %s", data.Battery.StartPct, data.Battery.CurrentPct, status)
		} else {
			text = fmt.Sprintf("%d%% • %s", data.Battery.CurrentPct, status)
		}
		fmt.Println(ui.RenderDataPoint("🔋", text))

		if data.Battery.PlugCount > 0 {
			plugText := fmt.Sprintf("%d plug event(s) today", data.Battery.PlugCount)
			fmt.Println(ui.RenderDataPoint("🔌", plugText))
		}
	}

	if data.Screen.Available && data.Screen.LockCount > 0 {
		var lockText string
		if data.Screen.AvgMinsBetweenLock > 0 {
			lockText = fmt.Sprintf("Screen locked %d time%s (avg %s between breaks)",
				data.Screen.LockCount,
				pluralize(data.Screen.LockCount),
				ui.FormatDuration(data.Screen.AvgMinsBetweenLock))
		} else {
			lockText = fmt.Sprintf("Screen locked %d time%s today",
				data.Screen.LockCount,
				pluralize(data.Screen.LockCount))
		}
		fmt.Println(ui.RenderDataPoint("🔒", lockText))
	}

	// Productivity Section
	if data.Focus.Available || (data.Apps.Available && len(data.Apps.TopApps) > 0) {
		fmt.Println()
		fmt.Println(ui.RenderHeader("PRODUCTIVITY"))

		if data.Focus.Available {
			text := fmt.Sprintf("Best focus: %s in %s", ui.FormatDuration(data.Focus.StreakMinutes), data.Focus.AppName)
			fmt.Println(ui.RenderHighlight("⏱️ ", text))
		}

		if data.Apps.Available && len(data.Apps.TopApps) > 0 {
			for i, app := range data.Apps.TopApps {
				if i >= 3 {
					break
				}
				appText := fmt.Sprintf("%s • %s", app.Name, ui.FormatDuration(app.Minutes))
				fmt.Println(ui.RenderDataPoint("📱", appText))
			}
		}
	}

	// Media Section
	if data.Media.Available && cfg.ShouldShowMedia() {
		fmt.Println()
		fmt.Println(ui.RenderHeader("NOW PLAYING"))
		text := fmt.Sprintf("\"%s\" in %s", data.Media.Track, data.Media.App)
		fmt.Println(ui.RenderDataPoint("🎵", text))
	}

	// Network Activity Section
	if data.Network.Available {
		fmt.Println()
		fmt.Println(ui.RenderHeader("NETWORK ACTIVITY"))

		text := fmt.Sprintf("%s: \"%s\" • %s down / %s up",
			data.Network.InterfaceName,
			data.Network.NetworkName,
			collectors.FormatBytes(data.Network.BytesReceived),
			collectors.FormatBytes(data.Network.BytesSent))
		fmt.Println(ui.RenderDataPoint("🌐", text))
	}

	// Browser Activity Section (tabs + history + domain breakdown)
	if data.Browsers.Available && (data.Browsers.TotalTabs > 0 || data.Browsers.TotalURLsVisited > 0) {
		fmt.Println()
		fmt.Println(ui.RenderHeader("BROWSER ACTIVITY"))

		if data.Browsers.TotalURLsVisited > 0 {
			historyText := fmt.Sprintf("%d URLs visited today", data.Browsers.TotalURLsVisited)
			if data.Browsers.TopHistoryDomain != "" {
				historyText += fmt.Sprintf(" • Top: %s (%d visit%s)",
					data.Browsers.TopHistoryDomain,
					data.Browsers.TopDomainVisits,
					pluralize(data.Browsers.TopDomainVisits))
			}
			fmt.Println(ui.RenderDataPoint("📊", historyText))

			if len(data.Browsers.AllIssueURLs) > 0 {
				issueText := fmt.Sprintf("Issues viewed: %s", collectors.FormatIssueURLs(data.Browsers.AllIssueURLs))
				fmt.Println(ui.RenderDataPoint("🎫", issueText))
			}
		}

		if data.Browsers.TotalTabs > 0 {
			text := fmt.Sprintf("%d tabs open", data.Browsers.TotalTabs)
			if data.Browsers.Chrome.Available {
				text += fmt.Sprintf(" • Chrome: %d", data.Browsers.Chrome.TabCount)
			}
			if data.Browsers.Safari.Available {
				text += fmt.Sprintf(" • Safari: %d", data.Browsers.Safari.TabCount)
			}
			if data.Browsers.Edge.Available {
				text += fmt.Sprintf(" • Edge: %d", data.Browsers.Edge.TabCount)
			}
			fmt.Println(ui.RenderDataPoint("🌐", text))

			if len(data.Browsers.TopDomains) > 0 {
				type domainCount struct {
					domain string
					count  int
				}
				var domains []domainCount
				for domain, count := range data.Browsers.TopDomains {
					domains = append(domains, domainCount{domain, count})
				}
				sort.Slice(domains, func(i, j int) bool {
					return domains[i].count > domains[j].count
				})

				fmt.Println(ui.RenderDataPoint("📑", "Top tab domains:"))
				for i, dc := range domains {
					if i >= 5 {
						break
					}
					domainText := fmt.Sprintf("   %s (%d tab%s)", dc.domain, dc.count, pluralize(dc.count))
					fmt.Println(ui.RenderSubItem(domainText))
				}
			}
		}

		// Domain breakdown (work/distraction/neutral)
		totalCategorized := data.Browsers.WorkVisits + data.Browsers.DistractionVisits + data.Browsers.NeutralVisits
		if totalCategorized > 0 {
			workPct := int(float64(data.Browsers.WorkVisits) / float64(totalCategorized) * 100)
			distractionPct := int(float64(data.Browsers.DistractionVisits) / float64(totalCategorized) * 100)
			neutralPct := int(float64(data.Browsers.NeutralVisits) / float64(totalCategorized) * 100)

			fmt.Println(ui.RenderDataPoint("📊", "Domain breakdown:"))
			fmt.Println(ui.RenderSubItem(fmt.Sprintf("   Work: %d visits (%d%%)", data.Browsers.WorkVisits, workPct)))
			fmt.Println(ui.RenderSubItem(fmt.Sprintf("   Distraction: %d visits (%d%%)", data.Browsers.DistractionVisits, distractionPct)))
			fmt.Println(ui.RenderSubItem(fmt.Sprintf("   Neutral: %d visits (%d%%)", data.Browsers.NeutralVisits, neutralPct)))
		}
	}

	// Notifications Section
	if data.Notifications.Available && data.Notifications.TotalNotifications > 0 {
		fmt.Println()
		fmt.Println(ui.RenderHeader("NOTIFICATIONS"))

		text := fmt.Sprintf("%d notification%s today", data.Notifications.TotalNotifications, pluralize(data.Notifications.TotalNotifications))
		fmt.Println(ui.RenderDataPoint("🔔", text))

		if len(data.Notifications.TopApps) > 0 {
			fmt.Println(ui.RenderDataPoint("📱", "Top interrupting apps:"))
			for i, app := range data.Notifications.TopApps {
				if i >= 3 {
					break
				}
				appText := fmt.Sprintf("   %s (%d notification%s)", app.Name, app.Count, pluralize(app.Count))
				fmt.Println(ui.RenderSubItem(appText))
			}
		}
	}

	// Context Fragmentation Section
	if data.Fragmentation.Available {
		fmt.Println()
		fmt.Println(ui.RenderHeader("CONTEXT FRAGMENTATION"))

		text := fmt.Sprintf("%d/100 (%s)", data.Fragmentation.Score, data.Fragmentation.Level)
		fmt.Println(ui.RenderDataPoint(data.Fragmentation.Emoji, text))
	}

	// Issues/Tickets Section
	if data.Issues.Available && len(data.Issues.Issues) > 0 {
		fmt.Println()
		fmt.Println(ui.RenderHeader("ISSUES/TICKETS"))

		fmt.Println(ui.RenderDataPoint("🎫", "Issues/Tickets viewed today:"))
		for i, issue := range data.Issues.Issues {
			if i >= 10 {
				break
			}
			issueText := fmt.Sprintf("   %s (%s, %d visit%s)", issue.ID, issue.Tracker, issue.VisitCount, pluralize(issue.VisitCount))
			fmt.Println(ui.RenderSubItem(issueText))
		}
	}

	// Burnout Warnings Section
	if data.Burnout.Available && len(data.Burnout.Warnings) > 0 {
		fmt.Println()
		fmt.Println(ui.RenderHeader("WELLNESS CHECK"))

		severityOrder := map[string]int{"high": 0, "medium": 1, "low": 2}
		sortedWarnings := make([]collectors.BurnoutWarning, len(data.Burnout.Warnings))
		copy(sortedWarnings, data.Burnout.Warnings)
		sort.Slice(sortedWarnings, func(i, j int) bool {
			return severityOrder[sortedWarnings[i].Severity] < severityOrder[sortedWarnings[j].Severity]
		})

		for _, warning := range sortedWarnings {
			icon := "⚠️"
			switch warning.Type {
			case "long_day":
				icon = "⏰"
			case "high_switching":
				icon = "🔄"
			case "tab_overload":
				icon = "📑"
			case "late_night":
				icon = "🌙"
			case "no_breaks":
				icon = "😰"
			}
			fmt.Println(ui.RenderBurnoutWarning(icon, warning.Message))
		}
	}

	fmt.Println()

	if !data.Apps.Available && data.Apps.Error != nil {
		fmt.Println(ui.RenderHint("Run 'rekap init' to enable Full Disk Access for app tracking"))
	}
}

func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
