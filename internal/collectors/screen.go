package collectors

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// ScreenResult contains screen-on time information
type ScreenResult struct {
	ScreenOnMinutes int
	Available       bool
	Error           error
}

// CollectScreen retrieves screen-on time since midnight
func CollectScreen(ctx context.Context) ScreenResult {
	result := ScreenResult{Available: false}

	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Parse pmset log for display events since midnight
	cmd := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf("pmset -g log | grep -i 'display' | grep '%s'", midnight.Format("2006-01-02")))
	output, err := cmd.Output()
	if err != nil {
		// pmset log might not be available or grep found nothing
		// Try alternative: assume screen has been on since midnight (rough estimate)
		result.ScreenOnMinutes = int(time.Since(midnight).Minutes())
		result.Available = true
		result.Error = fmt.Errorf("pmset log unavailable, using rough estimate: %w", err)
		return result
	}

	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	
	var totalMinutes int
	var lastOnTime time.Time
	isOn := false
	
	// Parse display on/off events
	timeRe := regexp.MustCompile(`(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})`)
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		matches := timeRe.FindStringSubmatch(line)
		if len(matches) < 2 {
			continue
		}
		
		eventTime, err := time.ParseInLocation("2006-01-02 15:04:05", matches[1], time.Local)
		if err != nil {
			continue
		}
		
		// Detect display on/off from log entries
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "display is turned on") || 
		   strings.Contains(lowerLine, "backlight level") && !strings.Contains(lowerLine, "level 0") {
			if !isOn {
				lastOnTime = eventTime
				isOn = true
			}
		} else if strings.Contains(lowerLine, "display is turned off") ||
		          strings.Contains(lowerLine, "display sleep") {
			if isOn && !lastOnTime.IsZero() {
				duration := eventTime.Sub(lastOnTime)
				totalMinutes += int(duration.Minutes())
				isOn = false
			}
		}
	}
	
	// If display is currently on, add time until now
	if isOn && !lastOnTime.IsZero() {
		duration := now.Sub(lastOnTime)
		totalMinutes += int(duration.Minutes())
	}
	
	// If we have no data, fall back to rough estimate
	if totalMinutes == 0 {
		totalMinutes = int(time.Since(midnight).Minutes())
		result.Error = fmt.Errorf("no display events parsed, using estimate")
	}
	
	result.ScreenOnMinutes = totalMinutes
	result.Available = true
	return result
}
