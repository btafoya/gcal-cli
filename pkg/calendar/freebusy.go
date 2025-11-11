package calendar

import (
	"context"
	"fmt"
	"time"

	"github.com/btafoya/gcal-cli/pkg/types"
	"google.golang.org/api/calendar/v3"
)

// FreeBusyPeriod represents a time period when a calendar is busy
type FreeBusyPeriod struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// FreeBusyInfo contains free/busy information for a calendar
type FreeBusyInfo struct {
	CalendarID string           `json:"calendarId"`
	Busy       []FreeBusyPeriod `json:"busy"`
	Errors     []string         `json:"errors,omitempty"`
}

// FreeBusyQueryRequest represents a free/busy query request
type FreeBusyQueryRequest struct {
	TimeMin     string   `json:"timeMin"`
	TimeMax     string   `json:"timeMax"`
	CalendarIDs []string `json:"calendarIds"`
	TimeZone    string   `json:"timeZone,omitempty"`
}

// FreeBusyQueryResponse contains the results of a free/busy query
type FreeBusyQueryResponse struct {
	Calendars map[string]FreeBusyInfo `json:"calendars"`
	TimeMin   string                  `json:"timeMin"`
	TimeMax   string                  `json:"timeMax"`
}

// QueryFreeBusy queries the free/busy status of one or more calendars
func (c *Client) QueryFreeBusy(ctx context.Context, request FreeBusyQueryRequest) (*FreeBusyQueryResponse, error) {
	if c.Service == nil {
		return nil, types.ErrAuthFailed("calendar service not initialized")
	}

	// Validate input
	if request.TimeMin == "" || request.TimeMax == "" {
		return nil, types.ErrInvalidInput("timeMin and timeMax", "both required for free/busy query")
	}

	if len(request.CalendarIDs) == 0 {
		return nil, types.ErrInvalidInput("calendarIds", "at least one calendar ID required")
	}

	// Parse times to validate
	timeMin, err := time.Parse(time.RFC3339, request.TimeMin)
	if err != nil {
		return nil, types.ErrInvalidInput("timeMin", fmt.Sprintf("invalid RFC3339 format: %v", err))
	}

	timeMax, err := time.Parse(time.RFC3339, request.TimeMax)
	if err != nil {
		return nil, types.ErrInvalidInput("timeMax", fmt.Sprintf("invalid RFC3339 format: %v", err))
	}

	if timeMax.Before(timeMin) {
		return nil, types.NewAppError(types.ErrCodeInvalidTimeRange,
			"end time must be after start time", true).
			WithDetails(fmt.Sprintf("start: %s, end: %s", request.TimeMin, request.TimeMax))
	}

	// Build calendar items for request
	items := make([]*calendar.FreeBusyRequestItem, len(request.CalendarIDs))
	for i, calID := range request.CalendarIDs {
		items[i] = &calendar.FreeBusyRequestItem{
			Id: calID,
		}
	}

	// Create free/busy request
	fbRequest := &calendar.FreeBusyRequest{
		TimeMin:  request.TimeMin,
		TimeMax:  request.TimeMax,
		TimeZone: request.TimeZone,
		Items:    items,
	}

	// Execute query
	fbResponse, err := c.Service.Freebusy.Query(fbRequest).Context(ctx).Do()
	if err != nil {
		return nil, types.ErrAPIError.WithDetails(fmt.Sprintf("free/busy query failed: %v", err))
	}

	// Convert response
	response := &FreeBusyQueryResponse{
		Calendars: make(map[string]FreeBusyInfo),
		TimeMin:   request.TimeMin,
		TimeMax:   request.TimeMax,
	}

	for calID, calInfo := range fbResponse.Calendars {
		info := FreeBusyInfo{
			CalendarID: calID,
			Busy:       make([]FreeBusyPeriod, 0),
		}

		// Add busy periods
		if calInfo.Busy != nil {
			for _, period := range calInfo.Busy {
				info.Busy = append(info.Busy, FreeBusyPeriod{
					Start: period.Start,
					End:   period.End,
				})
			}
		}

		// Add errors if any
		if calInfo.Errors != nil {
			for _, errItem := range calInfo.Errors {
				info.Errors = append(info.Errors, fmt.Sprintf("%s: %s", errItem.Domain, errItem.Reason))
			}
		}

		response.Calendars[calID] = info
	}

	return response, nil
}

