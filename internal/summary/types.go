package summary

import "github.com/alexinslc/rekap/internal/collectors"

// Data holds all collector results for a single run.
// Shared between cmd/rekap and internal/ui/tui to avoid duplication.
type Data struct {
	Uptime        collectors.UptimeResult
	Battery       collectors.BatteryResult
	Screen        collectors.ScreenResult
	Apps          collectors.AppsResult
	Focus         collectors.FocusResult
	Media         collectors.MediaResult
	Network       collectors.NetworkResult
	Browsers      collectors.BrowsersResult
	Notifications collectors.NotificationsResult
	Issues        collectors.IssuesResult
	Fragmentation collectors.FragmentationResult
	Burnout       collectors.BurnoutResult
}
