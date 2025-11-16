package collectors

import (
	"context"
	"testing"
)

func TestCalculateFragmentation(t *testing.T) {
	ctx := context.Background()
	thresholds := DefaultFragmentationThresholds()

	tests := []struct {
		name            string
		apps            AppsResult
		browsers        BrowsersResult
		uptime          UptimeResult
		expectedLevel   string
		expectedAvail   bool
		minScore        int
		maxScore        int
	}{
		{
			name: "focused - few apps and tabs",
			apps: AppsResult{
				TopApps: []AppUsage{
					{Name: "VS Code", Minutes: 120},
					{Name: "Safari", Minutes: 60},
				},
				Available: true,
			},
			browsers: BrowsersResult{
				TotalTabs: 5,
				TopDomains: map[string]int{
					"github.com": 3,
					"docs.go.dev": 2,
				},
				Available: true,
			},
			uptime: UptimeResult{
				AwakeMinutes: 240,
				Available:    true,
			},
			expectedLevel: "focused",
			expectedAvail: true,
			minScore:      0,
			maxScore:      30,
		},
		{
			name: "moderate - medium activity",
			apps: AppsResult{
				TopApps: []AppUsage{
					{Name: "VS Code", Minutes: 100},
					{Name: "Safari", Minutes: 80},
					{Name: "Slack", Minutes: 50},
					{Name: "Mail", Minutes: 40},
					{Name: "Notes", Minutes: 30},
					{Name: "Terminal", Minutes: 25},
					{Name: "Calendar", Minutes: 20},
				},
				Available: true,
			},
			browsers: BrowsersResult{
				TotalTabs: 20,
				TopDomains: map[string]int{
					"github.com": 5,
					"stackoverflow.com": 4,
					"mail.google.com": 3,
					"docs.go.dev": 2,
					"twitter.com": 2,
					"reddit.com": 2,
					"youtube.com": 2,
				},
				Available: true,
			},
			uptime: UptimeResult{
				AwakeMinutes: 300,
				Available:    true,
			},
			expectedLevel: "moderate",
			expectedAvail: true,
			minScore:      31,
			maxScore:      60,
		},
		{
			name: "fragmented - many apps and tabs",
			apps: AppsResult{
				TopApps: []AppUsage{
					{Name: "App1", Minutes: 50},
					{Name: "App2", Minutes: 45},
					{Name: "App3", Minutes: 40},
					{Name: "App4", Minutes: 35},
					{Name: "App5", Minutes: 30},
					{Name: "App6", Minutes: 25},
					{Name: "App7", Minutes: 20},
					{Name: "App8", Minutes: 18},
					{Name: "App9", Minutes: 15},
					{Name: "App10", Minutes: 12},
				},
				Available: true,
			},
			browsers: BrowsersResult{
				TotalTabs: 50,
				TopDomains: map[string]int{
					"github.com": 8,
					"stackoverflow.com": 6,
					"mail.google.com": 5,
					"docs.go.dev": 4,
					"twitter.com": 4,
					"reddit.com": 4,
					"youtube.com": 3,
					"medium.com": 3,
					"dev.to": 3,
					"linkedin.com": 2,
					"facebook.com": 2,
					"news.ycombinator.com": 2,
					"producthunt.com": 2,
					"slack.com": 2,
				},
				Available: true,
			},
			uptime: UptimeResult{
				AwakeMinutes: 300,
				Available:    true,
			},
			expectedLevel: "fragmented",
			expectedAvail: true,
			minScore:      61,
			maxScore:      100,
		},
		{
			name: "no data available",
			apps: AppsResult{
				Available: false,
			},
			browsers: BrowsersResult{
				Available: false,
			},
			uptime: UptimeResult{
				Available: false,
			},
			expectedAvail: false,
		},
		{
			name: "only apps data",
			apps: AppsResult{
				TopApps: []AppUsage{
					{Name: "VS Code", Minutes: 120},
					{Name: "Safari", Minutes: 60},
					{Name: "Slack", Minutes: 40},
				},
				Available: true,
			},
			browsers: BrowsersResult{
				Available: false,
			},
			uptime: UptimeResult{
				AwakeMinutes: 240,
				Available:    true,
			},
			expectedAvail: true,
			expectedLevel: "focused",
		},
		{
			name: "only browser data",
			apps: AppsResult{
				Available: false,
			},
			browsers: BrowsersResult{
				TotalTabs: 8,
				TopDomains: map[string]int{
					"github.com": 4,
					"docs.go.dev": 4,
				},
				Available: true,
			},
			uptime: UptimeResult{
				AwakeMinutes: 240,
				Available:    true,
			},
			expectedAvail: true,
			expectedLevel: "focused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateFragmentation(ctx, tt.apps, tt.browsers, tt.uptime, thresholds)

			if result.Available != tt.expectedAvail {
				t.Errorf("Available = %v, want %v", result.Available, tt.expectedAvail)
			}

			if !result.Available {
				return // Skip other checks if no data available
			}

			if result.Level != tt.expectedLevel {
				t.Errorf("Level = %v, want %v (score: %d)", result.Level, tt.expectedLevel, result.Score)
			}

			if result.Score < tt.minScore || result.Score > tt.maxScore {
				t.Errorf("Score %d not in expected range [%d, %d]", result.Score, tt.minScore, tt.maxScore)
			}

			// Verify emoji matches level
			expectedEmoji := ""
			switch tt.expectedLevel {
			case "focused":
				expectedEmoji = "🎯"
			case "moderate":
				expectedEmoji = "⚖️"
			case "fragmented":
				expectedEmoji = "🔀"
			}

			if result.Emoji != expectedEmoji {
				t.Errorf("Emoji = %v, want %v for level %s", result.Emoji, expectedEmoji, tt.expectedLevel)
			}

			// Log breakdown for debugging
			t.Logf("Breakdown: Apps=%d, Tabs=%d, Domains=%d, Switches/hr=%.2f",
				result.Breakdown.UniqueApps,
				result.Breakdown.TotalTabs,
				result.Breakdown.UniqueDomains,
				result.Breakdown.AppSwitchesPerHour)
		})
	}
}

