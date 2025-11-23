package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/alexinslc/rekap/internal/collectors"
	"github.com/alexinslc/rekap/internal/config"
	"github.com/alexinslc/rekap/internal/permissions"
	"github.com/alexinslc/rekap/internal/theme"
	"github.com/alexinslc/rekap/internal/tui"
	"github.com/alexinslc/rekap/internal/ui"
	"github.com/charmbracelet/fang"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

func main() {
	var quietFlag bool
	var themeFlag string
	var accessibleFlag bool

	rootCmd := &cobra.Command{
		Use:   "rekap",
		Short: "Daily Mac Activity Summary",
		Long:  `A single-binary macOS CLI that summarizes today's computer activity in a friendly, animated terminal UI.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			cfg, err := config.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
				cfg = config.Default()
			}

			// Apply theme if specified
			if themeFlag != "" {
				t, err := theme.Load(themeFlag)
				if err != nil {
					return fmt.Errorf("failed to load theme: %w", err)
				}
				cfg.ApplyTheme(t)
			}

			// Override config with flag if provided
			if accessibleFlag {
				cfg.Accessibility.Enabled = true
				cfg.Accessibility.HighContrast = true
			}

			runSummary(quietFlag, cfg)
			return nil
		},
	}

	rootCmd.Flags().BoolVarP(&quietFlag, "quiet", "q", false, "Output machine-parsable key=value format")
	rootCmd.Flags().StringVar(&themeFlag, "theme", "", "Color theme (built-in: default, minimal, hacker, pastel, nord, dracula, solarized) or path to theme file")
	rootCmd.PersistentFlags().BoolVar(&accessibleFlag, "accessible", false, "Enable accessibility mode (color-blind friendly, high contrast)")

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Permission setup wizard",
		Long:  `Run the guided permission setup wizard to enable Full Disk Access and other permissions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit()
		},
	}

	doctorCmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check capabilities and permissions",
		Long:  `Check the current status of permissions and capabilities.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			runDoctor()
			return nil
		},
	}

	var demoThemeFlag string
	demoCmd := &cobra.Command{
		Use:   "demo",
		Short: "See sample output with fake data",
		Long:  `Display a demo with randomized sample data to preview the output format.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config for demo too
			cfg, err := config.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
				cfg = config.Default()
			}

			// Apply theme if specified
			if demoThemeFlag != "" {
				t, err := theme.Load(demoThemeFlag)
				if err != nil {
					return fmt.Errorf("failed to load theme: %w", err)
				}
				cfg.ApplyTheme(t)
			}

			// Override config with flag if provided
			if accessibleFlag {
				cfg.Accessibility.Enabled = true
				cfg.Accessibility.HighContrast = true
			}

			runDemo(cfg)
			return nil
		},
	}
	demoCmd.Flags().StringVar(&demoThemeFlag, "theme", "", "Color theme (built-in: default, minimal, hacker, pastel, nord, dracula, solarized) or path to theme file")

	// Add themes command
	themesCmd := &cobra.Command{
		Use:   "themes",
		Short: "Manage color themes",
		Long:  "List, preview, and apply color themes.",
	}

	themesListCmd := &cobra.Command{
		Use:   "list",
		Short: "List available themes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return listThemes()
		},
	}

	themesPreviewCmd := &cobra.Command{
		Use:   "preview",
		Short: "Interactive theme previewer",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runThemePreview()
		},
	}

	themesSetCmd := &cobra.Command{
		Use:   "set [theme-name]",
		Short: "Apply theme to config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return setTheme(args[0])
		},
	}

	themesCmd.AddCommand(themesListCmd, themesPreviewCmd, themesSetCmd)
	rootCmd.AddCommand(initCmd, doctorCmd, demoCmd, themesCmd)

	// Add fang configuration
	fang.Configure(rootCmd, "REKAP")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runInit() error {
	return permissions.RequestFlow()
}

