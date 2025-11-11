package output

import (
	"strings"
	"testing"

	"github.com/btafoya/gcal-cli/pkg/types"
)

func TestTextFormatter_Format_Success(t *testing.T) {
	formatter := &TextFormatter{}

	data := &types.EventData{
		Message: "Event created successfully",
		Event: &types.Event{
			ID:      "test123",
			Summary: "Test Event",
			Start: types.EventTime{
				DateTime: "2024-01-15T10:00:00Z",
				TimeZone: "UTC",
			},
			End: types.EventTime{
				DateTime: "2024-01-15T11:00:00Z",
				TimeZone: "UTC",
			},
			Status: "confirmed",
		},
	}

	response := types.SuccessResponse("create", data)

	output, err := formatter.Format(response)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Verify output contains expected elements
	if !strings.Contains(output, "✓ Success") {
		t.Error("Output should contain success indicator")
	}

	if !strings.Contains(output, "Test Event") {
		t.Error("Output should contain event summary")
	}

	if !strings.Contains(output, "test123") {
		t.Error("Output should contain event ID")
	}
}

func TestTextFormatter_Format_Error(t *testing.T) {
	formatter := &TextFormatter{}

	appErr := types.ErrAuthFailed("Authentication required")
	response := types.ErrorResponse(appErr)

	output, err := formatter.Format(response)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Verify error output
	if !strings.Contains(output, "✗ Error") {
		t.Error("Output should contain error indicator")
	}

	if !strings.Contains(output, types.ErrCodeAuthFailed) {
		t.Error("Output should contain error code")
	}

	if !strings.Contains(output, "Authentication required") {
		t.Error("Output should contain error message")
	}

	if !strings.Contains(output, "Suggested Action") {
		t.Error("Output should contain suggested action")
	}
}

func TestTextFormatter_Format_EventList(t *testing.T) {
	formatter := &TextFormatter{}

	data := &types.EventListData{
		Events: []*types.Event{
			{
				ID:      "event1",
				Summary: "Event 1",
				Start: types.EventTime{
					DateTime: "2024-01-15T10:00:00Z",
				},
				End: types.EventTime{
					DateTime: "2024-01-15T11:00:00Z",
				},
			},
			{
				ID:      "event2",
				Summary: "Event 2",
				Start: types.EventTime{
					DateTime: "2024-01-16T10:00:00Z",
				},
				End: types.EventTime{
					DateTime: "2024-01-16T11:00:00Z",
				},
			},
		},
		Count: 2,
	}

	response := types.SuccessResponse("list", data)

	output, err := formatter.Format(response)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Verify list output
	if !strings.Contains(output, "Found 2 event(s)") {
		t.Error("Output should contain event count")
	}

	if !strings.Contains(output, "Event 1") {
		t.Error("Output should contain first event")
	}

	if !strings.Contains(output, "Event 2") {
		t.Error("Output should contain second event")
	}
}