func TestDefaultFragmentationThresholds(t *testing.T) {
	thresholds := DefaultFragmentationThresholds()

	if thresholds.FocusedMax != 30 {
		t.Errorf("FocusedMax = %d, want 30", thresholds.FocusedMax)
	}

	if thresholds.ModerateMax != 60 {
		t.Errorf("ModerateMax = %d, want 60", thresholds.ModerateMax)
	}

	if thresholds.FragmentedMin != 61 {
		t.Errorf("FragmentedMin = %d, want 61", thresholds.FragmentedMin)
	}
}

func TestNormalizeValue(t *testing.T) {
	tests := []struct {
		name         string
		value        float64
		minThreshold float64
		maxThreshold float64
		expected     float64
	}{
		{"below min", 2, 5, 15, 0.0},
		{"at min", 5, 5, 15, 0.0},
		{"mid range", 10, 5, 15, 0.5},
		{"at max", 15, 5, 15, 1.0},
		{"above max", 20, 5, 15, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeValue(tt.value, tt.minThreshold, tt.maxThreshold)
			if result != tt.expected {
				t.Errorf("normalizeValue(%v, %v, %v) = %v, want %v",
					tt.value, tt.minThreshold, tt.maxThreshold, result, tt.expected)
			}
		})
	}
}

