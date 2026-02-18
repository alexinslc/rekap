package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/alexinslc/rekap/internal/collectors"
)

// JSON output structs -- separate from internal collector structs to form a stable API contract.
// Only include fields consumers need; omit Available, Error, and other implementation details.

type JSONOutput struct {
	Version         string               `json:"version"`
	Date            string               `json:"date"`
	CollectedAt     string               `json:"collected_at"`
	Uptime          *UptimeJSON          `json:"uptime,omitempty"`
	Battery         *BatteryJSON         `json:"battery,omitempty"`
	Screen          *ScreenJSON          `json:"screen,omitempty"`
	Apps            *AppsJSON            `json:"apps,omitempty"`
	Focus           *FocusJSON           `json:"focus,omitempty"`
	Media           *MediaJSON           `json:"media,omitempty"`
	Network         *NetworkJSON         `json:"network,omitempty"`
	Browsers        *BrowsersJSON        `json:"browsers,omitempty"`
	Notifications   *NotificationsJSON   `json:"notifications,omitempty"`
	Fragmentation   *FragmentationJSON   `json:"fragmentation,omitempty"`
	Issues          *IssuesJSON          `json:"issues,omitempty"`
	Burnout         *BurnoutJSON         `json:"burnout,omitempty"`
	ContextOverload *ContextOverloadJSON `json:"context_overload,omitempty"`
}

type UptimeJSON struct {
	AwakeMinutes int   `json:"awake_minutes"`
	BootTimeUnix int64 `json:"boot_time_unix"`
}

type BatteryJSON struct {
	StartPct   int  `json:"start_pct"`
	CurrentPct int  `json:"current_pct"`
	PlugEvents int  `json:"plug_events"`
	IsPlugged  bool `json:"is_plugged"`
}

type ScreenJSON struct {
	ScreenOnMinutes    int `json:"screen_on_minutes"`
	LockCount          int `json:"lock_count"`
	AvgMinsBetweenLock int `json:"avg_mins_between_locks"`
}

type AppJSON struct {
	Name     string `json:"name"`
	Minutes  int    `json:"minutes"`
	BundleID string `json:"bundle_id"`
}

type AppsJSON struct {
	TopApps              []AppJSON `json:"top_apps"`
	TotalSwitches        int       `json:"total_switches"`
	SwitchesPerHour      float64   `json:"switches_per_hour"`
	AvgMinsBetweenSwitch float64   `json:"avg_mins_between_switches"`
}

type FocusJSON struct {
	StreakMinutes int    `json:"streak_minutes"`
	AppName       string `json:"app_name"`
}

type MediaJSON struct {
	Track string `json:"track"`
	App   string `json:"app"`
}

type NetworkJSON struct {
	Interface     string `json:"interface"`
	NetworkName   string `json:"network_name"`
	BytesReceived int64  `json:"bytes_received"`
	BytesSent     int64  `json:"bytes_sent"`
	SinceBoot     bool   `json:"since_boot"`
}

type BrowserJSON struct {
	Tabs int `json:"tabs"`
}

type BrowsersJSON struct {
	TotalTabs         int          `json:"total_tabs"`
	Chrome            *BrowserJSON `json:"chrome,omitempty"`
	Safari            *BrowserJSON `json:"safari,omitempty"`
	Edge              *BrowserJSON `json:"edge,omitempty"`
	URLsVisited       int          `json:"urls_visited"`
	TopDomain         string       `json:"top_domain,omitempty"`
	TopDomainVisits   int          `json:"top_domain_visits,omitempty"`
	WorkVisits        int          `json:"work_visits"`
	DistractionVisits int          `json:"distraction_visits"`
	NeutralVisits     int          `json:"neutral_visits"`
	IssuesViewed      []string     `json:"issues_viewed,omitempty"`
}

type NotificationAppJSON struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type NotificationsJSON struct {
	Total   int                   `json:"total"`
	TopApps []NotificationAppJSON `json:"top_apps,omitempty"`
}

type FragmentationJSON struct {
	Score int    `json:"score"`
	Level string `json:"level"`
}

type IssueJSON struct {
	ID         string `json:"id"`
	Tracker    string `json:"tracker"`
	URL        string `json:"url"`
	VisitCount int    `json:"visit_count"`
}

type IssuesJSON struct {
	Issues []IssueJSON `json:"issues"`
}

type BurnoutWarningJSON struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

type BurnoutJSON struct {
	Warnings []BurnoutWarningJSON `json:"warnings"`
}

type ContextOverloadJSON struct {
	IsOverloaded bool   `json:"is_overloaded"`
	Message      string `json:"message,omitempty"`
}

