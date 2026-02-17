package collectors

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// coreDataEpoch is Apple's Core Data epoch (2001-01-01 00:00:00 UTC)
var coreDataEpoch = time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)

// systemApps are excluded from focus streak and switching calculations
var systemApps = map[string]bool{
	"com.apple.finder":               true,
	"com.apple.systempreferences":    true,
	"com.apple.preferences":          true,
	"com.apple.dock":                 true,
	"com.apple.notificationcenterui": true,
	"com.apple.Spotlight":            true,
}

// openKnowledgeDB opens the macOS Screen Time knowledgeC.db database.
// Callers are responsible for closing the returned *sql.DB.
func openKnowledgeDB() (*sql.DB, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	dbPath := filepath.Join(homeDir, "Library", "Application Support", "Knowledge", "knowledgeC.db")

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Screen Time database not found (requires Full Disk Access)")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Screen Time database: %w", err)
	}

	return db, nil
}

// todayTimestampRange returns the Core Data timestamp range for today
// (from midnight to now), as seconds since the Core Data epoch (2001-01-01).
func todayTimestampRange() (start, end float64) {
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	start = midnight.Sub(coreDataEpoch).Seconds()
	end = now.Sub(coreDataEpoch).Seconds()
	return start, end
}
