package collectors

import (
	"context"
	"testing"
	"time"

	"github.com/alexinslc/rekap/internal/config"
)

func TestCollectUptime(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := CollectUptime(ctx)

	// Uptime collection may fail in some environments (e.g., CI, Linux)
	if !result.Available {
		t.Skip("Uptime not available in this environment")
	}

	if result.AwakeMinutes < 0 {
		t.Errorf("AwakeMinutes should be >= 0, got %d", result.AwakeMinutes)
	}

	if result.BootTime.IsZero() {
		t.Error("BootTime should not be zero")
	}

	if result.FormattedTime == "" {
		t.Error("FormattedTime should not be empty")
	}
}

func TestCollectBattery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := CollectBattery(ctx)

	if !result.Available {
		t.Skip("Battery not available (running on desktop?)")
	}

	if result.CurrentPct < 0 || result.CurrentPct > 100 {
		t.Errorf("CurrentPct should be 0-100, got %d", result.CurrentPct)
	}

	if result.StartPct < 0 || result.StartPct > 100 {
		t.Errorf("StartPct should be 0-100, got %d", result.StartPct)
	}
}

func TestCollectScreen(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := CollectScreen(ctx)

	// Screen collection is best-effort, may not always work
	if !result.Available {
		t.Log("Screen-on time not available")
		return
	}

	if result.ScreenOnMinutes < 0 {
		t.Errorf("ScreenOnMinutes should be >= 0, got %d", result.ScreenOnMinutes)
	}

	// Lock count should be non-negative
	if result.LockCount < 0 {
		t.Errorf("LockCount should be >= 0, got %d", result.LockCount)
	}

	// If there are locks, avg should be non-negative
	if result.LockCount > 0 && result.AvgMinsBetweenLock < 0 {
		t.Errorf("AvgMinsBetweenLock should be >= 0 when locks exist, got %d", result.AvgMinsBetweenLock)
	}

	// If there are no locks, avg should be 0
	if result.LockCount == 0 && result.AvgMinsBetweenLock != 0 {
		t.Errorf("AvgMinsBetweenLock should be 0 when no locks, got %d", result.AvgMinsBetweenLock)
	}

	t.Logf("Screen on: %d minutes, Locks: %d, Avg between: %d minutes",
		result.ScreenOnMinutes, result.LockCount, result.AvgMinsBetweenLock)
}

func TestCollectApps(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := CollectApps(ctx)

	// Apps require Full Disk Access, may not be available
	if !result.Available {
		t.Log("App tracking not available (needs Full Disk Access)")
		return
	}

	for _, app := range result.TopApps {
		if app.Minutes < 0 {
			t.Errorf("App minutes should be >= 0, got %d for %s", app.Minutes, app.Name)
		}
		if app.Name == "" {
			t.Error("App name should not be empty")
		}
	}
}

func TestCollectMedia(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := CollectMedia(ctx)

	// Media is optional, test if available
	if !result.Available {
		t.Log("No media playing")
		return
	}

	if result.Track == "" {
		t.Error("Track should not be empty when Available=true")
	}

	if result.App == "" {
		t.Error("App should not be empty when Available=true")
	}
}

func TestCollectFocus(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := CollectFocus(ctx)

	// Focus tracking requires Full Disk Access, may not be available
	if !result.Available {
		t.Log("Focus tracking not available (needs Full Disk Access)")
		return
	}

	if result.StreakMinutes < 0 {
		t.Errorf("StreakMinutes should be >= 0, got %d", result.StreakMinutes)
	}

	if result.AppName == "" {
		t.Error("AppName should not be empty when Available=true")
	}

	// Check that time window is set
	if result.StartTime.IsZero() {
		t.Error("StartTime should not be zero when Available=true")
	}

	if result.EndTime.IsZero() {
		t.Error("EndTime should not be zero when Available=true")
	}

	// Validate that EndTime is after StartTime
	if !result.EndTime.After(result.StartTime) {
		t.Errorf("EndTime (%v) should be after StartTime (%v)", result.EndTime, result.StartTime)
	}

	t.Logf("Best flow: %dm in %s (%v - %v)",
		result.StreakMinutes, result.AppName, result.StartTime, result.EndTime)
}

