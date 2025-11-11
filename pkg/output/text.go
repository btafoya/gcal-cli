package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/btafoya/gcal-cli/pkg/types"
)

// TextFormatter formats output as human-readable text
type TextFormatter struct{}

// Format formats a response as human-readable text
func (f *TextFormatter) Format(response *types.Response) (string, error) {
	var builder strings.Builder

	if !response.Success {
		// Format error response
		builder.WriteString("✗ Error\n\n")
		if response.Error != nil {
			builder.WriteString(fmt.Sprintf("Code:    %s\n", response.Error.Code))
			builder.WriteString(fmt.Sprintf("Message: %s\n", response.Error.Message))
			if response.Error.Details != "" {
				builder.WriteString(fmt.Sprintf("Details: %s\n", response.Error.Details))
			}
			if response.Error.SuggestedAction != "" {
				builder.WriteString(fmt.Sprintf("\nSuggested Action:\n  %s\n", response.Error.SuggestedAction))
			}
		}
		return builder.String(), nil
	}

	// Format success response
	builder.WriteString("✓ Success\n\n")

	// Handle different data types
	switch data := response.Data.(type) {
	case *types.EventData:
		f.formatEventData(&builder, data)
	case *types.EventListData:
		f.formatEventListData(&builder, data)
	case *types.AuthData:
		f.formatAuthData(&builder, data)
	case map[string]interface{}:
		f.formatGenericData(&builder, data)
	default:
		builder.WriteString(fmt.Sprintf("Data: %v\n", data))
	}

	return builder.String(), nil
}

func (f *TextFormatter) formatEventData(builder *strings.Builder, data *types.EventData) {
	if data.Message != "" {
		builder.WriteString(fmt.Sprintf("%s\n\n", data.Message))
	}

	if data.Event != nil {
		builder.WriteString("Event Details:\n")
		builder.WriteString(fmt.Sprintf("  ID:       %s\n", data.Event.ID))
		builder.WriteString(fmt.Sprintf("  Title:    %s\n", data.Event.Summary))
		if data.Event.Description != "" {
			builder.WriteString(fmt.Sprintf("  Description: %s\n", data.Event.Description))
		}
		builder.WriteString(fmt.Sprintf("  Start:    %s\n", f.formatEventTime(&data.Event.Start)))
		builder.WriteString(fmt.Sprintf("  End:      %s\n", f.formatEventTime(&data.Event.End)))
		builder.WriteString(fmt.Sprintf("  Status:   %s\n", data.Event.Status))
		if data.Event.Location != "" {
			builder.WriteString(fmt.Sprintf("  Location: %s\n", data.Event.Location))
		}

		if len(data.Event.Attendees) > 0 {
			builder.WriteString("\nAttendees:\n")
			for _, attendee := range data.Event.Attendees {
				status := ""
				switch attendee.ResponseStatus {
				case "accepted":
					status = "✓"
				case "declined":
					status = "✗"
				case "tentative":
					status = "?"
				default:
					status = "○"
				}
				builder.WriteString(fmt.Sprintf("  %s %s (%s)\n", status, attendee.Email, attendee.ResponseStatus))
			}
		}

		if data.Event.HTMLLink != "" {
			builder.WriteString(fmt.Sprintf("\nLink: %s\n", data.Event.HTMLLink))
		}
	} else if data.EventID != "" {
		builder.WriteString(fmt.Sprintf("Event ID: %s\n", data.EventID))
	}
}

func (f *TextFormatter) formatEventListData(builder *strings.Builder, data *types.EventListData) {
	builder.WriteString(fmt.Sprintf("Found %d event(s)\n\n", data.Count))

	for i, event := range data.Events {
		builder.WriteString(fmt.Sprintf("[%d] %s\n", i+1, event.Summary))
		builder.WriteString(fmt.Sprintf("    ID:    %s\n", event.ID))
		builder.WriteString(fmt.Sprintf("    Start: %s\n", f.formatEventTime(&event.Start)))
		builder.WriteString(fmt.Sprintf("    End:   %s\n", f.formatEventTime(&event.End)))
		if event.Location != "" {
			builder.WriteString(fmt.Sprintf("    Location: %s\n", event.Location))
		}
		builder.WriteString("\n")
	}

	if data.NextPageToken != "" {
		builder.WriteString(fmt.Sprintf("Next Page Token: %s\n", data.NextPageToken))
	}
}

func (f *TextFormatter) formatAuthData(builder *strings.Builder, data *types.AuthData) {
	builder.WriteString(fmt.Sprintf("%s\n", data.Message))
	if data.Email != "" {
		builder.WriteString(fmt.Sprintf("\nEmail:  %s\n", data.Email))
	}
	if len(data.Scopes) > 0 {
		builder.WriteString("\nScopes:\n")
		for _, scope := range data.Scopes {
			builder.WriteString(fmt.Sprintf("  • %s\n", scope))
		}
	}
}

func (f *TextFormatter) formatGenericData(builder *strings.Builder, data map[string]interface{}) {
	for key, value := range data {
		builder.WriteString(fmt.Sprintf("%s: %v\n", key, value))
	}
}

func (f *TextFormatter) formatEventTime(eventTime *types.EventTime) string {
	if eventTime.Date != "" {
		// All-day event
		return eventTime.Date
	}

	if eventTime.DateTime != "" {
		// Parse and format the datetime
		t, err := time.Parse(time.RFC3339, eventTime.DateTime)
		if err != nil {
			return eventTime.DateTime
		}

		// Format as: Mon, Jan 2 2006 3:04 PM MST
		return t.Format("Mon, Jan 2 2006 3:04 PM MST")
	}

	return "Unknown"
}
