package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alexinslc/rekap/internal/collectors"
	"github.com/alexinslc/rekap/internal/permissions"
	"github.com/alexinslc/rekap/internal/ui"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

func main() {
	var quietFlag bool

	rootCmd := &cobra.Command{
		Use:   "rekap",
		Short: "Daily Mac Activity Summary",
		Long:  `A single-binary macOS CLI that summarizes today's computer activity in a friendly, animated terminal UI.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			runSummary(quietFlag)
			return nil
		},
	}

	rootCmd.Flags().BoolVarP(&quietFlag, "quiet", "q", false, "Output machine-parsable key=value format")

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

	demoCmd := &cobra.Command{
		Use:   "demo",
		Short: "See sample output with fake data",
		Long:  `Display a demo with randomized sample data to preview the output format.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			runDemo()
			return nil
		},
	}

	rootCmd.AddCommand(initCmd, doctorCmd, demoCmd)

	if err := fang.Execute(context.Background(), rootCmd, fang.WithVersion(version)); err != nil {
		os.Exit(1)
	}
}

func runInit() error {
	return permissions.RequestFlow()
}

func runDoctor() {
	fmt.Println("🩺 rekap capabilities check")
	fmt.Println()
	
	caps := permissions.Check()
	fmt.Println(permissions.FormatCapabilities(caps))
	fmt.Println()
	
	if !caps.FullDiskAccess {
		fmt.Println("💡 Run 'rekap init' to enable Full Disk Access for app tracking")
	} else {
		fmt.Println("✅ All major permissions granted!")
	}
}

func runDemo() {
	fmt.Println("🎭 rekap demo mode")
	fmt.Println("   (Showing randomized sample data)")
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

	demoScreen := collectors.ScreenResult{
		ScreenOnMinutes: 215, // 3h 35m
		Available:       true,
	}

	demoApps := collectors.AppsResult{
		TopApps: []collectors.AppUsage{
			{Name: "VS Code", Minutes: 142, BundleID: "com.microsoft.VSCode"},
			{Name: "Safari", Minutes: 89, BundleID: "com.apple.Safari"},
			{Name: "Slack", Minutes: 52, BundleID: "com.tinyspeck.slackmacgap"},
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

	// Show in human-friendly format
	printHuman(demoUptime, demoBattery, demoScreen, demoApps, demoFocus, demoMedia)
}

func runSummary(quiet bool) {
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

	go func() { uptimeCh <- collectors.CollectUptime(ctx) }()
	go func() { batteryCh <- collectors.CollectBattery(ctx) }()
	go func() { screenCh <- collectors.CollectScreen(ctx) }()
	go func() { appsCh <- collectors.CollectApps(ctx) }()
	go func() { focusCh <- collectors.CollectFocus(ctx) }()
	go func() { mediaCh <- collectors.CollectMedia(ctx) }()

	// Wait for all results
	uptimeResult := <-uptimeCh
	batteryResult := <-batteryCh
	screenResult := <-screenCh
	appsResult := <-appsCh
	focusResult := <-focusCh
	mediaResult := <-mediaCh

	if quiet {
		// Machine-parsable output
		printQuiet(uptimeResult, batteryResult, screenResult, appsResult, focusResult, mediaResult)
	} else {
		// Human-friendly output
		printHuman(uptimeResult, batteryResult, screenResult, appsResult, focusResult, mediaResult)
	}
}

func printQuiet(uptime collectors.UptimeResult, battery collectors.BatteryResult, screen collectors.ScreenResult, apps collectors.AppsResult, focus collectors.FocusResult, media collectors.MediaResult) {
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
}

func printHuman(uptime collectors.UptimeResult, battery collectors.BatteryResult, screen collectors.ScreenResult, apps collectors.AppsResult, focus collectors.FocusResult, media collectors.MediaResult) {
	// Render title with animation if TTY
	title := ui.RenderTitle("📊 Today's rekap", ui.IsTTY())
	if title != "" {
		fmt.Println(title)
	}
	fmt.Println()

	// Build summary line
	var summaryParts []string
	
	if screen.Available {
		summaryParts = append(summaryParts, ui.FormatDuration(screen.ScreenOnMinutes)+" screen-on")
	}
	
	if battery.Available && battery.PlugCount > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d plug-ins", battery.PlugCount))
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

	// Uptime info
	if uptime.Available {
		text := fmt.Sprintf("Active since %s • %s", 
			uptime.BootTime.Format("3:04 PM"), 
			uptime.FormattedTime)
		fmt.Println(ui.RenderDataPoint("⏰", text))
	}

	// Battery info
	if battery.Available {
		status := "discharging"
		if battery.IsPlugged {
			status = "plugged in"
		}
		var text string
		if battery.StartPct != battery.CurrentPct {
			text = fmt.Sprintf("Started at %d%%, now %d%% • %s", battery.StartPct, battery.CurrentPct, status)
		} else {
			text = fmt.Sprintf("%d%% • %s", battery.CurrentPct, status)
		}
		fmt.Println(ui.RenderDataPoint("🔋", text))
	}

	// Focus streak
	if focus.Available {
		text := fmt.Sprintf("Best focus: %s in %s", ui.FormatDuration(focus.StreakMinutes), focus.AppName)
		fmt.Println(ui.RenderHighlight("⏱️ ", text))
	}

	// Media info
	if media.Available {
		text := fmt.Sprintf("Now playing: \"%s\" in %s", media.Track, media.App)
		fmt.Println(ui.RenderDataPoint("🎵", text))
	}

	fmt.Println()

	// Show hints for missing data
	if !apps.Available && apps.Error != nil {
		fmt.Println(ui.RenderHint("Screen Time unavailable—run 'rekap init' to enable app tracking"))
	}
}
