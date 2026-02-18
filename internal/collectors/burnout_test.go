package collectors

import (
	"context"
	"testing"
)

func TestDefaultBurnoutConfig(t *testing.T) {
	t.Parallel()
	cfg := DefaultBurnoutConfig()

	if cfg.LongDayHours != 10 {
		t.Errorf("Expected LongDayHours to be 10, got %d", cfg.LongDayHours)
	}
	if cfg.AppSwitchesPerHour != 50 {
		t.Errorf("Expected AppSwitchesPerHour to be 50, got %d", cfg.AppSwitchesPerHour)
	}
	if cfg.MaxTabs != 100 {
		t.Errorf("Expected MaxTabs to be 100, got %d", cfg.MaxTabs)
	}
	if cfg.LateNightHour != 0 {
		t.Errorf("Expected LateNightHour to be 0, got %d", cfg.LateNightHour)
	}
	if cfg.NoBreakHours != 4 {
		t.Errorf("Expected NoBreakHours to be 4, got %d", cfg.NoBreakHours)
	}
}

func TestCollectBurnout_LongDay(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	config := DefaultBurnoutConfig()

	// Test case: Long work day (11 hours)
	screen := ScreenResult{
		ScreenOnMinutes: 660, // 11 hours
		Available:       true,
	}

	browsers := BrowsersResult{
		TotalTabs: 50,
		Available: true,
	}

	result := CollectBurnout(ctx, screen, browsers, config)

	if !result.Available {
		t.Error("Expected burnout result to be available")
	}

	foundLongDay := false
	for _, warning := range result.Warnings {
		if warning.Type == "long_day" {
			foundLongDay = true
			if warning.Severity != "medium" {
				t.Errorf("Expected severity to be 'medium', got '%s'", warning.Severity)
			}
			if warning.MetricValue != 11 {
				t.Errorf("Expected metric value to be 11, got %d", warning.MetricValue)
			}
		}
	}

	if !foundLongDay {
		t.Error("Expected to find long_day warning")
	}
}

func TestCollectBurnout_TabOverload(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	config := DefaultBurnoutConfig()

	// Test case: Tab overload (125 tabs)
	screen := ScreenResult{
		ScreenOnMinutes: 240, // 4 hours (below threshold)
		Available:       true,
	}

	browsers := BrowsersResult{
		TotalTabs: 125,
		Available: true,
	}

	result := CollectBurnout(ctx, screen, browsers, config)

	if !result.Available {
		t.Error("Expected burnout result to be available")
	}

	foundTabOverload := false
	for _, warning := range result.Warnings {
		if warning.Type == "tab_overload" {
			foundTabOverload = true
			if warning.Severity != "low" {
				t.Errorf("Expected severity to be 'low', got '%s'", warning.Severity)
			}
			if warning.MetricValue != 125 {
				t.Errorf("Expected metric value to be 125, got %d", warning.MetricValue)
			}
		}
	}

	if !foundTabOverload {
		t.Error("Expected to find tab_overload warning")
	}
}

func TestCollectBurnout_NoWarnings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	config := DefaultBurnoutConfig()

	// Test case: Normal work day (no warnings)
	screen := ScreenResult{
		ScreenOnMinutes: 360, // 6 hours
		Available:       true,
	}

	browsers := BrowsersResult{
		TotalTabs: 35,
		Available: true,
	}

	result := CollectBurnout(ctx, screen, browsers, config)

	if !result.Available {
		t.Error("Expected burnout result to be available")
	}

	if len(result.Warnings) > 0 {
		t.Errorf("Expected no warnings for normal work day, got %d warnings", len(result.Warnings))
	}
}

func TestCollectBurnout_MultipleWarnings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	config := DefaultBurnoutConfig()

	// Test case: Multiple warnings (long day + tab overload)
	screen := ScreenResult{
		ScreenOnMinutes: 660, // 11 hours
		Available:       true,
	}

	browsers := BrowsersResult{
		TotalTabs: 150,
		Available: true,
	}

	result := CollectBurnout(ctx, screen, browsers, config)

	if !result.Available {
		t.Error("Expected burnout result to be available")
	}

	if len(result.Warnings) < 2 {
		t.Errorf("Expected at least 2 warnings, got %d", len(result.Warnings))
	}

	warningTypes := make(map[string]bool)
	for _, warning := range result.Warnings {
		warningTypes[warning.Type] = true
	}

	if !warningTypes["long_day"] {
		t.Error("Expected to find long_day warning")
	}

	if !warningTypes["tab_overload"] {
		t.Error("Expected to find tab_overload warning")
	}
}

func TestCollectBurnout_UnavailableData(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	config := DefaultBurnoutConfig()

	// Test case: Data not available
	screen := ScreenResult{
		Available: false,
	}

	browsers := BrowsersResult{
		Available: false,
	}

	result := CollectBurnout(ctx, screen, browsers, config)

	if !result.Available {
		t.Error("Expected burnout result to be available even when data is not")
	}

	// Should have no warnings when data is unavailable
	// (except for warnings that don't depend on screen/browsers)
	hasDataDependentWarning := false
	for _, warning := range result.Warnings {
		if warning.Type == "long_day" || warning.Type == "tab_overload" {
			hasDataDependentWarning = true
		}
	}

	if hasDataDependentWarning {
		t.Error("Should not have data-dependent warnings when data is unavailable")
	}
}
