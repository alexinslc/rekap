package collectors

import (
	"testing"
)

func TestCheckContextOverload(t *testing.T) {
	tests := []struct {
		name           string
		apps           AppsResult
		browsers       BrowsersResult
		expectOverload bool
		expectMessage  string
	}{
		{
			name: "no overload - few apps and tabs",
			apps: AppsResult{
				TopApps: []AppUsage{
					{Name: "App1", Minutes: 10},
					{Name: "App2", Minutes: 5},
				},
				Available: true,
			},
			browsers: BrowsersResult{
				TotalTabs: 10,
				TopDomains: map[string]int{
					"example.com": 5,
					"test.com":    5,
				},
				Available: true,
			},
			expectOverload: false,
			expectMessage:  "",
		},
		{
			name: "overload - too many apps (>5)",
			apps: AppsResult{
				TopApps: []AppUsage{
					{Name: "App1", Minutes: 10},
					{Name: "App2", Minutes: 10},
					{Name: "App3", Minutes: 10},
					{Name: "App4", Minutes: 10},
					{Name: "App5", Minutes: 10},
					{Name: "App6", Minutes: 10},
				},
				Available: true,
			},
			browsers: BrowsersResult{
				TotalTabs:  10,
				TopDomains: map[string]int{"example.com": 10},
				Available:  true,
			},
			expectOverload: true,
			expectMessage:  "6 apps active",
		},
		{
			name: "overload - too many tabs (>30)",
			apps: AppsResult{
				TopApps: []AppUsage{
					{Name: "App1", Minutes: 10},
				},
				Available: true,
			},
			browsers: BrowsersResult{
				TotalTabs:  45,
				TopDomains: map[string]int{"example.com": 45},
				Available:  true,
			},
			expectOverload: true,
			expectMessage:  "45 tabs active",
		},
		{
			name: "overload - too many domains (>10)",
			apps: AppsResult{
				TopApps: []AppUsage{
					{Name: "App1", Minutes: 10},
				},
				Available: true,
			},
			browsers: BrowsersResult{
				TotalTabs: 15,
				TopDomains: map[string]int{
					"domain1.com":  1,
					"domain2.com":  1,
					"domain3.com":  1,
					"domain4.com":  1,
					"domain5.com":  1,
					"domain6.com":  1,
					"domain7.com":  1,
					"domain8.com":  1,
					"domain9.com":  1,
					"domain10.com": 1,
					"domain11.com": 1,
				},
				Available: true,
			},
			expectOverload: true,
			expectMessage:  "11 domains active",
		},
		{
			name: "overload - apps and tabs",
			apps: AppsResult{
				TopApps: []AppUsage{
					{Name: "App1", Minutes: 10},
					{Name: "App2", Minutes: 10},
					{Name: "App3", Minutes: 10},
					{Name: "App4", Minutes: 10},
					{Name: "App5", Minutes: 10},
					{Name: "App6", Minutes: 10},
					{Name: "App7", Minutes: 10},
				},
				Available: true,
			},
			browsers: BrowsersResult{
				TotalTabs:  45,
				TopDomains: map[string]int{"example.com": 45},
				Available:  true,
			},
			expectOverload: true,
			expectMessage:  "7 apps + 45 tabs active",
		},
		{
			name: "edge case - exactly at threshold",
			apps: AppsResult{
				TopApps: []AppUsage{
					{Name: "App1", Minutes: 10},
					{Name: "App2", Minutes: 10},
					{Name: "App3", Minutes: 10},
					{Name: "App4", Minutes: 10},
					{Name: "App5", Minutes: 10},
				},
				Available: true,
			},
			browsers: BrowsersResult{
				TotalTabs: 30,
				TopDomains: map[string]int{
					"domain1.com":  1,
					"domain2.com":  1,
					"domain3.com":  1,
					"domain4.com":  1,
					"domain5.com":  1,
					"domain6.com":  1,
					"domain7.com":  1,
					"domain8.com":  1,
					"domain9.com":  1,
					"domain10.com": 1,
				},
				Available: true,
			},
			expectOverload: false,
			expectMessage:  "",
		},
		{
			name: "no data available",
			apps: AppsResult{
				Available: false,
			},
			browsers: BrowsersResult{
				Available: false,
			},
			expectOverload: false,
			expectMessage:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckContextOverload(tt.apps, tt.browsers)

			if result.IsOverloaded != tt.expectOverload {
				t.Errorf("IsOverloaded = %v, want %v", result.IsOverloaded, tt.expectOverload)
			}

			if result.IsOverloaded && result.WarningMessage != tt.expectMessage {
				t.Errorf("WarningMessage = %q, want %q", result.WarningMessage, tt.expectMessage)
			}
		})
	}
}

func TestFormatWithCount(t *testing.T) {
	tests := []struct {
		count    int
		singular string
		expected string
	}{
		{1, "app", "1 app"},
		{2, "app", "2 apps"},
		{10, "tab", "10 tabs"},
		{45, "tab", "45 tabs"},
		{1, "domain", "1 domain"},
		{11, "domain", "11 domains"},
	}

	for _, tt := range tests {
		result := formatWithCount(tt.count, tt.singular)
		if result != tt.expected {
			t.Errorf("formatWithCount(%d, %q) = %q, want %q", tt.count, tt.singular, result, tt.expected)
		}
	}
}
