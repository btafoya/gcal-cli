# gcal-cli User Instructions

Complete guide for using the Google Calendar CLI tool, designed for both human users and LLM agents.

**Version**: 1.0 (Phases 1-7 Complete)
**Last Updated**: 2025-11-11

---

## Table of Contents

1. [Getting Started](#getting-started)
2. [Installation](#installation)
3. [Authentication](#authentication)
4. [Basic Commands](#basic-commands)
5. [Event Management](#event-management)
6. [Calendar Management](#calendar-management)
7. [Advanced Features](#advanced-features)
8. [Configuration](#configuration)
9. [Output Formats](#output-formats)
10. [Error Handling](#error-handling)
11. [LLM Agent Integration](#llm-agent-integration)
12. [Troubleshooting](#troubleshooting)

---

## Getting Started

### What is gcal-cli?

gcal-cli is a command-line interface for Google Calendar specifically designed for:
- **Human Users**: Quick calendar management from the terminal
- **LLM Agents**: Programmatic calendar access with structured JSON output
- **Automation**: Reliable, scriptable calendar operations

### Key Features

✅ **Complete CRUD Operations**: Create, read, update, delete calendar events
✅ **Natural Language Dates**: Parse "tomorrow at 2pm", "next Monday", "in 2 hours"
✅ **LLM-Optimized Output**: Structured JSON with consistent schemas
✅ **Multi-Calendar Support**: Work with multiple calendars simultaneously
✅ **Smart Scheduling**: Free/busy queries, conflict detection, find meeting times
✅ **Event Templates**: Quick event creation with predefined templates
✅ **OAuth2 Authentication**: Secure, automatic token refresh
✅ **Error Recovery**: Automatic retry with exponential backoff

---

## Installation

### Prerequisites

- **Go 1.21+**: [Install Go](https://golang.org/dl/)
- **Google Cloud Account**: [Google Cloud Console](https://console.cloud.google.com)
- **Calendar API Enabled**: Enable in Cloud Console
- **OAuth2 Credentials**: Download client credentials

### Build from Source

```bash
# Clone repository
git clone https://github.com/btafoya/gcal-cli.git
cd gcal-cli

# Build binary
go build -o gcal-cli ./cmd/gcal-cli

# Verify installation
./gcal-cli version
```

### Install to PATH

```bash
# Option 1: Copy to system bin
sudo cp gcal-cli /usr/local/bin/

# Option 2: Add to PATH
export PATH=$PATH:$(pwd)

# Verify
gcal-cli --help
```

---

## Authentication

### Step 1: Get OAuth2 Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Create or select a project
3. Enable **Google Calendar API**
4. Go to **Credentials** → **Create Credentials** → **OAuth 2.0 Client ID**
5. Select **Desktop application**
6. Download `credentials.json`
7. Save to `~/.config/gcal-cli/credentials.json`

### Step 2: Authenticate

```bash
# First-time login (opens browser)
gcal-cli auth login

# Check authentication status
gcal-cli auth status --format json
```

**Expected Output**:
```json
{
  "success": true,
  "data": {
    "authenticated": true,
    "email": "your-email@gmail.com",
    "scopes": ["https://www.googleapis.com/auth/calendar"]
  }
}
```

### Step 3: Verify Access

```bash
# List your calendars
gcal-cli calendars list --format json
```

### Authentication Commands

```bash
# Login (opens browser for OAuth)
gcal-cli auth login

# Check status
gcal-cli auth status

# Logout (removes tokens)
gcal-cli auth logout
```

**Token Storage**: `~/.config/gcal-cli/tokens.json` (secure, auto-refreshed)

---

## Basic Commands

### Version Information

```bash
# Show version
gcal-cli version

# JSON format
gcal-cli version --format json
```

### Configuration

```bash
# Initialize config file
gcal-cli config init

# Show current configuration
gcal-cli config show

# Set configuration value
gcal-cli config set output.default_format json
gcal-cli config set calendar.default_calendar_id primary
```

### Help

```bash
# General help
gcal-cli --help

# Command-specific help
gcal-cli events --help
gcal-cli events create --help
```

---

## Event Management

### Create Events

#### Simple Event

```bash
gcal-cli events create \
  --title "Team Meeting" \
  --start "2024-01-15T10:00:00" \
  --end "2024-01-15T11:00:00"
```

#### Event with Details

```bash
gcal-cli events create \
  --title "Project Review" \
  --description "Q1 project status review" \
  --location "Conference Room A" \
  --start "2024-01-16 14:00" \
  --end "2024-01-16 15:30"
```

#### Event with Attendees

```bash
gcal-cli events create \
  --title "Client Call" \
  --start "2024-01-17T09:00:00" \
  --end "2024-01-17T10:00:00" \
  --attendees "client@example.com,teammate@company.com"
```

#### All-Day Event

```bash
gcal-cli events create \
  --title "Conference" \
  --start "2024-01-20" \
  --end "2024-01-21" \
  --all-day
```

#### Recurring Event

```bash
# Weekly meeting for 10 occurrences
gcal-cli events create \
  --title "Weekly Standup" \
  --start "2024-01-15T09:00:00" \
  --end "2024-01-15T09:30:00" \
  --recurrence "RRULE:FREQ=WEEKLY;COUNT=10"

# Daily meeting Monday-Friday
gcal-cli events create \
  --title "Daily Standup" \
  --start "2024-01-15T09:00:00" \
  --end "2024-01-15T09:15:00" \
  --recurrence "RRULE:FREQ=DAILY;BYDAY=MO,TU,WE,TH,FR"
```

### List Events

#### Basic Listing

```bash
# List events in date range
gcal-cli events list \
  --from "2024-01-15" \
  --to "2024-01-20"
```

#### With Search Query

```bash
# Search for events containing "meeting"
gcal-cli events list \
  --from "2024-01-01" \
  --to "2024-01-31" \
  --query "meeting"
```

#### With Limits

```bash
# Limit results
gcal-cli events list \
  --from "2024-01-15" \
  --to "2024-01-20" \
  --max-results 10
```

#### Sorted Listing

```bash
# Sort by last updated
gcal-cli events list \
  --from "2024-01-15" \
  --to "2024-01-20" \
  --order-by updated
```

### Get Event Details

```bash
# Get single event by ID
gcal-cli events get <event-id> --format json
```

### Update Events

```bash
# Update title only
gcal-cli events update <event-id> \
  --title "Updated Meeting Title"

# Update time
gcal-cli events update <event-id> \
  --start "2024-01-15T11:00:00" \
  --end "2024-01-15T12:00:00"

# Update multiple fields
gcal-cli events update <event-id> \
  --title "Revised Meeting" \
  --description "Updated agenda" \
  --location "Virtual" \
  --attendees "newperson@example.com"
```

### Delete Events

```bash
# Delete event (with confirmation)
gcal-cli events delete <event-id>

# Delete without confirmation
gcal-cli events delete <event-id> --confirm
```

---

## Calendar Management

### List Calendars

```bash
# List all accessible calendars
gcal-cli calendars list --format json
```

**Output**:
```json
{
  "success": true,
  "data": {
    "calendars": [
      {
        "id": "primary",
        "summary": "Your Calendar",
        "primary": true,
        "accessRole": "owner"
      },
      {
        "id": "work@company.com",
        "summary": "Work Calendar",
        "primary": false,
        "accessRole": "writer"
      }
    ],
    "count": 2
  }
}
```

### Get Calendar Details

```bash
# Get specific calendar
gcal-cli calendars get primary --format json
gcal-cli calendars get work@company.com --format json
```

---

## Advanced Features

### Natural Language Dates (Phase 7)

Parse human-friendly date/time strings automatically:

```bash
# Relative dates
gcal-cli events create \
  --title "Tomorrow's Meeting" \
  --start "tomorrow at 2pm" \
  --end "tomorrow at 3pm"

# Day of week
gcal-cli events create \
  --title "Monday Planning" \
  --start "next Monday at 9am" \
  --end "next Monday at 10am"

# Time offsets
gcal-cli events create \
  --title "Quick Sync" \
  --start "in 2 hours" \
  --end "in 3 hours"
```

**Supported Patterns**:
- `now`, `today`, `tomorrow`, `yesterday`
- `next Monday`, `this Friday`, `last Wednesday`
- `in 2 hours`, `in 30 minutes`, `in 1 week`
- `tomorrow at 2pm`, `Friday at 14:00`

### Free/Busy Queries (Phase 7)

Check calendar availability:

```bash
# Check if calendar is free (returns JSON)
# Note: Command integration pending
# API: client.IsBusy(ctx, "primary", start, end)
```

**API Usage** (for developers):
```go
// Check availability
isBusy, err := client.IsBusy(ctx, "primary", start, end)

// Find free 60-minute slots
freeSlots, err := client.FindFreeSlots(ctx, "primary",
    startTime, endTime, 60*time.Minute)

// Check for conflicts
hasConflict, conflicts, err := client.CheckConflicts(ctx,
    "primary", proposedStart, proposedEnd)
```

### Event Templates (Phase 7)

Use predefined templates for common events:

**Default Templates**:
- `meeting` - 60-minute team meeting
- `1on1` - 30-minute one-on-one
- `lunch` - 60-minute lunch break
- `focus` - 120-minute deep work
- `standup` - 15-minute daily standup (recurring)
- `interview` - 60-minute interview

**API Usage** (for developers):
```go
// Create event from template
event, err := client.CreateEventFromTemplate(ctx,
    "primary", "meeting", startTime, nil)

// With overrides
overrides := map[string]interface{}{
    "summary": "Custom Meeting Title",
    "location": "Conference Room B",
}
event, err := client.CreateEventFromTemplate(ctx,
    "primary", "meeting", startTime, overrides)
```

### Multi-Calendar Operations (Phase 7)

Work with multiple calendars simultaneously:

**API Usage** (for developers):
```go
// List events from multiple calendars (parallel)
calendarIDs := []string{"primary", "work@company.com"}
result, err := client.ListEventsMultiCalendar(ctx,
    calendarIDs, timeMin, timeMax, 50)

// Find common free time
freeSlots, err := client.FindCommonFreeTime(ctx,
    calendarIDs, start, end, 60*time.Minute)

// Create event in multiple calendars
results, err := client.CreateEventMultiCalendar(ctx,
    calendarIDs, event)
```

---

## Configuration

### Configuration File

**Location**: `~/.config/gcal-cli/config.yaml`

**Full Configuration**:
```yaml
# Calendar settings
calendar:
  default_calendar_id: "primary"
  default_timezone: "America/New_York"

# Output preferences
output:
  default_format: "json"        # json|text|minimal
  color_enabled: false          # terminal colors
  pretty_print: true            # format JSON

# Authentication
auth:
  credentials_path: "~/.config/gcal-cli/credentials.json"
  tokens_path: "~/.config/gcal-cli/tokens.json"
  auto_refresh: true

# API settings
api:
  retry_attempts: 3
  retry_delay_ms: 1000
  retry_max_delay_ms: 10000
  timeout_seconds: 30
  rate_limit_buffer: 0.9        # use 90% of rate limit

# Event defaults
events:
  default_duration_minutes: 60
  default_reminder_minutes: 10
  send_notifications: true
```

### Environment Variables

Override configuration with environment variables:

```bash
# Format: GCAL_<SECTION>_<KEY>
export GCAL_OUTPUT_DEFAULT_FORMAT=json
export GCAL_CALENDAR_DEFAULT_CALENDAR_ID=primary
export GCAL_API_TIMEOUT_SECONDS=60
```

### Configuration Commands

```bash
# Initialize default config
gcal-cli config init

# Show current configuration
gcal-cli config show

# Show specific section
gcal-cli config show --format json | jq '.data.config.output'

# Set value
gcal-cli config set output.default_format text
gcal-cli config set api.timeout_seconds 60
```

---

## Output Formats

### JSON Format (Default - LLM Optimized)

**Structured, parseable output for automation:**

```bash
gcal-cli events list --from "2024-01-15" --to "2024-01-20" --format json
```

**Success Response**:
```json
{
  "success": true,
  "operation": "list",
  "data": {
    "events": [
      {
        "id": "abc123xyz",
        "summary": "Team Meeting",
        "start": {
          "dateTime": "2024-01-15T10:00:00-05:00",
          "timeZone": "America/New_York"
        },
        "end": {
          "dateTime": "2024-01-15T11:00:00-05:00",
          "timeZone": "America/New_York"
        },
        "status": "confirmed"
      }
    ],
    "count": 1
  },
  "metadata": {
    "calendarId": "primary",
    "timestamp": "2024-01-15T09:30:00Z"
  }
}
```

**Error Response**:
```json
{
  "success": false,
  "error": {
    "code": "INVALID_INPUT",
    "message": "Start time must be before end time",
    "details": "start: 2024-01-15T14:00:00, end: 2024-01-15T13:00:00",
    "recoverable": true,
    "suggestedAction": "Adjust the --start and --end times"
  }
}
```

### Text Format (Human-Readable)

**Pretty-printed output for humans:**

```bash
gcal-cli events list --from "2024-01-15" --to "2024-01-20" --format text
```

**Output**:
```
Events from 2024-01-15 to 2024-01-20

Event: Team Meeting
  ID:     abc123xyz
  Start:  Mon, Jan 15 2024 10:00 AM EST
  End:    Mon, Jan 15 2024 11:00 AM EST
  Status: Confirmed

Found 1 event(s)
```

### Minimal Format (IDs Only)

**For piping and scripting:**

```bash
gcal-cli events list --from "2024-01-15" --to "2024-01-20" --format minimal
```

**Output**:
```
abc123xyz
def456uvw
ghi789rst
```

**Usage in Scripts**:
```bash
# Get all event IDs
EVENT_IDS=$(gcal-cli events list --from "2024-01-15" --to "2024-01-20" --format minimal)

# Delete all events
for id in $EVENT_IDS; do
  gcal-cli events delete $id --confirm
done
```

---

## Error Handling

### Error Codes

| Code | Description | Suggested Action |
|------|-------------|------------------|
| `AUTH_FAILED` | Authentication failed | Run `gcal-cli auth login` |
| `TOKEN_EXPIRED` | Token has expired | Re-authenticate |
| `INVALID_INPUT` | Invalid input value | Provide valid value |
| `MISSING_REQUIRED` | Required field missing | Add required flag |
| `INVALID_TIME_RANGE` | End before start | Fix time range |
| `NOT_FOUND` | Resource not found | Verify ID/resource |
| `RATE_LIMIT` | API rate limit exceeded | Wait and retry |
| `API_ERROR` | Google Calendar API error | Check API status |
| `PERMISSION_DENIED` | Insufficient permissions | Check calendar access |
| `CONFIG_ERROR` | Configuration error | Run `config init` |
| `NETWORK_ERROR` | Network failure | Check connection |

### Error Response Structure

All errors follow this structure:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable message",
    "details": "Additional context",
    "recoverable": true/false,
    "suggestedAction": "Corrective action"
  }
}
```

### Handling Errors in Scripts

```bash
# Capture output and check success
RESULT=$(gcal-cli events create \
  --title "Test" \
  --start "2024-01-15T10:00:00" \
  --end "2024-01-15T11:00:00" \
  --format json 2>&1)

SUCCESS=$(echo "$RESULT" | jq -r '.success // false')

if [ "$SUCCESS" = "true" ]; then
  EVENT_ID=$(echo "$RESULT" | jq -r '.data.event.id')
  echo "Created event: $EVENT_ID"
else
  ERROR_CODE=$(echo "$RESULT" | jq -r '.error.code')
  ERROR_MSG=$(echo "$RESULT" | jq -r '.error.message')
  echo "Error: $ERROR_CODE - $ERROR_MSG"
  exit 1
fi
```

---

## LLM Agent Integration

### Design Principles

gcal-cli is specifically designed for LLM agent integration:

1. **Structured Output**: Consistent JSON schemas
2. **Machine-Parseable Errors**: Error codes and suggested actions
3. **Idempotency**: Safe to retry operations
4. **Stateless**: No session state between commands
5. **Self-Contained**: Each command is independent

### LLM Agent Patterns

#### Pattern 1: Parse JSON Response

```bash
# Create event and extract ID
EVENT_ID=$(gcal-cli events create \
  --title "Automated Meeting" \
  --start "2024-01-15T10:00:00" \
  --end "2024-01-15T11:00:00" \
  --format json | jq -r '.data.event.id')

echo "Created event with ID: $EVENT_ID"
```

#### Pattern 2: Error Handling with Retry

```bash
perform_with_retry() {
  RESULT=$(gcal-cli events create \
    --title "Test Event" \
    --start "2024-01-15T10:00:00" \
    --end "2024-01-15T11:00:00" \
    --format json 2>&1)

  ERROR_CODE=$(echo $RESULT | jq -r '.error.code // empty')

  case $ERROR_CODE in
    RATE_LIMIT)
      echo "Rate limited, waiting..."
      sleep 5
      # Retry once
      gcal-cli events create \
        --title "Test Event" \
        --start "2024-01-15T10:00:00" \
        --end "2024-01-15T11:00:00"
      ;;
    AUTH_FAILED|TOKEN_EXPIRED)
      echo "Re-authenticating..."
      gcal-cli auth login
      # Retry operation
      ;;
    *)
      echo $RESULT
      ;;
  esac
}
```

#### Pattern 3: Verify Authentication Before Operations

```bash
# Check if authenticated
IS_AUTH=$(gcal-cli auth status --format json | jq -r '.data.authenticated // false')

if [ "$IS_AUTH" != "true" ]; then
  echo "Not authenticated. Please run: gcal-cli auth login"
  exit 1
fi

# Proceed with operation
gcal-cli events list --from "2024-01-15" --to "2024-01-20"
```

#### Pattern 4: Idempotent Event Creation

```bash
create_event_if_not_exists() {
  TITLE=$1
  START=$2
  END=$3

  # Check if event exists
  EXISTING=$(gcal-cli events list \
    --from "$(date -d "$START" +%Y-%m-%d)" \
    --to "$(date -d "$END" +%Y-%m-%d)" \
    --query "$TITLE" \
    --format json | jq -r '.data.events[] | select(.summary == "'$TITLE'") | .id')

  if [ -z "$EXISTING" ]; then
    # Create new event
    gcal-cli events create \
      --title "$TITLE" \
      --start "$START" \
      --end "$END"
    echo "Created new event"
  else
    echo "Event already exists: $EXISTING"
  fi
}

# Usage
create_event_if_not_exists "Daily Standup" "2024-01-15T09:00:00" "2024-01-15T09:15:00"
```

#### Pattern 5: Batch Operations

```bash
# Process multiple events
gcal-cli events list \
  --from "2024-01-15" \
  --to "2024-01-20" \
  --format json | \
  jq -r '.data.events[] | "\(.id)|\(.summary)|\(.start.dateTime)"' | \
  while IFS='|' read -r id title start; do
    echo "Processing event: $id - $title at $start"
    # Perform operations on each event
  done
```

### JSON Parsing Examples

```bash
# Extract specific fields
gcal-cli events list \
  --from "2024-01-15" \
  --to "2024-01-20" \
  --format json | \
  jq '.data.events[] | {id, summary, start: .start.dateTime}'

# Count events
gcal-cli events list \
  --from "2024-01-15" \
  --to "2024-01-20" \
  --format json | \
  jq '.data.count'

# Filter events by title
gcal-cli events list \
  --from "2024-01-15" \
  --to "2024-01-20" \
  --format json | \
  jq '.data.events[] | select(.summary | contains("Meeting"))'

# Get event IDs only
gcal-cli events list \
  --from "2024-01-15" \
  --to "2024-01-20" \
  --format json | \
  jq -r '.data.events[].id'
```

---

## Troubleshooting

### Common Issues

#### Authentication Problems

**Issue**: "AUTH_FAILED: authentication failed or not configured"

**Solution**:
```bash
# 1. Verify credentials file exists
ls -la ~/.config/gcal-cli/credentials.json

# 2. Check permissions
chmod 600 ~/.config/gcal-cli/credentials.json
chmod 600 ~/.config/gcal-cli/tokens.json

# 3. Re-authenticate
gcal-cli auth login
```

#### Token Expired

**Issue**: "TOKEN_EXPIRED: authentication token has expired"

**Solution**:
```bash
# Tokens auto-refresh, but if needed:
gcal-cli auth login
```

#### Rate Limits

**Issue**: "RATE_LIMIT: API rate limit exceeded"

**Solution**:
```bash
# Wait 5-10 seconds and retry
sleep 5
gcal-cli events list --from "2024-01-15" --to "2024-01-20"
```

#### Network Errors

**Issue**: "NETWORK_ERROR: network communication failed"

**Solution**:
```bash
# 1. Check internet connectivity
ping google.com

# 2. Try with increased timeout
export GCAL_API_TIMEOUT_SECONDS=60
gcal-cli events list --from "2024-01-15" --to "2024-01-20"
```

#### Permission Denied

**Issue**: "PERMISSION_DENIED: insufficient permissions for calendar"

**Solution**:
```bash
# List accessible calendars
gcal-cli calendars list --format json

# Use calendar you have access to
gcal-cli events list \
  --calendar-id "your-calendar-id" \
  --from "2024-01-15" \
  --to "2024-01-20"
```

### Debug Mode

Enable verbose logging:

```bash
# Set log level via environment
export GCAL_LOG_LEVEL=debug

# Run command
gcal-cli events list --from "2024-01-15" --to "2024-01-20"
```

### Get Help

```bash
# General help
gcal-cli --help

# Command help
gcal-cli events create --help

# Version info
gcal-cli version

# Configuration
gcal-cli config show
```

### Report Issues

1. **Gather Information**:
   ```bash
   gcal-cli version
   gcal-cli config show
   ```

2. **Capture Error Output**:
   ```bash
   gcal-cli <command> --format json 2>&1
   ```

3. **Report**: [GitHub Issues](https://github.com/btafoya/gcal-cli/issues)

---

## Additional Resources

### Documentation

- **README.md** - Project overview and quick start
- **PLAN.md** - Complete implementation plan
- **SCHEMAS.md** - JSON schema documentation
- **TROUBLESHOOTING.md** - Detailed troubleshooting guide
- **PHASE6_COMPLETE.md** - Testing & documentation phase
- **PHASE7_COMPLETE.md** - Advanced features phase

### Examples

- **pkg/examples/examples.go** - Comprehensive usage examples
- All commands include built-in examples via `--help`

### API Reference

- [Google Calendar API v3](https://developers.google.com/calendar/api/v3/reference)
- [OAuth2 for Go](https://pkg.go.dev/golang.org/x/oauth2)

---

## Quick Reference

### Essential Commands

```bash
# Authentication
gcal-cli auth login
gcal-cli auth status

# Create event
gcal-cli events create --title "Meeting" --start "2024-01-15T10:00:00" --end "2024-01-15T11:00:00"

# List events
gcal-cli events list --from "2024-01-15" --to "2024-01-20"

# Get event
gcal-cli events get <event-id>

# Update event
gcal-cli events update <event-id> --title "New Title"

# Delete event
gcal-cli events delete <event-id>

# List calendars
gcal-cli calendars list

# Configuration
gcal-cli config show
```

### Time Format Examples

```bash
# RFC3339 (preferred)
"2024-01-15T10:00:00Z"
"2024-01-15T10:00:00-05:00"

# ISO-like
"2024-01-15T10:00:00"
"2024-01-15T10:00"

# Space-separated
"2024-01-15 10:00:00"
"2024-01-15 10:00"

# Natural language (Phase 7)
"tomorrow at 2pm"
"next Monday at 9am"
"in 2 hours"
```

### Output Format Selection

```bash
# JSON (default - for LLM agents)
--format json

# Text (human-readable)
--format text

# Minimal (IDs only)
--format minimal
```

---

**End of User Instructions**

For additional help:
- Run `gcal-cli --help` for command reference
- See [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) for detailed solutions
- Report issues at [GitHub Issues](https://github.com/btafoya/gcal-cli/issues)
