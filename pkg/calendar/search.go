package calendar

import (
	"context"
	"strings"
	"time"

	"github.com/btafoya/gcal-cli/pkg/types"
)

// SearchFilter contains advanced search criteria
type SearchFilter struct {
	Query          string    // Text search query
	From           time.Time // Start of date range
	To             time.Time // End of date range
	Attendee       string    // Filter by attendee email
	Location       string    // Filter by location
	Status         string    // Filter by status (confirmed, tentative, cancelled)
	HasAttendees   *bool     // Filter by presence of attendees
	IsAllDay       *bool     // Filter by all-day events
	IsRecurring    *bool     // Filter by recurring events
	MaxResults     int64     // Maximum results to return
	OrderBy        string    // Sort order
}

// SearchEvents performs an advanced search for events
func (c *Client) SearchEvents(ctx context.Context, filter SearchFilter) ([]*types.Event, error) {
	// Validate filter
	if err := validateSearchFilter(filter); err != nil {
		return nil, err
	}

	// Get all events in the date range
	params := ListEventsParams{
		From:       filter.From,
		To:         filter.To,
		MaxResults: filter.MaxResults,
		Query:      filter.Query,
		OrderBy:    filter.OrderBy,
	}

	events, err := c.ListEvents(ctx, params)
	if err != nil {
		return nil, err
	}

	// Apply additional filters
	filtered := make([]*types.Event, 0, len(events))
	for _, event := range events {
		if matchesFilter(event, filter) {
			filtered = append(filtered, event)
		}
	}

	return filtered, nil
}

// SearchUpcoming searches for upcoming events
func (c *Client) SearchUpcoming(ctx context.Context, days int, query string) ([]*types.Event, error) {
	now := time.Now()
	future := now.AddDate(0, 0, days)

	filter := SearchFilter{
		From:       now,
		To:         future,
		Query:      query,
		MaxResults: 250,
		OrderBy:    "startTime",
	}

	return c.SearchEvents(ctx, filter)
}

// SearchByAttendee finds all events with a specific attendee
func (c *Client) SearchByAttendee(ctx context.Context, email string, from, to time.Time) ([]*types.Event, error) {
	if !isValidEmail(email) {
		return nil, types.ErrInvalidInput("attendee", "invalid email address")
	}

	filter := SearchFilter{
		From:       from,
		To:         to,
		Attendee:   email,
		MaxResults: 250,
		OrderBy:    "startTime",
	}

	return c.SearchEvents(ctx, filter)
}

// SearchByLocation finds all events at a specific location
func (c *Client) SearchByLocation(ctx context.Context, location string, from, to time.Time) ([]*types.Event, error) {
	filter := SearchFilter{
		From:       from,
		To:         to,
		Location:   location,
		MaxResults: 250,
		OrderBy:    "startTime",
	}

	return c.SearchEvents(ctx, filter)
}

// SearchRecurring finds all recurring events
func (c *Client) SearchRecurring(ctx context.Context, from, to time.Time) ([]*types.Event, error) {
	isRecurring := true
	filter := SearchFilter{
		From:        from,
		To:          to,
		IsRecurring: &isRecurring,
		MaxResults:  250,
		OrderBy:     "startTime",
	}

	return c.SearchEvents(ctx, filter)
}

// matchesFilter checks if an event matches the search filter
func matchesFilter(event *types.Event, filter SearchFilter) bool {
	// Filter by attendee
	if filter.Attendee != "" {
		found := false
		for _, att := range event.Attendees {
			if strings.EqualFold(att.Email, filter.Attendee) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by location
	if filter.Location != "" {
		if !strings.Contains(strings.ToLower(event.Location), strings.ToLower(filter.Location)) {
			return false
		}
	}

	// Filter by status
	if filter.Status != "" {
		if !strings.EqualFold(event.Status, filter.Status) {
			return false
		}
	}

	// Filter by presence of attendees
	if filter.HasAttendees != nil {
		hasAttendees := len(event.Attendees) > 0
		if hasAttendees != *filter.HasAttendees {
			return false
		}
	}

	// Filter by all-day events
	if filter.IsAllDay != nil {
		isAllDay := event.Start.Date != ""
		if isAllDay != *filter.IsAllDay {
			return false
		}
	}

	// Filter by recurring events
	if filter.IsRecurring != nil {
		isRecurring := len(event.Recurrence) > 0
		if isRecurring != *filter.IsRecurring {
			return false
		}
	}

	return true
}

// validateSearchFilter validates search filter parameters
func validateSearchFilter(filter SearchFilter) error {
	if filter.From.IsZero() {
		return types.ErrMissingRequired("from")
	}

	if filter.To.IsZero() {
		return types.ErrMissingRequired("to")
	}

	if !filter.From.Before(filter.To) {
		return types.NewAppError(types.ErrCodeInvalidTimeRange,
			"from date must be before to date", true).
			WithDetails("adjust the date range")
	}

	if filter.Attendee != "" && !isValidEmail(filter.Attendee) {
		return types.ErrInvalidInput("attendee", "invalid email address")
	}

	if filter.Status != "" {
		validStatuses := map[string]bool{
			"confirmed":  true,
			"tentative":  true,
			"cancelled":  true,
		}
		if !validStatuses[strings.ToLower(filter.Status)] {
			return types.ErrInvalidInput("status",
				"must be 'confirmed', 'tentative', or 'cancelled'")
		}
	}

	if filter.OrderBy != "" && filter.OrderBy != "startTime" && filter.OrderBy != "updated" {
		return types.ErrInvalidInput("order-by",
			"must be 'startTime' or 'updated'")
	}

	return nil
}
