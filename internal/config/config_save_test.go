package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexinslc/rekap/internal/theme"
)

func TestConfigSave(t *testing.T) {
	t.Parallel()
	
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	
	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})
	os.Setenv("HOME", tmpDir)
	
	// Create a test config
	cfg := Default()
	cfg.Display.TimeFormat = "24h"
	
	// Save the config
	err := cfg.Save()
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}
	
	// Verify file exists
	configPath := filepath.Join(tmpDir, ".config", "rekap", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created at %s", configPath)
	}
	
	// Load the config back
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Verify loaded config matches saved config
	if loaded.Display.TimeFormat != "24h" {
		t.Errorf("TimeFormat = %q, want %q", loaded.Display.TimeFormat, "24h")
	}
}

func TestConfigApplyTheme(t *testing.T) {
	t.Parallel()
	
	cfg := Default()
	originalPrimary := cfg.Colors.Primary
	
	// Load a theme
	nordTheme, err := theme.Load("nord")
	if err != nil {
		t.Fatalf("Failed to load nord theme: %v", err)
	}
	
	// Apply theme
	cfg.ApplyTheme(nordTheme)
	
	// Verify colors changed
	if cfg.Colors.Primary == originalPrimary {
		t.Error("ApplyTheme did not change Primary color")
	}
	
	if cfg.Colors.Primary != nordTheme.Colors.Primary {
		t.Errorf("Primary color = %q, want %q", cfg.Colors.Primary, nordTheme.Colors.Primary)
	}
}
