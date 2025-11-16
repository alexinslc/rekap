package collectors

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FocusResult contains focus streak information
type FocusResult struct {
	StreakMinutes int
	AppName       string
	StartTime     time.Time
	EndTime       time.Time
	Available     bool
	Error         error
}

// systemApps are excluded from focus streak calculation
var systemApps = map[string]bool{
	"com.apple.finder":               true,
	"com.apple.systempreferences":    true,
	"com.apple.preferences":          true,
	"com.apple.dock":                 true,
	"com.apple.notificationcenterui": true,
	"com.apple.Spotlight":            true,
}

// CollectFocus calculates the longest focus streak from app usage data
func CollectFocus(ctx context.Context) FocusResult {
	result := FocusResult{Available: false}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		result.Error = fmt.Errorf("failed to get home directory: %w", err)
		return result
	}

	dbPath := filepath.Join(homeDir, "Library", "Application Support", "Knowledge", "knowledgeC.db")

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		result.Error = fmt.Errorf("screen Time database not found")
		return result
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to open database: %w", err)
		return result
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil && result.Error == nil {
			result.Error = fmt.Errorf("failed to close database: %w", closeErr)
		}
	}()

	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	coreDataEpoch := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)

	startTimestamp := midnight.Sub(coreDataEpoch).Seconds()
	endTimestamp := now.Sub(coreDataEpoch).Seconds()

	// Get all app usage intervals ordered by time
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
		result.Error = fmt.Errorf("failed to query data: %w", err)
		return result
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && result.Error == nil {
			result.Error = fmt.Errorf("failed to close rows: %w", closeErr)
		}
	}()

	type interval struct {
		bundleID string
		start    float64
		end      float64
		minutes  int
	}

	var intervals []interval
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

		minutes := int((end - start) / 60)
		if minutes > 0 {
			intervals = append(intervals, interval{
				bundleID: bundleID,
				start:    start,
				end:      end,
				minutes:  minutes,
			})
		}
	}

	if len(intervals) == 0 {
		result.Error = fmt.Errorf("no app usage data found")
		return result
	}

	// Find longest continuous streak for same app
	maxStreak := 0
	maxStreakApp := ""
	maxStreakStart := 0.0
	maxStreakEnd := 0.0
	currentStreak := 0
	currentApp := ""
	currentStreakStart := 0.0
	lastEnd := 0.0

	for _, iv := range intervals {
		gap := int((iv.start - lastEnd) / 60) // gap in minutes

		// If same app and gap < 30 seconds (0.5 minutes), continue streak
		if iv.bundleID == currentApp && gap < 1 {
			currentStreak += iv.minutes
		} else {
			// New streak
			if currentStreak > maxStreak {
				maxStreak = currentStreak
				maxStreakApp = currentApp
				maxStreakStart = currentStreakStart
				maxStreakEnd = lastEnd
			}
			currentApp = iv.bundleID
			currentStreak = iv.minutes
			currentStreakStart = iv.start
		}

		lastEnd = iv.end
	}

	// Check final streak
	if currentStreak > maxStreak {
		maxStreak = currentStreak
		maxStreakApp = currentApp
		maxStreakStart = currentStreakStart
		maxStreakEnd = lastEnd
	}

	if maxStreak > 0 {
		result.StreakMinutes = maxStreak
		result.AppName = resolveAppName(maxStreakApp)
		// Convert Core Data timestamps back to Go time.Time
		result.StartTime = coreDataEpoch.Add(time.Duration(maxStreakStart) * time.Second)
		result.EndTime = coreDataEpoch.Add(time.Duration(maxStreakEnd) * time.Second)
		result.Available = true
	} else {
		result.Error = fmt.Errorf("no focus streaks found")
	}

	return result
}