func runDoctor() {
	// Enhanced styling for doctor command
	fmt.Println(ui.RenderTitle("🩺 rekap capabilities check", false))
	fmt.Println()

	caps := permissions.Check()
	fmt.Println(permissions.FormatCapabilities(caps))
	fmt.Println()

	if !caps.FullDiskAccess {
		fmt.Println(ui.RenderHint("Run 'rekap init' to enable Full Disk Access for app tracking"))
	} else {
		fmt.Println(ui.RenderSuccess("All major permissions granted!"))
	}
}

func runDemo(cfg *config.Config) {
	// Apply colors from config
	ui.ApplyColors(cfg)

	// Enhanced styling for demo mode
	fmt.Println(ui.RenderTitle("🎭 rekap demo mode", false))
	fmt.Println(ui.RenderHint("Showing randomized sample data"))
	fmt.Println()

	// Generate realistic demo data
	demoUptime := collectors.UptimeResult{
		BootTime:      time.Now().Add(-8 * time.Hour),
		AwakeMinutes:  287, // 4h 47m
		FormattedTime: "4h 47m awake",
		Available:     true,
	}

	demoBattery := collectors.BatteryResult{
		StartPct:   92,
		CurrentPct: 68,
		PlugCount:  1,
		Available:  true,
		IsPlugged:  false,
	}

	demoApps := collectors.AppsResult{
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
	}

	demoFocus := collectors.FocusResult{
		StreakMinutes: 87, // 1h 27m
		AppName:       "VS Code",
		Available:     true,
	}

	demoMedia := collectors.MediaResult{
		Track:           "Blinding Lights - The Weeknd",
		App:             "Spotify",
		DurationMinutes: 18,
		Available:       true,
	}

	demoNetwork := collectors.NetworkResult{
		InterfaceName: "en0",
		NetworkName:   "Home-5GHz",
		BytesReceived: 2469606195, // ~2.3 GB
		BytesSent:     471859200,  // ~450 MB
		Available:     true,
	}

	// Demo burnout warnings - create a scenario with 11h screen time to trigger long day warning
	demoScreenLongDay := collectors.ScreenResult{
		ScreenOnMinutes: 660, // 11h
		Available:       true,
	}

	demoBrowsers := collectors.BrowsersResult{
		Chrome: collectors.BrowserResult{
			Browser:         "Chrome",
			TabCount:        28,
			Available:       true,
			URLsVisited:     89,
			TopDomain:       "github.com",
			TopDomainVisits: 34,
			IssueURLs:       []string{"org/repo#89", "PROJ-123"},
		},
		Safari: collectors.BrowserResult{
			Browser:         "Safari",
			TabCount:        12,
			Available:       true,
			URLsVisited:     42,
			TopDomain:       "stackoverflow.com",
			TopDomainVisits: 18,
			IssueURLs:       []string{"PROJ-456"},
		},
		Edge: collectors.BrowserResult{
			Browser:         "Edge",
			TabCount:        5,
			Available:       true,
			URLsVisited:     16,
			TopDomain:       "mail.google.com",
			TopDomainVisits: 12,
		},
		TotalTabs: 45,
		TopDomains: map[string]int{
			"github.com":        8,
			"stackoverflow.com": 6,
			"mail.google.com":   5,
			"chatgpt.com":       4,
			"youtube.com":       3,
			"reddit.com":        3,
			"twitter.com":       2,
			"linkedin.com":      2,
			"docs.google.com":   2,
			"slack.com":         2,
			"notion.so":         2,
		},
		WorkVisits:        19, // github(8) + stackoverflow(6) + docs.python.org(3) + linear.app(2)
		DistractionVisits: 7,  // youtube(3) + reddit(2) + twitter(2)
		NeutralVisits:     9,  // mail.google.com(5) + chatgpt.com(4)
		TotalURLsVisited:  147,
		TopHistoryDomain:  "github.com",
		TopDomainVisits:   34,
		AllIssueURLs:      []string{"PROJ-123", "PROJ-456", "org/repo#89"},
		Available:         true,
	}

	// Demo with 125 tabs to trigger tab overload
	demoBrowsersOverload := collectors.BrowsersResult{
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
		WorkVisits:        19, // github(8) + stackoverflow(6) + docs.python.org(3) + linear.app(2)
		DistractionVisits: 7,  // youtube(3) + reddit(2) + twitter(2)
		NeutralVisits:     9,  // mail.google.com(5) + chatgpt.com(4)
		TotalURLsVisited:  147,
		TopHistoryDomain:  "github.com",
		TopDomainVisits:   34,
		AllIssueURLs:      []string{"PROJ-123", "PROJ-456", "org/repo#89"},
		Available:         true,
	}

	demoNotifications := collectors.NotificationsResult{
		TotalNotifications: 47,
		TopApps: []collectors.NotificationApp{
			{Name: "Slack", Count: 18, BundleID: "com.tinyspeck.slackmacgap"},
			{Name: "Mail", Count: 12, BundleID: "com.apple.mail"},
			{Name: "Messages", Count: 9, BundleID: "com.apple.MobileSMS"},
		},
		Available: true,
	}

	demoIssues := collectors.IssuesResult{
		Issues: []collectors.IssueVisit{
			{ID: "PROJ-123", Tracker: "Jira", URL: "https://company.atlassian.net/browse/PROJ-123", VisitCount: 8},
			{ID: "github.com/alexinslc/rekap/issues/42", Tracker: "GitHub", URL: "https://github.com/alexinslc/rekap/issues/42", VisitCount: 5},
			{ID: "ENG-789", Tracker: "Linear", URL: "https://linear.app/issue/ENG-789", VisitCount: 3},
		},
		Available: true,
	}

	// Calculate fragmentation for demo
	fragmentationThresholds := collectors.FragmentationThresholds{
		FocusedMax:    cfg.Fragmentation.FocusedMax,
		ModerateMax:   cfg.Fragmentation.ModerateMax,
		FragmentedMin: cfg.Fragmentation.FragmentedMin,
	}
	demoFragmentation := collectors.CalculateFragmentation(
		context.Background(),
		demoApps,
		demoBrowsers,
		demoUptime,
		fragmentationThresholds,
	)

	// Generate burnout warnings based on demo data
	ctx := context.Background()
	burnoutConfig := collectors.DefaultBurnoutConfig()
	demoBurnout := collectors.CollectBurnout(ctx, demoScreenLongDay, demoBrowsersOverload, burnoutConfig)

	// Show in human-friendly format (use the modified screen and browsers for demo)
	printHuman(cfg, demoUptime, demoBattery, demoScreenLongDay, demoApps, demoFocus, demoMedia, demoNetwork, demoBrowsersOverload, demoNotifications, demoIssues, demoFragmentation, demoBurnout)
}

