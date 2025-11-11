# JSON Schemas - gcal-cli

Comprehensive documentation of all JSON schemas for LLM agent integration.

## Core Response Structure

All responses follow a consistent structure for predictable parsing.

### Success Response Schema

```json
{
  "success": true,
  "operation": "<operation_name>",
  "data": {
    // Operation-specific data
  },
  "metadata": {
    "timestamp": "2024-01-15T09:30:00Z",
    // Additional context-specific metadata
  }
}
```

**Fields**:
- `success` (boolean, required): Always `true` for successful operations
- `operation` (string, required): Operation identifier (e.g., "create", "list", "update", "delete")
- `data` (object, required): Operation-specific payload
- `metadata` (object, required): Contextual information including timestamp

### Error Response Schema

```json
{
  "success": false,
  "error": {
    "code": "<ERROR_CODE>",
    "message": "Human-readable error message",
    "details": "Additional context about the error",
    "recoverable": true,
    "suggestedAction": "Specific action to resolve the error"
  },
  "metadata": {
    "timestamp": "2024-01-15T09:30:00Z"
  }
}
```

**Fields**:
- `success` (boolean, required): Always `false` for errors
- `error` (object, required): Error information
  - `code` (string, required): Machine-readable error code (see Error Codes section)
  - `message` (string, required): Human-readable error description
  - `details` (string, optional): Additional error context
  - `recoverable` (boolean, required): Whether error can be recovered from
  - `suggestedAction` (string, optional): Recommended resolution step
- `metadata` (object, required): Contextual information including timestamp

## Error Codes

All error codes with descriptions and suggested actions.

### Authentication Errors

| Code | Description | Recoverable | Suggested Action |
|------|-------------|-------------|------------------|
| `AUTH_FAILED` | Authentication failed | Yes | Run 'gcal-cli auth login' |
| `TOKEN_EXPIRED` | Auth token expired | Yes | Run 'gcal-cli auth login' |
| `INVALID_CREDENTIALS` | Invalid OAuth credentials | No | Check credentials file |

### Input Validation Errors

| Code | Description | Recoverable | Suggested Action |
|------|-------------|-------------|------------------|
| `INVALID_INPUT` | Invalid parameter value | Yes | Check parameter format |
| `MISSING_REQUIRED` | Required parameter missing | Yes | Add required parameter |
| `INVALID_FORMAT` | Incorrect format | Yes | Use correct format (see docs) |
| `INVALID_TIME_RANGE` | Invalid time range | Yes | Ensure start before end |

### API Errors

| Code | Description | Recoverable | Suggested Action |
|------|-------------|-------------|------------------|
| `NOT_FOUND` | Resource not found | No | Verify resource ID |
| `RATE_LIMIT` | API rate limit exceeded | Yes | Wait and retry |
| `API_ERROR` | Google API error | Maybe | Check Google Calendar status |
| `PERMISSION_DENIED` | Insufficient permissions | No | Check calendar sharing settings |

### System Errors

| Code | Description | Recoverable | Suggested Action |
|------|-------------|-------------|------------------|
| `CONFIG_ERROR` | Configuration error | Yes | Check config file |
| `NETWORK_ERROR` | Network connectivity issue | Yes | Check internet connection |
| `FILE_ERROR` | File operation error | Yes | Check file permissions |

## Operation Schemas

### Authentication Operations

#### auth login

**Success Response**:
```json
{
  "success": true,
  "operation": "auth_login",
  "data": {
    "message": "Authentication successful",
    "email": "user@example.com",
    "scopes": ["https://www.googleapis.com/auth/calendar"]
  },
  "metadata": {
    "timestamp": "2024-01-15T09:30:00Z"
  }
}
```

#### auth status

**Success Response**:
```json
{
  "success": true,
  "operation": "auth_status",
  "data": {
    "authenticated": true,
    "email": "user@example.com",
    "tokenExpiry": "2024-01-20T10:00:00Z",
    "scopes": ["https://www.googleapis.com/auth/calendar"]
  },
  "metadata": {
    "timestamp": "2024-01-15T09:30:00Z"
  }
}
```

#### auth logout

**Success Response**:
```json
{
  "success": true,
  "operation": "auth_logout",
  "data": {
    "message": "Logged out successfully"
  },
  "metadata": {
    "timestamp": "2024-01-15T09:30:00Z"
  }
}
```

