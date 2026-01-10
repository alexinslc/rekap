package theme

import (
	"testing"
)

func TestExists(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		themeName  string
		wantExists bool
	}{
		{"default theme exists", "default", true},
		{"minimal theme exists", "minimal", true},
		{"hacker theme exists", "hacker", true},
		{"pastel theme exists", "pastel", true},
		{"nord theme exists", "nord", true},
		{"dracula theme exists", "dracula", true},
		{"solarized theme exists", "solarized", true},
		{"nonexistent theme", "nonexistent", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists := Exists(tt.themeName)
			if exists != tt.wantExists {
				t.Errorf("Exists(%q) = %v, want %v", tt.themeName, exists, tt.wantExists)
			}
		})
	}
}

func TestThemeDescriptions(t *testing.T) {
	t.Parallel()
	// Verify all built-in themes have descriptions
	themes := ListBuiltIn()
	
	for _, name := range themes {
		t.Run(name, func(t *testing.T) {
			theme, ok := GetBuiltIn(name)
			if !ok {
				t.Errorf("Failed to get built-in theme %q", name)
				return
			}
			
			if theme.Description == "" {
				t.Errorf("Theme %q is missing description", name)
			}
			
			if theme.Name == "" {
				t.Errorf("Theme %q has empty Name field", name)
			}
		})
	}
}
