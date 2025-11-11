package examples

// EventsCreateExamples provides comprehensive examples for events create command
const EventsCreateExamples = `Examples:
  # Create a simple 1-hour meeting
  gcal-cli events create \
    --title "Team Meeting" \
    --start "2024-01-15T10:00:00" \
    --end "2024-01-15T11:00:00"

  # Create meeting with description and location
  gcal-cli events create \
    --title "Project Review" \
    --description "Q1 project status review" \
    --location "Conference Room A" \
    --start "2024-01-16 14:00" \
    --end "2024-01-16 15:30"

  # Create meeting with attendees
  gcal-cli events create \
    --title "Client Call" \
    --start "2024-01-17T09:00:00-05:00" \
    --end "2024-01-17T10:00:00-05:00" \
    --attendees "client@example.com,teammate@company.com"

  # Create all-day event
  gcal-cli events create \
    --title "Conference" \
    --start "2024-01-20" \
    --end "2024-01-21" \
    --all-day

  # Create recurring weekly meeting
  gcal-cli events create \
    --title "Weekly Standup" \
    --start "2024-01-15T09:00:00" \
    --end "2024-01-15T09:30:00" \
    --recurrence "RRULE:FREQ=WEEKLY;COUNT=10"

  # Create with custom timezone
  gcal-cli events create \
    --title "Remote Meeting" \
    --start "2024-01-18T14:00:00" \
    --end "2024-01-18T15:00:00" \
    --timezone "America/Los_Angeles"

  # LLM Agent Usage: Parse JSON response to get event ID
  EVENT_ID=$(gcal-cli events create \
    --title "Automated Event" \
    --start "2024-01-19T10:00:00" \
    --end "2024-01-19T11:00:00" \
    --format json | jq -r '.data.event.id')
`

// EventsListExamples provides comprehensive examples for events list command
const EventsListExamples = `Examples:
  # List events in date range
  gcal-cli events list \
    --from "2024-01-15" \
    --to "2024-01-20"

  # List with search query
  gcal-cli events list \
    --from "2024-01-01" \
    --to "2024-01-31" \
    --query "meeting"

  # List with max results
  gcal-cli events list \
    --from "2024-01-15" \
    --to "2024-01-20" \
    --max-results 10

  # List sorted by last updated
  gcal-cli events list \
    --from "2024-01-15" \
    --to "2024-01-20" \
    --order-by updated

  # LLM Agent Usage: Parse event list
  gcal-cli events list \
    --from "2024-01-15" \
    --to "2024-01-20" \
    --format json | jq '.data.events[] | {id, summary, start}'

  # LLM Agent Usage: Count events
  COUNT=$(gcal-cli events list \
    --from "2024-01-15" \
    --to "2024-01-20" \
    --format json | jq '.data.count')

  # Get only event IDs for piping
  gcal-cli events list \
    --from "2024-01-15" \
    --to "2024-01-20" \
    --format minimal
`

// EventsGetExamples provides comprehensive examples for events get command
const EventsGetExamples = `Examples:
  # Get event by ID
  gcal-cli events get abc123xyz

  # Get event with JSON output
  gcal-cli events get abc123xyz --format json

  # LLM Agent Usage: Extract specific fields
  gcal-cli events get abc123xyz --format json | jq '.data.event | {summary, start, end}'

  # LLM Agent Usage: Check event status
  STATUS=$(gcal-cli events get abc123xyz --format json | jq -r '.data.event.status')
`

// EventsUpdateExamples provides comprehensive examples for events update command
const EventsUpdateExamples = `Examples:
  # Update event title
  gcal-cli events update abc123xyz \
    --title "Updated Meeting Title"

  # Update event time
  gcal-cli events update abc123xyz \
    --start "2024-01-15T11:00:00" \
    --end "2024-01-15T12:00:00"

  # Update description and location
  gcal-cli events update abc123xyz \
    --description "Updated agenda" \
    --location "Room B"

  # Update attendees (replaces existing)
  gcal-cli events update abc123xyz \
    --attendees "new@example.com,another@example.com"

  # Update multiple fields at once
  gcal-cli events update abc123xyz \
    --title "Revised Meeting" \
    --start "2024-01-15T14:00:00" \
    --end "2024-01-15T15:00:00" \
    --location "Conference Room C"

  # LLM Agent Usage: Conditional update
  if [ "$STATUS" = "tentative" ]; then
    gcal-cli events update abc123xyz --title "CONFIRMED: $TITLE"
  fi
`

// EventsDeleteExamples provides comprehensive examples for events delete command
const EventsDeleteExamples = `Examples:
  # Delete event with confirmation prompt
  gcal-cli events delete abc123xyz

  # Delete without confirmation
  gcal-cli events delete abc123xyz --confirm

  # LLM Agent Usage: Delete and check success
  RESULT=$(gcal-cli events delete abc123xyz --confirm --format json)
  if [ "$(echo $RESULT | jq -r '.success')" = "true" ]; then
    echo "Event deleted"
  fi

  # LLM Agent Usage: Batch delete events
  gcal-cli events list --from "2024-01-15" --to "2024-01-20" --format minimal | \
    while read EVENT_ID; do
      gcal-cli events delete "$EVENT_ID" --confirm
    done
`

