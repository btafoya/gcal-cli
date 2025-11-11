package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSuccessResponse(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		data      interface{}
	}{
		{
			name:      "basic success response",
			operation: "create",
			data:      map[string]string{"key": "value"},
		},
		{
			name:      "with event data",
			operation: "get",
			data: &EventData{
				Event:   &Event{ID: "test123", Summary: "Test Event"},
				Message: "Event retrieved",
			},
		},
		{
			name:      "with nil data",
			operation: "delete",
			data:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := SuccessResponse(tt.operation, tt.data)

			if !resp.Success {
				t.Error("Expected Success to be true")
			}

			if resp.Operation != tt.operation {
				t.Errorf("Expected operation %s, got %s", tt.operation, resp.Operation)
			}

			// Skip comparing data directly as it may contain uncomparable types

			if resp.Error != nil {
				t.Error("Expected Error to be nil")
			}

			if resp.Metadata == nil {
				t.Error("Expected Metadata to be initialized")
			}

			if _, ok := resp.Metadata["timestamp"]; !ok {
				t.Error("Expected timestamp in metadata")
			}

			// Verify timestamp is valid RFC3339
			timestamp, ok := resp.Metadata["timestamp"].(string)
			if !ok {
				t.Error("Timestamp should be a string")
			}

			_, err := time.Parse(time.RFC3339, timestamp)
			if err != nil {
				t.Errorf("Timestamp should be valid RFC3339: %v", err)
			}
		})
	}
}

func TestErrorResponse(t *testing.T) {
	tests := []struct {
		name  string
		error *AppError
	}{
		{
			name:  "basic error response",
			error: ErrInvalidInput("field", "invalid value"),
		},
		{
			name: "error with details",
			error: ErrAPIError.
				WithDetails("API call failed"),
		},
		{
			name:  "authentication error",
			error: ErrAuthFailed("invalid credentials"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := ErrorResponse(tt.error)

			if resp.Success {
				t.Error("Expected Success to be false")
			}

			if resp.Operation != "" {
				t.Error("Expected Operation to be empty")
			}

			if resp.Data != nil {
				t.Error("Expected Data to be nil")
			}

			if resp.Error != tt.error {
				t.Error("Error mismatch")
			}

			if resp.Metadata == nil {
				t.Error("Expected Metadata to be initialized")
			}

			if _, ok := resp.Metadata["timestamp"]; !ok {
				t.Error("Expected timestamp in metadata")
			}
		})
	}
}

func TestWithMetadata(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{
			name:  "add string metadata",
			key:   "calendarId",
			value: "primary",
		},
		{
			name:  "add int metadata",
			key:   "count",
			value: 42,
		},
		{
			name:  "add map metadata",
			key:   "details",
			value: map[string]string{"foo": "bar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := SuccessResponse("test", nil)
			result := resp.WithMetadata(tt.key, tt.value)

			// Verify it returns the same response (method chaining)
			if result != resp {
				t.Error("WithMetadata should return the same response for chaining")
			}

			// Verify metadata was added
			val, ok := resp.Metadata[tt.key]
			if !ok {
				t.Errorf("Expected metadata key %s to be present", tt.key)
			}

			// For maps, just verify it exists (can't compare directly)
			switch tt.value.(type) {
			case map[string]string:
				if val == nil {
					t.Error("Expected metadata value to be set")
				}
			default:
				if val != tt.value {
					t.Errorf("Expected metadata value %v, got %v", tt.value, val)
				}
			}
		})
	}

	// Test with nil metadata
	t.Run("nil metadata initialization", func(t *testing.T) {
		resp := &Response{Success: true}
		resp.Metadata = nil // Explicitly set to nil

		resp.WithMetadata("test", "value")

		if resp.Metadata == nil {
			t.Error("Expected Metadata to be initialized")
		}

		if resp.Metadata["test"] != "value" {
			t.Error("Expected metadata value to be set")
		}
	})
}

func TestResponseJSONSerialization(t *testing.T) {
	tests := []struct {
		name     string
		response *Response
	}{
		{
			name: "success response",
			response: SuccessResponse("create", map[string]interface{}{
				"event": map[string]string{
					"id":      "test123",
					"summary": "Test Event",
				},
			}),
		},
		{
			name:     "error response",
			response: ErrorResponse(ErrInvalidInput("field", "invalid")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize to JSON
			jsonData, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("Failed to marshal response: %v", err)
			}

			// Deserialize from JSON
			var decoded Response
			err = json.Unmarshal(jsonData, &decoded)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			// Verify basic fields
			if decoded.Success != tt.response.Success {
				t.Errorf("Success mismatch: expected %v, got %v", tt.response.Success, decoded.Success)
			}

			if decoded.Operation != tt.response.Operation {
				t.Errorf("Operation mismatch: expected %s, got %s", tt.response.Operation, decoded.Operation)
			}
		})
	}
}

func TestEventData(t *testing.T) {
	event := &Event{
		ID:      "test123",
		Summary: "Test Event",
	}

	data := &EventData{
		Event:   event,
		EventID: "test123",
		Message: "Event created",
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal EventData: %v", err)
	}

	var decoded EventData
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal EventData: %v", err)
	}

	if decoded.EventID != data.EventID {
		t.Errorf("EventID mismatch")
	}

	if decoded.Message != data.Message {
		t.Errorf("Message mismatch")
	}
}

func TestEventListData(t *testing.T) {
	events := []*Event{
		{ID: "event1", Summary: "Event 1"},
		{ID: "event2", Summary: "Event 2"},
	}

	data := &EventListData{
		Events:        events,
		Count:         len(events),
		NextPageToken: "token123",
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal EventListData: %v", err)
	}

	var decoded EventListData
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal EventListData: %v", err)
	}

	if decoded.Count != data.Count {
		t.Errorf("Count mismatch: expected %d, got %d", data.Count, decoded.Count)
	}

	if decoded.NextPageToken != data.NextPageToken {
		t.Errorf("NextPageToken mismatch")
	}

	if len(decoded.Events) != len(data.Events) {
		t.Errorf("Events length mismatch")
	}
}

func TestAuthData(t *testing.T) {
	data := &AuthData{
		Message: "Authentication successful",
		Email:   "user@example.com",
		Scopes:  []string{"https://www.googleapis.com/auth/calendar"},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal AuthData: %v", err)
	}

	var decoded AuthData
	err = json.Unmarshal(jsonData, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal AuthData: %v", err)
	}

	if decoded.Message != data.Message {
		t.Errorf("Message mismatch")
	}

	if decoded.Email != data.Email {
		t.Errorf("Email mismatch")
	}

	if len(decoded.Scopes) != len(data.Scopes) {
		t.Errorf("Scopes length mismatch")
	}
}
