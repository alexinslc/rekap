package collectors

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

// UptimeResult contains system uptime information
type UptimeResult struct {
	BootTime      time.Time
	AwakeMinutes  int
	FormattedTime string
	Available     bool
	Error         error
}

// CollectUptime retrieves system boot time and calculates awake time since midnight
func CollectUptime(ctx context.Context) UptimeResult {
	result := UptimeResult{Available: false}

	// Read kernel boot time via sysctl
	cmd := exec.CommandContext(ctx, "sysctl", "-n", "kern.boottime")
	output, err := cmd.Output()
	if err != nil {
		result.Error = fmt.Errorf("failed to read boot time: %w", err)
		return result
	}

	// Parse output like: { sec = 1699300000, usec = 0 } Thu Nov  6 12:00:00 2024
	re := regexp.MustCompile(`sec = (\d+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		result.Error = fmt.Errorf("failed to parse boot time")
		return result
	}

	bootTimeSec, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		result.Error = fmt.Errorf("failed to parse boot time seconds: %w", err)
		return result
	}

	result.BootTime = time.Unix(bootTimeSec, 0)

	// Calculate awake time since midnight (basic version, no sleep tracking yet)
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// If system booted before midnight, use time since midnight
	// If system booted after midnight, use time since boot
	var awakeStart time.Time
	if result.BootTime.Before(midnight) {
		awakeStart = midnight
	} else {
		awakeStart = result.BootTime
	}

	awakeDuration := now.Sub(awakeStart)
	result.AwakeMinutes = int(awakeDuration.Minutes())

	// Format time
	hours := result.AwakeMinutes / 60
	mins := result.AwakeMinutes % 60
	if hours > 0 {
		result.FormattedTime = fmt.Sprintf("%dh %dm awake", hours, mins)
	} else {
		result.FormattedTime = fmt.Sprintf("%dm awake", mins)
	}

	result.Available = true
	return result
}
