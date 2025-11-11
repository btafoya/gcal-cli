package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/btafoya/gcal-cli/pkg/types"
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Calendar CalendarConfig `mapstructure:"calendar"`
	Output   OutputConfig   `mapstructure:"output"`
	Auth     AuthConfig     `mapstructure:"auth"`
	API      APIConfig      `mapstructure:"api"`
	Events   EventsConfig   `mapstructure:"events"`
}

// CalendarConfig holds calendar-related configuration
type CalendarConfig struct {
	DefaultCalendarID string `mapstructure:"default_calendar_id"`
	DefaultTimezone   string `mapstructure:"default_timezone"`
}

// OutputConfig holds output-related configuration
type OutputConfig struct {
	DefaultFormat string `mapstructure:"default_format"`
	ColorEnabled  bool   `mapstructure:"color_enabled"`
	PrettyPrint   bool   `mapstructure:"pretty_print"`
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	CredentialsPath string `mapstructure:"credentials_path"`
	TokensPath      string `mapstructure:"tokens_path"`
	AutoRefresh     bool   `mapstructure:"auto_refresh"`
}

// APIConfig holds API-related configuration
type APIConfig struct {
	RetryAttempts   int     `mapstructure:"retry_attempts"`
	RetryDelayMs    int     `mapstructure:"retry_delay_ms"`
	RetryMaxDelayMs int     `mapstructure:"retry_max_delay_ms"`
	TimeoutSeconds  int     `mapstructure:"timeout_seconds"`
	RateLimitBuffer float64 `mapstructure:"rate_limit_buffer"`
}

// EventsConfig holds event-related configuration
type EventsConfig struct {
	DefaultDurationMinutes int  `mapstructure:"default_duration_minutes"`
	DefaultReminderMinutes int  `mapstructure:"default_reminder_minutes"`
	SendNotifications      bool `mapstructure:"send_notifications"`
}

// GetConfigDir returns the configuration directory path
func GetConfigDir() (string, error) {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "gcal-cli"), nil
	}

	// Fall back to ~/.config/gcal-cli
	home, err := os.UserHomeDir()
	if err != nil {
		return "", types.ErrConfigError("could not determine home directory").
			WithWrappedError(err)
	}

	return filepath.Join(home, ".config", "gcal-cli"), nil
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", types.ErrConfigError("could not create config directory").
			WithDetails(configDir).
			WithWrappedError(err)
	}

	return configDir, nil
}

// Initialize sets up Viper with default values and config file paths
func Initialize(cfgFile string) error {
	// Set defaults
	setDefaults()

	// Set config file
	if cfgFile != "" {
		// Use config file from flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find config directory
		configDir, err := GetConfigDir()
		if err != nil {
			return err
		}

		// Search for config in config directory
		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// Enable environment variable support
	viper.SetEnvPrefix("GCAL")
	viper.AutomaticEnv()

	// Read config file (it's okay if it doesn't exist)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return types.ErrConfigError("error reading config file").
				WithWrappedError(err)
		}
	}

	return nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Calendar defaults
	viper.SetDefault("calendar.default_calendar_id", "primary")
	viper.SetDefault("calendar.default_timezone", "")

	// Output defaults
	viper.SetDefault("output.default_format", "json")
	viper.SetDefault("output.color_enabled", false)
	viper.SetDefault("output.pretty_print", true)

	// Auth defaults
	configDir, _ := GetConfigDir()
	viper.SetDefault("auth.credentials_path", filepath.Join(configDir, "credentials.json"))
	viper.SetDefault("auth.tokens_path", filepath.Join(configDir, "tokens.json"))
	viper.SetDefault("auth.auto_refresh", true)

	// API defaults
	viper.SetDefault("api.retry_attempts", 3)
	viper.SetDefault("api.retry_delay_ms", 1000)
	viper.SetDefault("api.retry_max_delay_ms", 10000)
	viper.SetDefault("api.timeout_seconds", 30)
	viper.SetDefault("api.rate_limit_buffer", 0.9)

	// Event defaults
	viper.SetDefault("events.default_duration_minutes", 60)
	viper.SetDefault("events.default_reminder_minutes", 10)
	viper.SetDefault("events.send_notifications", true)
}

// Load loads the configuration into a Config struct
func Load() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, types.ErrConfigError("error unmarshaling config").
			WithWrappedError(err)
	}
	return &cfg, nil
}

// Save saves the current configuration to file
func Save() error {
	configDir, err := EnsureConfigDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.yaml")

	if err := viper.WriteConfigAs(configPath); err != nil {
		return types.ErrConfigError("error writing config file").
			WithDetails(configPath).
			WithWrappedError(err)
	}

	return nil
}

// Get returns a configuration value by key
func Get(key string) interface{} {
	return viper.Get(key)
}

// GetString returns a string configuration value
func GetString(key string) string {
	return viper.GetString(key)
}

// GetBool returns a boolean configuration value
func GetBool(key string) bool {
	return viper.GetBool(key)
}

// GetInt returns an integer configuration value
func GetInt(key string) int {
	return viper.GetInt(key)
}

// Set sets a configuration value
func Set(key string, value interface{}) {
	viper.Set(key, value)
}

// DisplayConfig returns a formatted string of current configuration
func DisplayConfig() (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`Current Configuration:

Calendar:
  Default Calendar ID: %s
  Default Timezone:    %s

Output:
  Default Format:      %s
  Color Enabled:       %t
  Pretty Print:        %t

Authentication:
  Credentials Path:    %s
  Tokens Path:         %s
  Auto Refresh:        %t

API:
  Retry Attempts:      %d
  Retry Delay (ms):    %d
  Timeout (seconds):   %d

Events:
  Default Duration:    %d minutes
  Default Reminder:    %d minutes
  Send Notifications:  %t
`,
		cfg.Calendar.DefaultCalendarID,
		cfg.Calendar.DefaultTimezone,
		cfg.Output.DefaultFormat,
		cfg.Output.ColorEnabled,
		cfg.Output.PrettyPrint,
		cfg.Auth.CredentialsPath,
		cfg.Auth.TokensPath,
		cfg.Auth.AutoRefresh,
		cfg.API.RetryAttempts,
		cfg.API.RetryDelayMs,
		cfg.API.TimeoutSeconds,
		cfg.Events.DefaultDurationMinutes,
		cfg.Events.DefaultReminderMinutes,
		cfg.Events.SendNotifications,
	), nil
}
