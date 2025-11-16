package collectors

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BurnoutWarning represents a specific burnout indicator
type BurnoutWarning struct {
	Type        string // "long_day", "high_switching", "tab_overload", "late_night", "no_breaks"
	Message     string
	Severity    string // "low", "medium", "high"
	Detected    bool
	MetricValue int // The actual value that triggered the warning
}

// BurnoutResult contains burnout detection information
type BurnoutResult struct {
	Warnings  []BurnoutWarning
	Available bool
	Error     error
}

// BurnoutConfig contains thresholds for burnout detection
type BurnoutConfig struct {
	LongDayHours         int // Default: 10 hours
	AppSwitchesPerHour   int // Default: 50 switches/hour
	MaxTabs              int // Default: 100 tabs
	LateNightHour        int // Default: 0 (midnight)
	NoBreakHours         int // Default: 4 hours
}

// DefaultBurnoutConfig returns default burnout detection thresholds
func DefaultBurnoutConfig() BurnoutConfig {
	return BurnoutConfig{
		LongDayHours:       10,
		AppSwitchesPerHour: 50,
		MaxTabs:            100,
		LateNightHour:      0,
		NoBreakHours:       4,
	}
}

// CollectBurnout analyzes activity patterns for burnout indicators
func CollectBurnout(ctx context.Context, screen ScreenResult, browsers BrowsersResult, config BurnoutConfig) BurnoutResult {
	result := BurnoutResult{
		Warnings:  []BurnoutWarning{},
		Available: true,
	}

	// Check 1: Long work day (>10h screen-on)
	if screen.Available {
		longDayHours := screen.ScreenOnMinutes / 60
		if longDayHours >= config.LongDayHours {
			result.Warnings = append(result.Warnings, BurnoutWarning{
				Type:        "long_day",
				Message:     fmt.Sprintf("Long work day: %dh+ screen time", longDayHours),
				Severity:    "medium",
				Detected:    true,
				MetricValue: longDayHours,
			})
		}
	}

	// Check 2: High app switching rate (>50 switches/hour)
	appSwitchRate, err := calculateAppSwitchRate(ctx)
	if err == nil && appSwitchRate > 0 {
		if appSwitchRate >= config.AppSwitchesPerHour {
			result.Warnings = append(result.Warnings, BurnoutWarning{
				Type:        "high_switching",
				Message:     fmt.Sprintf("High task switching: %d app switches/hour", appSwitchRate),
				Severity:    "medium",
				Detected:    true,
				MetricValue: appSwitchRate,
			})
		}
	}

	// Check 3: Tab overload (>100 tabs)
	if browsers.Available && browsers.TotalTabs >= config.MaxTabs {
		result.Warnings = append(result.Warnings, BurnoutWarning{
			Type:        "tab_overload",
			Message:     fmt.Sprintf("Browser overload: %d open tabs", browsers.TotalTabs),
			Severity:    "low",
			Detected:    true,
			MetricValue: browsers.TotalTabs,
		})
	}

	// Check 4: Late night work (activity past midnight)
	lateNightMinutes, err := detectLateNightWork(ctx)
	if err == nil && lateNightMinutes > 0 {
		result.Warnings = append(result.Warnings, BurnoutWarning{
			Type:        "late_night",
			Message:     fmt.Sprintf("Late night work: %d minutes past midnight", lateNightMinutes),
			Severity:    "high",
			Detected:    true,
			MetricValue: lateNightMinutes,
		})
	}

	// Check 5: No breaks (continuous focus >4h)
	longestStreak, err := calculateLongestNoBreakPeriod(ctx)
	if err == nil && longestStreak >= config.NoBreakHours*60 {
		result.Warnings = append(result.Warnings, BurnoutWarning{
			Type:        "no_breaks",
			Message:     fmt.Sprintf("No breaks: %dh+ continuous focus", longestStreak/60),
			Severity:    "high",
			Detected:    true,
			MetricValue: longestStreak / 60,
		})
	}

	return result
}

