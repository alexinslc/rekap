package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexinslc/rekap/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newConfigCmd() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage rekap configuration",
		Long:  `Create, validate, and inspect your rekap configuration file.`,
	}

	configCmd.AddCommand(newConfigInitCmd(), newConfigValidateCmd(), newConfigShowCmd())
	return configCmd
}

func newConfigInitCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a starter config file",
		Long:  `Generate a commented config file at ~/.config/rekap/config.yaml with all available options.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := config.GetConfigPath()
			if err != nil {
				return fmt.Errorf("failed to determine config path: %w", err)
			}

			if _, err := os.Stat(configPath); err == nil && !force {
				return fmt.Errorf("config file already exists at %s\nUse --force to overwrite", configPath)
			}

			dir := filepath.Dir(configPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}

			if err := os.WriteFile(configPath, []byte(configTemplate), 0644); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			fmt.Printf("Config file created at %s\n", configPath)
			fmt.Println("Edit the file to customize your settings. Uncomment sections you want to change.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing config file")
	return cmd
}

func newConfigValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate your config file",
		Long:  `Check your config file for YAML syntax errors and invalid values.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := config.GetConfigPath()
			if err != nil {
				return fmt.Errorf("failed to determine config path: %w", err)
			}

			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				fmt.Printf("No config file found at %s\n", configPath)
				fmt.Println("Default settings will be used. Run 'rekap config init' to create one.")
				return nil
			}

			data, err := os.ReadFile(configPath)
			if err != nil {
				return fmt.Errorf("failed to read config file: %w", err)
			}

			var cfg config.Config
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return fmt.Errorf("YAML syntax error: %w", err)
			}

			errors := config.ValidateStrict(&cfg)
			if len(errors) > 0 {
				fmt.Printf("Config file: %s\n\n", configPath)
				for _, e := range errors {
					fmt.Printf("  Warning: %s\n", e)
				}
				fmt.Println()
				return fmt.Errorf("config has %d issue(s)", len(errors))
			}

			fmt.Printf("Config file: %s\n", configPath)
			fmt.Println("Config is valid.")
			return nil
		},
	}
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show the effective configuration",
		Long:  `Print the resolved configuration (defaults merged with your overrides).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
				cfg = config.Default()
			}

			out, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}

			fmt.Print(string(out))
			return nil
		},
	}
}

const configTemplate = `# rekap configuration
# Documentation: https://github.com/alexinslc/rekap/blob/main/docs/CONFIG.md

# Colors (hex "#RRGGBB" or ANSI codes "0"-"255")
# colors:
#   primary: "13"       # Main titles
#   secondary: "14"     # Labels
#   accent: "11"        # Highlights
#   success: "10"       # Success messages
#   warning: "9"        # Warnings
#   muted: "240"        # Subdued text
#   text: "255"         # Main text

# Display options
# display:
#   show_media: true    # Show "Now Playing" section
#   show_battery: true  # Show battery information
#   time_format: "12h"  # "12h" or "24h"

# App tracking
# tracking:
#   exclude_apps:
#     - "Activity Monitor"
#     - "System Preferences"

# Accessibility
# accessibility:
#   enabled: false
#   high_contrast: false
#   no_emoji: false

# Domain categorization (overrides defaults)
# domains:
#   work:
#     - "mycompany.atlassian.net"
#     - "internal.company.com"
#   distraction:
#     - "news.ycombinator.com"
#   neutral:
#     - "gmail.com"

# Fragmentation score thresholds
# fragmentation:
#   focused_max: 30     # 0-30 = Focused
#   moderate_max: 60    # 31-60 = Moderate
#   fragmented_min: 61  # 61-100 = Fragmented
`
