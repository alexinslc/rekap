package collectors

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
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

	// Calculate awake time since midnight, subtracting sleep periods
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var awakeStart time.Time
	if result.BootTime.Before(midnight) {
		awakeStart = midnight
	} else {
		awakeStart = result.BootTime
	}

	awakeDuration := now.Sub(awakeStart)

	// Subtract sleep time from awake duration
	sleepDuration := collectSleepDuration(ctx, awakeStart, now)
	awakeDuration -= sleepDuration
	if awakeDuration < 0 {
		awakeDuration = 0
	}

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

var sleepPattern = regexp.MustCompile(`\bSleep\b`)
var wakePattern = regexp.MustCompile(`\bWake\b`)

// collectSleepDuration runs pmset -g log and returns total sleep time between start and end.
func collectSleepDuration(ctx context.Context, start, end time.Time) time.Duration {
	cmd := exec.CommandContext(ctx, "bash", "-c", "pmset -g log 2>/dev/null | grep -E '(Sleep|Wake)\\b'")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	return parseSleepWakeEvents(string(output), start, end)
}

// parseSleepWakeEvents parses filtered pmset log output and returns total sleep duration
// between start and end times. Exported for testing.
func parseSleepWakeEvents(output string, start, end time.Time) time.Duration {
	today := start.Format("2006-01-02")
	var totalSleep time.Duration
	var sleepStart time.Time
	inSleep := false

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		tsMatches := timestampPattern.FindStringSubmatch(line)
		if len(tsMatches) < 2 {
			continue
		}
		if !strings.HasPrefix(tsMatches[1], today) {
			continue
		}

		ts, err := time.ParseInLocation("2006-01-02 15:04:05", tsMatches[1], start.Location())
		if err != nil {
			continue
		}

		if ts.Before(start) || ts.After(end) {
			continue
		}

		isSleep := sleepPattern.MatchString(line)
		isWake := wakePattern.MatchString(line)

		if isSleep && !inSleep {
			sleepStart = ts
			inSleep = true
		} else if isWake && inSleep {
			totalSleep += ts.Sub(sleepStart)
			inSleep = false
		}
	}

	// If still in sleep at end time, count up to end
	if inSleep {
		totalSleep += end.Sub(sleepStart)
	}

	return totalSleep
}
