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
	Name    string
	Minutes int
	BundleID string
}

// AppsResult contains app usage information
type AppsResult struct {
	TopApps   []AppUsage
	Source    string // "ScreenTime" or "Sampling"
	Available bool
	Error     error
}

// CollectApps retrieves top app usage from Screen Time database
func CollectApps(ctx context.Context) AppsResult {
	result := AppsResult{Available: false, Source: "ScreenTime"}

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
		if err := db.Close(); err != nil {
			result.Error = fmt.Errorf("failed to close database: %w", err)
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
		if err := rows.Close(); err != nil {
			result.Error = fmt.Errorf("failed to close rows: %w", err)
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
	return result
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
