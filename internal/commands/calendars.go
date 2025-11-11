package commands

import (
	"context"

	"github.com/btafoya/gcal-cli/pkg/examples"
	"github.com/btafoya/gcal-cli/pkg/output"
	"github.com/btafoya/gcal-cli/pkg/types"
	"github.com/spf13/cobra"
)

// NewCalendarsCommand creates the calendars command group
func NewCalendarsCommand(formatter output.Formatter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calendars",
		Short: "Manage calendars",
		Long:  "List and retrieve calendar information",
	}

	cmd.AddCommand(newCalendarsListCommand(formatter))
	cmd.AddCommand(newCalendarsGetCommand(formatter))

	return cmd
}

func newCalendarsListCommand(formatter output.Formatter) *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Short:   "List all calendars",
		Long:    "List all calendars accessible to the authenticated user with access roles and timezone information",
		Example: examples.CalendarsListExamples,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			// Get calendar client
			client, err := getCalendarClient(ctx)
			if err != nil {
				outputError(cmd, formatter, err)
				return
			}

			// List calendars
			calendars, err := client.ListCalendars(ctx)
			if err != nil {
				outputError(cmd, formatter, err)
				return
			}

			// Output success
			response := types.SuccessResponse("list_calendars", map[string]interface{}{
				"calendars": calendars,
				"count":     len(calendars),
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

func newCalendarsGetCommand(formatter output.Formatter) *cobra.Command {
	return &cobra.Command{
		Use:     "get [calendar-id]",
		Short:   "Get calendar information",
		Long:    "Retrieve metadata for a specific calendar including timezone and access role (defaults to primary)",
		Example: examples.CalendarsGetExamples,
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			var calendarID string
			if len(args) > 0 {
				calendarID = args[0]
			}

			// Get calendar client
			client, err := getCalendarClient(ctx)
			if err != nil {
				outputError(cmd, formatter, err)
				return
			}

			// Get calendar
			cal, err := client.GetCalendar(ctx, calendarID)
			if err != nil {
				outputError(cmd, formatter, err)
				return
			}

			// Output success
			response := types.SuccessResponse("get_calendar", map[string]interface{}{
				"calendar": cal,
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
