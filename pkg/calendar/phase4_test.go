package calendar

import (
	"testing"
	"time"

	"github.com/btafoya/gcal-cli/pkg/types"
)

// TestTimezoneConverter tests timezone conversion utilities
func TestTimezoneConverter(t *testing.T) {
	tc := NewTimezoneConverter("America/New_York")

	// Test ConvertTime - conversion preserves the same moment in time, just different timezone representation
	now := time.Now()
	converted, err := tc.ConvertTime(now, "America/New_York", "America/Los_Angeles")
	if err != nil {
		t.Fatalf("ConvertTime failed: %v", err)
	}

	// The times should represent the same instant (equal when compared)
	if !now.Equal(converted) {
		t.Errorf("Converted time should represent same instant: %v != %v", now, converted)
	}

	// Test that timezone is actually different
	nyLoc, _ := time.LoadLocation("America/New_York")
	laLoc, _ := time.LoadLocation("America/Los_Angeles")

	nyTime := now.In(nyLoc)
	laTime := converted.In(laLoc)

	// The hour representation should differ by ~3 hours (accounting for DST)
	hourDiff := nyTime.Hour() - laTime.Hour()
	if hourDiff < 2 || hourDiff > 4 {
		t.Errorf("Expected 3 hour difference in representation, got %d", hourDiff)
	}
}

func TestTimezoneConverter_InvalidTimezone(t *testing.T) {
	tc := NewTimezoneConverter("UTC")

	_, err := tc.ConvertTime(time.Now(), "Invalid/Timezone", "UTC")
	if err == nil {
		t.Error("Expected error for invalid timezone, got nil")
	}
}

