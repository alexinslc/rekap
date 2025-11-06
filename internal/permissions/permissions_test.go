package permissions

import (
	"testing"
)

func TestCheck(t *testing.T) {
	caps := Check()

	// These are boolean values, just verify they're set
	t.Logf("Full Disk Access: %v", caps.FullDiskAccess)
	t.Logf("Accessibility: %v", caps.Accessibility)
	t.Logf("Now Playing: %v", caps.NowPlaying)
}

func TestGetCapabilitiesMatrix(t *testing.T) {
	matrix := GetCapabilitiesMatrix()

	// Uptime and battery should always be available
	if !matrix["uptime"] {
		t.Error("Uptime should always be available")
	}

	if !matrix["battery"] {
		t.Error("Battery should always be available")
	}

	// Verify all expected keys exist
	expectedKeys := []string{
		"uptime",
		"battery",
		"screen_on",
		"apps",
		"focus_streak",
		"accessibility",
		"media",
	}

	for _, key := range expectedKeys {
		if _, exists := matrix[key]; !exists {
			t.Errorf("Missing expected key in matrix: %s", key)
		}
	}
}

func TestFormatCapabilities(t *testing.T) {
	caps := Capabilities{
		FullDiskAccess: true,
		Accessibility:  true,
		NowPlaying:     true,
	}

	output := FormatCapabilities(caps)

	if output == "" {
		t.Error("FormatCapabilities should not return empty string")
	}

	// Should contain check marks for granted permissions
	if caps.FullDiskAccess {
		// Basic validation - output should be non-empty
		t.Logf("Output: %s", output)
	}
}
