package ui

import (
	"strings"
	"testing"
)

func TestRenderWarning(t *testing.T) {
	result := RenderWarning("⚠️", "Test warning message")
	
	if !strings.Contains(result, "⚠️") {
		t.Error("Expected warning to contain icon")
	}
	
	if !strings.Contains(result, "Test warning message") {
		t.Error("Expected warning to contain message")
	}
	
	// Should have proper spacing
	if !strings.Contains(result, "  ⚠️  ") {
		t.Error("Expected warning to have proper icon spacing")
	}
}
