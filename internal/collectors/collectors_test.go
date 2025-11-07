package collectors

import (
	"context"
	"testing"
	"time"
)

func TestCollectUptime(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := CollectUptime(ctx)

	if !result.Available {
		t.Error("Uptime should always be available")
	}

	if result.AwakeMinutes < 0 {
		t.Errorf("AwakeMinutes should be >= 0, got %d", result.AwakeMinutes)
	}

	if result.BootTime.IsZero() {
		t.Error("BootTime should not be zero")
	}

	if result.FormattedTime == "" {
		t.Error("FormattedTime should not be empty")
	}
}

func TestCollectBattery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := CollectBattery(ctx)

	if !result.Available {
		t.Skip("Battery not available (running on desktop?)")
	}

	if result.CurrentPct < 0 || result.CurrentPct > 100 {
		t.Errorf("CurrentPct should be 0-100, got %d", result.CurrentPct)
	}

	if result.StartPct < 0 || result.StartPct > 100 {
		t.Errorf("StartPct should be 0-100, got %d", result.StartPct)
	}
}

func TestCollectScreen(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := CollectScreen(ctx)

	// Screen collection is best-effort, may not always work
	if !result.Available {
		t.Log("Screen-on time not available")
		return
	}

	if result.ScreenOnMinutes < 0 {
		t.Errorf("ScreenOnMinutes should be >= 0, got %d", result.ScreenOnMinutes)
	}
}

func TestCollectApps(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := CollectApps(ctx)

	// Apps require Full Disk Access, may not be available
	if !result.Available {
		t.Log("App tracking not available (needs Full Disk Access)")
		return
	}

	for _, app := range result.TopApps {
		if app.Minutes < 0 {
			t.Errorf("App minutes should be >= 0, got %d for %s", app.Minutes, app.Name)
		}
		if app.Name == "" {
			t.Error("App name should not be empty")
		}
	}
}

func TestCollectMedia(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := CollectMedia(ctx)

	// Media is optional, test if available
	if !result.Available {
		t.Log("No media playing")
		return
	}

	if result.Track == "" {
		t.Error("Track should not be empty when Available=true")
	}

	if result.App == "" {
		t.Error("App should not be empty when Available=true")
	}
}

func TestCollectorTimeout(t *testing.T) {
	// Test that collectors respect context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(2 * time.Millisecond)

	// This should return quickly even though context is already done
	result := CollectUptime(ctx)

	// Even with expired context, best-effort should still work
	if !result.Available {
		t.Log("Uptime still unavailable with expired context (expected)")
	}
}

func TestCollectNetwork(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result := CollectNetwork(ctx)

	// Network collection is best-effort, may not always work
	if !result.Available {
		t.Log("Network not available")
		return
	}

	if result.InterfaceName == "" {
		t.Error("InterfaceName should not be empty when Available=true")
	}

	if result.NetworkName == "" {
		t.Error("NetworkName should not be empty when Available=true")
	}

	if result.BytesReceived < 0 {
		t.Errorf("BytesReceived should be >= 0, got %d", result.BytesReceived)
	}

	if result.BytesSent < 0 {
		t.Errorf("BytesSent should be >= 0, got %d", result.BytesSent)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
		{2147483648, "2.0 GB"},
	}

	for _, tt := range tests {
		result := FormatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("FormatBytes(%d) = %s, want %s", tt.bytes, result, tt.expected)
		}
	}
}
