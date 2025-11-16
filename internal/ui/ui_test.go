package ui

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	t.Parallel()
	tests := []struct {
		minutes  int
		expected string
	}{
		{0, "0m"},
		{5, "5m"},
		{59, "59m"},
		{60, "1h 0m"},
		{65, "1h 5m"},
		{120, "2h 0m"},
		{125, "2h 5m"},
	}

	for _, tt := range tests {
		result := FormatDuration(tt.minutes)
		if result != tt.expected {
			t.Errorf("FormatDuration(%d) = %s, want %s", tt.minutes, result, tt.expected)
		}
	}
}

func TestFormatDurationCompact(t *testing.T) {
	t.Parallel()
	tests := []struct {
		minutes  int
		expected string
	}{
		{0, "0m"},
		{5, "5m"},
		{59, "59m"},
		{60, "1h"},
		{65, "1h5m"},
		{120, "2h"},
		{125, "2h5m"},
	}

	for _, tt := range tests {
		result := FormatDurationCompact(tt.minutes)
		if result != tt.expected {
			t.Errorf("FormatDurationCompact(%d) = %s, want %s", tt.minutes, result, tt.expected)
		}
	}
}

func TestRenderDataPoint(t *testing.T) {
	t.Parallel()
	result := RenderDataPoint("🔋", "Battery: 75%")
	if result == "" {
		t.Error("RenderDataPoint should not return empty string")
	}

	// Should contain both icon and text
	if len(result) < 5 {
		t.Error("RenderDataPoint output seems too short")
	}
}

func TestRenderHint(t *testing.T) {
	t.Parallel()
	result := RenderHint("Test hint message")
	if result == "" {
		t.Error("RenderHint should not return empty string")
	}
}

func TestRenderError(t *testing.T) {
	t.Parallel()
	result := RenderError("Test error message")
	if result == "" {
		t.Error("RenderError should not return empty string")
	}
}

func TestRenderWarning(t *testing.T) {
	t.Parallel()
	result := RenderWarning("Context overload: 7 apps + 45 tabs active")
	if result == "" {
		t.Error("RenderWarning should not return empty string")
	}
	// Should contain the warning emoji
	if len(result) < 10 {
		t.Error("RenderWarning output seems too short")
	}
}

func TestRenderSummaryLine(t *testing.T) {
	t.Parallel()
	parts := []string{"3h 35m screen-on", "2 plug-ins", "Top apps: VS Code"}
	result := RenderSummaryLine(parts)

	if result == "" {
		t.Error("RenderSummaryLine should not return empty string")
	}
}

func TestRenderSummaryLineEmpty(t *testing.T) {
	t.Parallel()
	result := RenderSummaryLine([]string{})

	if result != "" {
		t.Error("RenderSummaryLine should return empty string for empty input")
	}
}

func TestFormatTime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		timeFormat string
		hour       int
		minute     int
		expected   string
	}{
		{"12h format morning", "12h", 9, 30, "9:30 AM"},
		{"12h format afternoon", "12h", 15, 45, "3:45 PM"},
		{"12h format midnight", "12h", 0, 0, "12:00 AM"},
		{"12h format noon", "12h", 12, 0, "12:00 PM"},
		{"24h format morning", "24h", 9, 30, "09:30"},
		{"24h format afternoon", "24h", 15, 45, "15:45"},
		{"24h format midnight", "24h", 0, 0, "00:00"},
		{"24h format noon", "24h", 12, 0, "12:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime := time.Date(2024, 1, 1, tt.hour, tt.minute, 0, 0, time.UTC)
			result := FormatTime(testTime, tt.timeFormat)
			if result != tt.expected {
				t.Errorf("FormatTime(%v, %s) = %s, want %s", testTime, tt.timeFormat, result, tt.expected)
			}
		})
	}
}

func TestRemoveEmoji(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello 🎭 World", "Hello World"},
		{"📊 Data", "Data"},
		{"No emoji here", "No emoji here"},
		{"Multiple 🔋⏰ emojis", "Multiple emojis"},
		{"", ""},
	}

	for _, tt := range tests {
		result := removeEmoji(tt.input)
		if result != tt.expected {
			t.Errorf("removeEmoji(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGetAccessibleIcon(t *testing.T) {
	t.Parallel()
	tests := []struct {
		emoji    string
		expected string
	}{
		{"⏰", "[TIME]"},
		{"🔋", "[BAT]"},
		{"🔌", "[PWR]"},
		{"📱", "[APP]"},
		{"⏱️", "[FOCUS]"},
		{"🎵", "[MUSIC]"},
		{"🌐", "[NET]"},
		{"📊", "[DATA]"},
		{"💡", "[INFO]"},
		{"✓", "[OK]"},
		{"✗", "[ERR]"},
		{"🚀", "[*]"}, // Unknown emoji
	}

	for _, tt := range tests {
		result := getAccessibleIcon(tt.emoji)
		if result != tt.expected {
			t.Errorf("getAccessibleIcon(%q) = %q, want %q", tt.emoji, result, tt.expected)
		}
	}
}
