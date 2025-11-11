package main

import (
	"fmt"
	"os"

	"github.com/btafoya/gcal-cli/internal/commands"
	"github.com/btafoya/gcal-cli/pkg/config"
	"github.com/btafoya/gcal-cli/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile      string
	outputFormat string
	calendarID   string
	timezone     string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "gcal-cli",
	Short: "Google Calendar CLI for LLM agents",
	Long: `A Google Calendar CLI tool designed for LLM agent integration.
Provides programmatic access to Google Calendar operations through
a command-line interface optimized for machine-readable interactions.`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default: ~/.config/gcal-cli/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "format", "json",
		"output format (json|text|minimal)")
	rootCmd.PersistentFlags().StringVar(&calendarID, "calendar-id", "primary",
		"calendar ID to operate on")
	rootCmd.PersistentFlags().StringVar(&timezone, "timezone", "",
		"timezone for operations (default: system timezone)")

	// Bind flags to viper
	viper.BindPFlag("output.default_format", rootCmd.PersistentFlags().Lookup("format"))
	viper.BindPFlag("calendar.default_calendar_id", rootCmd.PersistentFlags().Lookup("calendar-id"))
	viper.BindPFlag("calendar.default_timezone", rootCmd.PersistentFlags().Lookup("timezone"))

	// Add subcommands
	formatter := getFormatter()
	rootCmd.AddCommand(commands.NewVersionCommand(formatter))
	rootCmd.AddCommand(commands.NewConfigCommand(formatter))
	rootCmd.AddCommand(commands.NewAuthCommand(formatter))
	rootCmd.AddCommand(commands.NewEventsCommand(formatter))
	rootCmd.AddCommand(commands.NewCalendarsCommand(formatter))
}

// initConfig reads in config file and ENV variables
func initConfig() {
	if err := config.Initialize(cfgFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
		os.Exit(1)
	}
}

// getFormatter returns the appropriate output formatter based on configuration
func getFormatter() output.Formatter {
	format := output.ParseFormat(viper.GetString("output.default_format"))
	return output.NewFormatter(format)
}