func TestCollectorTimeout(t *testing.T) {
	// Test that collectors respect context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(2 * time.Millisecond)

	// This should return quickly even though context is already done
	result := CollectUptime(ctx)

	// Even with expired context, best-effort should still work
	if !result.Available {
		t.Log("Uptime still unavailable with expired context (expected)")
	}
}

func TestCollectNetwork(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := CollectNetwork(ctx)

	// Network collection is best-effort, may not always work
	if !result.Available {
		t.Log("Network not available")
		return
	}

	if result.InterfaceName == "" {
		t.Error("InterfaceName should not be empty when Available=true")
	}

	if result.NetworkName == "" {
		t.Error("NetworkName should not be empty when Available=true")
	}

	if result.BytesReceived < 0 {
		t.Errorf("BytesReceived should be >= 0, got %d", result.BytesReceived)
	}

	if result.BytesSent < 0 {
		t.Errorf("BytesSent should be >= 0, got %d", result.BytesSent)
	}
}

func TestCollectBrowserTabs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := config.Default()
	result := CollectBrowserTabs(ctx, cfg)

	// Browser collection is best-effort and depends on running browsers
	if !result.Available {
		t.Log("No browsers running or available")
		return
	}

	if result.TotalTabs < 0 {
		t.Errorf("TotalTabs should be >= 0, got %d", result.TotalTabs)
	}

	// If any browser is available, it should contribute to total
	chromeContrib := 0
	if result.Chrome.Available {
		chromeContrib = result.Chrome.TabCount
		if result.Chrome.Browser != "Chrome" {
			t.Errorf("Chrome browser name should be 'Chrome', got %s", result.Chrome.Browser)
		}
	}

	safariContrib := 0
	if result.Safari.Available {
		safariContrib = result.Safari.TabCount
		if result.Safari.Browser != "Safari" {
			t.Errorf("Safari browser name should be 'Safari', got %s", result.Safari.Browser)
		}
	}

	edgeContrib := 0
	if result.Edge.Available {
		edgeContrib = result.Edge.TabCount
		if result.Edge.Browser != "Edge" {
			t.Errorf("Edge browser name should be 'Edge', got %s", result.Edge.Browser)
		}
	}

	expectedTotal := chromeContrib + safariContrib + edgeContrib
	if result.TotalTabs != expectedTotal {
		t.Errorf("TotalTabs should equal sum of individual browsers: expected %d, got %d",
			expectedTotal, result.TotalTabs)
	}

	t.Logf("Collected %d total tabs (Chrome: %d, Safari: %d, Edge: %d)",
		result.TotalTabs, chromeContrib, safariContrib, edgeContrib)

	if len(result.TopDomains) > 0 {
		t.Logf("Top domain: %v", result.TopDomains)
	}
}

func TestIsExcluded(t *testing.T) {
	excludedApps := []string{"Activity Monitor", "System Preferences", "Slack"}

	tests := []struct {
		appName  string
		expected bool
	}{
		{"Activity Monitor", true},
		{"System Preferences", true},
		{"Slack", true},
		{"VS Code", false},
		{"Safari", false},
		{"", false},
		{"activity monitor", false}, // Case-sensitive
	}

	for _, tt := range tests {
		result := isExcluded(tt.appName, excludedApps)
		if result != tt.expected {
			t.Errorf("isExcluded(%q) = %v, want %v", tt.appName, result, tt.expected)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
		{2147483648, "2.0 GB"},
	}

	for _, tt := range tests {
		result := FormatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("FormatBytes(%d) = %s, want %s", tt.bytes, result, tt.expected)
		}
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://www.github.com/user/repo", "github.com"},
		{"http://mail.google.com", "mail.google.com"},
		{"https://example.com:8080/path", "example.com:8080"},
		{"", ""},
		{"invalid-url", ""},
		{"file:///local/path", ""},
	}

	for _, tt := range tests {
		result := extractDomain(tt.url)
		if result != tt.expected {
			t.Errorf("extractDomain(%q) = %q, want %q", tt.url, result, tt.expected)
		}
	}
}

