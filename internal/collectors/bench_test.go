package collectors

import (
	"context"
	"testing"
)

// Benchmark for CalculateFragmentation - a critical path in the app
func BenchmarkCalculateFragmentation(b *testing.B) {
	ctx := context.Background()
	thresholds := DefaultFragmentationThresholds()

	// Setup realistic mock data
	apps := AppsResult{
		TopApps: []AppUsage{
			{Name: "VS Code", Minutes: 120},
			{Name: "Safari", Minutes: 80},
			{Name: "Slack", Minutes: 50},
			{Name: "Terminal", Minutes: 40},
			{Name: "Mail", Minutes: 30},
		},
		Available: true,
	}

	browsers := BrowsersResult{
		TotalTabs: 25,
		TopDomains: map[string]int{
			"github.com":        10,
			"stackoverflow.com": 5,
			"docs.go.dev":       5,
			"mail.google.com":   3,
			"twitter.com":       2,
		},
		Available: true,
	}

	uptime := UptimeResult{
		AwakeMinutes: 240,
		Available:    true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateFragmentation(ctx, apps, browsers, uptime, thresholds)
	}
}

// Benchmark for FormatBytes - frequently called function
func BenchmarkFormatBytes(b *testing.B) {
	testCases := []int64{
		500,
		1024,
		1048576,
		1073741824,
		2147483648,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, bytes := range testCases {
			FormatBytes(bytes)
		}
	}
}

// Benchmark for extractDomain - frequently called in browser tracking
func BenchmarkExtractDomain(b *testing.B) {
	testURLs := []string{
		"https://www.github.com/user/repo",
		"http://mail.google.com",
		"https://example.com:8080/path",
		"https://stackoverflow.com/questions/12345",
		"https://docs.python.org/3/library/string.html",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, url := range testURLs {
			extractDomain(url)
		}
	}
}

// Benchmark for CollectBurnout - called on every run
func BenchmarkCollectBurnout(b *testing.B) {
	ctx := context.Background()
	config := DefaultBurnoutConfig()

	screen := ScreenResult{
		ScreenOnMinutes: 420,
		Available:       true,
	}

	browsers := BrowsersResult{
		TotalTabs: 45,
		Available: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CollectBurnout(ctx, screen, browsers, config)
	}
}

// Benchmark for CheckContextOverload - called frequently
func BenchmarkCheckContextOverload(b *testing.B) {
	apps := AppsResult{
		TopApps: []AppUsage{
			{Name: "App1", Minutes: 60},
			{Name: "App2", Minutes: 50},
			{Name: "App3", Minutes: 40},
			{Name: "App4", Minutes: 30},
			{Name: "App5", Minutes: 20},
			{Name: "App6", Minutes: 10},
		},
		Available: true,
	}

	browsers := BrowsersResult{
		TotalTabs: 40,
		TopDomains: map[string]int{
			"domain1.com": 10,
			"domain2.com": 10,
			"domain3.com": 10,
			"domain4.com": 10,
		},
		Available: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckContextOverload(apps, browsers)
	}
}
