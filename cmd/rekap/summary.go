package main

import (
	"context"
	"time"

	"github.com/alexinslc/rekap/internal/collectors"
	"github.com/alexinslc/rekap/internal/config"
	"github.com/alexinslc/rekap/internal/ui"
)

// SummaryData holds all collector results for a single run
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

func runSummary(quiet bool, asJSON bool, cfg *config.Config) {
	if !quiet && !asJSON {
		ui.ApplyColors(cfg)
	}

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

	data := SummaryData{
		Uptime:        <-uptimeCh,
		Battery:       <-batteryCh,
		Screen:        <-screenCh,
		Apps:          <-appsCh,
		Focus:         <-focusCh,
		Media:         <-mediaCh,
		Network:       <-networkCh,
		Browsers:      <-browsersCh,
		Issues:        <-issuesCh,
		Notifications: <-notificationsCh,
	}

	// Calculate fragmentation score after collecting data
	fragmentationThresholds := collectors.FragmentationThresholds{
		FocusedMax:    cfg.Fragmentation.FocusedMax,
		ModerateMax:   cfg.Fragmentation.ModerateMax,
		FragmentedMin: cfg.Fragmentation.FragmentedMin,
	}
	data.Fragmentation = collectors.CalculateFragmentation(ctx, data.Apps, data.Browsers, data.Uptime, fragmentationThresholds)

	// Analyze burnout patterns after collecting primary data
	burnoutConfig := collectors.DefaultBurnoutConfig()
	data.Burnout = collectors.CollectBurnout(ctx, data.Screen, data.Browsers, burnoutConfig)

	switch {
	case asJSON:
		printJSON(&data)
	case quiet:
		printQuiet(cfg, &data)
	default:
		printHuman(cfg, &data)
	}
}
