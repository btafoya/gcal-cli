package types

// Event represents a calendar event
type Event struct {
	ID          string     `json:"id"`
	Summary     string     `json:"summary"`
	Description string     `json:"description,omitempty"`
	Start       EventTime  `json:"start"`
	End         EventTime  `json:"end"`
	Status      string     `json:"status"`
	Attendees   []Attendee `json:"attendees,omitempty"`
	Recurrence  []string   `json:"recurrence,omitempty"`
	Location    string     `json:"location,omitempty"`
	HTMLLink    string     `json:"htmlLink,omitempty"`
}

// EventTime represents a point in time for an event
type EventTime struct {
	DateTime string `json:"dateTime,omitempty"`
	Date     string `json:"date,omitempty"` // For all-day events
	TimeZone string `json:"timeZone,omitempty"`
}

// Attendee represents an event attendee
type Attendee struct {
	Email          string `json:"email"`
	ResponseStatus string `json:"responseStatus"`
	Organizer      bool   `json:"organizer,omitempty"`
	DisplayName    string `json:"displayName,omitempty"`
}
