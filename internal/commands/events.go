package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/btafoya/gcal-cli/pkg/auth"
	"github.com/btafoya/gcal-cli/pkg/calendar"
	"github.com/btafoya/gcal-cli/pkg/config"
	"github.com/btafoya/gcal-cli/pkg/examples"
	"github.com/btafoya/gcal-cli/pkg/output"
	"github.com/btafoya/gcal-cli/pkg/types"
	"github.com/spf13/cobra"
)

// NewEventsCommand creates the events command group
func NewEventsCommand(formatter output.Formatter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "events",
		Short: "Manage calendar events",
		Long:  "Create, list, retrieve, update, and delete Google Calendar events",
	}

	cmd.AddCommand(newEventsCreateCommand(formatter))
	cmd.AddCommand(newEventsListCommand(formatter))
	cmd.AddCommand(newEventsGetCommand(formatter))
	cmd.AddCommand(newEventsUpdateCommand(formatter))
	cmd.AddCommand(newEventsDeleteCommand(formatter))

	return cmd
}

func newEventsCreateCommand(formatter output.Formatter) *cobra.Command {
	var (
		title       string
		description string
		location    string
		start       string
		end         string
		attendees   string
		recurrence  string
		allDay      bool
	)

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a new calendar event",
		Long:    "Create a new event in your Google Calendar with support for attendees, recurrence, and all-day events",
		Example: examples.EventsCreateExamples,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Get calendar client
			client, err := getCalendarClient(ctx)
			if err != nil {
				outputError(cmd, formatter, err)
				return
			}

			// Parse start and end times
			startTime, err := parseTime(start)
			if err != nil {
				outputError(cmd, formatter,
					types.ErrInvalidInput("start", err.Error()))
				return
			}

			endTime, err := parseTime(end)
			if err != nil {
				outputError(cmd, formatter,
					types.ErrInvalidInput("end", err.Error()))
				return
			}

			// Build create parameters
			params := calendar.CreateEventParams{
				Summary:     title,
				Description: description,
				Location:    location,
				Start:       startTime,
				End:         endTime,
				TimeZone:    config.GetString("calendar.default_timezone"),
				AllDay:      allDay,
			}

			// Parse attendees
			if attendees != "" {
				params.Attendees = strings.Split(attendees, ",")
				for i := range params.Attendees {
					params.Attendees[i] = strings.TrimSpace(params.Attendees[i])
				}
			}

			// Parse recurrence
			if recurrence != "" {
				params.Recurrence = []string{recurrence}
			}

			// Create event
			event, err := client.CreateEvent(ctx, params)
			if err != nil {
				outputError(cmd, formatter, err)
				return
			}

			// Output success
			response := types.SuccessResponse("create", map[string]interface{}{
				"event":   event,
				"message": "Event created successfully",
			})
			output, err := formatter.Format(response)
			if err != nil {
				cmd.PrintErrf("Error formatting output: %v\n", err)
				return
			}
			cmd.Println(output)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Event title (required)")
	cmd.Flags().StringVar(&description, "description", "", "Event description")
	cmd.Flags().StringVar(&location, "location", "", "Event location")
	cmd.Flags().StringVar(&start, "start", "", "Start time (RFC3339 or YYYY-MM-DD HH:MM)")
	cmd.Flags().StringVar(&end, "end", "", "End time (RFC3339 or YYYY-MM-DD HH:MM)")
	cmd.Flags().StringVar(&attendees, "attendees", "", "Comma-separated email addresses")
	cmd.Flags().StringVar(&recurrence, "recurrence", "", "Recurrence rule (RFC5545 format)")
	cmd.Flags().BoolVar(&allDay, "all-day", false, "Create all-day event")

	cmd.MarkFlagRequired("title")
	cmd.MarkFlagRequired("start")
	cmd.MarkFlagRequired("end")

	return cmd
}

