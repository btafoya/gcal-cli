package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestGetConfigDir(t *testing.T) {
	// Save original environment
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	tests := []struct {
		name        string
		xdgConfig   string
		wantContain string
	}{
		{
			name:        "with XDG_CONFIG_HOME set",
			xdgConfig:   "/custom/config",
			wantContain: "/custom/config/gcal-cli",
		},
		{
			name:        "without XDG_CONFIG_HOME",
			xdgConfig:   "",
			wantContain: ".config/gcal-cli",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("XDG_CONFIG_HOME", tt.xdgConfig)

			got, err := GetConfigDir()
			if err != nil {
				t.Fatalf("GetConfigDir() error = %v", err)
			}

			if got == "" {
				t.Error("GetConfigDir() returned empty string")
			}

			if !contains(got, tt.wantContain) {
				t.Errorf("GetConfigDir() = %v, want to contain %v", got, tt.wantContain)
			}
		})
	}
}

func TestEnsureConfigDir(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	// Set temporary XDG_CONFIG_HOME
	os.Setenv("XDG_CONFIG_HOME", tempDir)

	configDir, err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() error = %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("Config directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Config path is not a directory")
	}

	// Verify permissions (0700)
	mode := info.Mode()
	if mode.Perm() != 0700 {
		t.Errorf("Config directory permissions = %o, want 0700", mode.Perm())
	}

	// Test idempotency - calling again should not error
	configDir2, err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() second call error = %v", err)
	}

	if configDir != configDir2 {
		t.Errorf("Config dir changed between calls: %s != %s", configDir, configDir2)
	}
}

func TestInitialize(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()

	tests := []struct {
		name      string
		cfgFile   string
		wantError bool
	}{
		{
			name:      "without config file",
			cfgFile:   "",
			wantError: false, // Should succeed even without config file
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Reset()

			err := Initialize(tt.cfgFile)
			if (err != nil) != tt.wantError {
				t.Errorf("Initialize() error = %v, wantError %v", err, tt.wantError)
			}

			// Verify defaults are set
			if viper.GetString("calendar.default_calendar_id") == "" {
				t.Error("calendar.default_calendar_id default not set")
			}

			if viper.GetString("output.default_format") == "" {
				t.Error("output.default_format default not set")
			}
		})
	}
}

