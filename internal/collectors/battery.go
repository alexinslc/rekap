package collectors

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// BatteryResult contains battery usage information
type BatteryResult struct {
	StartPct   int
	CurrentPct int
	PlugCount  int
	Available  bool
	IsPlugged  bool
	Error      error
}

// CollectBattery retrieves current battery status
func CollectBattery(ctx context.Context) BatteryResult {
	result := BatteryResult{Available: false}

	// Get current battery percentage using pmset
	cmd := exec.CommandContext(ctx, "pmset", "-g", "batt")
	output, err := cmd.Output()
	if err != nil {
		result.Error = fmt.Errorf("failed to read battery status: %w", err)
		return result
	}

	outputStr := string(output)
	
	// Parse current percentage
	// Output looks like: "Now drawing from 'Battery Power' -InternalBattery-0 (id=1234567)	85%; discharging; 3:45 remaining"
	// or: "Now drawing from 'AC Power' -InternalBattery-0 (id=1234567)	100%; charged; 0:00 remaining present: true"
	re := regexp.MustCompile(`(\d+)%`)
	matches := re.FindStringSubmatch(outputStr)
	if len(matches) < 2 {
		result.Error = fmt.Errorf("failed to parse battery percentage")
		return result
	}

	currentPct, err := strconv.Atoi(matches[1])
	if err != nil {
		result.Error = fmt.Errorf("failed to parse battery percentage: %w", err)
		return result
	}

	result.CurrentPct = currentPct
	result.IsPlugged = strings.Contains(outputStr, "AC Power") || strings.Contains(outputStr, "charged")
	
	// TODO: Parse pmset -g log for start percentage and plug events since midnight
	// For now, use current percentage as start (will be improved)
	result.StartPct = currentPct
	result.PlugCount = 0
	
	result.Available = true
	return result
}