// CalendarsListExamples provides comprehensive examples for calendars list command
const CalendarsListExamples = `Examples:
  # List all accessible calendars
  gcal-cli calendars list

  # List with JSON output
  gcal-cli calendars list --format json

  # LLM Agent Usage: Extract calendar IDs
  gcal-cli calendars list --format json | jq -r '.data.calendars[] | .id'

  # LLM Agent Usage: Find primary calendar
  PRIMARY=$(gcal-cli calendars list --format json | \
    jq -r '.data.calendars[] | select(.primary == true) | .id')
`

// CalendarsGetExamples provides comprehensive examples for calendars get command
const CalendarsGetExamples = `Examples:
  # Get primary calendar metadata
  gcal-cli calendars get

  # Get specific calendar by ID
  gcal-cli calendars get calendar@group.calendar.google.com

  # LLM Agent Usage: Get calendar timezone
  TZ=$(gcal-cli calendars get --format json | jq -r '.data.calendar.timeZone')
`

// AuthLoginExamples provides comprehensive examples for auth login command
const AuthLoginExamples = `Examples:
  # Authenticate with default credentials location
  gcal-cli auth login

  # Authenticate with custom credentials file
  gcal-cli auth login --credentials /path/to/credentials.json

  # LLM Agent Usage: Check authentication success
  RESULT=$(gcal-cli auth login --format json)
  if [ "$(echo $RESULT | jq -r '.success')" = "true" ]; then
    echo "Authenticated as $(echo $RESULT | jq -r '.data.email')"
  fi
`

// AuthStatusExamples provides comprehensive examples for auth status command
const AuthStatusExamples = `Examples:
  # Check authentication status
  gcal-cli auth status

  # Check with JSON output
  gcal-cli auth status --format json

  # LLM Agent Usage: Verify authentication before operations
  IS_AUTH=$(gcal-cli auth status --format json | jq -r '.data.authenticated')
  if [ "$IS_AUTH" != "true" ]; then
    gcal-cli auth login
  fi

  # LLM Agent Usage: Get token expiry
  EXPIRY=$(gcal-cli auth status --format json | jq -r '.data.tokenExpiry')
`

// AuthLogoutExamples provides comprehensive examples for auth logout command
const AuthLogoutExamples = `Examples:
  # Logout and clear tokens
  gcal-cli auth logout

  # LLM Agent Usage: Logout with confirmation
  gcal-cli auth logout --format json | jq '.data.message'
`

// ConfigShowExamples provides comprehensive examples for config show command
const ConfigShowExamples = `Examples:
  # Show current configuration
  gcal-cli config show

  # Show with JSON output
  gcal-cli config show --format json

  # LLM Agent Usage: Get specific config value
  DEFAULT_CAL=$(gcal-cli config show --format json | \
    jq -r '.data.config.calendar.default_calendar_id')
`

// ErrorHandlingExamples provides error handling examples for LLM agents
const ErrorHandlingExamples = `Error Handling for LLM Agents:

# Basic error detection
RESULT=$(gcal-cli events create --title "Test" --start "invalid" --end "invalid" --format json 2>&1)
if [ "$(echo $RESULT | jq -r '.success')" != "true" ]; then
  ERROR_CODE=$(echo $RESULT | jq -r '.error.code')
  ERROR_MSG=$(echo $RESULT | jq -r '.error.message')
  echo "Error: $ERROR_CODE - $ERROR_MSG"
fi

# Handle authentication errors
perform_operation() {
  RESULT=$(gcal-cli events list --from "2024-01-15" --to "2024-01-20" --format json 2>&1)
  ERROR_CODE=$(echo $RESULT | jq -r '.error.code // empty')

  case $ERROR_CODE in
    AUTH_FAILED|TOKEN_EXPIRED)
      echo "Re-authenticating..."
      gcal-cli auth login
      # Retry operation
      gcal-cli events list --from "2024-01-15" --to "2024-01-20"
      ;;
    RATE_LIMIT)
      echo "Rate limited, waiting..."
      sleep 5
      # Retry operation
      $0
      ;;
    "")
      # Success
      echo $RESULT
      ;;
    *)
      echo "Error: $ERROR_CODE"
      echo $(echo $RESULT | jq -r '.error.suggestedAction')
      ;;
  esac
}

# Idempotent event creation
create_event_if_not_exists() {
  TITLE=$1
  START=$2
  END=$3

  # Check if event exists
  EXISTING=$(gcal-cli events list \
    --from "$(date -d "$START" +%Y-%m-%d)" \
    --to "$(date -d "$END" +%Y-%m-%d)" \
    --query "$TITLE" \
    --format json | jq -r '.data.events[] | select(.summary == "'"$TITLE"'") | .id')

  if [ -z "$EXISTING" ]; then
    # Create new event
    gcal-cli events create --title "$TITLE" --start "$START" --end "$END"
  else
    echo "Event already exists: $EXISTING"
  fi
}
`