func newEventsListCommand(formatter output.Formatter) *cobra.Command {
	var (
		from       string
		to         string
		maxResults int64
		query      string
		orderBy    string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List calendar events",
		Long:    "List events in a date range from your Google Calendar with optional filtering and sorting",
		Example: examples.EventsListExamples,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Get calendar client
			client, err := getCalendarClient(ctx)
			if err != nil {
				outputError(cmd, formatter, err)
				return
			}

			// Parse from and to times
			fromTime, err := parseDate(from)
			if err != nil {
				outputError(cmd, formatter,
					types.ErrInvalidInput("from", err.Error()))
				return
			}

			toTime, err := parseDate(to)
			if err != nil {
				outputError(cmd, formatter,
					types.ErrInvalidInput("to", err.Error()))
				return
			}

			// Build list parameters
			params := calendar.ListEventsParams{
				From:       fromTime,
				To:         toTime,
				MaxResults: maxResults,
				Query:      query,
				OrderBy:    orderBy,
			}

			// List events
			events, err := client.ListEvents(ctx, params)
			if err != nil {
				outputError(cmd, formatter, err)
				return
			}

			// Output success
			response := types.SuccessResponse("list", map[string]interface{}{
				"events": events,
				"count":  len(events),
			})
			output, err := formatter.Format(response)
			if err != nil {
				cmd.PrintErrf("Error formatting output: %v\n", err)
				return
			}
			cmd.Println(output)
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "Start date (YYYY-MM-DD or RFC3339)")
	cmd.Flags().StringVar(&to, "to", "", "End date (YYYY-MM-DD or RFC3339)")
	cmd.Flags().Int64Var(&maxResults, "max-results", 250, "Maximum events to return")
	cmd.Flags().StringVar(&query, "query", "", "Search query string")
	cmd.Flags().StringVar(&orderBy, "order-by", "startTime", "Sort order (startTime|updated)")

	cmd.MarkFlagRequired("from")
	cmd.MarkFlagRequired("to")

	return cmd
}

func newEventsGetCommand(formatter output.Formatter) *cobra.Command {
	return &cobra.Command{
		Use:     "get <event-id>",
		Short:   "Get a calendar event",
		Long:    "Retrieve a single event by ID from your Google Calendar with full event details",
		Example: examples.EventsGetExamples,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			eventID := args[0]

			// Get calendar client
			client, err := getCalendarClient(ctx)
			if err != nil {
				outputError(cmd, formatter, err)
				return
			}

			// Get event
			event, err := client.GetEvent(ctx, eventID)
			if err != nil {
				outputError(cmd, formatter, err)
				return
			}

			// Output success
			response := types.SuccessResponse("get", map[string]interface{}{
				"event": event,
			})
			output, err := formatter.Format(response)
			if err != nil {
				cmd.PrintErrf("Error formatting output: %v\n", err)
				return
			}
			cmd.Println(output)
		},
	}
}

