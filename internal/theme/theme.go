package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Theme represents a complete color theme
type Theme struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description,omitempty"`
	Author      string      `yaml:"author,omitempty"`
	Colors      ThemeColors `yaml:"colors"`
}

// ThemeColors defines all color values for a theme
type ThemeColors struct {
	Primary   string `yaml:"primary"`
	Secondary string `yaml:"secondary"`
	Accent    string `yaml:"accent"`
	Success   string `yaml:"success"`
	Error     string `yaml:"error,omitempty"` // Support both error and warning
	Warning   string `yaml:"warning,omitempty"`
	Muted     string `yaml:"muted"`
	Text      string `yaml:"text"`
}

// builtInThemes contains all the built-in themes
var builtInThemes = map[string]Theme{
	"default": {
		Name:        "Default",
		Description: "Default colorful theme",
		Author:      "rekap",
		Colors: ThemeColors{
			Primary:   "13",  // Bright magenta/pink
			Secondary: "14",  // Cyan
			Accent:    "11",  // Bright yellow
			Success:   "10",  // Bright green
			Warning:   "9",   // Bright red
			Muted:     "240", // Darker gray
			Text:      "255", // White
		},
	},
	"minimal": {
		Name:        "Minimal",
		Description: "Minimalist monochrome",
		Author:      "rekap",
		Colors: ThemeColors{
			Primary:   "255", // White
			Secondary: "250", // Light gray
			Accent:    "255", // White
			Success:   "252", // Gray
			Warning:   "244", // Medium gray
			Muted:     "240", // Dark gray
			Text:      "255", // White
		},
	},
	"hacker": {
		Name:        "Hacker",
		Description: "Matrix green aesthetic",
		Author:      "rekap",
		Colors: ThemeColors{
			Primary:   "10", // Bright green
			Secondary: "2",  // Green
			Accent:    "10", // Bright green
			Success:   "10", // Bright green
			Warning:   "2",  // Green
			Muted:     "22", // Dark green
			Text:      "2",  // Green
		},
	},
	"pastel": {
		Name:        "Pastel",
		Description: "Soft pastel colors",
		Author:      "rekap",
		Colors: ThemeColors{
			Primary:   "#ff99cc", // Soft pink
			Secondary: "#99ccff", // Soft blue
			Accent:    "#ffcc99", // Soft orange
			Success:   "#99ff99", // Soft green
			Warning:   "#ff9999", // Soft red
			Muted:     "#cccccc", // Light gray
			Text:      "#ffffff", // White
		},
	},
	"nord": {
		Name:        "Nord",
		Description: "Arctic inspired palette",
		Author:      "rekap",
		Colors: ThemeColors{
			Primary:   "#88c0d0", // Nord frost
			Secondary: "#81a1c1", // Nord frost
			Accent:    "#ebcb8b", // Nord aurora yellow
			Success:   "#a3be8c", // Nord aurora green
			Warning:   "#bf616a", // Nord aurora red
			Muted:     "#4c566a", // Nord polar night
			Text:      "#eceff4", // Nord snow storm
		},
	},
	"dracula": {
		Name:        "Dracula",
		Description: "Dark purple theme",
		Author:      "rekap",
		Colors: ThemeColors{
			Primary:   "#ff79c6", // Pink
			Secondary: "#8be9fd", // Cyan
			Accent:    "#f1fa8c", // Yellow
			Success:   "#50fa7b", // Green
			Warning:   "#ff5555", // Red
			Muted:     "#6272a4", // Comment
			Text:      "#f8f8f2", // Foreground
		},
	},
	"solarized": {
		Name:        "Solarized Dark",
		Description: "Solarized color scheme",
		Author:      "rekap",
		Colors: ThemeColors{
			Primary:   "#268bd2", // Blue
			Secondary: "#2aa198", // Cyan
			Accent:    "#b58900", // Yellow
			Success:   "#859900", // Green
			Warning:   "#dc322f", // Red
			Muted:     "#586e75", // Base01
			Text:      "#93a1a1", // Base1
		},
	},
}

// GetBuiltIn returns a built-in theme by name
func GetBuiltIn(name string) (Theme, bool) {
	theme, ok := builtInThemes[name]
	return theme, ok
}

// ListBuiltIn returns all built-in theme names
func ListBuiltIn() []string {
	names := make([]string, 0, len(builtInThemes))
	for name := range builtInThemes {
		names = append(names, name)
	}
	return names
}

// LoadFromFile loads a theme from a YAML file
func LoadFromFile(path string) (Theme, error) {
	var theme Theme

	data, err := os.ReadFile(path)
	if err != nil {
		return theme, fmt.Errorf("failed to read theme file: %w", err)
	}

	if err := yaml.Unmarshal(data, &theme); err != nil {
		return theme, fmt.Errorf("failed to parse theme file: %w", err)
	}

	// Validate theme has required colors
	if err := theme.Validate(); err != nil {
		return theme, err
	}

	return theme, nil
}

// Load loads a theme by name or path
// First checks for built-in themes, then tries to load from filesystem
func Load(nameOrPath string) (Theme, error) {
	// Check if it's a built-in theme first
	if theme, ok := GetBuiltIn(nameOrPath); ok {
		return theme, nil
	}

	// Try to load from file path
	path := nameOrPath

	// If it's an absolute path or starts with ./ or ../, use it as-is
	if filepath.IsAbs(path) || strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
		return LoadFromFile(path)
	}

	// Otherwise, check in themes directory
	homeDir, err := os.UserHomeDir()
	if err == nil {
		themesDir := filepath.Join(homeDir, ".config", "rekap", "themes")
		// Try with .yaml extension
		if filepath.Ext(path) == "" {
			path = filepath.Join(themesDir, path+".yaml")
		} else {
			path = filepath.Join(themesDir, path)
		}
	}

	return LoadFromFile(path)
}

// Validate checks that all required color fields are set
func (t *Theme) Validate() error {
	if t.Colors.Primary == "" {
		return fmt.Errorf("theme missing required color: primary")
	}
	if t.Colors.Secondary == "" {
		return fmt.Errorf("theme missing required color: secondary")
	}
	if t.Colors.Accent == "" {
		return fmt.Errorf("theme missing required color: accent")
	}
	if t.Colors.Success == "" {
		return fmt.Errorf("theme missing required color: success")
	}
	// Support both error and warning fields for compatibility
	if t.Colors.Warning == "" && t.Colors.Error == "" {
		return fmt.Errorf("theme missing required color: warning or error")
	}
	// If error is specified but warning isn't, use error for warning
	if t.Colors.Warning == "" && t.Colors.Error != "" {
		t.Colors.Warning = t.Colors.Error
	}
	if t.Colors.Muted == "" {
		return fmt.Errorf("theme missing required color: muted")
	}
	if t.Colors.Text == "" {
		return fmt.Errorf("theme missing required color: text")
	}

	return nil
}

// Exists checks if a theme exists (built-in or custom)
func Exists(name string) bool {
	// Check built-in themes
	if _, ok := GetBuiltIn(name); ok {
		return true
	}

	// Check custom themes in user directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	themesDir := filepath.Join(homeDir, ".config", "rekap", "themes")
	
	// Try with .yaml extension
	if filepath.Ext(name) == "" {
		themePath := filepath.Join(themesDir, name+".yaml")
		if _, err := os.Stat(themePath); err == nil {
			return true
		}
	} else {
		themePath := filepath.Join(themesDir, name)
		if _, err := os.Stat(themePath); err == nil {
			return true
		}
	}

	return false
}