func TestValidateTimezone(t *testing.T) {
	tests := []struct {
		name    string
		tz      string
		wantErr bool
	}{
		{"valid UTC", "UTC", false},
		{"valid New York", "America/New_York", false},
		{"valid Tokyo", "Asia/Tokyo", false},
		{"invalid", "Invalid/Timezone", true},
		{"empty (allowed)", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimezone(tt.tz)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTimezone() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetCommonTimezones(t *testing.T) {
	timezones := GetCommonTimezones()

	if len(timezones) == 0 {
		t.Error("Expected non-empty timezone list")
	}

	// Check for some expected timezones
	expectedTz := map[string]bool{
		"UTC":                true,
		"America/New_York":   true,
		"Europe/London":      true,
		"Asia/Tokyo":         true,
	}

	for tz := range expectedTz {
		found := false
		for _, t := range timezones {
			if t == tz {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected timezone %s not found in common timezones", tz)
		}
	}
}

// TestMatchesFilter tests event filtering logic
func TestMatchesFilter(t *testing.T) {
	event := &types.Event{
		ID:       "test-event",
		Summary:  "Test Event",
		Location: "Conference Room A",
		Status:   "confirmed",
		Attendees: []types.Attendee{
			{Email: "user1@example.com", ResponseStatus: "accepted"},
			{Email: "user2@example.com", ResponseStatus: "needsAction"},
		},
		Recurrence: []string{"RRULE:FREQ=WEEKLY"},
		Start: types.EventTime{
			DateTime: "2024-01-15T10:00:00-05:00",
		},
	}

	tests := []struct {
		name    string
		filter  SearchFilter
		matches bool
	}{
		{
			name: "matches attendee",
			filter: SearchFilter{
				From:     time.Now().Add(-24 * time.Hour),
				To:       time.Now().Add(24 * time.Hour),
				Attendee: "user1@example.com",
			},
			matches: true,
		},
		{
			name: "doesn't match attendee",
			filter: SearchFilter{
				From:     time.Now().Add(-24 * time.Hour),
				To:       time.Now().Add(24 * time.Hour),
				Attendee: "other@example.com",
			},
			matches: false,
		},
		{
			name: "matches location",
			filter: SearchFilter{
				From:     time.Now().Add(-24 * time.Hour),
				To:       time.Now().Add(24 * time.Hour),
				Location: "Conference",
			},
			matches: true,
		},
		{
			name: "matches status",
			filter: SearchFilter{
				From:   time.Now().Add(-24 * time.Hour),
				To:     time.Now().Add(24 * time.Hour),
				Status: "confirmed",
			},
			matches: true,
		},
		{
			name: "doesn't match status",
			filter: SearchFilter{
				From:   time.Now().Add(-24 * time.Hour),
				To:     time.Now().Add(24 * time.Hour),
				Status: "cancelled",
			},
			matches: false,
		},
		{
			name: "matches has attendees",
			filter: SearchFilter{
				From:         time.Now().Add(-24 * time.Hour),
				To:           time.Now().Add(24 * time.Hour),
				HasAttendees: boolPtr(true),
			},
			matches: true,
		},
		{
			name: "matches is recurring",
			filter: SearchFilter{
				From:        time.Now().Add(-24 * time.Hour),
				To:          time.Now().Add(24 * time.Hour),
				IsRecurring: boolPtr(true),
			},
			matches: true,
		},
		{
			name: "doesn't match is all day",
			filter: SearchFilter{
				From:     time.Now().Add(-24 * time.Hour),
				To:       time.Now().Add(24 * time.Hour),
				IsAllDay: boolPtr(true),
			},
			matches: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesFilter(event, tt.filter)
			if result != tt.matches {
				t.Errorf("matchesFilter() = %v, want %v", result, tt.matches)
			}
		})
	}
}

// TestValidateSearchFilter tests search filter validation
func TestValidateSearchFilter(t *testing.T) {
	now := time.Now()
	later := now.Add(24 * time.Hour)

	tests := []struct {
		name    string
		filter  SearchFilter
		wantErr bool
	}{
		{
			name: "valid filter",
			filter: SearchFilter{
				From: now,
				To:   later,
			},
			wantErr: false,
		},
		{
			name: "missing from",
			filter: SearchFilter{
				To: later,
			},
			wantErr: true,
		},
		{
			name: "missing to",
			filter: SearchFilter{
				From: now,
			},
			wantErr: true,
		},
		{
			name: "to before from",
			filter: SearchFilter{
				From: later,
				To:   now,
			},
			wantErr: true,
		},
		{
			name: "invalid attendee email",
			filter: SearchFilter{
				From:     now,
				To:       later,
				Attendee: "invalid-email",
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			filter: SearchFilter{
				From:   now,
				To:     later,
				Status: "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid status confirmed",
			filter: SearchFilter{
				From:   now,
				To:     later,
				Status: "confirmed",
			},
			wantErr: false,
		},
		{
			name: "valid status tentative",
			filter: SearchFilter{
				From:   now,
				To:     later,
				Status: "tentative",
			},
			wantErr: false,
		},
		{
			name: "invalid order by",
			filter: SearchFilter{
				From:    now,
				To:      later,
				OrderBy: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSearchFilter(tt.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSearchFilter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGetBatchSummary tests batch operation summary
func TestGetBatchSummary(t *testing.T) {
	results := []*BatchResult{
		{Success: true, Index: 0, EventID: "event1"},
		{Success: true, Index: 1, EventID: "event2"},
		{Success: false, Index: 2, EventID: "event3", Error: types.ErrAPIError},
		{Success: true, Index: 3, EventID: "event4"},
		{Success: false, Index: 4, EventID: "event5", Error: types.ErrAPIError},
	}

	summary := GetBatchSummary(results)

	if summary["total"] != 5 {
		t.Errorf("Expected total 5, got %d", summary["total"])
	}

	if summary["success"] != 3 {
		t.Errorf("Expected success 3, got %d", summary["success"])
	}

	if summary["failed"] != 2 {
		t.Errorf("Expected failed 2, got %d", summary["failed"])
	}
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}