func TestIssuePatterns(t *testing.T) {
	tests := []struct {
		url           string
		expectMatch   bool
		expectedID    string
		expectedType  string
	}{
		// GitHub
		{"https://github.com/alexinslc/rekap/issues/42", true, "github.com/alexinslc/rekap/issues/42", "GitHub"},
		{"https://github.com/org/repo/issues/123", true, "github.com/org/repo/issues/123", "GitHub"},
		
		// Jira
		{"https://company.atlassian.net/browse/PROJ-123", true, "PROJ-123", "Jira"},
		{"https://myorg.atlassian.net/browse/ABC-456", true, "ABC-456", "Jira"},
		
		// Linear
		{"https://linear.app/issue/ENG-789", true, "ENG-789", "Linear"},
		{"https://linear.app/workspace/issue/TEAM-123", true, "TEAM-123", "Linear"},
		
		// GitLab
		{"https://gitlab.com/group/project/-/issues/99", true, "gitlab.com/group/project/-/issues/99", "GitLab"},
		
		// Azure DevOps
		{"https://dev.azure.com/org/project/_workitems/edit/555", true, "555", "Azure DevOps"},
		
		// Non-matching URLs
		{"https://github.com/user/repo", false, "", ""},
		{"https://example.com", false, "", ""},
	}

	for _, tt := range tests {
		var matched bool
		var matchedID string
		var matchedType string

		for _, pattern := range issuePatterns {
			matches := pattern.pattern.FindStringSubmatch(tt.url)
			if matches != nil {
				matched = true
				matchedType = pattern.tracker
				if pattern.idGroup == 0 {
					matchedID = matches[0]
				} else {
					matchedID = matches[pattern.idGroup]
				}
				break
			}
		}

		if matched != tt.expectMatch {
			t.Errorf("URL %q: expected match=%v, got match=%v", tt.url, tt.expectMatch, matched)
		}

		if tt.expectMatch {
			if matchedID != tt.expectedID {
				t.Errorf("URL %q: expected ID=%q, got ID=%q", tt.url, tt.expectedID, matchedID)
			}
			if matchedType != tt.expectedType {
				t.Errorf("URL %q: expected type=%q, got type=%q", tt.url, tt.expectedType, matchedType)
			}
		}
	}
}

func TestIsIssueURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		// Jira patterns
		{"https://company.atlassian.net/browse/PROJ-123", true},
		{"https://jira.example.com/browse/ABC-456", true},
		
		// GitHub patterns
		{"https://github.com/owner/repo/issues/123", true},
		{"https://github.com/owner/repo/pull/456", true},
		
		// Linear patterns
		{"https://linear.app/team/issue/ABC-123", true},
		
		// GitLab patterns
		{"https://gitlab.com/owner/repo/issues/123", true},
		{"https://gitlab.com/owner/repo/merge_requests/456", true},
		
		// Bitbucket patterns
		{"https://bitbucket.org/owner/repo/issues/123", true},
		
		// Azure DevOps patterns
		{"https://dev.azure.com/org/project/_workitems/123", true},
		
		// Non-issue URLs
		{"https://github.com", false},
		{"https://stackoverflow.com/questions/12345", false},
		{"https://google.com", false},
		{"", false},
	}

	for _, tt := range tests {
		result := isIssueURL(tt.url)
		if result != tt.expected {
			t.Errorf("isIssueURL(%q) = %v, want %v", tt.url, result, tt.expected)
		}
	}
}

func TestCollectIssues(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := CollectIssues(ctx)

	// This is best-effort and may not find any issues
	// Just verify the structure is correct
	// Available should be true if and only if there are issues
	if result.Available && len(result.Issues) == 0 {
		t.Error("Available is true but no issues found")
	} else if !result.Available && len(result.Issues) > 0 {
		t.Error("Available is false but issues were found")
	}

	for _, issue := range result.Issues {
		if issue.ID == "" {
			t.Error("Issue ID should not be empty")
		}
		if issue.Tracker == "" {
			t.Error("Issue Tracker should not be empty")
		}
		if issue.URL == "" {
			t.Error("Issue URL should not be empty")
		}
		if issue.VisitCount <= 0 {
			t.Errorf("Issue visit count should be > 0, got %d", issue.VisitCount)
		}
	}
}

