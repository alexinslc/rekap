package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Colors.Primary != "13" {
		t.Errorf("Expected primary color 13, got %s", cfg.Colors.Primary)
	}

	if cfg.Display.TimeFormat != "12h" {
		t.Errorf("Expected 12h time format, got %s", cfg.Display.TimeFormat)
	}

	if !cfg.ShouldShowMedia() {
		t.Error("Expected media to be shown by default")
	}

	if !cfg.ShouldShowBattery() {
		t.Error("Expected battery to be shown by default")
	}

	if len(cfg.Tracking.ExcludeApps) != 0 {
		t.Errorf("Expected empty exclude list, got %d items", len(cfg.Tracking.ExcludeApps))
	}
}

func TestLoadNonExistent(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() should not error for non-existent config: %v", err)
	}

	// Should return defaults
	if cfg.Display.TimeFormat != "12h" {
		t.Errorf("Expected default time format 12h, got %s", cfg.Display.TimeFormat)
	}
}

func TestLoadValidConfig(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "rekap")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Write a test config
	configPath := filepath.Join(configDir, "config.yaml")
	configContent := `colors:
  primary: "#ff00ff"
  secondary: "#00ffff"

display:
  show_media: false
  show_battery: true
  time_format: "24h"

tracking:
  exclude_apps:
    - "Activity Monitor"
    - "System Preferences"
`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify loaded values
	if cfg.Colors.Primary != "#ff00ff" {
		t.Errorf("Expected primary color #ff00ff, got %s", cfg.Colors.Primary)
	}

	if cfg.Colors.Secondary != "#00ffff" {
		t.Errorf("Expected secondary color #00ffff, got %s", cfg.Colors.Secondary)
	}

	if cfg.Display.TimeFormat != "24h" {
		t.Errorf("Expected time format 24h, got %s", cfg.Display.TimeFormat)
	}

	if cfg.ShouldShowMedia() {
		t.Error("Expected media to be hidden")
	}

	if !cfg.ShouldShowBattery() {
		t.Error("Expected battery to be shown")
	}

	if len(cfg.Tracking.ExcludeApps) != 2 {
		t.Errorf("Expected 2 excluded apps, got %d", len(cfg.Tracking.ExcludeApps))
	}

	if !cfg.IsAppExcluded("Activity Monitor") {
		t.Error("Expected 'Activity Monitor' to be excluded")
	}

	if cfg.IsAppExcluded("VS Code") {
		t.Error("Expected 'VS Code' to not be excluded")
	}
}

func TestValidate(t *testing.T) {
	cfg := &Config{
		Display: DisplayConfig{
			TimeFormat: "invalid",
		},
	}

	cfg.Validate()

	// Should default to 12h for invalid format
	if cfg.Display.TimeFormat != "12h" {
		t.Errorf("Expected time format to default to 12h, got %s", cfg.Display.TimeFormat)
	}

	// Should have default colors
	if cfg.Colors.Primary == "" {
		t.Error("Expected primary color to have default value")
	}
}

func TestIsAppExcluded(t *testing.T) {
	cfg := &Config{
		Tracking: TrackingConfig{
			ExcludeApps: []string{"App1", "App2"},
		},
	}

	tests := []struct {
		appName  string
		expected bool
	}{
		{"App1", true},
		{"App2", true},
		{"App3", false},
		{"", false},
	}

	for _, tt := range tests {
		result := cfg.IsAppExcluded(tt.appName)
		if result != tt.expected {
			t.Errorf("IsAppExcluded(%q) = %v, want %v", tt.appName, result, tt.expected)
		}
	}
}

func TestCategorizeDomain(t *testing.T) {
	cfg := Default()

	tests := []struct {
		domain   string
		expected string
	}{
		// Work domains
		{"github.com", "work"},
		{"gitlab.com", "work"},
		{"stackoverflow.com", "work"},
		{"docs.python.org", "work"},
		{"docs.microsoft.com", "work"},
		{"developer.mozilla.org", "work"},
		{"api.github.com", "work"},
		{"mycompany.atlassian.net", "work"},
		{"linear.app", "work"},
		{"notion.so", "work"},
		
		// Distraction domains
		{"twitter.com", "distraction"},
		{"x.com", "distraction"},
		{"reddit.com", "distraction"},
		{"facebook.com", "distraction"},
		{"youtube.com", "distraction"},
		{"tiktok.com", "distraction"},
		
		// Neutral/uncategorized domains
		{"gmail.com", "neutral"},
		{"example.com", "neutral"},
		{"", ""},
	}

	for _, tt := range tests {
		result := cfg.CategorizeDomain(tt.domain)
		if result != tt.expected {
			t.Errorf("CategorizeDomain(%q) = %q, want %q", tt.domain, result, tt.expected)
		}
	}
}

func TestCategorizeDomainCustomConfig(t *testing.T) {
	cfg := &Config{
		Domains: DomainsConfig{
			Work:        []string{"mycompany.com", "internal.*"},
			Distraction: []string{"news.ycombinator.com"},
			Neutral:     []string{"gmail.com"},
		},
	}

	tests := []struct {
		domain   string
		expected string
	}{
		{"mycompany.com", "work"},
		{"internal.example.com", "work"},
		{"news.ycombinator.com", "distraction"},
		{"gmail.com", "neutral"},
		{"other.com", "neutral"},
	}

	for _, tt := range tests {
		result := cfg.CategorizeDomain(tt.domain)
		if result != tt.expected {
			t.Errorf("CategorizeDomain(%q) = %q, want %q", tt.domain, result, tt.expected)
		}
	}
}

func TestMatchDomainPattern(t *testing.T) {
	tests := []struct {
		domain   string
		pattern  string
		expected bool
	}{
		// Exact matches
		{"github.com", "github.com", true},
		{"gitlab.com", "github.com", false},
		
		// Wildcard prefix patterns (docs.*)
		{"docs.python.org", "docs.*", true},
		{"docs.microsoft.com", "docs.*", true},
		{"api.github.com", "docs.*", false},
		
		// Wildcard suffix patterns (*.google.com)
		{"mail.google.com", "*.google.com", true},
		{"drive.google.com", "*.google.com", true},
		{"google.com", "*.google.com", false},
		{"gmail.com", "*.google.com", false},
		
		// Suffix matching (for patterns like "atlassian.net")
		{"example.com", "example.com", true},
		{"mycompany.atlassian.net", "atlassian.net", true},
		{"subdomain.example.com", "example.com", true},
		{"notexample.com", "example.com", false},
	}

	for _, tt := range tests {
		result := matchDomainPattern(tt.domain, tt.pattern)
		if result != tt.expected {
			t.Errorf("matchDomainPattern(%q, %q) = %v, want %v", tt.domain, tt.pattern, result, tt.expected)
		}
	}
}