### Event Operations

#### events create

**Success Response**:
```json
{
  "success": true,
  "operation": "create",
  "data": {
    "event": {
      "id": "abc123xyz",
      "summary": "Team Meeting",
      "description": "Weekly sync meeting",
      "location": "Conference Room A",
      "status": "confirmed",
      "start": {
        "dateTime": "2024-01-15T10:00:00-05:00",
        "timeZone": "America/New_York",
        "date": null
      },
      "end": {
        "dateTime": "2024-01-15T11:00:00-05:00",
        "timeZone": "America/New_York",
        "date": null
      },
      "attendees": [
        {
          "email": "user1@example.com",
          "responseStatus": "needsAction",
          "organizer": false
        }
      ],
      "recurrence": ["RRULE:FREQ=WEEKLY;COUNT=10"],
      "htmlLink": "https://www.google.com/calendar/event?eid=...",
      "created": "2024-01-15T09:30:00Z",
      "updated": "2024-01-15T09:30:00Z"
    },
    "message": "Event created successfully"
  },
  "metadata": {
    "timestamp": "2024-01-15T09:30:00Z",
    "calendarId": "primary"
  }
}
```

**Event Object Schema**:
```typescript
{
  id: string,                    // Google Calendar event ID
  summary: string,               // Event title
  description?: string,          // Event description (optional)
  location?: string,             // Event location (optional)
  status: "confirmed" | "tentative" | "cancelled",
  start: EventTime,              // Start time
  end: EventTime,                // End time
  attendees?: Attendee[],        // Attendees list (optional)
  recurrence?: string[],         // Recurrence rules (optional)
  htmlLink?: string,             // Google Calendar link (optional)
  created?: string,              // Creation timestamp (optional)
  updated?: string               // Last update timestamp (optional)
}
```

**EventTime Schema**:
```typescript
{
  dateTime?: string,             // RFC3339 datetime (for timed events)
  timeZone?: string,             // IANA timezone
  date?: string                  // YYYY-MM-DD date (for all-day events)
}
```

**Attendee Schema**:
```typescript
{
  email: string,                 // Email address
  responseStatus: "needsAction" | "accepted" | "declined" | "tentative",
  organizer?: boolean            // True if event organizer
}
```

#### events list

**Success Response**:
```json
{
  "success": true,
  "operation": "list",
  "data": {
    "events": [
      {
        // Event object (see Event Object Schema)
      }
    ],
    "count": 2,
    "nextPageToken": null
  },
  "metadata": {
    "timestamp": "2024-01-15T09:30:00Z",
    "calendarId": "primary",
    "from": "2024-01-15T00:00:00Z",
    "to": "2024-01-20T23:59:59Z"
  }
}
```

#### events get

**Success Response**:
```json
{
  "success": true,
  "operation": "get",
  "data": {
    "event": {
      // Event object (see Event Object Schema)
    }
  },
  "metadata": {
    "timestamp": "2024-01-15T09:30:00Z",
    "calendarId": "primary"
  }
}
```

#### events update

**Success Response**:
```json
{
  "success": true,
  "operation": "update",
  "data": {
    "event": {
      // Updated event object (see Event Object Schema)
    },
    "message": "Event updated successfully"
  },
  "metadata": {
    "timestamp": "2024-01-15T09:30:00Z",
    "calendarId": "primary"
  }
}
```

#### events delete

**Success Response**:
```json
{
  "success": true,
  "operation": "delete",
  "data": {
    "eventId": "abc123xyz",
    "message": "Event deleted successfully"
  },
  "metadata": {
    "timestamp": "2024-01-15T09:30:00Z",
    "calendarId": "primary"
  }
}
```

### Calendar Operations

#### calendars list

**Success Response**:
```json
{
  "success": true,
  "operation": "list_calendars",
  "data": {
    "calendars": [
      {
        "id": "primary",
        "summary": "user@example.com",
        "description": "Primary calendar",
        "timeZone": "America/New_York",
        "primary": true,
        "accessRole": "owner"
      }
    ],
    "count": 1
  },
  "metadata": {
    "timestamp": "2024-01-15T09:30:00Z"
  }
}
```

