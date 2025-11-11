package calendar

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/btafoya/gcal-cli/pkg/types"
	"google.golang.org/api/calendar/v3"
)

// MultiCalendarEvent represents an event with its source calendar
type MultiCalendarEvent struct {
	CalendarID string       `json:"calendarId"`
	Event      *types.Event `json:"event"`
}

// MultiCalendarListResult contains events from multiple calendars
type MultiCalendarListResult struct {
	Events     []MultiCalendarEvent `json:"events"`
	TotalCount int                  `json:"totalCount"`
	ByCalendar map[string]int       `json:"byCalendar"` // Count per calendar
}

// ListEventsMultiCalendar lists events from multiple calendars in parallel
func (c *Client) ListEventsMultiCalendar(ctx context.Context, calendarIDs []string, timeMin, timeMax time.Time, maxResults int) (*MultiCalendarListResult, error) {
	if c.Service == nil {
		return nil, types.ErrAuthFailed("calendar service not initialized")
	}

	if len(calendarIDs) == 0 {
		return nil, types.ErrInvalidInput("calendarIds", "at least one calendar ID required")
	}

	result := &MultiCalendarListResult{
		Events:     make([]MultiCalendarEvent, 0),
		ByCalendar: make(map[string]int),
	}

	// Use WaitGroup for concurrent calendar queries
	var wg sync.WaitGroup
	var mu sync.Mutex
	errorsChan := make(chan error, len(calendarIDs))

	for _, calID := range calendarIDs {
		wg.Add(1)
		go func(calendarID string) {
			defer wg.Done()

			// List events for this calendar
			call := c.Service.Events.List(calendarID).
				TimeMin(timeMin.Format(time.RFC3339)).
				TimeMax(timeMax.Format(time.RFC3339)).
				SingleEvents(true).
				OrderBy("startTime")

			if maxResults > 0 {
				call = call.MaxResults(int64(maxResults))
			}

			events, err := call.Context(ctx).Do()
			if err != nil {
				errorsChan <- fmt.Errorf("calendar %s: %w", calendarID, err)
				return
			}

			// Add events to result
			mu.Lock()
			for _, event := range events.Items {
				result.Events = append(result.Events, MultiCalendarEvent{
					CalendarID: calendarID,
					Event:      convertEvent(event),
				})
			}
			result.ByCalendar[calendarID] = len(events.Items)
			mu.Unlock()
		}(calID)
	}

	wg.Wait()
	close(errorsChan)

	// Check for errors
	if len(errorsChan) > 0 {
		err := <-errorsChan
		return nil, types.ErrAPIError.WithDetails(fmt.Sprintf("multi-calendar query failed: %v", err))
	}

	result.TotalCount = len(result.Events)

	// Sort events by start time
	sortMultiCalendarEvents(result.Events)

	return result, nil
}

// CreateEventMultiCalendar creates the same event in multiple calendars
func (c *Client) CreateEventMultiCalendar(ctx context.Context, calendarIDs []string, event *calendar.Event) (map[string]*types.Event, error) {
	if c.Service == nil {
		return nil, types.ErrAuthFailed("calendar service not initialized")
	}

	if len(calendarIDs) == 0 {
		return nil, types.ErrInvalidInput("calendarIds", "at least one calendar ID required")
	}

	results := make(map[string]*types.Event)
	var wg sync.WaitGroup
	var mu sync.Mutex
	errorsChan := make(chan error, len(calendarIDs))

	for _, calID := range calendarIDs {
		wg.Add(1)
		go func(calendarID string) {
			defer wg.Done()

			createdEvent, err := c.Service.Events.Insert(calendarID, event).
				Context(ctx).
				Do()
			if err != nil {
				errorsChan <- fmt.Errorf("calendar %s: %w", calendarID, err)
				return
			}

			mu.Lock()
			results[calendarID] = convertEvent(createdEvent)
			mu.Unlock()
		}(calID)
	}

	wg.Wait()
	close(errorsChan)

	// Check for errors
	if len(errorsChan) > 0 {
		err := <-errorsChan
		return nil, types.ErrAPIError.WithDetails(fmt.Sprintf("multi-calendar create failed: %v", err))
	}

	return results, nil
}

