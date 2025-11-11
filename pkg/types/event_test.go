package types

import (
	"encoding/json"
	"testing"
)

func TestEventJSONSerialization(t *testing.T) {
	tests := []struct {
		name  string
		event *Event
	}{
		{
			name: "basic event",
			event: &Event{
				ID:      "test123",
				Summary: "Test Event",
				Status:  "confirmed",
			},
		},
		{
			name: "event with all fields",
			event: &Event{
				ID:          "test456",
				Summary:     "Full Event",
				Description: "Event description",
				Start: EventTime{
					DateTime: "2024-01-15T10:00:00Z",
					TimeZone: "America/New_York",
				},
				End: EventTime{
					DateTime: "2024-01-15T11:00:00Z",
					TimeZone: "America/New_York",
				},
				Status: "confirmed",
				Attendees: []Attendee{
					{
						Email:          "user@example.com",
						ResponseStatus: "needsAction",
						Organizer:      false,
						DisplayName:    "Test User",
					},
				},
				Recurrence: []string{"RRULE:FREQ=WEEKLY;COUNT=10"},
				Location:   "Conference Room A",
				HTMLLink:   "https://calendar.google.com/event?eid=test456",
			},
		},
		{
			name: "all-day event",
			event: &Event{
				ID:      "allday123",
				Summary: "All Day Event",
				Start: EventTime{
					Date: "2024-01-15",
				},
				End: EventTime{
					Date: "2024-01-16",
				},
				Status: "confirmed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			jsonData, err := json.Marshal(tt.event)
			if err != nil {
				t.Fatalf("Failed to marshal event: %v", err)
			}

			// Unmarshal from JSON
			var decoded Event
			err = json.Unmarshal(jsonData, &decoded)
			if err != nil {
				t.Fatalf("Failed to unmarshal event: %v", err)
			}

			// Verify basic fields
			if decoded.ID != tt.event.ID {
				t.Errorf("ID mismatch: expected %s, got %s", tt.event.ID, decoded.ID)
			}

			if decoded.Summary != tt.event.Summary {
				t.Errorf("Summary mismatch: expected %s, got %s", tt.event.Summary, decoded.Summary)
			}

			if decoded.Status != tt.event.Status {
				t.Errorf("Status mismatch: expected %s, got %s", tt.event.Status, decoded.Status)
			}

			if decoded.Description != tt.event.Description {
				t.Errorf("Description mismatch: expected %s, got %s", tt.event.Description, decoded.Description)
			}

			if decoded.Location != tt.event.Location {
				t.Errorf("Location mismatch: expected %s, got %s", tt.event.Location, decoded.Location)
			}

			if decoded.HTMLLink != tt.event.HTMLLink {
				t.Errorf("HTMLLink mismatch: expected %s, got %s", tt.event.HTMLLink, decoded.HTMLLink)
			}
		})
	}
}

func TestEventTimeJSONSerialization(t *testing.T) {
	tests := []struct {
		name string
		time EventTime
	}{
		{
			name: "datetime with timezone",
			time: EventTime{
				DateTime: "2024-01-15T10:00:00Z",
				TimeZone: "America/New_York",
			},
		},
		{
			name: "date only (all-day)",
			time: EventTime{
				Date: "2024-01-15",
			},
		},
		{
			name: "datetime without timezone",
			time: EventTime{
				DateTime: "2024-01-15T10:00:00",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			jsonData, err := json.Marshal(tt.time)
			if err != nil {
				t.Fatalf("Failed to marshal EventTime: %v", err)
			}

			// Unmarshal from JSON
			var decoded EventTime
			err = json.Unmarshal(jsonData, &decoded)
			if err != nil {
				t.Fatalf("Failed to unmarshal EventTime: %v", err)
			}

			// Verify fields
			if decoded.DateTime != tt.time.DateTime {
				t.Errorf("DateTime mismatch: expected %s, got %s", tt.time.DateTime, decoded.DateTime)
			}

			if decoded.Date != tt.time.Date {
				t.Errorf("Date mismatch: expected %s, got %s", tt.time.Date, decoded.Date)
			}

			if decoded.TimeZone != tt.time.TimeZone {
				t.Errorf("TimeZone mismatch: expected %s, got %s", tt.time.TimeZone, decoded.TimeZone)
			}
		})
	}
}

func TestAttendeeJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		attendee Attendee
	}{
		{
			name: "basic attendee",
			attendee: Attendee{
				Email:          "user@example.com",
				ResponseStatus: "needsAction",
			},
		},
		{
			name: "organizer attendee",
			attendee: Attendee{
				Email:          "organizer@example.com",
				ResponseStatus: "accepted",
				Organizer:      true,
				DisplayName:    "Event Organizer",
			},
		},
		{
			name: "accepted attendee",
			attendee: Attendee{
				Email:          "attendee@example.com",
				ResponseStatus: "accepted",
				DisplayName:    "Attendee Name",
			},
		},
		{
			name: "declined attendee",
			attendee: Attendee{
				Email:          "declined@example.com",
				ResponseStatus: "declined",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			jsonData, err := json.Marshal(tt.attendee)
			if err != nil {
				t.Fatalf("Failed to marshal Attendee: %v", err)
			}

			// Unmarshal from JSON
			var decoded Attendee
			err = json.Unmarshal(jsonData, &decoded)
			if err != nil {
				t.Fatalf("Failed to unmarshal Attendee: %v", err)
			}

			// Verify fields
			if decoded.Email != tt.attendee.Email {
				t.Errorf("Email mismatch: expected %s, got %s", tt.attendee.Email, decoded.Email)
			}

			if decoded.ResponseStatus != tt.attendee.ResponseStatus {
				t.Errorf("ResponseStatus mismatch: expected %s, got %s", tt.attendee.ResponseStatus, decoded.ResponseStatus)
			}

			if decoded.Organizer != tt.attendee.Organizer {
				t.Errorf("Organizer mismatch: expected %v, got %v", tt.attendee.Organizer, decoded.Organizer)
			}

			if decoded.DisplayName != tt.attendee.DisplayName {
				t.Errorf("DisplayName mismatch: expected %s, got %s", tt.attendee.DisplayName, decoded.DisplayName)
			}
		})
	}
}

func TestEventOmitEmpty(t *testing.T) {
	// Test that omitempty fields are not included in JSON when empty
	event := &Event{
		ID:      "test123",
		Summary: "Basic Event",
		Status:  "confirmed",
	}

	jsonData, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	jsonStr := string(jsonData)

	// These fields should not be in JSON when empty
	omitFields := []string{"description", "attendees", "recurrence", "location", "htmlLink"}
	for _, field := range omitFields {
		if contains(jsonStr, field) {
			t.Errorf("Expected field %s to be omitted from JSON, but it was included", field)
		}
	}
}

func TestEventTimeOmitEmpty(t *testing.T) {
	// Test all-day event (only date, no datetime/timezone)
	allDay := EventTime{
		Date: "2024-01-15",
	}

	jsonData, err := json.Marshal(allDay)
	if err != nil {
		t.Fatalf("Failed to marshal EventTime: %v", err)
	}

	jsonStr := string(jsonData)

	if contains(jsonStr, "dateTime") {
		t.Error("Expected dateTime to be omitted for all-day event")
	}

	if contains(jsonStr, "timeZone") {
		t.Error("Expected timeZone to be omitted for all-day event")
	}

	// Test datetime event (no date field)
	datetime := EventTime{
		DateTime: "2024-01-15T10:00:00Z",
		TimeZone: "America/New_York",
	}

	jsonData, err = json.Marshal(datetime)
	if err != nil {
		t.Fatalf("Failed to marshal EventTime: %v", err)
	}

	jsonStr = string(jsonData)

	if contains(jsonStr, "\"date\"") {
		t.Error("Expected date to be omitted for datetime event")
	}
}

func TestAttendeeOmitEmpty(t *testing.T) {
	// Minimal attendee
	attendee := Attendee{
		Email:          "user@example.com",
		ResponseStatus: "needsAction",
	}

	jsonData, err := json.Marshal(attendee)
	if err != nil {
		t.Fatalf("Failed to marshal Attendee: %v", err)
	}

	jsonStr := string(jsonData)

	// organizer should be omitted when false
	if contains(jsonStr, "organizer") {
		t.Error("Expected organizer to be omitted when false")
	}

	// displayName should be omitted when empty
	if contains(jsonStr, "displayName") {
		t.Error("Expected displayName to be omitted when empty")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