**Calendar Object Schema**:
```typescript
{
  id: string,                    // Calendar ID
  summary: string,               // Calendar name
  description?: string,          // Calendar description (optional)
  timeZone: string,              // IANA timezone
  primary?: boolean,             // True if primary calendar
  accessRole: "owner" | "reader" | "writer" | "freeBusyReader"
}
```

#### calendars get

**Success Response**:
```json
{
  "success": true,
  "operation": "get_calendar",
  "data": {
    "calendar": {
      "id": "primary",
      "summary": "user@example.com",
      "description": "Primary calendar",
      "timeZone": "America/New_York"
    }
  },
  "metadata": {
    "timestamp": "2024-01-15T09:30:00Z"
  }
}
```

### Configuration Operations

#### config show

**Success Response**:
```json
{
  "success": true,
  "operation": "config_show",
  "data": {
    "config": {
      "calendar": {
        "default_calendar_id": "primary",
        "default_timezone": "America/New_York"
      },
      "output": {
        "default_format": "json",
        "color_enabled": false
      },
      "auth": {
        "credentials_path": "~/.config/gcal-cli/credentials.json",
        "tokens_path": "~/.config/gcal-cli/tokens.json"
      },
      "api": {
        "retry_attempts": 3,
        "retry_delay_ms": 1000,
        "timeout_seconds": 30
      }
    }
  },
  "metadata": {
    "timestamp": "2024-01-15T09:30:00Z",
    "configPath": "~/.config/gcal-cli/config.yaml"
  }
}
```

## Parsing Guidelines for LLM Agents

### Success Detection

```python
response = json.loads(output)
if response["success"]:
    # Handle success
    data = response["data"]
    operation = response["operation"]
else:
    # Handle error
    error = response["error"]
    code = error["code"]
    message = error["message"]
```

### Error Handling

```python
if not response["success"]:
    error_code = response["error"]["code"]

    if error_code == "AUTH_FAILED" or error_code == "TOKEN_EXPIRED":
        # Re-authenticate
        run_command("gcal-cli auth login")

    elif error_code == "RATE_LIMIT":
        # Wait and retry
        time.sleep(5)
        retry_operation()

    elif error_code in ["INVALID_INPUT", "MISSING_REQUIRED"]:
        # Fix input and retry
        suggested_action = response["error"]["suggestedAction"]
        # Parse and apply suggested action

    elif not response["error"]["recoverable"]:
        # Permanent error, don't retry
        log_error(response["error"]["message"])
```

### Event Extraction

```python
# From create/update/get responses
event = response["data"]["event"]
event_id = event["id"]
title = event["summary"]
start_time = event["start"]["dateTime"]

# From list responses
events = response["data"]["events"]
for event in events:
    process_event(event)
```

### Pagination

```python
# Check for more pages
next_page_token = response["data"].get("nextPageToken")
if next_page_token:
    # Fetch next page (not yet implemented in CLI)
    pass
```

## Schema Validation

All responses can be validated against these schemas. Required fields MUST be present. Optional fields MAY be present.

### Required Fields by Operation

| Operation | Required Response Fields |
|-----------|-------------------------|
| All | `success`, `metadata.timestamp` |
| Success | `operation`, `data` |
| Error | `error.code`, `error.message`, `error.recoverable` |
| create | `data.event`, `data.event.id`, `data.message` |
| list | `data.events`, `data.count` |
| get | `data.event` |
| update | `data.event`, `data.message` |
| delete | `data.eventId`, `data.message` |

## Idempotency

### Idempotent Operations

- `GET` operations (events get, calendars get, etc.)
- `DELETE` operations (returns success even if already deleted)
- `auth status`
- `config show`

### Non-Idempotent Operations

- `events create` - Creates new event each time
- `events update` - Last write wins
- `auth login` - Creates new token each time

### Idempotency Best Practices

For LLM agents implementing idempotent workflows:

1. **Check before create**: Use `events list` to check if event exists before creating
2. **Use event IDs**: Store and reuse event IDs for updates
3. **Verify deletions**: Check `events get` after delete to confirm

## Version Compatibility

**Schema Version**: 1.0
**Last Updated**: 2025-11-10
**Backward Compatibility**: All schema changes will be backward compatible
**Deprecation Policy**: Deprecated fields will be marked 6 months before removal
