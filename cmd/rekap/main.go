package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alexinslc/rekap/internal/collectors"
	"github.com/alexinslc/rekap/internal/permissions"
)

const version = "0.1.0"

func main() {
	quietFlag := flag.Bool("quiet", false, "Output machine-parsable key=value format")
	versionFlag := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("rekap v%s\n", version)
		os.Exit(0)
	}

	// Parse subcommands
	args := flag.Args()
	if len(args) > 0 {
		switch args[0] {
		case "init":
			runInit()
		case "doctor":
			runDoctor()
		case "demo":
			runDemo()
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
			os.Exit(1)
		}
		return
	}

	// Default: run summary
	runSummary(*quietFlag)
}

func runInit() {
	err := permissions.RequestFlow()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
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
	fmt.Println("Coming soon...")
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
	fmt.Println("📊 Today's rekap")
	fmt.Println()

	// Build summary line
	var summaryParts []string
	
	if screen.Available {
		hours := screen.ScreenOnMinutes / 60
		mins := screen.ScreenOnMinutes % 60
		summaryParts = append(summaryParts, fmt.Sprintf("%dh %dm screen-on", hours, mins))
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
			hours := app.Minutes / 60
			mins := app.Minutes % 60
			if hours > 0 {
				appList = append(appList, fmt.Sprintf("%s (%dh%dm)", app.Name, hours, mins))
			} else {
				appList = append(appList, fmt.Sprintf("%s (%dm)", app.Name, mins))
			}
		}
		if len(appList) > 0 {
			summaryParts = append(summaryParts, "Top apps: "+strings.Join(appList, ", "))
		}
	}

	if len(summaryParts) > 0 {
		fmt.Println(strings.Join(summaryParts, " • "))
		fmt.Println()
	}

	// Uptime info
	if uptime.Available {
		fmt.Printf("⏰ Active since %s • %s\n", 
			uptime.BootTime.Format("3:04 PM"), 
			uptime.FormattedTime)
	}

	// Battery info
	if battery.Available {
		status := "discharging"
		if battery.IsPlugged {
			status = "plugged in"
		}
		if battery.StartPct != battery.CurrentPct {
			fmt.Printf("🔋 Battery: Started at %d%%, now %d%% • %s\n", battery.StartPct, battery.CurrentPct, status)
		} else {
			fmt.Printf("🔋 Battery: %d%% • %s\n", battery.CurrentPct, status)
		}
	}

	// Focus streak
	if focus.Available {
		hours := focus.StreakMinutes / 60
		mins := focus.StreakMinutes % 60
		if hours > 0 {
			fmt.Printf("⏱️  Best focus: %dh %dm in %s\n", hours, mins, focus.AppName)
		} else {
			fmt.Printf("⏱️  Best focus: %dm in %s\n", mins, focus.AppName)
		}
	}

	// Media info
	if media.Available {
		fmt.Printf("🎵 Now playing: \"%s\" in %s\n", media.Track, media.App)
	}

	fmt.Println()

	// Show hints for missing data
	if !apps.Available && apps.Error != nil {
		fmt.Println("💡 Screen Time unavailable—run 'rekap init' to enable app tracking")
	}
}
