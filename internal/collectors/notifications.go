package collectors

import (
	"context"
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"time"
)

// NotificationApp represents notification count for a single app
type NotificationApp struct {
	Name     string
	Count    int
	BundleID string
}

// NotificationsResult contains notification interruption information
type NotificationsResult struct {
	TotalNotifications int
	TopApps            []NotificationApp
	Available          bool
	Error              error
}

// CollectNotifications retrieves notification counts from Screen Time database
func CollectNotifications(ctx context.Context) NotificationsResult {
	result := NotificationsResult{Available: false}

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		result.Error = fmt.Errorf("failed to get home directory: %w", err)
		return result
	}

	dbPath := filepath.Join(homeDir, "Library", "Application Support", "Knowledge", "knowledgeC.db")

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		result.Error = fmt.Errorf("Screen Time database not found (requires Full Disk Access)")
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

	// Query for notification events
	// ZSTREAMNAME = '/notification/usage' contains notification events
	// ZVALUESTRING contains event types like 'Receive', 'DefaultAction', etc.
	// We want to count 'Receive' events which represent incoming notifications
	query := `
		SELECT 
			COALESCE(sm.Z_DKNOTIFICATIONAPPMETADATAKEY__BUNDLEIDENTIFIER, 'unknown') as bundle_id,
			COUNT(*) as notification_count
		FROM ZOBJECT zo
		LEFT JOIN ZSTRUCTUREDMETADATA sm ON zo.ZSTRUCTUREDMETADATA = sm.Z_PK
		WHERE zo.ZSTREAMNAME = '/notification/usage'
			AND zo.ZSTARTDATE >= ?
			AND zo.ZSTARTDATE <= ?
			AND zo.ZVALUESTRING = 'Receive'
		GROUP BY bundle_id
		ORDER BY notification_count DESC
	`

	rows, err := db.QueryContext(ctx, query, startTimestamp, endTimestamp)
	if err != nil {
		result.Error = fmt.Errorf("failed to query notification data: %w", err)
		return result
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && result.Error == nil {
			result.Error = fmt.Errorf("failed to close rows: %w", closeErr)
		}
	}()

	var apps []NotificationApp
	totalCount := 0

	for rows.Next() {
		var bundleID string
		var count int

		if err := rows.Scan(&bundleID, &count); err != nil {
			continue
		}

		totalCount += count

		// Resolve bundle ID to app name
		appName := resolveAppName(bundleID)

		apps = append(apps, NotificationApp{
			Name:     appName,
			Count:    count,
			BundleID: bundleID,
		})
	}

	// Check for errors encountered during iteration
	if err := rows.Err(); err != nil {
		result.Error = fmt.Errorf("error iterating notification data: %w", err)
		return result
	}
	result.TotalNotifications = totalCount
	result.TopApps = apps
	result.Available = true

	// Get notifications during focus periods (optional enhancement)
	// This would require correlating notification timestamps with focus streaks
	// For now, we skip this calculation to keep the implementation simple

	return result
}
