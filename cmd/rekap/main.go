package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/alexinslc/rekap/internal/collectors"
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
	fmt.Println("🔐 rekap permission setup")
	fmt.Println("This wizard will help you grant permissions for full functionality.")
	fmt.Println("\nComing soon...")
}

func runDoctor() {
	fmt.Println("🩺 rekap capabilities check")
	fmt.Println("Coming soon...")
}

func runDemo() {
	fmt.Println("🎭 rekap demo mode")
	fmt.Println("Coming soon...")
}

func runSummary(quiet bool) {
	// Create context with timeout for all collectors
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Collect data from all sources
	uptimeResult := collectors.CollectUptime(ctx)
	batteryResult := collectors.CollectBattery(ctx)

	if quiet {
		// Machine-parsable output
		printQuiet(uptimeResult, batteryResult)
	} else {
		// Human-friendly output
		printHuman(uptimeResult, batteryResult)
	}
}

func printQuiet(uptime collectors.UptimeResult, battery collectors.BatteryResult) {
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
}

func printHuman(uptime collectors.UptimeResult, battery collectors.BatteryResult) {
	fmt.Println("📊 Today's rekap")
	fmt.Println()

	// Uptime info
	if uptime.Available {
		fmt.Printf("⏰ Active since %s • %s\n", 
			uptime.BootTime.Format("3:04 PM"), 
			uptime.FormattedTime)
	} else if uptime.Error != nil {
		fmt.Printf("⏰ Uptime unavailable: %v\n", uptime.Error)
	}

	// Battery info
	if battery.Available {
		status := "discharging"
		if battery.IsPlugged {
			status = "plugged in"
		}
		fmt.Printf("🔋 Battery: %d%% • %s\n", battery.CurrentPct, status)
		
		if battery.PlugCount > 0 {
			fmt.Printf("   Plugged in %d times today\n", battery.PlugCount)
		}
	} else if battery.Error != nil {
		fmt.Printf("🔋 Battery unavailable: %v\n", battery.Error)
	}

	fmt.Println()
	fmt.Println("💡 More data sources coming soon - run 'rekap doctor' to check capabilities")
}
