package calendar

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"
	"github.com/btafoya/gcal-cli/pkg/types"
)

// CreateEventParams contains parameters for creating an event
type CreateEventParams struct {
	Summary     string
	Description string
	Location    string
	Start       time.Time
	End         time.Time
	TimeZone    string
	Attendees   []string
	Recurrence  []string
	AllDay      bool
}

// ListEventsParams contains parameters for listing events
type ListEventsParams struct {
	From       time.Time
	To         time.Time
	MaxResults int64
	Query      string
	OrderBy    string
}

// CreateEvent creates a new calendar event
func (c *Client) CreateEvent(ctx context.Context, params CreateEventParams) (*types.Event, error) {
	// Validate parameters
	if err := validateCreateParams(params); err != nil {
		return nil, err
	}

	// Build Google Calendar event
	event := &calendar.Event{
		Summary:     params.Summary,
		Description: params.Description,
		Location:    params.Location,
	}

	// Set start and end times
	if params.AllDay {
		event.Start = &calendar.EventDateTime{
			Date: params.Start.Format("2006-01-02"),
		}
		event.End = &calendar.EventDateTime{
			Date: params.End.Format("2006-01-02"),
		}
	} else {
		tz := params.TimeZone
		if tz == "" {
			tz = "UTC"
		}
		event.Start = &calendar.EventDateTime{
			DateTime: params.Start.Format(time.RFC3339),
			TimeZone: tz,
		}
		event.End = &calendar.EventDateTime{
			DateTime: params.End.Format(time.RFC3339),
			TimeZone: tz,
		}
	}

	// Add attendees
	if len(params.Attendees) > 0 {
		event.Attendees = make([]*calendar.EventAttendee, len(params.Attendees))
		for i, email := range params.Attendees {
			event.Attendees[i] = &calendar.EventAttendee{
				Email: email,
			}
		}
	}

	// Add recurrence rules
	if len(params.Recurrence) > 0 {
		event.Recurrence = params.Recurrence
	}

	// Create event with retry logic
	var created *calendar.Event
	err := c.withRetry(ctx, "create event", func() error {
		var err error
		created, err = c.Service.Events.Insert(c.CalendarID, event).
			Context(ctx).
			Do()
		return err
	})

	if err != nil {
		return nil, handleAPIError(err, "create event")
	}

	return convertEvent(created), nil
}

// ListEvents lists events in a date range
func (c *Client) ListEvents(ctx context.Context, params ListEventsParams) ([]*types.Event, error) {
	// Validate parameters
	if err := validateListParams(params); err != nil {
		return nil, err
	}

	// Set default max results
	if params.MaxResults == 0 {
		params.MaxResults = 250
	}

	// Build list request
	call := c.Service.Events.List(c.CalendarID).
		Context(ctx).
		TimeMin(params.From.Format(time.RFC3339)).
		TimeMax(params.To.Format(time.RFC3339)).
		MaxResults(params.MaxResults).
		SingleEvents(true)

	if params.Query != "" {
		call = call.Q(params.Query)
	}

	if params.OrderBy != "" {
		call = call.OrderBy(params.OrderBy)
	} else {
		call = call.OrderBy("startTime")
	}

	// Execute with retry logic
	var eventsList *calendar.Events
	err := c.withRetry(ctx, "list events", func() error {
		var err error
		eventsList, err = call.Do()
		return err
	})

	if err != nil {
		return nil, handleAPIError(err, "list events")
	}

	// Convert events
	events := make([]*types.Event, len(eventsList.Items))
	for i, item := range eventsList.Items {
		events[i] = convertEvent(item)
	}

	return events, nil
}

// GetEvent retrieves a single event by ID
func (c *Client) GetEvent(ctx context.Context, eventID string) (*types.Event, error) {
	if eventID == "" {
		return nil, types.ErrMissingRequired("event-id")
	}

	var event *calendar.Event
	err := c.withRetry(ctx, "get event", func() error {
		var err error
		event, err = c.Service.Events.Get(c.CalendarID, eventID).
			Context(ctx).
			Do()
		return err
	})

	if err != nil {
		return nil, handleAPIError(err, "get event")
	}

	return convertEvent(event), nil
}

// UpdateEvent updates an existing event
func (c *Client) UpdateEvent(ctx context.Context, eventID string, params CreateEventParams) (*types.Event, error) {
	if eventID == "" {
		return nil, types.ErrMissingRequired("event-id")
	}

	// Get existing event first
	existing, err := c.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Build updated event (merge with existing)
	event := &calendar.Event{
		Summary:     params.Summary,
		Description: params.Description,
		Location:    params.Location,
	}

	// Use existing values if not provided
	if params.Summary == "" {
		event.Summary = existing.Summary
	}
	if params.Description == "" {
		event.Description = existing.Description
	}
	if params.Location == "" {
		event.Location = existing.Location
	}

	// Set start and end times
	if !params.Start.IsZero() && !params.End.IsZero() {
		if params.AllDay {
			event.Start = &calendar.EventDateTime{
				Date: params.Start.Format("2006-01-02"),
			}
			event.End = &calendar.EventDateTime{
				Date: params.End.Format("2006-01-02"),
			}
		} else {
			tz := params.TimeZone
			if tz == "" {
				tz = existing.Start.TimeZone
			}
			event.Start = &calendar.EventDateTime{
				DateTime: params.Start.Format(time.RFC3339),
				TimeZone: tz,
			}
			event.End = &calendar.EventDateTime{
				DateTime: params.End.Format(time.RFC3339),
				TimeZone: tz,
			}
		}
	} else {
		// Keep existing times
		event.Start = &calendar.EventDateTime{
			DateTime: existing.Start.DateTime,
			TimeZone: existing.Start.TimeZone,
			Date:     existing.Start.Date,
		}
		event.End = &calendar.EventDateTime{
			DateTime: existing.End.DateTime,
			TimeZone: existing.End.TimeZone,
			Date:     existing.End.Date,
		}
	}

	// Update attendees if provided
	if len(params.Attendees) > 0 {
		event.Attendees = make([]*calendar.EventAttendee, len(params.Attendees))
		for i, email := range params.Attendees {
			event.Attendees[i] = &calendar.EventAttendee{
				Email: email,
			}
		}
	} else {
		// Keep existing attendees
		event.Attendees = make([]*calendar.EventAttendee, len(existing.Attendees))
		for i, att := range existing.Attendees {
			event.Attendees[i] = &calendar.EventAttendee{
				Email:          att.Email,
				ResponseStatus: att.ResponseStatus,
			}
		}
	}

	// Update recurrence if provided
	if len(params.Recurrence) > 0 {
		event.Recurrence = params.Recurrence
	} else if len(existing.Recurrence) > 0 {
		event.Recurrence = existing.Recurrence
	}

	// Update with retry logic
	var updated *calendar.Event
	err = c.withRetry(ctx, "update event", func() error {
		var err error
		updated, err = c.Service.Events.Update(c.CalendarID, eventID, event).
			Context(ctx).
			Do()
		return err
	})

	if err != nil {
		return nil, handleAPIError(err, "update event")
	}

	return convertEvent(updated), nil
}