func TestCalculateWeightedScore(t *testing.T) {
	tests := []struct {
		name      string
		breakdown FragmentationBreakdown
		minScore  float64
		maxScore  float64
	}{
		{
			name: "minimal activity",
			breakdown: FragmentationBreakdown{
				UniqueApps:         2,
				TotalTabs:          5,
				UniqueDomains:      2,
				AppSwitchesPerHour: 1.0,
			},
			minScore: 0,
			maxScore: 30,
		},
		{
			name: "moderate activity",
			breakdown: FragmentationBreakdown{
				UniqueApps:         7,
				TotalTabs:          20,
				UniqueDomains:      10,
				AppSwitchesPerHour: 4.0,
			},
			minScore: 30,
			maxScore: 70,
		},
		{
			name: "high activity",
			breakdown: FragmentationBreakdown{
				UniqueApps:         15,
				TotalTabs:          50,
				UniqueDomains:      20,
				AppSwitchesPerHour: 8.0,
			},
			minScore: 70,
			maxScore: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateWeightedScore(tt.breakdown)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("Score %.2f not in expected range [%.2f, %.2f]", score, tt.minScore, tt.maxScore)
			}

			t.Logf("Score: %.2f for breakdown: %+v", score, tt.breakdown)
		})
	}
}

func TestFragmentationWithCustomThresholds(t *testing.T) {
	ctx := context.Background()

	// Custom thresholds: more relaxed
	customThresholds := FragmentationThresholds{
		FocusedMax:    40,
		ModerateMax:   70,
		FragmentedMin: 71,
	}

	apps := AppsResult{
		TopApps: []AppUsage{
			{Name: "App1", Minutes: 60},
			{Name: "App2", Minutes: 50},
			{Name: "App3", Minutes: 40},
			{Name: "App4", Minutes: 30},
			{Name: "App5", Minutes: 20},
		},
		Available: true,
	}

	browsers := BrowsersResult{
		TotalTabs: 15,
		TopDomains: map[string]int{
			"github.com": 5,
			"docs.go.dev": 5,
			"stackoverflow.com": 5,
		},
		Available: true,
	}

	uptime := UptimeResult{
		AwakeMinutes: 240,
		Available:    true,
	}

	// With default thresholds
	defaultResult := CalculateFragmentation(ctx, apps, browsers, uptime, DefaultFragmentationThresholds())

	// With custom thresholds
	customResult := CalculateFragmentation(ctx, apps, browsers, uptime, customThresholds)

	// Scores should be the same
	if defaultResult.Score != customResult.Score {
		t.Errorf("Scores differ: default=%d, custom=%d", defaultResult.Score, customResult.Score)
	}

	// But levels might differ due to different thresholds
	t.Logf("Default: level=%s, score=%d", defaultResult.Level, defaultResult.Score)
	t.Logf("Custom: level=%s, score=%d", customResult.Level, customResult.Score)
}

func TestFragmentationBreakdownPopulation(t *testing.T) {
	ctx := context.Background()
	thresholds := DefaultFragmentationThresholds()

	apps := AppsResult{
		TopApps: []AppUsage{
			{Name: "App1", Minutes: 60},
			{Name: "App2", Minutes: 50},
			{Name: "App3", Minutes: 40},
		},
		Available: true,
	}

	browsers := BrowsersResult{
		TotalTabs: 25,
		TopDomains: map[string]int{
			"domain1.com": 10,
			"domain2.com": 8,
			"domain3.com": 7,
		},
		Available: true,
	}

	uptime := UptimeResult{
		AwakeMinutes: 180, // 3 hours
		Available:    true,
	}

	result := CalculateFragmentation(ctx, apps, browsers, uptime, thresholds)

	if !result.Available {
		t.Fatal("Result should be available")
	}

	// Check breakdown values
	if result.Breakdown.UniqueApps != 3 {
		t.Errorf("UniqueApps = %d, want 3", result.Breakdown.UniqueApps)
	}

	if result.Breakdown.TotalTabs != 25 {
		t.Errorf("TotalTabs = %d, want 25", result.Breakdown.TotalTabs)
	}

	if result.Breakdown.UniqueDomains != 3 {
		t.Errorf("UniqueDomains = %d, want 3", result.Breakdown.UniqueDomains)
	}

	// 3 unique apps / 3 hours = 1.0 apps per hour
	expectedAppsPerHour := 1.0
	if result.Breakdown.AppSwitchesPerHour != expectedAppsPerHour {
		t.Errorf("AppSwitchesPerHour = %.2f, want %.2f",
			result.Breakdown.AppSwitchesPerHour, expectedAppsPerHour)
	}
}
