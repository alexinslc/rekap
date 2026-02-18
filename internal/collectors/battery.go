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
	result.Available = true

	// Parse pmset log for start percentage and plug events since midnight
	startPct, plugCount := parsePmsetLog(ctx)
	if startPct >= 0 {
		result.StartPct = startPct
	} else {
		result.StartPct = currentPct
	}
	result.PlugCount = plugCount

	return result
}

// pmset log charge pattern: "Using AC(Charge: 80)" or "Using AC (Charge:80%)" or "Using Batt(Charge: 100)"
var chargePattern = regexp.MustCompile(`Using (AC|Batt).*?Charge:\s*(\d+)`)

// pmset log timestamp pattern: "2026-02-17 14:30:22 -0700"
var timestampPattern = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})`)

// parsePmsetLog reads pmset log to find the first battery charge after midnight
// and count AC plug-in events. Returns (startPct, plugCount) where startPct is -1
// if no data found.
func parsePmsetLog(ctx context.Context) (int, int) {
	// Use grep to filter relevant lines before processing (keeps it fast on large logs)
	cmd := exec.CommandContext(ctx, "bash", "-c", "pmset -g log 2>/dev/null | grep -E 'Using (AC|Batt)'")
	output, err := cmd.Output()
	if err != nil {
		return -1, 0
	}

	return parsePmsetLogOutput(string(output))
}

// parsePmsetLogOutput parses filtered pmset log output for today's battery data.
// Returns (startPct, plugCount) where startPct is -1 if no data found.
func parsePmsetLogOutput(output string) (int, int) {
	today := time.Now().Format("2006-01-02")
	startPct := -1
	plugCount := 0
	lastSource := "" // "AC" or "Batt"

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		// Only process lines from today
		tsMatches := timestampPattern.FindStringSubmatch(line)
		if len(tsMatches) < 2 {
			continue
		}
		if !strings.HasPrefix(tsMatches[1], today) {
			continue
		}

		chargeMatches := chargePattern.FindStringSubmatch(line)
		if len(chargeMatches) < 3 {
			continue
		}

		source := chargeMatches[1] // "AC" or "Batt"
		pct, err := strconv.Atoi(chargeMatches[2])
		if err != nil {
			continue
		}

		// First charge reading of the day is our start percentage
		if startPct < 0 {
			startPct = pct
		}

		// Count transitions from Batt to AC as plug events
		if source == "AC" && lastSource == "Batt" {
			plugCount++
		}
		lastSource = source
	}

	return startPct, plugCount
}