// DeleteEvent deletes an event
func (c *Client) DeleteEvent(ctx context.Context, eventID string) error {
	if eventID == "" {
		return types.ErrMissingRequired("event-id")
	}

	err := c.withRetry(ctx, "delete event", func() error {
		return c.Service.Events.Delete(c.CalendarID, eventID).
			Context(ctx).
			Do()
	})

	if err != nil {
		return handleAPIError(err, "delete event")
	}

	return nil
}

// convertEvent converts Google Calendar event to our Event type
func convertEvent(event *calendar.Event) *types.Event {
	if event == nil {
		return nil
	}

	result := &types.Event{
		ID:          event.Id,
		Summary:     event.Summary,
		Description: event.Description,
		Location:    event.Location,
		Status:      event.Status,
	}

	// Convert start time
	if event.Start != nil {
		result.Start = types.EventTime{
			DateTime: event.Start.DateTime,
			Date:     event.Start.Date,
			TimeZone: event.Start.TimeZone,
		}
	}

	// Convert end time
	if event.End != nil {
		result.End = types.EventTime{
			DateTime: event.End.DateTime,
			Date:     event.End.Date,
			TimeZone: event.End.TimeZone,
		}
	}

	// Convert attendees
	if len(event.Attendees) > 0 {
		result.Attendees = make([]types.Attendee, len(event.Attendees))
		for i, att := range event.Attendees {
			result.Attendees[i] = types.Attendee{
				Email:          att.Email,
				ResponseStatus: att.ResponseStatus,
			}
		}
	}

	// Copy recurrence rules
	if len(event.Recurrence) > 0 {
		result.Recurrence = event.Recurrence
	}

	return result
}

// validateCreateParams validates event creation parameters
func validateCreateParams(params CreateEventParams) error {
	if params.Summary == "" {
		return types.ErrMissingRequired("title")
	}

	if params.Start.IsZero() {
		return types.ErrMissingRequired("start")
	}

	if params.End.IsZero() {
		return types.ErrMissingRequired("end")
	}

	if !params.Start.Before(params.End) {
		return types.NewAppError(types.ErrCodeInvalidTimeRange,
			"start time must be before end time", true).
			WithDetails(fmt.Sprintf("start: %s, end: %s",
				params.Start.Format(time.RFC3339),
				params.End.Format(time.RFC3339))).
			WithSuggestedAction("Adjust the start and end times")
	}

	// Validate attendee emails
	for _, email := range params.Attendees {
		if !isValidEmail(email) {
			return types.ErrInvalidInput("attendees",
				fmt.Sprintf("invalid email address: %s", email))
		}
	}

	return nil
}

// validateListParams validates event listing parameters
func validateListParams(params ListEventsParams) error {
	if params.From.IsZero() {
		return types.ErrMissingRequired("from")
	}

	if params.To.IsZero() {
		return types.ErrMissingRequired("to")
	}

	if !params.From.Before(params.To) {
		return types.NewAppError(types.ErrCodeInvalidTimeRange,
			"from date must be before to date", true).
			WithDetails(fmt.Sprintf("from: %s, to: %s",
				params.From.Format(time.RFC3339),
				params.To.Format(time.RFC3339))).
			WithSuggestedAction("Adjust the date range")
	}

	if params.MaxResults < 0 {
		return types.ErrInvalidInput("max-results",
			"must be a positive number")
	}

	if params.OrderBy != "" && params.OrderBy != "startTime" && params.OrderBy != "updated" {
		return types.ErrInvalidInput("order-by",
			"must be 'startTime' or 'updated'")
	}

	return nil
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	// Basic validation - check for @ and domain
	if len(email) < 3 {
		return false
	}

	atIndex := -1
	for i, c := range email {
		if c == '@' {
			if atIndex != -1 {
				return false // Multiple @ signs
			}
			atIndex = i
		}
	}

	if atIndex == -1 || atIndex == 0 || atIndex == len(email)-1 {
		return false
	}

	// Check for domain part after @
	domain := email[atIndex+1:]
	if len(domain) < 3 {
		return false
	}

	// Check for at least one dot in domain
	hasDot := false
	for _, c := range domain {
		if c == '.' {
			hasDot = true
			break
		}
	}

	return hasDot
}
