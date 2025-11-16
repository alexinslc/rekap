package collectors

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// AppUsage represents usage time for a single app
type AppUsage struct {
	Name     string
	Minutes  int
	BundleID string
}

// AppsResult contains app usage information
type AppsResult struct {
	TopApps           []AppUsage
	Source            string // "ScreenTime" or "Sampling"
	Available         bool
	Error             error
	ExcludedApps      []string // Apps that were filtered out
	TotalSwitches     int      // Total number of app switches today
	AvgMinsBetween    float64  // Average minutes between switches
	SwitchesPerHour   float64  // Switches per hour rate
	SwitchingAvailable bool    // Whether switching data is available
}

// CollectApps retrieves top app usage from Screen Time database
// excludedApps is an optional list of app names to filter out
func CollectApps(ctx context.Context, excludedApps ...[]string) AppsResult {
	result := AppsResult{Available: false, Source: "ScreenTime"}

	// Flatten excluded apps list if provided
	var excluded []string
	if len(excludedApps) > 0 {
		excluded = excludedApps[0]
		result.ExcludedApps = excluded
	}

	// Try KnowledgeC database first
	homeDir, err := os.UserHomeDir()
	if err != nil {
		result.Error = fmt.Errorf("failed to get home directory: %w", err)
		return result
	}

	dbPath := filepath.Join(homeDir, "Library", "Application Support", "Knowledge", "knowledgeC.db")

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		result.Error = fmt.Errorf("screen Time database not found (requires Full Disk Access)")
		return result
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to open Screen Time database: %w", err)
		return result
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil && result.Error == nil {
			result.Error = fmt.Errorf("failed to close database: %w", closeErr)
		}
	}()

	// Calculate today's timestamp range in Core Data format (seconds since 2001-01-01)
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	coreDataEpoch := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)

	startTimestamp := midnight.Sub(coreDataEpoch).Seconds()
	endTimestamp := now.Sub(coreDataEpoch).Seconds()

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

// resolveAppName converts a bundle ID to a human-readable app name
func resolveAppName(bundleID string) string {
	// Try to get app name from system
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "Finder" to get name of application file id "%s"`, bundleID))
	output, err := cmd.Output()
	if err == nil {
		name := strings.TrimSpace(string(output))
		if name != "" {
			return strings.TrimSuffix(name, ".app")
		}
	}

	// Fallback: try to extract from bundle ID
	// com.apple.Safari -> Safari
	// com.microsoft.VSCode -> VSCode
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

	// System apps to exclude from switching calculation
	systemApps := map[string]bool{
		"com.apple.finder":               true,
		"com.apple.systempreferences":    true,
		"com.apple.preferences":          true,
		"com.apple.dock":                 true,
		"com.apple.notificationcenterui": true,
		"com.apple.Spotlight":            true,
	}

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

		// Skip excluded apps
		appName := resolveAppName(bundleID)
		if isExcluded(appName, excludedApps) {
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
	var totalGapSeconds float64
	lastBundleID := events[0].bundleID

	for i := 1; i < len(events); i++ {
		if events[i].bundleID != lastBundleID {
			switches++
			// Calculate gap between end of last app and start of new app
			gap := events[i].start - events[i-1].end
			totalGapSeconds += gap
		}
		lastBundleID = events[i].bundleID
	}

	if switches == 0 {
		return stats
	}

	// Calculate statistics
	stats.totalSwitches = switches
	stats.avgMinsBetween = (totalGapSeconds / float64(switches)) / 60.0
	
	// Calculate switches per hour based on total active time
	totalActiveSeconds := events[len(events)-1].end - events[0].start
	if totalActiveSeconds > 0 {
		totalActiveHours := totalActiveSeconds / 3600.0
		stats.switchesPerHour = float64(switches) / totalActiveHours
	}

	stats.available = true
	return stats
}
