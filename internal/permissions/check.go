package permissions

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Capabilities represents the available permissions and capabilities
type Capabilities struct {
	FullDiskAccess bool
	Accessibility  bool
	NowPlaying     bool
}

// Check returns the current permission status for all capabilities
func Check() Capabilities {
	return Capabilities{
		FullDiskAccess: checkFullDiskAccess(),
		Accessibility:  checkAccessibility(),
		NowPlaying:     checkNowPlaying(),
	}
}

// checkFullDiskAccess tests if we can read the Screen Time database
func checkFullDiskAccess() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	dbPath := filepath.Join(homeDir, "Library", "Application Support", "Knowledge", "knowledgeC.db")
	
	// Try to open the file for reading
	file, err := os.Open(dbPath)
	if err != nil {
		return false
	}
	defer file.Close()

	// Try to read a byte to ensure we have actual read access
	buf := make([]byte, 1)
	_, err = file.Read(buf)
	return err == nil
}

// checkAccessibility tests if we have Accessibility permission
// Note: This is a rough check - real Accessibility API would need CGo
func checkAccessibility() bool {
	// Try to use AppleScript to check if we can access UI elements
	cmd := exec.Command("osascript", "-e", `
		tell application "System Events"
			try
				get name of first process
				return "true"
			on error
				return "false"
			end try
		end tell
	`)
	
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(output)) == "true"
}

// checkNowPlaying tests if we can access media information
func checkNowPlaying() bool {
	// Try to query Music app
	cmd := exec.Command("osascript", "-e", `
		tell application "Music"
			try
				get player state
				return "true"
			on error
				return "false"
			end try
		end tell
	`)
	
	output, err := cmd.Output()
	if err != nil {
		// This might fail if Music isn't installed, but that's ok
		// Check if nowplaying-cli is available as alternative
		cmd = exec.Command("which", "nowplaying-cli")
		err = cmd.Run()
		return err == nil
	}

	return strings.TrimSpace(string(output)) == "true"
}

// GetCapabilitiesMatrix returns a map of capability names to status
func GetCapabilitiesMatrix() map[string]bool {
	caps := Check()
	return map[string]bool{
		"uptime":          true, // Always available
		"battery":         true, // Always available
		"screen_on":       caps.FullDiskAccess,
		"apps":            caps.FullDiskAccess,
		"focus_streak":    caps.FullDiskAccess,
		"accessibility":   caps.Accessibility,
		"media":           caps.NowPlaying,
	}
}

// FormatCapabilities returns a human-readable string of capabilities
func FormatCapabilities(caps Capabilities) string {
	var lines []string
	
	lines = append(lines, fmt.Sprintf("✓ uptime          (kernel boot time)"))
	lines = append(lines, fmt.Sprintf("✓ battery         (power management)"))
	
	if caps.FullDiskAccess {
		lines = append(lines, fmt.Sprintf("✓ screen_on       (Full Disk Access)"))
		lines = append(lines, fmt.Sprintf("✓ apps            (Screen Time data)"))
		lines = append(lines, fmt.Sprintf("✓ focus_streak    (Screen Time data)"))
	} else {
		lines = append(lines, fmt.Sprintf("✗ screen_on       (needs Full Disk Access)"))
		lines = append(lines, fmt.Sprintf("✗ apps            (needs Full Disk Access)"))
		lines = append(lines, fmt.Sprintf("✗ focus_streak    (needs Full Disk Access)"))
	}
	
	if caps.Accessibility {
		lines = append(lines, fmt.Sprintf("✓ accessibility   (UI element access)"))
	} else {
		lines = append(lines, fmt.Sprintf("✗ accessibility   (not granted)"))
	}
	
	if caps.NowPlaying {
		lines = append(lines, fmt.Sprintf("✓ media           (Now Playing)"))
	} else {
		lines = append(lines, fmt.Sprintf("✗ media           (Music app or nowplaying-cli)"))
	}
	
	return strings.Join(lines, "\n")
}
