package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/alexinslc/rekap/internal/theme"
	"gopkg.in/yaml.v3"
)

// Config holds all user preferences
type Config struct {
	Colors        ColorConfig                   `yaml:"colors"`
	Display       DisplayConfig                 `yaml:"display"`
	Tracking      TrackingConfig                `yaml:"tracking"`
	Accessibility AccessibilityConfig           `yaml:"accessibility"`
	Domains       DomainsConfig                 `yaml:"domains"`
	Fragmentation FragmentationThresholdsConfig `yaml:"fragmentation"`
}

// ColorConfig holds color customization settings
type ColorConfig struct {
	Primary   string `yaml:"primary"`
	Secondary string `yaml:"secondary"`
	Accent    string `yaml:"accent"`
	Success   string `yaml:"success"`
	Warning   string `yaml:"warning"`
	Muted     string `yaml:"muted"`
	Text      string `yaml:"text"`
}

// DisplayConfig holds display preferences
type DisplayConfig struct {
	ShowMedia   *bool  `yaml:"show_media"`   // pointer to distinguish unset from false
	ShowBattery *bool  `yaml:"show_battery"` // pointer to distinguish unset from false
	TimeFormat  string `yaml:"time_format"`  // "12h" or "24h"
}

// TrackingConfig holds tracking preferences
type TrackingConfig struct {
	ExcludeApps []string `yaml:"exclude_apps"`
}

// AccessibilityConfig holds accessibility preferences
type AccessibilityConfig struct {
	Enabled      bool `yaml:"enabled"`
	HighContrast bool `yaml:"high_contrast"`
	NoEmoji      bool `yaml:"no_emoji"`
}

// DomainsConfig holds domain categorization configuration
type DomainsConfig struct {
	Work        []string `yaml:"work"`
	Distraction []string `yaml:"distraction"`
	Neutral     []string `yaml:"neutral"`
}

// FragmentationThresholdsConfig holds configurable thresholds for fragmentation scoring
type FragmentationThresholdsConfig struct {
	FocusedMax    int `yaml:"focused_max"`    // 0-30 = Focused
	ModerateMax   int `yaml:"moderate_max"`   // 31-60 = Moderate
	FragmentedMin int `yaml:"fragmented_min"` // 61-100 = Fragmented
}

// Default returns a config with sensible defaults
func Default() *Config {
	showMedia := true
	showBattery := true

	return &Config{
		Colors: ColorConfig{
			Primary:   "13",  // Bright magenta/pink
			Secondary: "14",  // Cyan
			Accent:    "11",  // Bright yellow
			Success:   "10",  // Bright green
			Warning:   "9",   // Bright red
			Muted:     "240", // Darker gray
			Text:      "255", // White
		},
		Display: DisplayConfig{
			ShowMedia:   &showMedia,
			ShowBattery: &showBattery,
			TimeFormat:  "12h",
		},
		Tracking: TrackingConfig{
			ExcludeApps: []string{},
		},
		Accessibility: AccessibilityConfig{
			Enabled:      false,
			HighContrast: false,
			NoEmoji:      false,
		},
		Domains: DomainsConfig{
			Work: []string{
				"github.com",
				"gitlab.com",
				"bitbucket.org",
				"stackoverflow.com",
				"stackexchange.com",
				"docs.*",
				"developer.*",
				"api.*",
				"atlassian.net",
				"linear.app",
				"asana.com",
				"notion.so",
				"aws.amazon.com",
				"console.cloud.google.com",
				"portal.azure.com",
			},
			Distraction: []string{
				"twitter.com",
				"x.com",
				"reddit.com",
				"facebook.com",
				"instagram.com",
				"youtube.com",
				"tiktok.com",
				"twitch.tv",
			},
			Neutral: []string{},
		},
		Fragmentation: FragmentationThresholdsConfig{
			FocusedMax:    30,
			ModerateMax:   60,
			FragmentedMin: 61,
		},
	}
}

// Load reads config from ~/.config/rekap/config.yaml
// If file doesn't exist, returns default config
func Load() (*Config, error) {
	cfg := Default()

	// Get config file path
	configPath, err := GetConfigPath()
	if err != nil {
		return cfg, nil // Use defaults if we can't determine path
	}

	// If config file doesn't exist, use defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return cfg, nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return cfg, err
	}

	// Parse YAML, merging with defaults
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return cfg, err
	}

	// Validate and apply defaults for unset values
	cfg.Validate()

	return cfg, nil
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".config", "rekap", "config.yaml"), nil
}

