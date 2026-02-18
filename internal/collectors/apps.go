package collectors

import (
	"context"
	"database/sql"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

// AppUsage represents usage time for a single app
type AppUsage struct {
	Name     string
	Minutes  int
	BundleID string
}

// AppsResult contains app usage information
type AppsResult struct {
	TopApps            []AppUsage
	Source             string // "ScreenTime" or "Sampling"
	Available          bool
	Error              error
	ExcludedApps       []string // Apps that were filtered out
	TotalSwitches      int      // Total number of app switches today
	AvgMinsBetween     float64  // Average minutes between switches
	SwitchesPerHour    float64  // Switches per hour rate
	SwitchingAvailable bool     // Whether switching data is available
}

// CollectApps retrieves top app usage from Screen Time database
func CollectApps(ctx context.Context, excludedApps []string) AppsResult {
	result := AppsResult{Available: false, Source: "ScreenTime"}
	result.ExcludedApps = excludedApps

	db, err := openKnowledgeDB()
	if err != nil {
		result.Error = err
		return result
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil && result.Error == nil {
			result.Error = fmt.Errorf("failed to close database: %w", closeErr)
		}
	}()

	startTimestamp, endTimestamp := todayTimestampRange()

	// Query app usage from ZOBJECT table
	// ZSTREAMNAME contains app usage data, ZVALUESTRING has bundle IDs
	query := `
		SELECT 
			ZVALUESTRING as bundle_id,
			SUM((ZENDDATE - ZSTARTDATE)) as duration_seconds
		FROM ZOBJECT
		WHERE ZSTREAMNAME = '/app/usage'
			AND ZSTARTDATE >= ?
			AND ZENDDATE <= ?
			AND ZVALUESTRING IS NOT NULL
			AND ZVALUESTRING != ''
		GROUP BY ZVALUESTRING
		ORDER BY duration_seconds DESC
		LIMIT 10
	`

	rows, err := db.QueryContext(ctx, query, startTimestamp, endTimestamp)
	if err != nil {
		result.Error = fmt.Errorf("failed to query Screen Time data: %w", err)
		return result
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && result.Error == nil {
			result.Error = fmt.Errorf("failed to close rows: %w", closeErr)
		}
	}()

	var apps []AppUsage
	for rows.Next() {
		var bundleID string
		var durationSec float64

		if err := rows.Scan(&bundleID, &durationSec); err != nil {
			continue
		}

		// Resolve bundle ID to app name
		appName := resolveAppName(bundleID)

		// Skip if app is in exclusion list
		if isExcluded(appName, excluded) {
			continue
		}

		minutes := int(durationSec / 60)

		if minutes > 0 {
			apps = append(apps, AppUsage{
				Name:     appName,
				Minutes:  minutes,
				BundleID: bundleID,
			})
		}
	}

	result.TopApps = apps
	result.Available = len(apps) > 0

	// Calculate app switching statistics
	switchStats := calculateAppSwitching(ctx, db, startTimestamp, endTimestamp, excluded)
	result.TotalSwitches = switchStats.totalSwitches
	result.AvgMinsBetween = switchStats.avgMinsBetween
	result.SwitchesPerHour = switchStats.switchesPerHour
	result.SwitchingAvailable = switchStats.available

	return result
}

// isExcluded checks if an app name is in the exclusion list
func isExcluded(appName string, excludedApps []string) bool {
	for _, excluded := range excludedApps {
		if excluded == appName {
			return true
		}
	}
	return false
}

// validBundleID matches reverse-DNS bundle identifiers (alphanumeric, dots, hyphens, underscores)
var validBundleID = regexp.MustCompile(`^[a-zA-Z0-9.\-_]+$`)

// appNameCache stores resolved app names to avoid repeated osascript calls
var appNameCache sync.Map

// resolveAppName converts a bundle ID to a human-readable app name.
// Results are cached globally so each bundle ID is resolved at most once per run.
func resolveAppName(bundleID string) string {
	if cached, ok := appNameCache.Load(bundleID); ok {
		return cached.(string)
	}

	name := resolveAppNameUncached(bundleID)
	appNameCache.Store(bundleID, name)
	return name
}

func resolveAppNameUncached(bundleID string) string {
	// Only shell out to osascript if the bundle ID is safe (no injection risk)
	if validBundleID.MatchString(bundleID) {
		cmd := exec.Command("osascript", "-e",
			fmt.Sprintf(`tell application "Finder" to get name of application file id "%s"`, bundleID))
		output, err := cmd.Output()
		if err == nil {
			name := strings.TrimSpace(string(output))
			if name != "" {
				return strings.TrimSuffix(name, ".app")
			}
		}
	}

	// Fallback: extract last component from bundle ID
	parts := strings.Split(bundleID, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}

	return bundleID
}

type appSwitchingStats struct {
	totalSwitches   int
	avgMinsBetween  float64
	switchesPerHour float64
	available       bool
}

// calculateAppSwitching calculates app switching frequency and patterns
func calculateAppSwitching(ctx context.Context, db *sql.DB, startTimestamp, endTimestamp float64, excludedApps []string) appSwitchingStats {
	stats := appSwitchingStats{available: false}

	// Query all app usage intervals ordered by time
	query := `
		SELECT 
			ZVALUESTRING as bundle_id,
			ZSTARTDATE,
			ZENDDATE
		FROM ZOBJECT
		WHERE ZSTREAMNAME = '/app/usage'
			AND ZSTARTDATE >= ?
			AND ZENDDATE <= ?
			AND ZVALUESTRING IS NOT NULL
			AND ZVALUESTRING != ''
		ORDER BY ZSTARTDATE ASC
	`

	rows, err := db.QueryContext(ctx, query, startTimestamp, endTimestamp)
	if err != nil {
		return stats
	}
	defer rows.Close()

	type focusEvent struct {
		bundleID string
		start    float64
		end      float64
	}

	var events []focusEvent

	for rows.Next() {
		var bundleID string
		var start, end float64

		if err := rows.Scan(&bundleID, &start, &end); err != nil {
			continue
		}

		// Skip system apps
		if systemApps[bundleID] {
			continue
		}

		// Skip excluded apps (resolveAppName is globally cached)
		if isExcluded(resolveAppName(bundleID), excludedApps) {
			continue
		}

		events = append(events, focusEvent{
			bundleID: bundleID,
			start:    start,
			end:      end,
		})
	}

	if len(events) < 2 {
		// Need at least 2 events to calculate switches
		return stats
	}

	// Count app switches (when bundle ID changes)
	var switches int
	var switchTimestamps []float64
	lastBundleID := events[0].bundleID

	// Initialize with the first event's timestamp as the starting point
	switchTimestamps = append(switchTimestamps, events[0].start)

	for i := 1; i < len(events); i++ {
		if events[i].bundleID != lastBundleID {
			switches++
			switchTimestamps = append(switchTimestamps, events[i].start)
		}
		lastBundleID = events[i].bundleID
	}

	if switches == 0 {
		return stats
	}

	// Calculate average time between switches
	var totalIntervalSeconds float64
	for i := 1; i < len(switchTimestamps); i++ {
		interval := switchTimestamps[i] - switchTimestamps[i-1]
		totalIntervalSeconds += interval
	}

	stats.totalSwitches = switches
	stats.avgMinsBetween = (totalIntervalSeconds / float64(switches)) / 60.0

	// Calculate switches per hour based on total active time
	totalActiveSeconds := events[len(events)-1].end - events[0].start
	if totalActiveSeconds > 0 {
		totalActiveHours := totalActiveSeconds / 3600.0
		stats.switchesPerHour = float64(switches) / totalActiveHours
	}

	stats.available = true
	return stats
}