func TestExtractIssueIdentifier(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		// Jira
		{"https://company.atlassian.net/browse/PROJ-123", "PROJ-123"},
		{"https://jira.example.com/browse/ABC-456", "ABC-456"},
		
		// GitHub issues
		{"https://github.com/owner/repo/issues/123", "owner/repo#123"},
		
		// GitHub pull requests
		{"https://github.com/owner/repo/pull/456", "owner/repo#456"},
		
		// Linear
		{"https://linear.app/team/issue/ABC-123", "ABC-123"},
		
		// GitLab issues
		{"https://gitlab.com/owner/repo/issues/123", "owner/repo#123"},
		
		// GitLab merge requests
		{"https://gitlab.com/owner/repo/merge_requests/456", "owner/repo!456"},
		
		// Bitbucket
		{"https://bitbucket.org/owner/repo/issues/123", "owner/repo#123"},
		
		// Azure DevOps
		{"https://dev.azure.com/org/project/_workitems/123", "WI-123"},
	}

	for _, tt := range tests {
		result := extractIssueIdentifier(tt.url)
		if result != tt.expected {
			t.Errorf("extractIssueIdentifier(%q) = %q, want %q", tt.url, result, tt.expected)
		}
	}
}

func TestFormatIssueURLs(t *testing.T) {
	tests := []struct {
		name     string
		urls     []string
		expected string
	}{
		{
			name:     "empty list",
			urls:     []string{},
			expected: "",
		},
		{
			name:     "single issue",
			urls:     []string{"PROJ-123"},
			expected: "PROJ-123",
		},
		{
			name:     "three issues",
			urls:     []string{"PROJ-123", "PROJ-456", "ABC-789"},
			expected: "PROJ-123, PROJ-456, ABC-789",
		},
		{
			name:     "more than three issues",
			urls:     []string{"PROJ-123", "PROJ-456", "ABC-789", "XYZ-999", "DEF-111"},
			expected: "PROJ-123, PROJ-456, ABC-789 (+2 more)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatIssueURLs(tt.urls)
			if result != tt.expected {
				t.Errorf("FormatIssueURLs() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCollectNotifications(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := CollectNotifications(ctx)

	// Notifications require Full Disk Access, may not be available
	if !result.Available {
		t.Log("Notification tracking not available (needs Full Disk Access)")
		return
	}

	if result.TotalNotifications < 0 {
		t.Errorf("TotalNotifications should be >= 0, got %d", result.TotalNotifications)
	}

	for _, app := range result.TopApps {
		if app.Count < 0 {
			t.Errorf("Notification count should be >= 0, got %d for %s", app.Count, app.Name)
		}
		if app.Name == "" {
			t.Error("App name should not be empty")
		}
		if app.Count > result.TotalNotifications {
			t.Errorf("App notification count (%d) should not exceed total (%d)", app.Count, result.TotalNotifications)
		}
	}

	t.Logf("Collected %d total notifications from %d apps", result.TotalNotifications, len(result.TopApps))
}

func TestCollectAppsWithSwitching(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := CollectApps(ctx)

	// Apps require Full Disk Access, may not be available
	if !result.Available {
		t.Log("App tracking not available (needs Full Disk Access)")
		return
	}

	// Validate switching metrics if available
	if result.SwitchingAvailable {
		if result.TotalSwitches < 0 {
			t.Errorf("TotalSwitches should be >= 0, got %d", result.TotalSwitches)
		}
		if result.AvgMinsBetween < 0 {
			t.Errorf("AvgMinsBetween should be >= 0, got %.2f", result.AvgMinsBetween)
		}
		if result.SwitchesPerHour < 0 {
			t.Errorf("SwitchesPerHour should be >= 0, got %.2f", result.SwitchesPerHour)
		}
		t.Logf("App switching: %d switches, avg %.2f mins between, %.2f per hour",
			result.TotalSwitches, result.AvgMinsBetween, result.SwitchesPerHour)
	}
}