func runSummary(quiet bool, cfg *config.Config) {
	// Apply colors from config (for non-quiet mode)
	if !quiet {
		ui.ApplyColors(cfg)
	}

	// Create context with timeout for all collectors
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Collect data from all sources concurrently
	uptimeCh := make(chan collectors.UptimeResult, 1)
	batteryCh := make(chan collectors.BatteryResult, 1)
	screenCh := make(chan collectors.ScreenResult, 1)
	appsCh := make(chan collectors.AppsResult, 1)
	focusCh := make(chan collectors.FocusResult, 1)
	mediaCh := make(chan collectors.MediaResult, 1)
	networkCh := make(chan collectors.NetworkResult, 1)
	browsersCh := make(chan collectors.BrowsersResult, 1)
	issuesCh := make(chan collectors.IssuesResult, 1)
	notificationsCh := make(chan collectors.NotificationsResult, 1)

	go func() { uptimeCh <- collectors.CollectUptime(ctx) }()
	go func() { batteryCh <- collectors.CollectBattery(ctx) }()
	go func() { screenCh <- collectors.CollectScreen(ctx) }()
	go func() { appsCh <- collectors.CollectApps(ctx, cfg.Tracking.ExcludeApps) }()
	go func() { focusCh <- collectors.CollectFocus(ctx) }()
	go func() { mediaCh <- collectors.CollectMedia(ctx) }()
	go func() { networkCh <- collectors.CollectNetwork(ctx) }()
	go func() { browsersCh <- collectors.CollectBrowserTabs(ctx, cfg) }()
	go func() { issuesCh <- collectors.CollectIssues(ctx) }()
	go func() { notificationsCh <- collectors.CollectNotifications(ctx) }()

	// Wait for all results
	uptimeResult := <-uptimeCh
	batteryResult := <-batteryCh
	screenResult := <-screenCh
	appsResult := <-appsCh
	focusResult := <-focusCh
	mediaResult := <-mediaCh
	networkResult := <-networkCh
	browsersResult := <-browsersCh
	issuesResult := <-issuesCh
	notificationsResult := <-notificationsCh

	// Calculate fragmentation score after collecting data
	fragmentationThresholds := collectors.FragmentationThresholds{
		FocusedMax:    cfg.Fragmentation.FocusedMax,
		ModerateMax:   cfg.Fragmentation.ModerateMax,
		FragmentedMin: cfg.Fragmentation.FragmentedMin,
	}
	fragmentationResult := collectors.CalculateFragmentation(ctx, appsResult, browsersResult, uptimeResult, fragmentationThresholds)

	// Analyze burnout patterns after collecting primary data
	burnoutConfig := collectors.DefaultBurnoutConfig()
	burnoutResult := collectors.CollectBurnout(ctx, screenResult, browsersResult, burnoutConfig)

	if quiet {
		// Machine-parsable output
		printQuiet(uptimeResult, batteryResult, screenResult, appsResult, focusResult, mediaResult, networkResult, browsersResult, issuesResult, notificationsResult, fragmentationResult)
	} else {
		// Human-friendly output
		printHuman(cfg, uptimeResult, batteryResult, screenResult, appsResult, focusResult, mediaResult, networkResult, browsersResult, notificationsResult, issuesResult, fragmentationResult, burnoutResult)
	}
}

