package calendar

import (
	"context"
	"testing"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	"github.com/btafoya/gcal-cli/pkg/types"
)

// Test helpers

func createTestEvent() *calendar.Event {
	return &calendar.Event{
		Id:          "test-event-id",
		Summary:     "Test Event",
		Description: "Test Description",
		Location:    "Test Location",
		Status:      "confirmed",
		Start: &calendar.EventDateTime{
			DateTime: "2024-01-15T10:00:00-05:00",
			TimeZone: "America/New_York",
		},
		End: &calendar.EventDateTime{
			DateTime: "2024-01-15T11:00:00-05:00",
			TimeZone: "America/New_York",
		},
		Attendees: []*calendar.EventAttendee{
			{
				Email:          "user@example.com",
				ResponseStatus: "needsAction",
			},
		},
	}
}

// TestConvertEvent tests event conversion from Google Calendar to our type
func TestConvertEvent(t *testing.T) {
	gcalEvent := createTestEvent()
	event := convertEvent(gcalEvent)

	if event.ID != gcalEvent.Id {
		t.Errorf("Expected ID %s, got %s", gcalEvent.Id, event.ID)
	}

	if event.Summary != gcalEvent.Summary {
		t.Errorf("Expected summary %s, got %s", gcalEvent.Summary, event.Summary)
	}

	if event.Description != gcalEvent.Description {
		t.Errorf("Expected description %s, got %s", gcalEvent.Description, event.Description)
	}

	if event.Location != gcalEvent.Location {
		t.Errorf("Expected location %s, got %s", gcalEvent.Location, event.Location)
	}

	if event.Status != gcalEvent.Status {
		t.Errorf("Expected status %s, got %s", gcalEvent.Status, event.Status)
	}

	if event.Start.DateTime != gcalEvent.Start.DateTime {
		t.Errorf("Expected start time %s, got %s", gcalEvent.Start.DateTime, event.Start.DateTime)
	}

	if event.End.DateTime != gcalEvent.End.DateTime {
		t.Errorf("Expected end time %s, got %s", gcalEvent.End.DateTime, event.End.DateTime)
	}

	if len(event.Attendees) != len(gcalEvent.Attendees) {
		t.Errorf("Expected %d attendees, got %d", len(gcalEvent.Attendees), len(event.Attendees))
	}

	if event.Attendees[0].Email != gcalEvent.Attendees[0].Email {
		t.Errorf("Expected attendee email %s, got %s",
			gcalEvent.Attendees[0].Email, event.Attendees[0].Email)
	}
}

// TestConvertEvent_Nil tests nil event conversion
func TestConvertEvent_Nil(t *testing.T) {
	event := convertEvent(nil)
	if event != nil {
		t.Error("Expected nil event, got non-nil")
	}
}

// TestConvertEvent_AllDayEvent tests all-day event conversion
func TestConvertEvent_AllDayEvent(t *testing.T) {
	gcalEvent := &calendar.Event{
		Id:      "all-day-event",
		Summary: "All Day Event",
		Start: &calendar.EventDateTime{
			Date: "2024-01-15",
		},
		End: &calendar.EventDateTime{
			Date: "2024-01-16",
		},
	}

	event := convertEvent(gcalEvent)

	if event.Start.Date != "2024-01-15" {
		t.Errorf("Expected start date 2024-01-15, got %s", event.Start.Date)
	}

	if event.End.Date != "2024-01-16" {
		t.Errorf("Expected end date 2024-01-16, got %s", event.End.Date)
	}
}