// FindCommonFreeTime finds time slots when all calendars are free
func (c *Client) FindCommonFreeTime(ctx context.Context, calendarIDs []string, start, end time.Time, slotDuration time.Duration) ([]time.Time, error) {
	if c.Service == nil {
		return nil, types.ErrAuthFailed("calendar service not initialized")
	}

	if len(calendarIDs) == 0 {
		return nil, types.ErrInvalidInput("calendarIds", "at least one calendar ID required")
	}

	// Query free/busy for all calendars
	request := FreeBusyQueryRequest{
		TimeMin:     start.Format(time.RFC3339),
		TimeMax:     end.Format(time.RFC3339),
		CalendarIDs: calendarIDs,
	}

	response, err := c.QueryFreeBusy(ctx, request)
	if err != nil {
		return nil, err
	}

	// Collect all busy periods from all calendars
	allBusyPeriods := make([]struct{ Start, End time.Time }, 0)

	for _, calInfo := range response.Calendars {
		for _, period := range calInfo.Busy {
			periodStart, _ := time.Parse(time.RFC3339, period.Start)
			periodEnd, _ := time.Parse(time.RFC3339, period.End)
			allBusyPeriods = append(allBusyPeriods, struct{ Start, End time.Time }{periodStart, periodEnd})
		}
	}

	// Find slots where NO calendar is busy
	commonFreeSlots := make([]time.Time, 0)
	current := start

	for current.Add(slotDuration).Before(end) || current.Add(slotDuration).Equal(end) {
		slotEnd := current.Add(slotDuration)
		isFree := true

		// Check if this slot conflicts with ANY busy period
		for _, busy := range allBusyPeriods {
			if (current.Before(busy.End) && slotEnd.After(busy.Start)) {
				isFree = false
				break
			}
		}

		if isFree {
			commonFreeSlots = append(commonFreeSlots, current)
		}

		current = current.Add(slotDuration)
	}

	return commonFreeSlots, nil
}

// SyncEventAcrossCalendars synchronizes an event across multiple calendars
func (c *Client) SyncEventAcrossCalendars(ctx context.Context, sourceCalendarID, eventID string, targetCalendarIDs []string) (map[string]*types.Event, error) {
	// Get the source event
	sourceEvent, err := c.Service.Events.Get(sourceCalendarID, eventID).Context(ctx).Do()
	if err != nil {
		return nil, types.ErrAPIError.WithDetails(fmt.Sprintf("failed to get source event: %v", err))
	}

	// Create copies in target calendars
	return c.CreateEventMultiCalendar(ctx, targetCalendarIDs, sourceEvent)
}

// sortMultiCalendarEvents sorts events by start time
func sortMultiCalendarEvents(events []MultiCalendarEvent) {
	// Simple bubble sort for now - can be optimized with sort.Slice
	for i := 0; i < len(events); i++ {
		for j := i + 1; j < len(events); j++ {
			// Compare start times
			startI := events[i].Event.Start.DateTime
			if events[i].Event.Start.Date != "" {
				startI = events[i].Event.Start.Date
			}
			startJ := events[j].Event.Start.DateTime
			if events[j].Event.Start.Date != "" {
				startJ = events[j].Event.Start.Date
			}

			if startI > startJ {
				events[i], events[j] = events[j], events[i]
			}
		}
	}
}

// GetCalendarPermissions retrieves the access control list for a calendar
func (c *Client) GetCalendarPermissions(ctx context.Context, calendarID string) ([]*calendar.AclRule, error) {
	if c.Service == nil {
		return nil, types.ErrAuthFailed("calendar service not initialized")
	}

	acl, err := c.Service.Acl.List(calendarID).Context(ctx).Do()
	if err != nil {
		return nil, types.ErrAPIError.WithDetails(fmt.Sprintf("failed to get calendar permissions: %v", err))
	}

	return acl.Items, nil
}

// ShareCalendar adds a user to a calendar's access control list
func (c *Client) ShareCalendar(ctx context.Context, calendarID, email, role string) error {
	if c.Service == nil {
		return types.ErrAuthFailed("calendar service not initialized")
	}

	// Validate role
	validRoles := map[string]bool{
		"owner":      true,
		"writer":     true,
		"reader":     true,
		"freeBusyReader": true,
	}

	if !validRoles[role] {
		return types.ErrInvalidInput("role", "must be one of: owner, writer, reader, freeBusyReader")
	}

	rule := &calendar.AclRule{
		Role: role,
		Scope: &calendar.AclRuleScope{
			Type:  "user",
			Value: email,
		},
	}

	_, err := c.Service.Acl.Insert(calendarID, rule).Context(ctx).Do()
	if err != nil {
		return types.ErrAPIError.WithDetails(fmt.Sprintf("failed to share calendar: %v", err))
	}

	return nil
}

// UnshareCalendar removes a user from a calendar's access control list
func (c *Client) UnshareCalendar(ctx context.Context, calendarID, ruleID string) error {
	if c.Service == nil {
		return types.ErrAuthFailed("calendar service not initialized")
	}

	err := c.Service.Acl.Delete(calendarID, ruleID).Context(ctx).Do()
	if err != nil {
		return types.ErrAPIError.WithDetails(fmt.Sprintf("failed to unshare calendar: %v", err))
	}

	return nil
}