func printQuiet(uptime collectors.UptimeResult, battery collectors.BatteryResult, screen collectors.ScreenResult, apps collectors.AppsResult, focus collectors.FocusResult, media collectors.MediaResult, network collectors.NetworkResult, browsers collectors.BrowsersResult, issues collectors.IssuesResult, notifications collectors.NotificationsResult, fragmentation collectors.FragmentationResult) {
	if uptime.Available {
		fmt.Printf("awake_minutes=%d\n", uptime.AwakeMinutes)
		fmt.Printf("boot_time=%d\n", uptime.BootTime.Unix())
	}

	if battery.Available {
		fmt.Printf("battery_start_pct=%d\n", battery.StartPct)
		fmt.Printf("battery_now_pct=%d\n", battery.CurrentPct)
		fmt.Printf("plug_events=%d\n", battery.PlugCount)
		if battery.IsPlugged {
			fmt.Printf("is_plugged=1\n")
		} else {
			fmt.Printf("is_plugged=0\n")
		}
	}

	if screen.Available {
		fmt.Printf("screen_on_minutes=%d\n", screen.ScreenOnMinutes)
		if screen.LockCount > 0 {
			fmt.Printf("screen_lock_count=%d\n", screen.LockCount)
			fmt.Printf("avg_mins_between_locks=%d\n", screen.AvgMinsBetweenLock)
		}
	}

	if apps.Available {
		for i, app := range apps.TopApps {
			if i >= 3 {
				break
			}
			fmt.Printf("top_app_%d=%s\n", i+1, app.Name)
			fmt.Printf("top_app_%d_minutes=%d\n", i+1, app.Minutes)
		}
	}

	if focus.Available {
		fmt.Printf("focus_streak_minutes=%d\n", focus.StreakMinutes)
		fmt.Printf("focus_streak_app=%s\n", focus.AppName)
	}

	if media.Available {
		fmt.Printf("media_track=%s\n", media.Track)
		fmt.Printf("media_app=%s\n", media.App)
	}

	if network.Available {
		fmt.Printf("network_interface=%s\n", network.InterfaceName)
		fmt.Printf("network_name=%s\n", network.NetworkName)
		fmt.Printf("network_bytes_received=%d\n", network.BytesReceived)
		fmt.Printf("network_bytes_sent=%d\n", network.BytesSent)
	}

	if browsers.Available {
		fmt.Printf("browser_total_tabs=%d\n", browsers.TotalTabs)
		if browsers.Chrome.Available {
			fmt.Printf("browser_chrome_tabs=%d\n", browsers.Chrome.TabCount)
		}
		if browsers.Safari.Available {
			fmt.Printf("browser_safari_tabs=%d\n", browsers.Safari.TabCount)
		}
		if browsers.Edge.Available {
			fmt.Printf("browser_edge_tabs=%d\n", browsers.Edge.TabCount)
		}
		// Domain categorization stats
		totalCategorized := browsers.WorkVisits + browsers.DistractionVisits + browsers.NeutralVisits
		if totalCategorized > 0 {
			fmt.Printf("browser_work_visits=%d\n", browsers.WorkVisits)
			fmt.Printf("browser_distraction_visits=%d\n", browsers.DistractionVisits)
			fmt.Printf("browser_neutral_visits=%d\n", browsers.NeutralVisits)
		}
		// History data
		if browsers.TotalURLsVisited > 0 {
			fmt.Printf("browser_urls_visited=%d\n", browsers.TotalURLsVisited)
		}
		if browsers.TopHistoryDomain != "" {
			fmt.Printf("browser_top_domain=%s\n", browsers.TopHistoryDomain)
			fmt.Printf("browser_top_domain_visits=%d\n", browsers.TopDomainVisits)
		}
		if len(browsers.AllIssueURLs) > 0 {
			fmt.Printf("browser_issues_viewed=%d\n", len(browsers.AllIssueURLs))
		}
	}

	if notifications.Available {
		fmt.Printf("notifications_total=%d\n", notifications.TotalNotifications)
		for i, app := range notifications.TopApps {
			if i >= 3 {
				break
			}
			fmt.Printf("notification_app_%d=%s\n", i+1, app.Name)
			fmt.Printf("notification_app_%d_count=%d\n", i+1, app.Count)
		}
	}

	if fragmentation.Available {
		fmt.Printf("fragmentation_score=%d\n", fragmentation.Score)
		fmt.Printf("fragmentation_level=%s\n", fragmentation.Level)
	}

	if issues.Available {
		fmt.Printf("issues_count=%d\n", len(issues.Issues))
		for i, issue := range issues.Issues {
			if i >= 10 {
				break
			}
			fmt.Printf("issue_%d_id=%s\n", i+1, issue.ID)
			fmt.Printf("issue_%d_tracker=%s\n", i+1, issue.Tracker)
			fmt.Printf("issue_%d_visits=%d\n", i+1, issue.VisitCount)
		}
	}

	// Check for context overload
	overload := collectors.CheckContextOverload(apps, browsers)
	if overload.IsOverloaded {
		fmt.Printf("context_overload=1\n")
		fmt.Printf("context_overload_message=%s\n", overload.WarningMessage)
	} else {
		fmt.Printf("context_overload=0\n")
	}
}