func printJSON(data *SummaryData) {
	out := JSONOutput{
		Version:     version,
		Date:        time.Now().Format("2006-01-02"),
		CollectedAt: time.Now().Format(time.RFC3339),
	}

	if data.Uptime.Available {
		out.Uptime = &UptimeJSON{
			AwakeMinutes: data.Uptime.AwakeMinutes,
			BootTimeUnix: data.Uptime.BootTime.Unix(),
		}
	}

	if data.Battery.Available {
		out.Battery = &BatteryJSON{
			StartPct:   data.Battery.StartPct,
			CurrentPct: data.Battery.CurrentPct,
			PlugEvents: data.Battery.PlugCount,
			IsPlugged:  data.Battery.IsPlugged,
		}
	}

	if data.Screen.Available {
		out.Screen = &ScreenJSON{
			ScreenOnMinutes:    data.Screen.ScreenOnMinutes,
			LockCount:          data.Screen.LockCount,
			AvgMinsBetweenLock: data.Screen.AvgMinsBetweenLock,
		}
	}

	if data.Apps.Available {
		appsJSON := &AppsJSON{}
		for _, app := range data.Apps.TopApps {
			appsJSON.TopApps = append(appsJSON.TopApps, AppJSON{
				Name:     app.Name,
				Minutes:  app.Minutes,
				BundleID: app.BundleID,
			})
		}
		if data.Apps.SwitchingAvailable {
			appsJSON.TotalSwitches = data.Apps.TotalSwitches
			appsJSON.SwitchesPerHour = data.Apps.SwitchesPerHour
			appsJSON.AvgMinsBetweenSwitch = data.Apps.AvgMinsBetween
		}
		out.Apps = appsJSON
	}

	if data.Focus.Available {
		out.Focus = &FocusJSON{
			StreakMinutes: data.Focus.StreakMinutes,
			AppName:       data.Focus.AppName,
		}
	}

	if data.Media.Available {
		out.Media = &MediaJSON{
			Track: data.Media.Track,
			App:   data.Media.App,
		}
	}

	if data.Network.Available {
		out.Network = &NetworkJSON{
			Interface:     data.Network.InterfaceName,
			NetworkName:   data.Network.NetworkName,
			BytesReceived: data.Network.BytesReceived,
			BytesSent:     data.Network.BytesSent,
			SinceBoot:     data.Network.SinceBoot,
		}
	}

	if data.Browsers.Available {
		browsersJSON := &BrowsersJSON{
			TotalTabs:         data.Browsers.TotalTabs,
			URLsVisited:       data.Browsers.TotalURLsVisited,
			TopDomain:         data.Browsers.TopHistoryDomain,
			TopDomainVisits:   data.Browsers.TopDomainVisits,
			WorkVisits:        data.Browsers.WorkVisits,
			DistractionVisits: data.Browsers.DistractionVisits,
			NeutralVisits:     data.Browsers.NeutralVisits,
			IssuesViewed:      data.Browsers.AllIssueURLs,
		}
		if data.Browsers.Chrome.Available {
			browsersJSON.Chrome = &BrowserJSON{Tabs: data.Browsers.Chrome.TabCount}
		}
		if data.Browsers.Safari.Available {
			browsersJSON.Safari = &BrowserJSON{Tabs: data.Browsers.Safari.TabCount}
		}
		if data.Browsers.Edge.Available {
			browsersJSON.Edge = &BrowserJSON{Tabs: data.Browsers.Edge.TabCount}
		}
		out.Browsers = browsersJSON
	}

	if data.Notifications.Available {
		notifJSON := &NotificationsJSON{
			Total: data.Notifications.TotalNotifications,
		}
		for _, app := range data.Notifications.TopApps {
			notifJSON.TopApps = append(notifJSON.TopApps, NotificationAppJSON{
				Name:  app.Name,
				Count: app.Count,
			})
		}
		out.Notifications = notifJSON
	}

	if data.Fragmentation.Available {
		out.Fragmentation = &FragmentationJSON{
			Score: data.Fragmentation.Score,
			Level: data.Fragmentation.Level,
		}
	}

	if data.Issues.Available && len(data.Issues.Issues) > 0 {
		issuesJSON := &IssuesJSON{}
		for _, issue := range data.Issues.Issues {
			issuesJSON.Issues = append(issuesJSON.Issues, IssueJSON{
				ID:         issue.ID,
				Tracker:    issue.Tracker,
				URL:        issue.URL,
				VisitCount: issue.VisitCount,
			})
		}
		out.Issues = issuesJSON
	}

	if data.Burnout.Available && len(data.Burnout.Warnings) > 0 {
		burnoutJSON := &BurnoutJSON{}
		for _, w := range data.Burnout.Warnings {
			burnoutJSON.Warnings = append(burnoutJSON.Warnings, BurnoutWarningJSON{
				Type:     w.Type,
				Severity: w.Severity,
				Message:  w.Message,
			})
		}
		out.Burnout = burnoutJSON
	}

	overload := collectors.CheckContextOverload(data.Apps, data.Browsers)
	out.ContextOverload = &ContextOverloadJSON{
		IsOverloaded: overload.IsOverloaded,
		Message:      overload.WarningMessage,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "rekap: json encode error: %v\n", err)
		os.Exit(1)
	}
}
