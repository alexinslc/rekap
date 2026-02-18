package main

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alexinslc/rekap/internal/collectors"
	"github.com/alexinslc/rekap/internal/config"
	"github.com/alexinslc/rekap/internal/ui"
	"github.com/alexinslc/rekap/internal/ui/tui"
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

func runSummary(quiet bool, asJSON bool, print bool, cfg *config.Config) {
	ui.ApplyColors(cfg)

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
	case print || !ui.IsTTY():
		printHuman(cfg, &data)
	default:
		runTUI(cfg, &data)
	}
}

func runTUI(cfg *config.Config, data *SummaryData) {
	tuiData := &tui.SummaryData{
		Uptime:        data.Uptime,
		Battery:       data.Battery,
		Screen:        data.Screen,
		Apps:          data.Apps,
		Focus:         data.Focus,
		Media:         data.Media,
		Network:       data.Network,
		Browsers:      data.Browsers,
		Notifications: data.Notifications,
		Issues:        data.Issues,
		Fragmentation: data.Fragmentation,
		Burnout:       data.Burnout,
	}
	sections := tui.BuildSections(tuiData, cfg)
	m := tui.New(sections, cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}
}
