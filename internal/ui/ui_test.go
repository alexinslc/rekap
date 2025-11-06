package ui

import (
	"testing"
)

func TestFormatDuration(t *testing.T) {
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
	result := RenderHint("Test hint message")
	if result == "" {
		t.Error("RenderHint should not return empty string")
	}
}

func TestRenderError(t *testing.T) {
	result := RenderError("Test error message")
	if result == "" {
		t.Error("RenderError should not return empty string")
	}
}

func TestRenderSummaryLine(t *testing.T) {
	parts := []string{"3h 35m screen-on", "2 plug-ins", "Top apps: VS Code"}
	result := RenderSummaryLine(parts)
	
	if result == "" {
		t.Error("RenderSummaryLine should not return empty string")
	}
}

func TestRenderSummaryLineEmpty(t *testing.T) {
	result := RenderSummaryLine([]string{})
	
	if result != "" {
		t.Error("RenderSummaryLine should return empty string for empty input")
	}
}