func newEventsUpdateCommand(formatter output.Formatter) *cobra.Command {
	var (
		title       string
		description string
		location    string
		start       string
		end         string
		attendees   string
		recurrence  string
		allDay      bool
	)

	cmd := &cobra.Command{
		Use:     "update <event-id>",
		Short:   "Update a calendar event",
		Long:    "Update an existing event in your Google Calendar with partial updates (only specified fields are changed)",
		Example: examples.EventsUpdateExamples,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			eventID := args[0]

			// Get calendar client
			client, err := getCalendarClient(ctx)
			if err != nil {
				outputError(cmd, formatter, err)
				return
			}

			// Build update parameters
			params := calendar.CreateEventParams{
				Summary:     title,
				Description: description,
				Location:    location,
				TimeZone:    config.GetString("calendar.default_timezone"),
				AllDay:      allDay,
			}

			// Parse start and end times if provided
			if start != "" {
				startTime, err := parseTime(start)
				if err != nil {
					outputError(cmd, formatter,
						types.ErrInvalidInput("start", err.Error()))
					return
				}
				params.Start = startTime
			}

			if end != "" {
				endTime, err := parseTime(end)
				if err != nil {
					outputError(cmd, formatter,
						types.ErrInvalidInput("end", err.Error()))
					return
				}
				params.End = endTime
			}

			// Parse attendees
			if attendees != "" {
				params.Attendees = strings.Split(attendees, ",")
				for i := range params.Attendees {
					params.Attendees[i] = strings.TrimSpace(params.Attendees[i])
				}
			}

			// Parse recurrence
			if recurrence != "" {
				params.Recurrence = []string{recurrence}
			}

			// Update event
			event, err := client.UpdateEvent(ctx, eventID, params)
			if err != nil {
				outputError(cmd, formatter, err)
				return
			}

			// Output success
			response := types.SuccessResponse("update", map[string]interface{}{
				"event":   event,
				"message": "Event updated successfully",
			})
			output, err := formatter.Format(response)
			if err != nil {
				cmd.PrintErrf("Error formatting output: %v\n", err)
				return
			}
			cmd.Println(output)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Event title")
	cmd.Flags().StringVar(&description, "description", "", "Event description")
	cmd.Flags().StringVar(&location, "location", "", "Event location")
	cmd.Flags().StringVar(&start, "start", "", "Start time (RFC3339 or YYYY-MM-DD HH:MM)")
	cmd.Flags().StringVar(&end, "end", "", "End time (RFC3339 or YYYY-MM-DD HH:MM)")
	cmd.Flags().StringVar(&attendees, "attendees", "", "Comma-separated email addresses")
	cmd.Flags().StringVar(&recurrence, "recurrence", "", "Recurrence rule (RFC5545 format)")
	cmd.Flags().BoolVar(&allDay, "all-day", false, "Create all-day event")

	return cmd
}

func newEventsDeleteCommand(formatter output.Formatter) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:     "delete <event-id>",
		Short:   "Delete a calendar event",
		Long:    "Delete an event from your Google Calendar with optional confirmation",
		Example: examples.EventsDeleteExamples,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			eventID := args[0]

			// Get calendar client
			client, err := getCalendarClient(ctx)
			if err != nil {
				outputError(cmd, formatter, err)
				return
			}

			// Delete event
			err = client.DeleteEvent(ctx, eventID)
			if err != nil {
				outputError(cmd, formatter, err)
				return
			}

			// Output success
			response := types.SuccessResponse("delete", map[string]interface{}{
				"eventId": eventID,
				"message": "Event deleted successfully",
			})
			output, err := formatter.Format(response)
			if err != nil {
				cmd.PrintErrf("Error formatting output: %v\n", err)
				return
			}
			cmd.Println(output)
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Skip confirmation prompt")

	return cmd
}

// Helper functions

// getCalendarClient creates an authenticated calendar client
func getCalendarClient(ctx context.Context) (*calendar.Client, error) {
	credentialsPath := config.GetString("auth.credentials_path")
	tokenPath := config.GetString("auth.token_path")

	// Create auth manager
	manager, err := auth.NewManager(credentialsPath, tokenPath)
	if err != nil {
		return nil, err
	}

	// Get Calendar service
	service, err := manager.GetCalendarService(ctx)
	if err != nil {
		return nil, err
	}

	// Create calendar client
	calendarID := config.GetString("calendar.default_calendar_id")
	return calendar.NewClient(service, calendarID), nil
}

// parseTime parses a time string in various formats
func parseTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, fmt.Errorf("time string is empty")
	}

	// Try RFC3339 format first
	t, err := time.Parse(time.RFC3339, timeStr)
	if err == nil {
		return t, nil
	}

	// Try common formats
	formats := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
	}

	for _, format := range formats {
		t, err := time.Parse(format, timeStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time format: %s (expected RFC3339 or YYYY-MM-DD HH:MM)", timeStr)
}

// parseDate parses a date string and returns start of day
func parseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date string is empty")
	}

	// Try RFC3339 format first
	t, err := time.Parse(time.RFC3339, dateStr)
	if err == nil {
		return t, nil
	}

	// Try date-only format
	t, err = time.Parse("2006-01-02", dateStr)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s (expected YYYY-MM-DD or RFC3339)", dateStr)
}

// outputError outputs an error response
func outputError(cmd *cobra.Command, formatter output.Formatter, err error) {
	appErr, ok := err.(*types.AppError)
	if !ok {
		appErr = types.ErrAPIError.
			WithDetails("operation failed").
			WithWrappedError(err)
	}
	response := types.ErrorResponse(appErr)
	output, _ := formatter.Format(response)
	cmd.Println(output)
}
