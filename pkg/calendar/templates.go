package calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/btafoya/gcal-cli/pkg/config"
	"github.com/btafoya/gcal-cli/pkg/types"
	"google.golang.org/api/calendar/v3"
)

// EventTemplate represents a predefined event template
type EventTemplate struct {
	Name              string   `json:"name"`
	Summary           string   `json:"summary"`
	Description       string   `json:"description,omitempty"`
	Location          string   `json:"location,omitempty"`
	DurationMinutes   int      `json:"durationMinutes"`
	Attendees         []string `json:"attendees,omitempty"`
	Recurrence        []string `json:"recurrence,omitempty"`
	ReminderMinutes   int      `json:"reminderMinutes,omitempty"`
	ColorID           string   `json:"colorId,omitempty"`
	Visibility        string   `json:"visibility,omitempty"`
	SendNotifications bool     `json:"sendNotifications"`
}

// TemplateManager manages event templates
type TemplateManager struct {
	templatesPath string
	templates     map[string]EventTemplate
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() (*TemplateManager, error) {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return nil, err
	}

	templatesPath := filepath.Join(configDir, "templates.json")

	tm := &TemplateManager{
		templatesPath: templatesPath,
		templates:     make(map[string]EventTemplate),
	}

	// Load existing templates if file exists
	if _, err := os.Stat(templatesPath); err == nil {
		if err := tm.Load(); err != nil {
			return nil, err
		}
	}

	return tm, nil
}

// Load loads templates from the templates file
func (tm *TemplateManager) Load() error {
	data, err := os.ReadFile(tm.templatesPath)
	if err != nil {
		return fmt.Errorf("failed to read templates file: %w", err)
	}

	if err := json.Unmarshal(data, &tm.templates); err != nil {
		return fmt.Errorf("failed to parse templates file: %w", err)
	}

	return nil
}

// Save saves templates to the templates file
func (tm *TemplateManager) Save() error {
	data, err := json.MarshalIndent(tm.templates, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal templates: %w", err)
	}

	if err := os.WriteFile(tm.templatesPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write templates file: %w", err)
	}

	return nil
}

// Get retrieves a template by name
func (tm *TemplateManager) Get(name string) (EventTemplate, error) {
	template, ok := tm.templates[name]
	if !ok {
		return EventTemplate{}, types.ErrNotFound("template", name)
	}
	return template, nil
}

// List returns all available templates
func (tm *TemplateManager) List() map[string]EventTemplate {
	return tm.templates
}

// Add adds a new template
func (tm *TemplateManager) Add(name string, template EventTemplate) error {
	if name == "" {
		return types.ErrInvalidInput("name", "template name cannot be empty")
	}

	template.Name = name
	tm.templates[name] = template

	return tm.Save()
}

// Delete removes a template
func (tm *TemplateManager) Delete(name string) error {
	if _, ok := tm.templates[name]; !ok {
		return types.ErrNotFound("template", name)
	}

	delete(tm.templates, name)
	return tm.Save()
}

// CreateEventFromTemplate creates a calendar event from a template
func (c *Client) CreateEventFromTemplate(ctx context.Context, calendarID string, templateName string, start time.Time, overrides map[string]interface{}) (*types.Event, error) {
	// Get template
	tm, err := NewTemplateManager()
	if err != nil {
		return nil, err
	}

	template, err := tm.Get(templateName)
	if err != nil {
		return nil, err
	}

	// Calculate end time based on duration
	end := start.Add(time.Duration(template.DurationMinutes) * time.Minute)

	// Build event from template
	event := &calendar.Event{
		Summary:     template.Summary,
		Description: template.Description,
		Location:    template.Location,
		Start: &calendar.EventDateTime{
			DateTime: start.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: end.Format(time.RFC3339),
		},
	}

	// Add attendees
	if len(template.Attendees) > 0 {
		event.Attendees = make([]*calendar.EventAttendee, len(template.Attendees))
		for i, email := range template.Attendees {
			event.Attendees[i] = &calendar.EventAttendee{
				Email: email,
			}
		}
	}

	// Add recurrence
	if len(template.Recurrence) > 0 {
		event.Recurrence = template.Recurrence
	}

	// Add reminders
	if template.ReminderMinutes > 0 {
		event.Reminders = &calendar.EventReminders{
			UseDefault: false,
			Overrides: []*calendar.EventReminder{
				{
					Method:  "popup",
					Minutes: int64(template.ReminderMinutes),
				},
			},
		}
	}

	// Add color
	if template.ColorID != "" {
		event.ColorId = template.ColorID
	}

	// Add visibility
	if template.Visibility != "" {
		event.Visibility = template.Visibility
	}

	// Apply overrides
	if overrides != nil {
		if summary, ok := overrides["summary"].(string); ok {
			event.Summary = summary
		}
		if description, ok := overrides["description"].(string); ok {
			event.Description = description
		}
		if location, ok := overrides["location"].(string); ok {
			event.Location = location
		}
	}

	// Create event
	createdEvent, err := c.Service.Events.Insert(calendarID, event).
		SendUpdates("all").
		Context(ctx).
		Do()
	if err != nil {
		return nil, types.ErrAPIError.WithDetails(fmt.Sprintf("failed to create event from template: %v", err))
	}

	return convertEvent(createdEvent), nil
}

// DefaultTemplates returns a set of common default templates
func DefaultTemplates() map[string]EventTemplate {
	return map[string]EventTemplate{
		"meeting": {
			Name:              "meeting",
			Summary:           "Team Meeting",
			Description:       "Regular team sync meeting",
			DurationMinutes:   60,
			ReminderMinutes:   10,
			SendNotifications: true,
		},
		"1on1": {
			Name:              "1on1",
			Summary:           "1:1 Meeting",
			Description:       "One-on-one check-in",
			DurationMinutes:   30,
			ReminderMinutes:   10,
			SendNotifications: true,
		},
		"lunch": {
			Name:              "lunch",
			Summary:           "Lunch Break",
			DurationMinutes:   60,
			ReminderMinutes:   15,
			SendNotifications: false,
		},
		"focus": {
			Name:              "focus",
			Summary:           "Focus Time",
			Description:       "Deep work - no interruptions",
			DurationMinutes:   120,
			ReminderMinutes:   0,
			Visibility:        "private",
			SendNotifications: false,
		},
		"standup": {
			Name:              "standup",
			Summary:           "Daily Standup",
			Description:       "Daily team standup",
			DurationMinutes:   15,
			Recurrence:        []string{"RRULE:FREQ=DAILY;BYDAY=MO,TU,WE,TH,FR"},
			ReminderMinutes:   5,
			SendNotifications: true,
		},
		"interview": {
			Name:              "interview",
			Summary:           "Interview",
			Description:       "Candidate interview",
			DurationMinutes:   60,
			ReminderMinutes:   30,
			SendNotifications: true,
		},
	}
}

// InitializeDefaultTemplates creates the default templates file
func InitializeDefaultTemplates() error {
	tm, err := NewTemplateManager()
	if err != nil {
		return err
	}

	defaults := DefaultTemplates()
	for name, template := range defaults {
		tm.templates[name] = template
	}

	return tm.Save()
}
