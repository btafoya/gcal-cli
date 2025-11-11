package calendar

import (
	"context"
	"fmt"

	"github.com/btafoya/gcal-cli/pkg/types"
)

// AttendeeOperation represents an operation on event attendees
type AttendeeOperation struct {
	Add    []string // Email addresses to add
	Remove []string // Email addresses to remove
}

// ManageAttendees adds or removes attendees from an event
func (c *Client) ManageAttendees(ctx context.Context, eventID string, operation AttendeeOperation) (*types.Event, error) {
	if eventID == "" {
		return nil, types.ErrMissingRequired("event-id")
	}

	// Validate email addresses
	for _, email := range operation.Add {
		if !isValidEmail(email) {
			return nil, types.ErrInvalidInput("attendees",
				fmt.Sprintf("invalid email address: %s", email))
		}
	}

	for _, email := range operation.Remove {
		if !isValidEmail(email) {
			return nil, types.ErrInvalidInput("attendees",
				fmt.Sprintf("invalid email address: %s", email))
		}
	}

	// Get current event
	event, err := c.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Build new attendee list
	newAttendees := make(map[string]types.Attendee)

	// Add existing attendees
	for _, att := range event.Attendees {
		newAttendees[att.Email] = att
	}

	// Remove specified attendees
	for _, email := range operation.Remove {
		delete(newAttendees, email)
	}

	// Add new attendees
	for _, email := range operation.Add {
		if _, exists := newAttendees[email]; !exists {
			newAttendees[email] = types.Attendee{
				Email:          email,
				ResponseStatus: "needsAction",
			}
		}
	}

	// Convert map to slice
	attendeeEmails := make([]string, 0, len(newAttendees))
	for email := range newAttendees {
		attendeeEmails = append(attendeeEmails, email)
	}

	// Update event with new attendee list
	updateParams := CreateEventParams{
		Attendees: attendeeEmails,
	}

	return c.UpdateEvent(ctx, eventID, updateParams)
}

// AddAttendees adds attendees to an event
func (c *Client) AddAttendees(ctx context.Context, eventID string, emails []string) (*types.Event, error) {
	return c.ManageAttendees(ctx, eventID, AttendeeOperation{
		Add: emails,
	})
}

// RemoveAttendees removes attendees from an event
func (c *Client) RemoveAttendees(ctx context.Context, eventID string, emails []string) (*types.Event, error) {
	return c.ManageAttendees(ctx, eventID, AttendeeOperation{
		Remove: emails,
	})
}

// ReplaceAttendees replaces all attendees on an event
func (c *Client) ReplaceAttendees(ctx context.Context, eventID string, emails []string) (*types.Event, error) {
	if eventID == "" {
		return nil, types.ErrMissingRequired("event-id")
	}

	// Validate all emails
	for _, email := range emails {
		if !isValidEmail(email) {
			return nil, types.ErrInvalidInput("attendees",
				fmt.Sprintf("invalid email address: %s", email))
		}
	}

	// Update event with new attendee list
	updateParams := CreateEventParams{
		Attendees: emails,
	}

	return c.UpdateEvent(ctx, eventID, updateParams)
}

// GetAttendees retrieves the attendee list for an event
func (c *Client) GetAttendees(ctx context.Context, eventID string) ([]types.Attendee, error) {
	event, err := c.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	return event.Attendees, nil
}

// FindAttendee checks if an attendee is invited to an event
func (c *Client) FindAttendee(ctx context.Context, eventID, email string) (*types.Attendee, error) {
	attendees, err := c.GetAttendees(ctx, eventID)
	if err != nil {
		return nil, err
	}

	for _, att := range attendees {
		if att.Email == email {
			return &att, nil
		}
	}

	return nil, types.ErrNotFound("attendee", email)
}