// Validate ensures config values are valid, applying defaults where needed
func (c *Config) Validate() {
	// Ensure time format is valid
	if c.Display.TimeFormat != "12h" && c.Display.TimeFormat != "24h" {
		c.Display.TimeFormat = "12h"
	}

	// Ensure display booleans have defaults if not set
	if c.Display.ShowMedia == nil {
		showMedia := true
		c.Display.ShowMedia = &showMedia
	}
	if c.Display.ShowBattery == nil {
		showBattery := true
		c.Display.ShowBattery = &showBattery
	}

	// Color validation - ensure they're not empty
	defaults := Default()
	if c.Colors.Primary == "" {
		c.Colors.Primary = defaults.Colors.Primary
	}
	if c.Colors.Secondary == "" {
		c.Colors.Secondary = defaults.Colors.Secondary
	}
	if c.Colors.Accent == "" {
		c.Colors.Accent = defaults.Colors.Accent
	}
	if c.Colors.Success == "" {
		c.Colors.Success = defaults.Colors.Success
	}
	if c.Colors.Warning == "" {
		c.Colors.Warning = defaults.Colors.Warning
	}
	if c.Colors.Muted == "" {
		c.Colors.Muted = defaults.Colors.Muted
	}
	if c.Colors.Text == "" {
		c.Colors.Text = defaults.Colors.Text
	}

	// Validate fragmentation thresholds
	if c.Fragmentation.FocusedMax <= 0 {
		c.Fragmentation.FocusedMax = defaults.Fragmentation.FocusedMax
	}
	if c.Fragmentation.ModerateMax <= 0 {
		c.Fragmentation.ModerateMax = defaults.Fragmentation.ModerateMax
	}
	if c.Fragmentation.FragmentedMin <= 0 {
		c.Fragmentation.FragmentedMin = defaults.Fragmentation.FragmentedMin
	}
	// Ensure logical ordering: FocusedMax <= ModerateMax < FragmentedMin
	if !(c.Fragmentation.FocusedMax <= c.Fragmentation.ModerateMax &&
		c.Fragmentation.ModerateMax < c.Fragmentation.FragmentedMin) {
		c.Fragmentation.FocusedMax = defaults.Fragmentation.FocusedMax
		c.Fragmentation.ModerateMax = defaults.Fragmentation.ModerateMax
		c.Fragmentation.FragmentedMin = defaults.Fragmentation.FragmentedMin
	}
}

// ShouldShowMedia returns whether to show media section
func (c *Config) ShouldShowMedia() bool {
	if c.Display.ShowMedia == nil {
		return true
	}
	return *c.Display.ShowMedia
}

// ShouldShowBattery returns whether to show battery section
func (c *Config) ShouldShowBattery() bool {
	if c.Display.ShowBattery == nil {
		return true
	}
	return *c.Display.ShowBattery
}

// ApplyTheme applies a theme's colors to the config, overriding existing colors
func (c *Config) ApplyTheme(t theme.Theme) {
	c.Colors.Primary = t.Colors.Primary
	c.Colors.Secondary = t.Colors.Secondary
	c.Colors.Accent = t.Colors.Accent
	c.Colors.Success = t.Colors.Success
	c.Colors.Warning = t.Colors.Warning
	c.Colors.Muted = t.Colors.Muted
	c.Colors.Text = t.Colors.Text
}

// CategorizeDomain returns "work", "distraction", "neutral", or "" (uncategorized)
func (c *Config) CategorizeDomain(domain string) string {
	if domain == "" {
		return ""
	}

	// Check work domains
	for _, pattern := range c.Domains.Work {
		if matchDomainPattern(domain, pattern) {
			return "work"
		}
	}

	// Check distraction domains
	for _, pattern := range c.Domains.Distraction {
		if matchDomainPattern(domain, pattern) {
			return "distraction"
		}
	}

	// Check neutral domains
	for _, pattern := range c.Domains.Neutral {
		if matchDomainPattern(domain, pattern) {
			return "neutral"
		}
	}

	// Default to neutral if not categorized
	return "neutral"
}

// matchDomainPattern matches a domain against a pattern
// Supports wildcards like "docs.*" or "*.google.com"
func matchDomainPattern(domain, pattern string) bool {
	// Exact match
	if domain == pattern {
		return true
	}

	// Wildcard pattern matching
	if strings.Contains(pattern, "*") {
		// Convert pattern to regex-like matching
		// docs.* matches docs.python.org, docs.microsoft.com, etc.
		// *.google.com matches mail.google.com, drive.google.com, etc.

		if strings.HasPrefix(pattern, "*.") {
			// *.example.com pattern
			suffix := pattern[1:] // Remove the *
			return strings.HasSuffix(domain, suffix)
		} else if strings.HasSuffix(pattern, ".*") {
			// docs.* pattern
			prefix := pattern[:len(pattern)-1] // Remove the *
			return strings.HasPrefix(domain, prefix)
		}
	}

	// Check if domain ends with .pattern (e.g., "atlassian.net" matches "mycompany.atlassian.net")
	if strings.HasSuffix(domain, "."+pattern) {
		return true
	}

	return false
}
