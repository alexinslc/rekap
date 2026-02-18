package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alexinslc/rekap/internal/config"
	"github.com/alexinslc/rekap/internal/permissions"
	"github.com/alexinslc/rekap/internal/theme"
	"github.com/alexinslc/rekap/internal/ui"
	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

func main() {
	var quietFlag bool
	var jsonFlag bool
	var themeFlag string
	var accessibleFlag bool

	rootCmd := &cobra.Command{
		Use:   "rekap",
		Short: "Daily Mac Activity Summary",
		Long:  `A single-binary macOS CLI that summarizes today's computer activity in a friendly, animated terminal UI.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
				cfg = config.Default()
			}

			if themeFlag != "" {
				t, err := theme.Load(themeFlag)
				if err != nil {
					return fmt.Errorf("failed to load theme: %w", err)
				}
				cfg.ApplyTheme(t)
			}

			if accessibleFlag {
				cfg.Accessibility.Enabled = true
				cfg.Accessibility.HighContrast = true
			}

			runSummary(quietFlag, jsonFlag, cfg)
			return nil
		},
	}

	rootCmd.Flags().BoolVarP(&quietFlag, "quiet", "q", false, "Output machine-parsable key=value format")
	rootCmd.Flags().BoolVar(&jsonFlag, "json", false, "Output structured JSON to stdout")
	rootCmd.Flags().StringVar(&themeFlag, "theme", "", "Color theme (built-in: default, minimal, hacker, pastel, nord, dracula, solarized) or path to theme file")
	rootCmd.MarkFlagsMutuallyExclusive("quiet", "json")
	rootCmd.PersistentFlags().BoolVar(&accessibleFlag, "accessible", false, "Enable accessibility mode (color-blind friendly, high contrast)")

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Permission setup wizard",
		Long:  `Run the guided permission setup wizard to enable Full Disk Access and other permissions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return permissions.RequestFlow()
		},
	}

	doctorCmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check capabilities and permissions",
		Long:  `Check the current status of permissions and capabilities.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			runDoctor()
			return nil
		},
	}

	var demoThemeFlag string
	demoCmd := &cobra.Command{
		Use:   "demo",
		Short: "See sample output with fake data",
		Long:  `Display a demo with randomized sample data to preview the output format.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
				cfg = config.Default()
			}

			if demoThemeFlag != "" {
				t, err := theme.Load(demoThemeFlag)
				if err != nil {
					return fmt.Errorf("failed to load theme: %w", err)
				}
				cfg.ApplyTheme(t)
			}

			if accessibleFlag {
				cfg.Accessibility.Enabled = true
				cfg.Accessibility.HighContrast = true
			}

			runDemo(cfg)
			return nil
		},
	}
	demoCmd.Flags().StringVar(&demoThemeFlag, "theme", "", "Color theme (built-in: default, minimal, hacker, pastel, nord, dracula, solarized) or path to theme file")

	rootCmd.AddCommand(initCmd, doctorCmd, demoCmd, newConfigCmd())

	if err := fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithVersion(version),
		fang.WithNotifySignal(os.Interrupt),
	); err != nil {
		os.Exit(1)
	}
}

func runDoctor() {
	fmt.Println(ui.RenderTitle("🩺 rekap capabilities check", false))
	fmt.Println()

	caps := permissions.Check()
	fmt.Println(permissions.FormatCapabilities(caps))
	fmt.Println()

	if !caps.FullDiskAccess {
		fmt.Println(ui.RenderHint("Run 'rekap init' to enable Full Disk Access for app tracking"))
	} else {
		fmt.Println(ui.RenderSuccess("All major permissions granted!"))
	}
}
