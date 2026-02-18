package collectors

import (
	"context"
	"fmt"
)

// FocusResult contains focus streak information
type FocusResult struct {
	StreakMinutes int
	AppName       string
	Available     bool
	Error         error
}

// CollectFocus calculates the longest focus streak from app usage data
func CollectFocus(ctx context.Context) FocusResult {
	result := FocusResult{Available: false}

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
	currentStreak := 0
	currentApp := ""
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
			}
			currentApp = iv.bundleID
			currentStreak = iv.minutes
		}

		lastEnd = iv.end
	}

	// Check final streak
	if currentStreak > maxStreak {
		maxStreak = currentStreak
		maxStreakApp = currentApp
	}

	if maxStreak > 0 {
		result.StreakMinutes = maxStreak
		result.AppName = resolveAppName(maxStreakApp)
		result.Available = true
	} else {
		result.Error = fmt.Errorf("no focus streaks found")
	}

	return result
}