func TestSetDefaults(t *testing.T) {
	viper.Reset()
	setDefaults()

	tests := []struct {
		key      string
		expected interface{}
		checkFn  func(string) interface{}
	}{
		{
			key:      "calendar.default_calendar_id",
			expected: "primary",
			checkFn:  func(k string) interface{} { return viper.GetString(k) },
		},
		{
			key:      "output.default_format",
			expected: "json",
			checkFn:  func(k string) interface{} { return viper.GetString(k) },
		},
		{
			key:      "output.color_enabled",
			expected: false,
			checkFn:  func(k string) interface{} { return viper.GetBool(k) },
		},
		{
			key:      "output.pretty_print",
			expected: true,
			checkFn:  func(k string) interface{} { return viper.GetBool(k) },
		},
		{
			key:      "auth.auto_refresh",
			expected: true,
			checkFn:  func(k string) interface{} { return viper.GetBool(k) },
		},
		{
			key:      "api.retry_attempts",
			expected: 3,
			checkFn:  func(k string) interface{} { return viper.GetInt(k) },
		},
		{
			key:      "api.retry_delay_ms",
			expected: 1000,
			checkFn:  func(k string) interface{} { return viper.GetInt(k) },
		},
		{
			key:      "api.timeout_seconds",
			expected: 30,
			checkFn:  func(k string) interface{} { return viper.GetInt(k) },
		},
		{
			key:      "events.default_duration_minutes",
			expected: 60,
			checkFn:  func(k string) interface{} { return viper.GetInt(k) },
		},
		{
			key:      "events.send_notifications",
			expected: true,
			checkFn:  func(k string) interface{} { return viper.GetBool(k) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got := tt.checkFn(tt.key)
			if got != tt.expected {
				t.Errorf("Default for %s = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	viper.Reset()
	setDefaults()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	// Verify default values are loaded
	if cfg.Calendar.DefaultCalendarID != "primary" {
		t.Errorf("DefaultCalendarID = %s, want primary", cfg.Calendar.DefaultCalendarID)
	}

	if cfg.Output.DefaultFormat != "json" {
		t.Errorf("DefaultFormat = %s, want json", cfg.Output.DefaultFormat)
	}

	if cfg.API.RetryAttempts != 3 {
		t.Errorf("RetryAttempts = %d, want 3", cfg.API.RetryAttempts)
	}

	if cfg.Events.DefaultDurationMinutes != 60 {
		t.Errorf("DefaultDurationMinutes = %d, want 60", cfg.Events.DefaultDurationMinutes)
	}
}

func TestSave(t *testing.T) {
	// Use temporary directory
	tempDir := t.TempDir()
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer os.Setenv("XDG_CONFIG_HOME", origXDG)

	os.Setenv("XDG_CONFIG_HOME", tempDir)

	viper.Reset()
	setDefaults()

	// Set a custom value
	viper.Set("calendar.default_calendar_id", "test@example.com")

	err := Save()
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file was created
	configPath := filepath.Join(tempDir, "gcal-cli", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created at %s", configPath)
	}

	// Read back and verify
	viper.Reset()
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("Failed to read saved config: %v", err)
	}

	if viper.GetString("calendar.default_calendar_id") != "test@example.com" {
		t.Error("Saved config value not persisted correctly")
	}
}

func TestGetString(t *testing.T) {
	viper.Reset()
	viper.Set("test.key", "test-value")

	got := GetString("test.key")
	if got != "test-value" {
		t.Errorf("GetString() = %s, want test-value", got)
	}

	// Test non-existent key
	got = GetString("non.existent")
	if got != "" {
		t.Errorf("GetString() for non-existent key = %s, want empty string", got)
	}
}

func TestGetBool(t *testing.T) {
	viper.Reset()
	viper.Set("test.bool", true)

	got := GetBool("test.bool")
	if !got {
		t.Error("GetBool() = false, want true")
	}

	// Test non-existent key
	got = GetBool("non.existent")
	if got {
		t.Error("GetBool() for non-existent key = true, want false")
	}
}

func TestGetInt(t *testing.T) {
	viper.Reset()
	viper.Set("test.int", 42)

	got := GetInt("test.int")
	if got != 42 {
		t.Errorf("GetInt() = %d, want 42", got)
	}

	// Test non-existent key
	got = GetInt("non.existent")
	if got != 0 {
		t.Errorf("GetInt() for non-existent key = %d, want 0", got)
	}
}

func TestGet(t *testing.T) {
	viper.Reset()
	viper.Set("test.key", "value")

	got := Get("test.key")
	if got != "value" {
		t.Errorf("Get() = %v, want value", got)
	}

	// Test non-existent key
	got = Get("non.existent")
	if got != nil {
		t.Errorf("Get() for non-existent key = %v, want nil", got)
	}
}

func TestSet(t *testing.T) {
	viper.Reset()

	Set("test.key", "new-value")

	got := viper.GetString("test.key")
	if got != "new-value" {
		t.Errorf("After Set(), value = %s, want new-value", got)
	}
}

func TestDisplayConfig(t *testing.T) {
	viper.Reset()
	setDefaults()

	output, err := DisplayConfig()
	if err != nil {
		t.Fatalf("DisplayConfig() error = %v", err)
	}

	if output == "" {
		t.Error("DisplayConfig() returned empty string")
	}

	// Verify output contains key sections
	expectedSections := []string{
		"Current Configuration:",
		"Calendar:",
		"Output:",
		"Authentication:",
		"API:",
		"Events:",
		"primary", // default calendar ID
		"json",    // default format
	}

	for _, section := range expectedSections {
		if !contains(output, section) {
			t.Errorf("DisplayConfig() output missing section: %s", section)
		}
	}
}

func TestConfigStructs(t *testing.T) {
	// Test that config structs can be created and marshaled
	cfg := &Config{
		Calendar: CalendarConfig{
			DefaultCalendarID: "test@example.com",
			DefaultTimezone:   "America/New_York",
		},
		Output: OutputConfig{
			DefaultFormat: "json",
			ColorEnabled:  false,
			PrettyPrint:   true,
		},
		Auth: AuthConfig{
			CredentialsPath: "/path/to/credentials.json",
			TokensPath:      "/path/to/tokens.json",
			AutoRefresh:     true,
		},
		API: APIConfig{
			RetryAttempts:   3,
			RetryDelayMs:    1000,
			RetryMaxDelayMs: 10000,
			TimeoutSeconds:  30,
			RateLimitBuffer: 0.9,
		},
		Events: EventsConfig{
			DefaultDurationMinutes: 60,
			DefaultReminderMinutes: 10,
			SendNotifications:      true,
		},
	}

	// Verify all fields are accessible
	if cfg.Calendar.DefaultCalendarID != "test@example.com" {
		t.Error("Calendar config mismatch")
	}

	if cfg.Output.DefaultFormat != "json" {
		t.Error("Output config mismatch")
	}

	if cfg.Auth.AutoRefresh != true {
		t.Error("Auth config mismatch")
	}

	if cfg.API.RetryAttempts != 3 {
		t.Error("API config mismatch")
	}

	if cfg.Events.DefaultDurationMinutes != 60 {
		t.Error("Events config mismatch")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
