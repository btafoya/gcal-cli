package calendar

import (
	"context"

	"google.golang.org/api/calendar/v3"
)

// CalendarInfo represents calendar metadata
type CalendarInfo struct {
	ID          string `json:"id"`
	Summary     string `json:"summary"`
	Description string `json:"description,omitempty"`
	TimeZone    string `json:"timeZone"`
	Primary     bool   `json:"primary,omitempty"`
	AccessRole  string `json:"accessRole"`
}

// ListCalendars retrieves all calendars accessible to the user
func (c *Client) ListCalendars(ctx context.Context) ([]*CalendarInfo, error) {
	var calendarList *calendar.CalendarList
	err := c.withRetry(ctx, "list calendars", func() error {
		var err error
		calendarList, err = c.Service.CalendarList.List().
			Context(ctx).
			Do()
		return err
	})

	if err != nil {
		return nil, handleAPIError(err, "list calendars")
	}

	// Convert to our type
	calendars := make([]*CalendarInfo, len(calendarList.Items))
	for i, item := range calendarList.Items {
		calendars[i] = &CalendarInfo{
			ID:          item.Id,
			Summary:     item.Summary,
			Description: item.Description,
			TimeZone:    item.TimeZone,
			Primary:     item.Primary,
			AccessRole:  item.AccessRole,
		}
	}

	return calendars, nil
}

// GetCalendar retrieves metadata for a specific calendar
func (c *Client) GetCalendar(ctx context.Context, calendarID string) (*CalendarInfo, error) {
	if calendarID == "" {
		calendarID = c.CalendarID
	}

	var cal *calendar.Calendar
	err := c.withRetry(ctx, "get calendar", func() error {
		var err error
		cal, err = c.Service.Calendars.Get(calendarID).
			Context(ctx).
			Do()
		return err
	})

	if err != nil {
		return nil, handleAPIError(err, "get calendar")
	}

	return &CalendarInfo{
		ID:          cal.Id,
		Summary:     cal.Summary,
		Description: cal.Description,
		TimeZone:    cal.TimeZone,
	}, nil
}

// GetPrimaryCalendar retrieves the user's primary calendar
func (c *Client) GetPrimaryCalendar(ctx context.Context) (*CalendarInfo, error) {
	calendars, err := c.ListCalendars(ctx)
	if err != nil {
		return nil, err
	}

	for _, cal := range calendars {
		if cal.Primary {
			return cal, nil
		}
	}

	// Fallback to "primary" calendar ID
	return c.GetCalendar(ctx, "primary")
}