// TestValidateCreateParams tests event creation parameter validation
func TestValidateCreateParams(t *testing.T) {
	tests := []struct {
		name    string
		params  CreateEventParams
		wantErr bool
	}{
		{
			name: "valid params",
			params: CreateEventParams{
				Summary: "Test Event",
				Start:   time.Now(),
				End:     time.Now().Add(1 * time.Hour),
			},
			wantErr: false,
		},
		{
			name: "missing summary",
			params: CreateEventParams{
				Summary: "",
				Start:   time.Now(),
				End:     time.Now().Add(1 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "missing start",
			params: CreateEventParams{
				Summary: "Test Event",
				Start:   time.Time{},
				End:     time.Now().Add(1 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "missing end",
			params: CreateEventParams{
				Summary: "Test Event",
				Start:   time.Now(),
				End:     time.Time{},
			},
			wantErr: true,
		},
		{
			name: "end before start",
			params: CreateEventParams{
				Summary: "Test Event",
				Start:   time.Now(),
				End:     time.Now().Add(-1 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "invalid email",
			params: CreateEventParams{
				Summary:   "Test Event",
				Start:     time.Now(),
				End:       time.Now().Add(1 * time.Hour),
				Attendees: []string{"invalid-email"},
			},
			wantErr: true,
		},
		{
			name: "valid email",
			params: CreateEventParams{
				Summary:   "Test Event",
				Start:     time.Now(),
				End:       time.Now().Add(1 * time.Hour),
				Attendees: []string{"user@example.com"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreateParams(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateCreateParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateListParams tests event listing parameter validation
func TestValidateListParams(t *testing.T) {
	now := time.Now()
	later := now.Add(24 * time.Hour)

	tests := []struct {
		name    string
		params  ListEventsParams
		wantErr bool
	}{
		{
			name: "valid params",
			params: ListEventsParams{
				From: now,
				To:   later,
			},
			wantErr: false,
		},
		{
			name: "missing from",
			params: ListEventsParams{
				From: time.Time{},
				To:   later,
			},
			wantErr: true,
		},
		{
			name: "missing to",
			params: ListEventsParams{
				From: now,
				To:   time.Time{},
			},
			wantErr: true,
		},
		{
			name: "to before from",
			params: ListEventsParams{
				From: later,
				To:   now,
			},
			wantErr: true,
		},
		{
			name: "negative max results",
			params: ListEventsParams{
				From:       now,
				To:         later,
				MaxResults: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid order by",
			params: ListEventsParams{
				From:    now,
				To:      later,
				OrderBy: "invalid",
			},
			wantErr: true,
		},
		{
			name: "valid order by startTime",
			params: ListEventsParams{
				From:    now,
				To:      later,
				OrderBy: "startTime",
			},
			wantErr: false,
		},
		{
			name: "valid order by updated",
			params: ListEventsParams{
				From:    now,
				To:      later,
				OrderBy: "updated",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateListParams(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateListParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestIsValidEmail tests email validation
func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"user.name@example.com", true},
		{"user+tag@example.co.uk", true},
		{"invalid", false},
		{"@example.com", false},
		{"user@", false},
		{"user", false},
		{"", false},
		{"user@domain", false}, // No dot in domain
		{"user@@example.com", false},
		{"user@example", false}, // No dot in domain
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			result := isValidEmail(tt.email)
			if result != tt.valid {
				t.Errorf("isValidEmail(%s) = %v, want %v", tt.email, result, tt.valid)
			}
		})
	}
}

// TestIsRetryable tests retry logic decision
func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		retryable bool
	}{
		{
			name:       "rate limit",
			err:        &googleapi.Error{Code: 429},
			retryable: true,
		},
		{
			name:       "server error 500",
			err:        &googleapi.Error{Code: 500},
			retryable: true,
		},
		{
			name:       "server error 502",
			err:        &googleapi.Error{Code: 502},
			retryable: true,
		},
		{
			name:       "server error 503",
			err:        &googleapi.Error{Code: 503},
			retryable: true,
		},
		{
			name:       "server error 504",
			err:        &googleapi.Error{Code: 504},
			retryable: true,
		},
		{
			name:       "bad request",
			err:        &googleapi.Error{Code: 400},
			retryable: false,
		},
		{
			name:       "unauthorized",
			err:        &googleapi.Error{Code: 401},
			retryable: false,
		},
		{
			name:       "not found",
			err:        &googleapi.Error{Code: 404},
			retryable: false,
		},
		{
			name:       "non-api error",
			err:        context.Canceled,
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryable(tt.err)
			if result != tt.retryable {
				t.Errorf("isRetryable() = %v, want %v", result, tt.retryable)
			}
		})
	}
}

// TestHandleAPIError tests API error conversion
func TestHandleAPIError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		expectedCode  string
	}{
		{
			name:         "bad request",
			err:          &googleapi.Error{Code: 400, Message: "Bad request"},
			expectedCode: types.ErrCodeInvalidInput,
		},
		{
			name:         "unauthorized",
			err:          &googleapi.Error{Code: 401, Message: "Unauthorized"},
			expectedCode: types.ErrCodeAuthFailed,
		},
		{
			name:         "forbidden",
			err:          &googleapi.Error{Code: 403, Message: "Forbidden"},
			expectedCode: types.ErrCodePermissionDenied,
		},
		{
			name:         "not found",
			err:          &googleapi.Error{Code: 404, Message: "Not found"},
			expectedCode: types.ErrCodeNotFound,
		},
		{
			name:         "conflict",
			err:          &googleapi.Error{Code: 409, Message: "Conflict"},
			expectedCode: types.ErrCodeInvalidInput,
		},
		{
			name:         "rate limit",
			err:          &googleapi.Error{Code: 429, Message: "Too many requests"},
			expectedCode: types.ErrCodeRateLimit,
		},
		{
			name:         "server error",
			err:          &googleapi.Error{Code: 500, Message: "Internal server error"},
			expectedCode: types.ErrCodeAPIError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := handleAPIError(tt.err, "test operation")
			if appErr == nil {
				t.Fatal("Expected error, got nil")
			}

			typedErr, ok := appErr.(*types.AppError)
			if !ok {
				t.Fatalf("Expected *types.AppError, got %T", appErr)
			}

			if typedErr.Code != tt.expectedCode {
				t.Errorf("Expected error code %s, got %s", tt.expectedCode, typedErr.Code)
			}
		})
	}
}

// TestHandleAPIError_Nil tests nil error handling
func TestHandleAPIError_Nil(t *testing.T) {
	err := handleAPIError(nil, "test operation")
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}
