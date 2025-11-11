package calendar

import (
	"strings"
	"testing"
	"time"
)

func TestParseRelativeDate(t *testing.T) {
	tz, _ := time.LoadLocation("America/New_York")
	now := time.Date(2024, 1, 15, 10, 0, 0, 0, tz) // Monday, Jan 15, 2024, 10:00 AM

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "now",
			input:   "now",
			want:    now.Format(time.RFC3339),
			wantErr: false,
		},
		{
			name:  "today",
			input: "today",
			want:  "2024-01-15T00:00:00-05:00",
			wantErr: false,
		},
		{
			name:  "tomorrow",
			input: "tomorrow",
			want:  "2024-01-16T00:00:00-05:00",
			wantErr: false,
		},
		{
			name:  "yesterday",
			input: "yesterday",
			want:  "2024-01-14T00:00:00-05:00",
			wantErr: false,
		},
		{
			name:  "tomorrow at 2pm",
			input: "tomorrow at 2pm",
			want:  "2024-01-16T14:00:00-05:00",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRelativeDate(tt.input, now, tz)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRelativeDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseRelativeDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTimeOffset(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(string) bool
	}{
		{
			name:    "in 2 hours",
			input:   "in 2 hours",
			wantErr: false,
			check: func(result string) bool {
				parsed, _ := time.Parse(time.RFC3339, result)
				expected := now.Add(2 * time.Hour)
				return parsed.Equal(expected)
			},
		},
		{
			name:    "in 30 minutes",
			input:   "in 30 minutes",
			wantErr: false,
			check: func(result string) bool {
				parsed, _ := time.Parse(time.RFC3339, result)
				expected := now.Add(30 * time.Minute)
				return parsed.Equal(expected)
			},
		},
		{
			name:    "in 1 day",
			input:   "in 1 day",
			wantErr: false,
			check: func(result string) bool {
				parsed, _ := time.Parse(time.RFC3339, result)
				expected := now.AddDate(0, 0, 1)
				return parsed.Equal(expected)
			},
		},
		{
			name:    "invalid format",
			input:   "2 hours",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimeOffset(tt.input, now)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTimeOffset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(got) {
				t.Errorf("parseTimeOffset() result check failed for %v", got)
			}
		})
	}
}

func TestParseDayOfWeek(t *testing.T) {
	tz, _ := time.LoadLocation("America/New_York")
	now := time.Date(2024, 1, 15, 10, 0, 0, 0, tz) // Monday, Jan 15, 2024

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "next Monday",
			input: "next monday",
			want:  "2024-01-22T00:00:00-05:00", // Next Monday
			wantErr: false,
		},
		{
			name:  "next Friday",
			input: "next friday",
			want:  "2024-01-19T00:00:00-05:00", // This week's Friday
			wantErr: false,
		},
		{
			name:  "Monday at 2pm",
			input: "monday at 2pm",
			want:  "2024-01-22T14:00:00-05:00", // Next Monday at 2pm
			wantErr: false,
		},
		{
			name:  "this Wednesday",
			input: "this wednesday",
			want:  "2024-01-17T00:00:00-05:00", // This week's Wednesday
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDayOfWeek(tt.input, now, tz)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDayOfWeek() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseDayOfWeek() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTimeOfDay(t *testing.T) {
	tz, _ := time.LoadLocation("America/New_York")

	tests := []struct {
		name    string
		input   string
		wantHour int
		wantMin int
		wantErr bool
	}{
		{
			name:    "2pm",
			input:   "2pm",
			wantHour: 14,
			wantMin: 0,
			wantErr: false,
		},
		{
			name:    "3:30pm",
			input:   "3:30pm",
			wantHour: 15,
			wantMin: 30,
			wantErr: false,
		},
		{
			name:    "14:30",
			input:   "14:30",
			wantHour: 14,
			wantMin: 30,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimeOfDay(tt.input, tz)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTimeOfDay() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Hour() != tt.wantHour || got.Minute() != tt.wantMin {
					t.Errorf("parseTimeOfDay() = %d:%02d, want %d:%02d",
						got.Hour(), got.Minute(), tt.wantHour, tt.wantMin)
				}
			}
		})
	}
}

func TestIsNaturalLanguageDate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "today",
			input: "today",
			want:  true,
		},
		{
			name:  "tomorrow at 2pm",
			input: "tomorrow at 2pm",
			want:  true,
		},
		{
			name:  "next Monday",
			input: "next Monday",
			want:  true,
		},
		{
			name:  "in 2 hours",
			input: "in 2 hours",
			want:  true,
		},
		{
			name:  "RFC3339",
			input: "2024-01-15T10:00:00Z",
			want:  false,
		},
		{
			name:  "regular date",
			input: "2024-01-15",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNaturalLanguageDate(tt.input)
			if got != tt.want {
				t.Errorf("IsNaturalLanguageDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseNaturalLanguageDate(t *testing.T) {
	tz, _ := time.LoadLocation("America/New_York")

	tests := []struct {
		name    string
		input   string
		wantErr bool
		checkFn func(string) bool
	}{
		{
			name:    "tomorrow",
			input:   "tomorrow",
			wantErr: false,
			checkFn: func(result string) bool {
				return strings.Contains(result, "T00:00:00")
			},
		},
		{
			name:    "in 2 hours",
			input:   "in 2 hours",
			wantErr: false,
			checkFn: func(result string) bool {
				return len(result) > 0
			},
		},
		{
			name:    "next Monday",
			input:   "next Monday",
			wantErr: false,
			checkFn: func(result string) bool {
				return len(result) > 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseNaturalLanguageDate(tt.input, tz)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNaturalLanguageDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFn != nil && !tt.checkFn(got) {
				t.Errorf("ParseNaturalLanguageDate() check failed for result: %v", got)
			}
		})
	}
}
