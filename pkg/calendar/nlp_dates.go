package calendar

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ParseNaturalLanguageDate parses natural language date/time strings into RFC3339 format
// Examples: "tomorrow at 2pm", "next Monday at 3:30pm", "in 2 hours"
func ParseNaturalLanguageDate(input string, timezone *time.Location) (string, error) {
	if timezone == nil {
		timezone = time.Local
	}

	input = strings.ToLower(strings.TrimSpace(input))
	now := time.Now().In(timezone)

	// Try relative time first (tomorrow, next week, etc.)
	if parsed, err := parseRelativeDate(input, now, timezone); err == nil {
		return parsed, nil
	}

	// Try time offsets (in 2 hours, in 30 minutes)
	if parsed, err := parseTimeOffset(input, now); err == nil {
		return parsed, nil
	}

	// Try day of week (Monday, next Tuesday, etc.)
	if parsed, err := parseDayOfWeek(input, now, timezone); err == nil {
		return parsed, nil
	}

	// Try specific date with time (Jan 15 at 2pm, 2024-01-15 at 14:00)
	if parsed, err := parseSpecificDateTime(input, now, timezone); err == nil {
		return parsed, nil
	}

	return "", fmt.Errorf("unable to parse natural language date: %s", input)
}

// parseRelativeDate handles "today", "tomorrow", "yesterday"
func parseRelativeDate(input string, now time.Time, tz *time.Location) (string, error) {
	// Extract time component if present
	timeStr := ""
	dateStr := input

	// Pattern: "tomorrow at 2pm" or "today at 14:00"
	atRegex := regexp.MustCompile(`(.*?)\s+at\s+(.+)`)
	if matches := atRegex.FindStringSubmatch(input); len(matches) == 3 {
		dateStr = matches[1]
		timeStr = matches[2]
	}

	var targetDate time.Time

	switch strings.TrimSpace(dateStr) {
	case "now":
		return now.Format(time.RFC3339), nil
	case "today":
		targetDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, tz)
	case "tomorrow":
		targetDate = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, tz)
	case "yesterday":
		targetDate = time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, tz)
	default:
		return "", fmt.Errorf("not a relative date")
	}

	// If time was specified, parse and apply it
	if timeStr != "" {
		parsedTime, err := parseTimeOfDay(timeStr, tz)
		if err != nil {
			return "", err
		}
		targetDate = time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
			parsedTime.Hour(), parsedTime.Minute(), 0, 0, tz)
	}

	return targetDate.Format(time.RFC3339), nil
}

// parseTimeOffset handles "in 2 hours", "in 30 minutes", "in 1 week"
func parseTimeOffset(input string, now time.Time) (string, error) {
	// Pattern: "in <number> <unit>"
	offsetRegex := regexp.MustCompile(`^in\s+(\d+)\s+(minute|minutes|hour|hours|day|days|week|weeks|month|months)$`)
	matches := offsetRegex.FindStringSubmatch(input)

	if len(matches) != 3 {
		return "", fmt.Errorf("not a time offset")
	}

	var amount int
	fmt.Sscanf(matches[1], "%d", &amount)
	unit := matches[2]

	var result time.Time
	switch unit {
	case "minute", "minutes":
		result = now.Add(time.Duration(amount) * time.Minute)
	case "hour", "hours":
		result = now.Add(time.Duration(amount) * time.Hour)
	case "day", "days":
		result = now.AddDate(0, 0, amount)
	case "week", "weeks":
		result = now.AddDate(0, 0, amount*7)
	case "month", "months":
		result = now.AddDate(0, amount, 0)
	default:
		return "", fmt.Errorf("unknown time unit: %s", unit)
	}

	return result.Format(time.RFC3339), nil
}

