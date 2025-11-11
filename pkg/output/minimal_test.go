package output

import (
	"strings"
	"testing"

	"github.com/btafoya/gcal-cli/pkg/types"
)

func TestMinimalFormatter_Format_Success(t *testing.T) {
	formatter := &MinimalFormatter{}

	data := &types.EventData{
		Event: &types.Event{
			ID: "test123",
		},
	}

	response := types.SuccessResponse("create", data)

	output, err := formatter.Format(response)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Should output only the ID
	output = strings.TrimSpace(output)
	if output != "test123" {
		t.Errorf("Expected 'test123', got '%s'", output)
	}
}

func TestMinimalFormatter_Format_EventList(t *testing.T) {
	formatter := &MinimalFormatter{}

	data := &types.EventListData{
		Events: []*types.Event{
			{ID: "event1"},
			{ID: "event2"},
			{ID: "event3"},
		},
		Count: 3,
	}

	response := types.SuccessResponse("list", data)

	output, err := formatter.Format(response)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Should output each ID on a separate line
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}

	if lines[0] != "event1" || lines[1] != "event2" || lines[2] != "event3" {
		t.Error("Output should contain all event IDs in order")
	}
}

func TestMinimalFormatter_Format_Error(t *testing.T) {
	formatter := &MinimalFormatter{}

	appErr := types.ErrNotFound("Event", "123")
	response := types.ErrorResponse(appErr)

	output, err := formatter.Format(response)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Should output error code
	output = strings.TrimSpace(output)
	expected := "ERROR: " + types.ErrCodeNotFound
	if output != expected {
		t.Errorf("Expected '%s', got '%s'", expected, output)
	}
}

func TestMinimalFormatter_Format_AuthData(t *testing.T) {
	formatter := &MinimalFormatter{}

	data := &types.AuthData{
		Message: "Authentication successful",
		Email:   "user@example.com",
	}

	response := types.SuccessResponse("auth_login", data)

	output, err := formatter.Format(response)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Auth operations should output OK
	output = strings.TrimSpace(output)
	if output != "OK" {
		t.Errorf("Expected 'OK', got '%s'", output)
	}
}
