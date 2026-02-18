package main

import (
	"context"
	"fmt"
	"time"

	"github.com/alexinslc/rekap/internal/collectors"
	"github.com/alexinslc/rekap/internal/config"
	"github.com/alexinslc/rekap/internal/ui"
)

func runDemo(cfg *config.Config) {
	ui.ApplyColors(cfg)

	fmt.Println(ui.RenderTitle("🎭 rekap demo mode", false))
	fmt.Println(ui.RenderHint("Showing randomized sample data"))
	fmt.Println()

	data := buildDemoData(cfg)
	printHuman(cfg, &data)
}

func buildDemoData(cfg *config.Config) SummaryData {
	data := SummaryData{
		Uptime: collectors.UptimeResult{
			BootTime:      time.Now().Add(-8 * time.Hour),
			AwakeMinutes:  287,
			FormattedTime: "4h 47m awake",
			Available:     true,
		},
		Battery: collectors.BatteryResult{
			StartPct:   92,
			CurrentPct: 68,
			PlugCount:  1,
			Available:  true,
			IsPlugged:  false,
		},
		Screen: collectors.ScreenResult{
			ScreenOnMinutes: 660, // 11h - triggers long day warning
			Available:       true,
		},
		Apps: collectors.AppsResult{
			TopApps: []collectors.AppUsage{
				{Name: "VS Code", Minutes: 142, BundleID: "com.microsoft.VSCode"},
				{Name: "Safari", Minutes: 89, BundleID: "com.apple.Safari"},
				{Name: "Slack", Minutes: 52, BundleID: "com.tinyspeck.slackmacgap"},
				{Name: "Terminal", Minutes: 38, BundleID: "com.apple.Terminal"},
				{Name: "Chrome", Minutes: 27, BundleID: "com.google.Chrome"},
				{Name: "Notion", Minutes: 18, BundleID: "com.notion.Notion"},
				{Name: "Discord", Minutes: 12, BundleID: "com.discord.Discord"},
			},
			Source:    "ScreenTime",
			Available: true,
		},
		Focus: collectors.FocusResult{
			StreakMinutes: 87,
			AppName:       "VS Code",
			Available:     true,
		},
		Media: collectors.MediaResult{
			Track:     "Blinding Lights - The Weeknd",
			App:       "Spotify",
			Available: true,
		},
		Network: collectors.NetworkResult{
			InterfaceName: "en0",
			NetworkName:   "Home-5GHz",
			BytesReceived: 2469606195,
			BytesSent:     471859200,
			Available:     true,
		},
		Browsers: collectors.BrowsersResult{
			Chrome: collectors.BrowserResult{
				Browser:   "Chrome",
				TabCount:  58,
				Available: true,
			},
			Safari: collectors.BrowserResult{
				Browser:   "Safari",
				TabCount:  42,
				Available: true,
			},
			Edge: collectors.BrowserResult{
				Browser:   "Edge",
				TabCount:  25,
				Available: true,
			},
			TotalTabs: 125,
			TopDomains: map[string]int{
				"github.com":        8,
				"stackoverflow.com": 6,
				"mail.google.com":   5,
				"chatgpt.com":       4,
				"youtube.com":       3,
				"reddit.com":        2,
				"twitter.com":       2,
				"docs.python.org":   3,
				"linear.app":        2,
			},
			WorkVisits:        19,
			DistractionVisits: 7,
			NeutralVisits:     9,
			TotalURLsVisited:  147,
			TopHistoryDomain:  "github.com",
			TopDomainVisits:   34,
			AllIssueURLs:      []string{"PROJ-123", "PROJ-456", "org/repo#89"},
			Available:         true,
		},
		Notifications: collectors.NotificationsResult{
			TotalNotifications: 47,
			TopApps: []collectors.NotificationApp{
				{Name: "Slack", Count: 18, BundleID: "com.tinyspeck.slackmacgap"},
				{Name: "Mail", Count: 12, BundleID: "com.apple.mail"},
				{Name: "Messages", Count: 9, BundleID: "com.apple.MobileSMS"},
			},
			Available: true,
		},
		Issues: collectors.IssuesResult{
			Issues: []collectors.IssueVisit{
				{ID: "PROJ-123", Tracker: "Jira", URL: "https://company.atlassian.net/browse/PROJ-123", VisitCount: 8},
				{ID: "github.com/alexinslc/rekap/issues/42", Tracker: "GitHub", URL: "https://github.com/alexinslc/rekap/issues/42", VisitCount: 5},
				{ID: "ENG-789", Tracker: "Linear", URL: "https://linear.app/issue/ENG-789", VisitCount: 3},
			},
			Available: true,
		},
	}

	// Calculate fragmentation for demo
	fragmentationThresholds := collectors.FragmentationThresholds{
		FocusedMax:    cfg.Fragmentation.FocusedMax,
		ModerateMax:   cfg.Fragmentation.ModerateMax,
		FragmentedMin: cfg.Fragmentation.FragmentedMin,
	}
	data.Fragmentation = collectors.CalculateFragmentation(
		context.Background(),
		data.Apps,
		data.Browsers,
		data.Uptime,
		fragmentationThresholds,
	)

	// Generate burnout warnings based on demo data
	burnoutConfig := collectors.DefaultBurnoutConfig()
	data.Burnout = collectors.CollectBurnout(context.Background(), data.Screen, data.Browsers, burnoutConfig)

	return data
}