// calculateAppSwitchRate calculates the number of app switches per hour
func calculateAppSwitchRate(ctx context.Context) (int, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return 0, fmt.Errorf("failed to get home directory: %w", err)
	}

	dbPath := filepath.Join(homeDir, "Library", "Application Support", "Knowledge", "knowledgeC.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return 0, fmt.Errorf("screen Time database not found")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	coreDataEpoch := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)

	startTimestamp := midnight.Sub(coreDataEpoch).Seconds()
	endTimestamp := now.Sub(coreDataEpoch).Seconds()

	// Count distinct app usage events (each represents a switch)
	query := `
		SELECT COUNT(*) as switch_count
		FROM ZOBJECT
		WHERE ZSTREAMNAME = '/app/usage'
			AND ZSTARTDATE >= ?
			AND ZENDDATE <= ?
			AND ZVALUESTRING IS NOT NULL
			AND ZVALUESTRING != ''
	`

	var switchCount int
	err = db.QueryRowContext(ctx, query, startTimestamp, endTimestamp).Scan(&switchCount)
	if err != nil {
		return 0, fmt.Errorf("failed to query switch count: %w", err)
	}

	// Calculate rate per hour
	hoursActive := time.Since(midnight).Hours()
	if hoursActive < 1 {
		hoursActive = 1
	}

	rate := int(float64(switchCount) / hoursActive)
	return rate, nil
}

// detectLateNightWork detects app usage past midnight (00:00-06:00)
func detectLateNightWork(ctx context.Context) (int, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return 0, fmt.Errorf("failed to get home directory: %w", err)
	}

	dbPath := filepath.Join(homeDir, "Library", "Application Support", "Knowledge", "knowledgeC.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return 0, fmt.Errorf("screen Time database not found")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	earlyMorning := midnight.Add(6 * time.Hour) // 06:00
	coreDataEpoch := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)

	// Only check if current time is within the late night window
	if now.Hour() >= 6 {
		return 0, nil // Not in late night period
	}

	startTimestamp := midnight.Sub(coreDataEpoch).Seconds()
	endTimestamp := earlyMorning.Sub(coreDataEpoch).Seconds()

	// Sum up activity time in late night hours
	query := `
		SELECT SUM(ZENDDATE - ZSTARTDATE) as total_seconds
		FROM ZOBJECT
		WHERE ZSTREAMNAME = '/app/usage'
			AND ZSTARTDATE >= ?
			AND ZENDDATE <= ?
			AND ZVALUESTRING IS NOT NULL
			AND ZVALUESTRING != ''
	`

	var totalSeconds sql.NullFloat64
	err = db.QueryRowContext(ctx, query, startTimestamp, endTimestamp).Scan(&totalSeconds)
	if err != nil {
		return 0, fmt.Errorf("failed to query late night activity: %w", err)
	}

	if !totalSeconds.Valid {
		return 0, nil
	}

	return int(totalSeconds.Float64 / 60), nil // Return minutes
}

// calculateLongestNoBreakPeriod finds the longest continuous work period without breaks
func calculateLongestNoBreakPeriod(ctx context.Context) (int, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return 0, fmt.Errorf("failed to get home directory: %w", err)
	}

	dbPath := filepath.Join(homeDir, "Library", "Application Support", "Knowledge", "knowledgeC.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return 0, fmt.Errorf("screen Time database not found")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	coreDataEpoch := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)

	startTimestamp := midnight.Sub(coreDataEpoch).Seconds()
	endTimestamp := now.Sub(coreDataEpoch).Seconds()

	// Get all app usage intervals ordered by time
	query := `
		SELECT ZSTARTDATE, ZENDDATE
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
		return 0, fmt.Errorf("failed to query intervals: %w", err)
	}
	defer rows.Close()

	type interval struct {
		start float64
		end   float64
	}

	var intervals []interval
	for rows.Next() {
		var start, end float64
		if err := rows.Scan(&start, &end); err != nil {
			continue
		}
		intervals = append(intervals, interval{start, end})
	}

	if len(intervals) == 0 {
		return 0, nil
	}

	// Find longest continuous period (gaps < 15 minutes are considered continuous)
	const maxGapMinutes = 15
	maxPeriod := 0
	currentPeriodStart := intervals[0].start
	currentPeriodEnd := intervals[0].end

	for i := 1; i < len(intervals); i++ {
		gap := int((intervals[i].start - currentPeriodEnd) / 60)

		if gap <= maxGapMinutes {
			// Continue current period
			currentPeriodEnd = intervals[i].end
		} else {
			// End current period, calculate duration
			periodMinutes := int((currentPeriodEnd - currentPeriodStart) / 60)
			if periodMinutes > maxPeriod {
				maxPeriod = periodMinutes
			}

			// Start new period
			currentPeriodStart = intervals[i].start
			currentPeriodEnd = intervals[i].end
		}
	}

	// Check final period
	periodMinutes := int((currentPeriodEnd - currentPeriodStart) / 60)
	if periodMinutes > maxPeriod {
		maxPeriod = periodMinutes
	}

	return maxPeriod, nil
}