func printHuman(cfg *config.Config, uptime collectors.UptimeResult, battery collectors.BatteryResult, screen collectors.ScreenResult, apps collectors.AppsResult, focus collectors.FocusResult, media collectors.MediaResult, network collectors.NetworkResult, browsers collectors.BrowsersResult, notifications collectors.NotificationsResult, issues collectors.IssuesResult, fragmentation collectors.FragmentationResult, burnout collectors.BurnoutResult) {
	// Render title
	title := ui.RenderTitle("📊 Today's rekap", ui.IsTTY())
	if title != "" {
		fmt.Println(title)
	}
	fmt.Println()

	// Check for context overload
	overload := collectors.CheckContextOverload(apps, browsers)
	if overload.IsOverloaded {
		fmt.Println(ui.RenderWarning("Context overload: " + overload.WarningMessage))
		fmt.Println()
	}

	// Build summary line
	var summaryParts []string

	if screen.Available {
		summaryParts = append(summaryParts, ui.FormatDuration(screen.ScreenOnMinutes)+" screen-on")
	}

	if apps.Available && len(apps.TopApps) > 0 {
		appList := []string{}
		for i, app := range apps.TopApps {
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

	if uptime.Available {
		text := fmt.Sprintf("Active since %s • %s",
			ui.FormatTime(uptime.BootTime, cfg.Display.TimeFormat),
			uptime.FormattedTime)
		fmt.Println(ui.RenderDataPoint("⏰", text))
	}

	if battery.Available && cfg.ShouldShowBattery() {
		status := "discharging"
		if battery.IsPlugged {
			status = "plugged in"
		}
		var text string
		if battery.StartPct != battery.CurrentPct {
			text = fmt.Sprintf("%d%% → %d%% • %s", battery.StartPct, battery.CurrentPct, status)
		} else {
			text = fmt.Sprintf("%d%% • %s", battery.CurrentPct, status)
		}
		fmt.Println(ui.RenderDataPoint("🔋", text))

		if battery.PlugCount > 0 {
			plugText := fmt.Sprintf("%d plug event(s) today", battery.PlugCount)
			fmt.Println(ui.RenderDataPoint("🔌", plugText))
		}
	}

	// Screen lock events
	if screen.Available && screen.LockCount > 0 {
		var lockText string
		if screen.AvgMinsBetweenLock > 0 {
			lockText = fmt.Sprintf("Screen locked %d time%s (avg %s between breaks)",
				screen.LockCount,
				pluralize(screen.LockCount),
				ui.FormatDuration(screen.AvgMinsBetweenLock))
		} else {
			lockText = fmt.Sprintf("Screen locked %d time%s today",
				screen.LockCount,
				pluralize(screen.LockCount))
		}
		fmt.Println(ui.RenderDataPoint("🔒", lockText))
	}

	// Productivity Section
	if focus.Available || (apps.Available && len(apps.TopApps) > 0) {
		fmt.Println()
		fmt.Println(ui.RenderHeader("PRODUCTIVITY"))

		if focus.Available {
			text := fmt.Sprintf("Best focus: %s in %s", ui.FormatDuration(focus.StreakMinutes), focus.AppName)
			fmt.Println(ui.RenderHighlight("⏱️ ", text))
		}

		if apps.Available && len(apps.TopApps) > 0 {
			for i, app := range apps.TopApps {
				if i >= 3 {
					break
				}
				appText := fmt.Sprintf("%s • %s", app.Name, ui.FormatDuration(app.Minutes))
				fmt.Println(ui.RenderDataPoint("📱", appText))
			}
		}
	}

	// Media Section
	if media.Available && cfg.ShouldShowMedia() {
		fmt.Println()
		fmt.Println(ui.RenderHeader("NOW PLAYING"))
		text := fmt.Sprintf("\"%s\" in %s", media.Track, media.App)
		fmt.Println(ui.RenderDataPoint("🎵", text))
	}

	// Network Activity Section
	if network.Available {
		fmt.Println()
		fmt.Println(ui.RenderHeader("NETWORK ACTIVITY"))

		// Display network name and data transfer
		text := fmt.Sprintf("%s: \"%s\" • %s down / %s up",
			network.InterfaceName,
			network.NetworkName,
			collectors.FormatBytes(network.BytesReceived),
			collectors.FormatBytes(network.BytesSent))
		fmt.Println(ui.RenderDataPoint("🌐", text))
	}

	// Browser Activity Section (tabs + history)
	if browsers.Available && (browsers.TotalTabs > 0 || browsers.TotalURLsVisited > 0) {
		fmt.Println()
		fmt.Println(ui.RenderHeader("BROWSER ACTIVITY"))

		// History summary
		if browsers.TotalURLsVisited > 0 {
			historyText := fmt.Sprintf("%d URLs visited today", browsers.TotalURLsVisited)
			if browsers.TopHistoryDomain != "" {
				historyText += fmt.Sprintf(" • Top: %s (%d visit%s)",
					browsers.TopHistoryDomain,
					browsers.TopDomainVisits,
					pluralize(browsers.TopDomainVisits))
			}
			fmt.Println(ui.RenderDataPoint("📊", historyText))

			// Show issue URLs if any
			if len(browsers.AllIssueURLs) > 0 {
				issueText := fmt.Sprintf("Issues viewed: %s", collectors.FormatIssueURLs(browsers.AllIssueURLs))
				fmt.Println(ui.RenderDataPoint("🎫", issueText))
			}
		}

		// Open tabs count
		if browsers.TotalTabs > 0 {
			text := fmt.Sprintf("%d tabs open", browsers.TotalTabs)
			if browsers.Chrome.Available {
				text += fmt.Sprintf(" • Chrome: %d", browsers.Chrome.TabCount)
			}
			if browsers.Safari.Available {
				text += fmt.Sprintf(" • Safari: %d", browsers.Safari.TabCount)
			}
			if browsers.Edge.Available {
				text += fmt.Sprintf(" • Edge: %d", browsers.Edge.TabCount)
			}
			fmt.Println(ui.RenderDataPoint("🌐", text))

			// Top domains from tabs
			if len(browsers.TopDomains) > 0 {
				type domainCount struct {
					domain string
					count  int
				}
				var domains []domainCount
				for domain, count := range browsers.TopDomains {
					domains = append(domains, domainCount{domain, count})
				}
				// Sort by count descending
				sort.Slice(domains, func(i, j int) bool {
					return domains[i].count > domains[j].count
				})

				// Show top 5 domains
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
	}

	// Notifications Section
	if notifications.Available && notifications.TotalNotifications > 0 {
		fmt.Println()
		fmt.Println(ui.RenderHeader("NOTIFICATIONS"))

		// Total notifications
		text := fmt.Sprintf("%d notification%s today", notifications.TotalNotifications, pluralize(notifications.TotalNotifications))
		fmt.Println(ui.RenderDataPoint("🔔", text))

		// Top apps by notification count
		if len(notifications.TopApps) > 0 {
			fmt.Println(ui.RenderDataPoint("📱", "Top interrupting apps:"))
			for i, app := range notifications.TopApps {
				if i >= 3 {
					break
				}
				appText := fmt.Sprintf("   %s (%d notification%s)", app.Name, app.Count, pluralize(app.Count))
				fmt.Println(ui.RenderSubItem(appText))
			}
		}

		// Domain breakdown
		totalCategorized := browsers.WorkVisits + browsers.DistractionVisits + browsers.NeutralVisits
		if totalCategorized > 0 {
			workPct := int(float64(browsers.WorkVisits) / float64(totalCategorized) * 100)
			distractionPct := int(float64(browsers.DistractionVisits) / float64(totalCategorized) * 100)
			neutralPct := int(float64(browsers.NeutralVisits) / float64(totalCategorized) * 100)

			fmt.Println(ui.RenderDataPoint("📊", "Domain breakdown:"))
			workText := fmt.Sprintf("   Work: %d visits (%d%%)", browsers.WorkVisits, workPct)
			fmt.Println(ui.RenderSubItem(workText))
			distractionText := fmt.Sprintf("   Distraction: %d visits (%d%%)", browsers.DistractionVisits, distractionPct)
			fmt.Println(ui.RenderSubItem(distractionText))
			neutralText := fmt.Sprintf("   Neutral: %d visits (%d%%)", browsers.NeutralVisits, neutralPct)
			fmt.Println(ui.RenderSubItem(neutralText))
		}
	}

	// Context Fragmentation Section
	if fragmentation.Available {
		fmt.Println()
		fmt.Println(ui.RenderHeader("CONTEXT FRAGMENTATION"))

		text := fmt.Sprintf("%d/100 (%s)", fragmentation.Score, fragmentation.Level)
		fmt.Println(ui.RenderDataPoint(fragmentation.Emoji, text))
	}

	// Issues/Tickets Section
	if issues.Available && len(issues.Issues) > 0 {
		fmt.Println()
		fmt.Println(ui.RenderHeader("ISSUES/TICKETS"))

		fmt.Println(ui.RenderDataPoint("🎫", "Issues/Tickets viewed today:"))
		for i, issue := range issues.Issues {
			if i >= 10 {
				break
			}
			issueText := fmt.Sprintf("   %s (%s, %d visit%s)", issue.ID, issue.Tracker, issue.VisitCount, pluralize(issue.VisitCount))
			fmt.Println(ui.RenderSubItem(issueText))
		}
	}

	// Burnout Warnings Section (subtle, only if warnings exist)
	if burnout.Available && len(burnout.Warnings) > 0 {
		fmt.Println()
		fmt.Println(ui.RenderHeader("WELLNESS CHECK"))

		// Sort warnings by severity (high > medium > low)
		severityOrder := map[string]int{"high": 0, "medium": 1, "low": 2}
		sortedWarnings := make([]collectors.BurnoutWarning, len(burnout.Warnings))
		copy(sortedWarnings, burnout.Warnings)
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

	// Show hints for missing data
	if !apps.Available && apps.Error != nil {
		fmt.Println(ui.RenderHint("Run 'rekap init' to enable Full Disk Access for app tracking"))
	}
}

func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// listThemes lists all available themes with color samples
func listThemes() error {
	themes := theme.ListBuiltIn()
	
	// Create a temporary config to apply colors
	cfg := config.Default()
	ui.ApplyColors(cfg)
	
	fmt.Println(ui.RenderTitle("Available Themes", false))
	fmt.Println()
	
	for _, name := range themes {
		t, err := theme.Load(name)
		if err != nil {
			continue
		}
		
		// Apply theme temporarily to get colors
		tempCfg := config.Default()
		tempCfg.ApplyTheme(t)
		ui.ApplyColors(tempCfg)
		
		// Create color samples
		primarySample := lipgloss.NewStyle().
			Background(lipgloss.Color(t.Colors.Primary)).
			Foreground(lipgloss.Color(t.Colors.Text)).
			Render("  Primary  ")
		
		secondarySample := lipgloss.NewStyle().
			Background(lipgloss.Color(t.Colors.Secondary)).
			Foreground(lipgloss.Color(t.Colors.Text)).
			Render("  Secondary  ")
		
		accentSample := lipgloss.NewStyle().
			Background(lipgloss.Color(t.Colors.Accent)).
			Foreground(lipgloss.Color("#000")). // Use black text for better contrast on accent
			Render("  Accent  ")
		
		// Display theme info
		fmt.Printf("%s %s\n", ui.RenderBold(name), t.Description)
		fmt.Printf("  %s %s %s\n\n", primarySample, secondarySample, accentSample)
	}
	
	// Reset colors
	ui.ApplyColors(cfg)
	
	fmt.Println(ui.RenderHint("Run 'rekap themes preview' to see full previews"))
	return nil
}

// runThemePreview starts the interactive theme previewer
func runThemePreview() error {
	// Load config to get current theme
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}
	
	// Start the interactive preview
	return tui.RunThemePreview()
}

// setTheme applies a theme to the config
func setTheme(themeName string) error {
	// Validate theme exists
	if !theme.Exists(themeName) {
		return fmt.Errorf("theme '%s' not found. Use 'rekap themes list' to see available themes", themeName)
	}
	
	// Load the theme
	t, err := theme.Load(themeName)
	if err != nil {
		return fmt.Errorf("failed to load theme: %w", err)
	}
	
	// Load or create config
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}
	
	// Apply theme
	cfg.ApplyTheme(t)
	
	// Save config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	
	// Apply colors to UI for immediate feedback
	ui.ApplyColors(cfg)
	
	fmt.Println(ui.RenderSuccess(fmt.Sprintf("Theme '%s' applied successfully!", themeName)))
	fmt.Println(ui.RenderHint("Run 'rekap' to see your new theme in action"))
	return nil
}
