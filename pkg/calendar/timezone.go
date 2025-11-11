package calendar

import (
	"fmt"
	"time"

	"github.com/btafoya/gcal-cli/pkg/types"
)

// TimezoneConverter handles timezone conversions for events
type TimezoneConverter struct {
	DefaultTimezone string
}

// NewTimezoneConverter creates a new timezone converter
func NewTimezoneConverter(defaultTimezone string) *TimezoneConverter {
	if defaultTimezone == "" {
		defaultTimezone = "UTC"
	}
	return &TimezoneConverter{
		DefaultTimezone: defaultTimezone,
	}
}

// ConvertTime converts a time from one timezone to another
func (tc *TimezoneConverter) ConvertTime(t time.Time, fromTz, toTz string) (time.Time, error) {
	// Load source timezone
	fromLoc, err := time.LoadLocation(fromTz)
	if err != nil {
		return time.Time{}, types.ErrInvalidInput("timezone",
			fmt.Sprintf("invalid source timezone: %s", fromTz))
	}

	// Load destination timezone
	toLoc, err := time.LoadLocation(toTz)
	if err != nil {
		return time.Time{}, types.ErrInvalidInput("timezone",
			fmt.Sprintf("invalid destination timezone: %s", toTz))
	}

	// Convert time to source timezone first
	tInFrom := t.In(fromLoc)

	// Then convert to destination timezone
	return tInFrom.In(toLoc), nil
}

// ConvertToLocal converts a time to the local system timezone
func (tc *TimezoneConverter) ConvertToLocal(t time.Time) time.Time {
	return t.Local()
}

// ConvertToUTC converts a time to UTC
func (tc *TimezoneConverter) ConvertToUTC(t time.Time) time.Time {
	return t.UTC()
}

// ParseTimeInTimezone parses a time string in a specific timezone
func (tc *TimezoneConverter) ParseTimeInTimezone(timeStr, tz string) (time.Time, error) {
	// Try RFC3339 first (already has timezone info)
	t, err := time.Parse(time.RFC3339, timeStr)
	if err == nil {
		return t, nil
	}

	// Load specified timezone
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.Time{}, types.ErrInvalidInput("timezone",
			fmt.Sprintf("invalid timezone: %s", tz))
	}

	// Try common formats
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
	}

	for _, format := range formats {
		t, err := time.ParseInLocation(format, timeStr, loc)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, types.ErrInvalidInput("time",
		fmt.Sprintf("invalid time format: %s (expected RFC3339 or YYYY-MM-DD HH:MM)", timeStr))
}

// FormatTimeInTimezone formats a time for display in a specific timezone
func (tc *TimezoneConverter) FormatTimeInTimezone(t time.Time, tz string) (string, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return "", types.ErrInvalidInput("timezone",
			fmt.Sprintf("invalid timezone: %s", tz))
	}

	return t.In(loc).Format(time.RFC3339), nil
}

// GetTimezoneOffset returns the offset from UTC for a timezone at a specific time
func (tc *TimezoneConverter) GetTimezoneOffset(tz string, at time.Time) (int, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return 0, types.ErrInvalidInput("timezone",
			fmt.Sprintf("invalid timezone: %s", tz))
	}

	_, offset := at.In(loc).Zone()
	return offset, nil
}

// ValidateTimezone checks if a timezone string is valid
func ValidateTimezone(tz string) error {
	if tz == "" {
		return nil // Empty timezone is allowed (will use default)
	}

	_, err := time.LoadLocation(tz)
	if err != nil {
		return types.ErrInvalidInput("timezone",
			fmt.Sprintf("invalid timezone: %s", tz))
	}

	return nil
}

// GetCommonTimezones returns a list of commonly used timezones
func GetCommonTimezones() []string {
	return []string{
		"UTC",
		"America/New_York",
		"America/Chicago",
		"America/Denver",
		"America/Los_Angeles",
		"America/Toronto",
		"America/Vancouver",
		"Europe/London",
		"Europe/Paris",
		"Europe/Berlin",
		"Asia/Tokyo",
		"Asia/Shanghai",
		"Asia/Singapore",
		"Asia/Kolkata",
		"Australia/Sydney",
		"Pacific/Auckland",
	}
}
