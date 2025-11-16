package collectors

import (
	"context"
	"math"
)

// FragmentationResult contains context fragmentation analysis
type FragmentationResult struct {
	Score     int    // 0-100 score
	Level     string // "focused", "moderate", or "fragmented"
	Emoji     string // Visual indicator
	Available bool
	Error     error
	Breakdown FragmentationBreakdown // Details on how score was calculated
}

// FragmentationBreakdown provides detailed metrics used in calculation
type FragmentationBreakdown struct {
	UniqueApps         int
	TotalTabs          int
	UniqueDomains      int
	AppSwitchesPerHour float64
}

// FragmentationThresholds defines configurable thresholds
type FragmentationThresholds struct {
	FocusedMax    int // 0-30 = Focused
	ModerateMax   int // 31-60 = Moderate
	FragmentedMin int // 61-100 = Fragmented
}

// DefaultFragmentationThresholds returns default threshold values
func DefaultFragmentationThresholds() FragmentationThresholds {
	return FragmentationThresholds{
		FocusedMax:    30,
		ModerateMax:   60,
		FragmentedMin: 61,
	}
}

// CalculateFragmentation computes the context fragmentation score
// Inputs: apps data, browser data, uptime data
func CalculateFragmentation(ctx context.Context, apps AppsResult, browsers BrowsersResult, uptime UptimeResult, thresholds FragmentationThresholds) FragmentationResult {
	result := FragmentationResult{Available: false}

	// Need at least some data to calculate
	if !apps.Available && !browsers.Available {
		result.Error = nil // Not an error, just no data
		return result
	}

	breakdown := FragmentationBreakdown{}

	// Count unique apps
	breakdown.UniqueApps = len(apps.TopApps)

	// Count total tabs
	breakdown.TotalTabs = browsers.TotalTabs

	// Count unique domains
	breakdown.UniqueDomains = len(browsers.TopDomains)

	// Calculate app switches per hour (estimate based on app count and time awake)
	if uptime.Available && uptime.AwakeMinutes > 0 {
		hoursAwake := float64(uptime.AwakeMinutes) / 60.0
		if hoursAwake > 0 {
			// Rough estimate: assumes each unique app represents switching activity distributed over awake hours
			breakdown.AppSwitchesPerHour = float64(breakdown.UniqueApps) / hoursAwake
		}
	} else {
		// If no uptime data, assume 4 hours as default
		breakdown.AppSwitchesPerHour = float64(breakdown.UniqueApps) / 4.0
	}

	result.Breakdown = breakdown

	// Calculate weighted score (0-100)
	score := calculateWeightedScore(breakdown)
	result.Score = int(math.Round(score))

	// Ensure score is in valid range
	if result.Score < 0 {
		result.Score = 0
	}
	if result.Score > 100 {
		result.Score = 100
	}

	// Determine level and emoji
	if result.Score <= thresholds.FocusedMax {
		result.Level = "focused"
		result.Emoji = "🎯"
	} else if result.Score <= thresholds.ModerateMax {
		result.Level = "moderate"
		result.Emoji = "⚖️"
	} else {
		result.Level = "fragmented"
		result.Emoji = "🔀"
	}

	result.Available = true
	return result
}

// calculateWeightedScore computes a weighted score based on multiple factors
func calculateWeightedScore(breakdown FragmentationBreakdown) float64 {
	var score float64

	// Factor 1: Unique apps (weight: 30%)
	// 0-3 apps = low, 4-8 = medium, 9+ = high
	appsScore := normalizeValue(float64(breakdown.UniqueApps), 3, 9) * 30

	// Factor 2: Total tabs (weight: 25%)
	// 0-10 tabs = low, 11-25 = medium, 26+ = high
	tabsScore := normalizeValue(float64(breakdown.TotalTabs), 10, 30) * 25

	// Factor 3: Unique domains (weight: 25%)
	// 0-5 domains = low, 6-12 = medium, 13+ = high
	domainsScore := normalizeValue(float64(breakdown.UniqueDomains), 5, 13) * 25

	// Factor 4: App switches per hour (weight: 20%)
	// 0-1 switches/hr = low, 2-3 = medium, 4+ = high
	switchesScore := normalizeValue(breakdown.AppSwitchesPerHour, 1, 4) * 20

	score = appsScore + tabsScore + domainsScore + switchesScore

	return score
}

// normalizeValue converts a value to 0-1 scale based on min/max thresholds
// Returns 0 at minThreshold, 1 at maxThreshold, interpolates in between
func normalizeValue(value, minThreshold, maxThreshold float64) float64 {
	if value <= minThreshold {
		return 0
	}
	if value >= maxThreshold {
		return 1
	}
	return (value - minThreshold) / (maxThreshold - minThreshold)
}
