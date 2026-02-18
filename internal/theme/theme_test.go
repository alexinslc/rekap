package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetBuiltIn(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		themeName string
		wantOk    bool
	}{
		{"default theme exists", "default", true},
		{"minimal theme exists", "minimal", true},
		{"hacker theme exists", "hacker", true},
		{"pastel theme exists", "pastel", true},
		{"nord theme exists", "nord", true},
		{"dracula theme exists", "dracula", true},
		{"solarized theme exists", "solarized", true},
		{"nonexistent theme", "nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme, ok := GetBuiltIn(tt.themeName)
			if ok != tt.wantOk {
				t.Errorf("GetBuiltIn(%q) ok = %v, want %v", tt.themeName, ok, tt.wantOk)
			}
			if ok {
				if theme.Name == "" {
					t.Errorf("GetBuiltIn(%q) returned theme with empty name", tt.themeName)
				}
				if err := theme.Validate(); err != nil {
					t.Errorf("GetBuiltIn(%q) returned invalid theme: %v", tt.themeName, err)
				}
			}
		})
	}
}

func TestListBuiltIn(t *testing.T) {
	t.Parallel()
	themes := ListBuiltIn()
	if len(themes) < 7 {
		t.Errorf("ListBuiltIn() returned %d themes, want at least 7", len(themes))
	}

	// Check that default theme is in the list
	found := false
	for _, name := range themes {
		if name == "default" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ListBuiltIn() did not include 'default' theme")
	}
}

func TestThemeValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		theme   Theme
		wantErr bool
	}{
		{
			name: "valid theme",
			theme: Theme{
				Name: "Test",
				Colors: ThemeColors{
					Primary:   "#ff00ff",
					Secondary: "#00ffff",
					Accent:    "#ffff00",
					Success:   "#00ff00",
					Warning:   "#ff0000",
					Muted:     "#808080",
					Text:      "#ffffff",
				},
			},
			wantErr: false,
		},
		{
			name: "missing primary",
			theme: Theme{
				Name: "Test",
				Colors: ThemeColors{
					Secondary: "#00ffff",
					Accent:    "#ffff00",
					Success:   "#00ff00",
					Warning:   "#ff0000",
					Muted:     "#808080",
					Text:      "#ffffff",
				},
			},
			wantErr: true,
		},
		{
			name: "missing warning",
			theme: Theme{
				Name: "Test",
				Colors: ThemeColors{
					Primary:   "#ff00ff",
					Secondary: "#00ffff",
					Accent:    "#ffff00",
					Success:   "#00ff00",
					Muted:     "#808080",
					Text:      "#ffffff",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.theme.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Theme.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	t.Parallel()
	// Create a temporary directory for test theme files
	tmpDir, err := os.MkdirTemp("", "rekap-theme-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a valid theme file
	validTheme := `name: "Test Theme"
author: "Test Author"
colors:
  primary: "#ff00ff"
  secondary: "#00ffff"
  accent: "#ffff00"
  success: "#00ff00"
  warning: "#ff0000"
  muted: "#808080"
  text: "#ffffff"
`
	validPath := filepath.Join(tmpDir, "valid.yaml")
	if err := os.WriteFile(validPath, []byte(validTheme), 0644); err != nil {
		t.Fatalf("Failed to write valid theme file: %v", err)
	}

	// Create an invalid theme file
	invalidTheme := `name: "Invalid Theme"
colors:
  primary: "#ff00ff"
  # missing required colors
`
	invalidPath := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(invalidPath, []byte(invalidTheme), 0644); err != nil {
		t.Fatalf("Failed to write invalid theme file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"valid theme file", validPath, false},
		{"invalid theme file", invalidPath, true},
		{"nonexistent file", filepath.Join(tmpDir, "nonexistent.yaml"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			theme, err := LoadFromFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if theme.Name == "" {
					t.Error("LoadFromFile() returned theme with empty name")
				}
			}
		})
	}
}

func TestLoad(t *testing.T) {
	// Test loading built-in themes
	t.Run("load built-in theme", func(t *testing.T) {
		theme, err := Load("default")
		if err != nil {
			t.Errorf("Load(\"default\") error = %v, want nil", err)
		}
		if theme.Name != "Default" {
			t.Errorf("Load(\"default\") name = %q, want \"Default\"", theme.Name)
		}
	})

	t.Run("load nonexistent theme", func(t *testing.T) {
		_, err := Load("nonexistent-theme-xyz")
		if err == nil {
			t.Error("Load(\"nonexistent-theme-xyz\") error = nil, want error")
		}
	})
}