// parseDayOfWeek handles "Monday", "next Tuesday", "this Friday"
func parseDayOfWeek(input string, now time.Time, tz *time.Location) (string, error) {
	// Extract time component if present
	timeStr := ""
	dateStr := input

	atRegex := regexp.MustCompile(`(.*?)\s+at\s+(.+)`)
	if matches := atRegex.FindStringSubmatch(input); len(matches) == 3 {
		dateStr = matches[1]
		timeStr = matches[2]
	}

	// Determine modifier (next, this, last)
	modifier := "next"
	dayStr := dateStr

	if strings.HasPrefix(dateStr, "next ") {
		dayStr = strings.TrimPrefix(dateStr, "next ")
	} else if strings.HasPrefix(dateStr, "this ") {
		modifier = "this"
		dayStr = strings.TrimPrefix(dateStr, "this ")
	} else if strings.HasPrefix(dateStr, "last ") {
		modifier = "last"
		dayStr = strings.TrimPrefix(dateStr, "last ")
	}

	// Map day name to weekday
	dayMap := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
		"sun":       time.Sunday,
		"mon":       time.Monday,
		"tue":       time.Tuesday,
		"wed":       time.Wednesday,
		"thu":       time.Thursday,
		"fri":       time.Friday,
		"sat":       time.Saturday,
	}

	targetWeekday, ok := dayMap[strings.ToLower(dayStr)]
	if !ok {
		return "", fmt.Errorf("not a day of week")
	}

	// Calculate target date
	var targetDate time.Time
	currentWeekday := now.Weekday()

	switch modifier {
	case "next":
		// Next occurrence of this weekday
		daysUntil := int(targetWeekday - currentWeekday)
		if daysUntil <= 0 {
			daysUntil += 7
		}
		targetDate = now.AddDate(0, 0, daysUntil)
	case "this":
		// This week's occurrence (or next week if already passed)
		daysUntil := int(targetWeekday - currentWeekday)
		if daysUntil < 0 {
			daysUntil += 7
		}
		targetDate = now.AddDate(0, 0, daysUntil)
	case "last":
		// Last occurrence
		daysAgo := int(currentWeekday - targetWeekday)
		if daysAgo <= 0 {
			daysAgo += 7
		}
		targetDate = now.AddDate(0, 0, -daysAgo)
	}

	targetDate = time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, tz)

	// Apply time if specified
	if timeStr != "" {
		parsedTime, err := parseTimeOfDay(timeStr, tz)
		if err != nil {
			return "", err
		}
		targetDate = time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(),
			parsedTime.Hour(), parsedTime.Minute(), 0, 0, tz)
	}

	return targetDate.Format(time.RFC3339), nil
}

// parseSpecificDateTime handles "Jan 15 at 2pm", "2024-01-15 at 14:00"
func parseSpecificDateTime(input string, now time.Time, tz *time.Location) (string, error) {
	// This is a simplified version - a full implementation would handle more formats
	// For now, we delegate to standard time parsing
	return "", fmt.Errorf("specific date/time parsing not implemented")
}

// parseTimeOfDay parses time strings like "2pm", "14:30", "3:30pm"
func parseTimeOfDay(timeStr string, tz *time.Location) (time.Time, error) {
	timeStr = strings.TrimSpace(strings.ToLower(timeStr))

	// Try formats: "2pm", "14:30", "2:30pm", "14:30:00"
	formats := []string{
		"3pm",
		"3:04pm",
		"15:04",
		"15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", timeStr)
}

// IsNaturalLanguageDate checks if a string appears to be a natural language date
func IsNaturalLanguageDate(input string) bool {
	input = strings.ToLower(strings.TrimSpace(input))

	// Check for common natural language patterns
	nlPatterns := []string{
		"now", "today", "tomorrow", "yesterday",
		"next ", "this ", "last ",
		"in ", // for "in 2 hours"
		"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday",
		"mon", "tue", "wed", "thu", "fri", "sat", "sun",
	}

	for _, pattern := range nlPatterns {
		if strings.Contains(input, pattern) {
			return true
		}
	}

	return false
}