// IsBusy checks if a specific calendar is busy during a given time period
func (c *Client) IsBusy(ctx context.Context, calendarID string, start, end time.Time) (bool, error) {
	request := FreeBusyQueryRequest{
		TimeMin:     start.Format(time.RFC3339),
		TimeMax:     end.Format(time.RFC3339),
		CalendarIDs: []string{calendarID},
	}

	response, err := c.QueryFreeBusy(ctx, request)
	if err != nil {
		return false, err
	}

	calInfo, ok := response.Calendars[calendarID]
	if !ok {
		return false, types.ErrNotFound("calendar", calendarID)
	}

	// If there are any busy periods, calendar is busy
	return len(calInfo.Busy) > 0, nil
}

// FindFreeSlots finds available time slots within a given time range
func (c *Client) FindFreeSlots(ctx context.Context, calendarID string, start, end time.Time, slotDuration time.Duration) ([]time.Time, error) {
	request := FreeBusyQueryRequest{
		TimeMin:     start.Format(time.RFC3339),
		TimeMax:     end.Format(time.RFC3339),
		CalendarIDs: []string{calendarID},
	}

	response, err := c.QueryFreeBusy(ctx, request)
	if err != nil {
		return nil, err
	}

	calInfo, ok := response.Calendars[calendarID]
	if !ok {
		return nil, types.ErrNotFound("calendar", calendarID)
	}

	// Convert busy periods to time intervals
	busyPeriods := make([]struct{ Start, End time.Time }, 0)
	for _, period := range calInfo.Busy {
		periodStart, _ := time.Parse(time.RFC3339, period.Start)
		periodEnd, _ := time.Parse(time.RFC3339, period.End)
		busyPeriods = append(busyPeriods, struct{ Start, End time.Time }{periodStart, periodEnd})
	}

	// Find free slots
	freeSlots := make([]time.Time, 0)
	current := start

	for current.Add(slotDuration).Before(end) || current.Add(slotDuration).Equal(end) {
		slotEnd := current.Add(slotDuration)
		isFree := true

		// Check if this slot conflicts with any busy period
		for _, busy := range busyPeriods {
			// Check for overlap
			if (current.Before(busy.End) && slotEnd.After(busy.Start)) {
				isFree = false
				break
			}
		}

		if isFree {
			freeSlots = append(freeSlots, current)
		}

		// Move to next potential slot (increment by slotDuration)
		current = current.Add(slotDuration)
	}

	return freeSlots, nil
}

// CheckConflicts checks if a proposed event conflicts with existing events
func (c *Client) CheckConflicts(ctx context.Context, calendarID string, start, end time.Time) (bool, []FreeBusyPeriod, error) {
	request := FreeBusyQueryRequest{
		TimeMin:     start.Format(time.RFC3339),
		TimeMax:     end.Format(time.RFC3339),
		CalendarIDs: []string{calendarID},
	}

	response, err := c.QueryFreeBusy(ctx, request)
	if err != nil {
		return false, nil, err
	}

	calInfo, ok := response.Calendars[calendarID]
	if !ok {
		return false, nil, types.ErrNotFound("calendar", calendarID)
	}

	// Check for conflicts
	conflicts := make([]FreeBusyPeriod, 0)
	for _, period := range calInfo.Busy {
		periodStart, _ := time.Parse(time.RFC3339, period.Start)
		periodEnd, _ := time.Parse(time.RFC3339, period.End)

		// Check for overlap
		if (start.Before(periodEnd) && end.After(periodStart)) {
			conflicts = append(conflicts, period)
		}
	}

	return len(conflicts) > 0, conflicts, nil
}
