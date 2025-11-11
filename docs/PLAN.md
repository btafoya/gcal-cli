# PLAN.md - Google Calendar CLI Implementation Plan

## Project Overview

A Google Calendar CLI tool written in Go, specifically designed for LLM agent integration. This tool provides programmatic access to Google Calendar operations through a command-line interface optimized for machine-readable interactions.

### Core Objectives

1. **LLM-First Design**: Structured JSON output, clear error codes, idempotent operations
2. **Complete CRUD**: Full event lifecycle management (Create, Read, Update, Delete)
3. **Seamless Auth**: OAuth2 with automatic token refresh
4. **Agent-Friendly**: Stateless commands, self-contained operations, parseable responses

## User Stories

### LLM Agent Perspectives

- **US-1**: As an LLM agent, I need to create calendar events with structured JSON responses so I can parse results reliably
- **US-2**: As an LLM agent, I need to list events in a date range with consistent schema so I can process multiple events
- **US-3**: As an LLM agent, I need clear error messages with codes so I can handle failures appropriately
- **US-4**: As an LLM agent, I need to update and delete events by ID with confirmation responses
- **US-5**: As an LLM agent, I need idempotent operations so repeated commands don't cause unexpected duplicates

### User Perspectives

- **US-6**: As a user, I need OAuth authentication to be handled automatically so agents don't need manual intervention
- **US-7**: As a user, I need configuration to persist across sessions so I don't have to re-authenticate constantly
- **US-8**: As a user, I need clear documentation on how to set up credentials so I can get started quickly

## Architecture

### High-Level Component Structure

```
gcal-cli/
├── cmd/
│   └── gcal-cli/
│       ├── main.go              # Application entry point
│       └── root.go              # Root Cobra command setup
├── pkg/
│   ├── auth/                    # Authentication layer
│   │   ├── oauth.go            # OAuth2 flow implementation
│   │   ├── token.go            # Token management & refresh
│   │   └── storage.go          # Secure token storage
│   ├── calendar/                # Google Calendar operations
│   │   ├── client.go           # Calendar API client wrapper
│   │   ├── events.go           # Event CRUD operations
│   │   ├── list.go             # List/search functionality
│   │   └── errors.go           # API error handling
│   ├── config/                  # Configuration management
│   │   ├── config.go           # Viper integration
│   │   ├── defaults.go         # Default values
│   │   └── validate.go         # Config validation
│   ├── output/                  # Output formatting
│   │   ├── json.go             # JSON formatter
│   │   ├── text.go             # Human-readable formatter
│   │   ├── minimal.go          # Minimal output (IDs only)
│   │   └── schema.go           # Response schemas
│   └── types/                   # Shared types
│       ├── event.go            # Event structures
│       ├── response.go         # Response schemas
│       └── errors.go           # Error types & codes
├── internal/
│   └── commands/                # Cobra command implementations
│       ├── auth.go             # Auth commands
│       ├── events.go           # Event commands
│       └── config.go           # Config commands
└── test/
    ├── integration/             # Integration tests
    └── fixtures/                # Test data
```

### Component Responsibilities

#### 1. cmd/gcal-cli - Main Application

**Purpose**: Application entry point and Cobra command orchestration

**Key Files**:
- `main.go` - Initializes application, executes root command
- `root.go` - Defines root command, global flags, configuration loading

**Dependencies**:
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management

#### 2. pkg/auth - Authentication Layer

**Purpose**: OAuth2 authentication and token management

**Responsibilities**:
- OAuth2 authorization code flow
- Token storage in `~/.config/gcal-cli/tokens.json`
- Automatic token refresh before expiry
- Credentials validation

**Key Patterns** (from Context7):
```go
// OAuth2 configuration
config := &oauth2.Config{
    ClientID:     viper.GetString("client_id"),
    ClientSecret: viper.GetString("client_secret"),
    Endpoint:     google.Endpoint,
    Scopes:       []string{calendar.CalendarScope},
}

// Token source with automatic refresh
tokenSource := config.TokenSource(ctx, token)
client := oauth2.NewClient(ctx, tokenSource)
```

**Security**:
- Token files stored with 0600 permissions
- Credentials never logged or printed
- In-memory token handling during refresh

#### 3. pkg/calendar - Google Calendar Operations

**Purpose**: Wrapper around Google Calendar API v3

**Responsibilities**:
- Event creation with validation
- Event listing with filtering (date range, max results)
- Event retrieval by ID
- Event updates (partial and full)
- Event deletion
- Retry logic for transient failures
- Rate limit handling

**Key Patterns** (from Context7):
```go
// Initialize service
service, err := calendar.NewService(ctx,
    option.WithHTTPClient(authClient))

// Create event
event := &calendar.Event{
    Summary:     "Team Meeting",
    Description: "Weekly sync",
    Start: &calendar.EventDateTime{
        DateTime: startTime,
        TimeZone: timezone,
    },
    End: &calendar.EventDateTime{
        DateTime: endTime,
        TimeZone: timezone,
    },
}

createdEvent, err := service.Events.
    Insert(calendarID, event).
    Do()
```

**Error Handling**:
- Exponential backoff for rate limits
- Clear error messages for validation failures
- Wrap Google API errors with context

#### 4. pkg/config - Configuration Management

**Purpose**: Centralized configuration using Viper

**Configuration Hierarchy** (from Context7 patterns):
1. Command-line flags (highest priority)
2. Environment variables (`GCAL_*`)
3. Config file (`~/.config/gcal-cli/config.yaml`)
4. Default values (lowest priority)

**Config Schema**:
```yaml
# ~/.config/gcal-cli/config.yaml
calendar:
  default_calendar_id: "primary"
  default_timezone: "America/New_York"

output:
  default_format: "json"  # json|text|minimal
  color_enabled: false    # for text output

auth:
  credentials_path: "~/.config/gcal-cli/credentials.json"
  tokens_path: "~/.config/gcal-cli/tokens.json"

api:
  retry_attempts: 3
  retry_delay_ms: 1000
  timeout_seconds: 30
```

**Viper Integration** (from Context7):
```go
// Watch config for changes
viper.WatchConfig()
viper.OnConfigChange(func(e fsnotify.Event) {
    log.Printf("Config file changed: %s", e.Name)
})

// Bind flags to config
viper.BindPFlag("output.format",
    rootCmd.PersistentFlags().Lookup("format"))
```

#### 5. pkg/output - Output Formatting

**Purpose**: Consistent, parseable output for LLM agents

**Formatters**:

**JSON Format** (Default for LLMs):
```json
{
  "success": true,
  "operation": "create",
  "data": {
    "event": {
      "id": "abc123xyz",
      "summary": "Team Meeting",
      "start": "2024-01-15T10:00:00-05:00",
      "end": "2024-01-15T11:00:00-05:00",
      "status": "confirmed",
      "attendees": [
        {
          "email": "user@example.com",
          "responseStatus": "needsAction"
        }
      ]
    },
    "message": "Event created successfully"
  },
  "metadata": {
    "calendarId": "primary",
    "timestamp": "2024-01-15T09:30:00Z",
    "timezone": "America/New_York"
  }
}
```

**Error Format**:
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

**Text Format** (Human-readable):
```
✓ Event created successfully

Event Details:
  ID:        abc123xyz
  Title:     Team Meeting
  Start:     Mon, Jan 15 2024 10:00 AM EST
  End:       Mon, Jan 15 2024 11:00 AM EST
  Status:    Confirmed
  Calendar:  primary

Attendees:
  • user@example.com (awaiting response)
```

**Minimal Format** (IDs only for piping):
```
abc123xyz
```

#### 6. pkg/types - Shared Types

**Purpose**: Common data structures across packages

**Event Schema**:
```go
type Event struct {
    ID          string       `json:"id"`
    Summary     string       `json:"summary"`
    Description string       `json:"description,omitempty"`
    Start       EventTime    `json:"start"`
    End         EventTime    `json:"end"`
    Status      string       `json:"status"`
    Attendees   []Attendee   `json:"attendees,omitempty"`
    Recurrence  []string     `json:"recurrence,omitempty"`
    Location    string       `json:"location,omitempty"`
}

type EventTime struct {
    DateTime string `json:"dateTime"`
    TimeZone string `json:"timeZone"`
}

type Attendee struct {
    Email          string `json:"email"`
    ResponseStatus string `json:"responseStatus"`
    Organizer      bool   `json:"organizer,omitempty"`
}
```

**Error Codes**:
```go
const (
    ErrCodeAuthFailed    = "AUTH_FAILED"
    ErrCodeNotFound      = "NOT_FOUND"
    ErrCodeInvalidInput  = "INVALID_INPUT"
    ErrCodeRateLimit     = "RATE_LIMIT"
    ErrCodeAPIError      = "API_ERROR"
    ErrCodeConfigError   = "CONFIG_ERROR"
    ErrCodeNetworkError  = "NETWORK_ERROR"
)
```

## Command Structure

### Command Hierarchy

Based on Cobra best practices from Context7:

```
gcal-cli
├── auth                         # Authentication commands
│   ├── login                    # Initiate OAuth flow
│   ├── logout                   # Clear stored tokens
│   └── status                   # Check auth status
├── events                       # Event management
│   ├── create                   # Create new event
│   ├── list                     # List events in range
│   ├── get                      # Get single event
│   ├── update                   # Update event
│   └── delete                   # Delete event
└── config                       # Configuration
    ├── init                     # Initialize config
    ├── show                     # Display config
    └── set                      # Set config value
```

### Global Flags

```go
// Persistent flags (available to all commands)
rootCmd.PersistentFlags().String("format", "json",
    "Output format (json|text|minimal)")
rootCmd.PersistentFlags().String("calendar-id", "primary",
    "Calendar ID to operate on")
rootCmd.PersistentFlags().String("config", "",
    "Config file path (default: ~/.config/gcal-cli/config.yaml)")
rootCmd.PersistentFlags().String("timezone", "",
    "Timezone for operations (default: system timezone)")

// Bind to Viper
viper.BindPFlag("output.format",
    rootCmd.PersistentFlags().Lookup("format"))
viper.BindPFlag("calendar.default_calendar_id",
    rootCmd.PersistentFlags().Lookup("calendar-id"))
```

### Command Details

#### auth login

**Purpose**: Initiate OAuth2 flow and store tokens

**Usage**:
```bash
gcal-cli auth login [--credentials <path>]
```

**Flags**:
- `--credentials` - Path to credentials.json (default: ~/.config/gcal-cli/credentials.json)

**Flow**:
1. Load OAuth2 credentials from file
2. Start local HTTP server for callback
3. Open browser for user authorization
4. Receive authorization code via callback
5. Exchange code for tokens
6. Store tokens securely
7. Output success message

**Output** (JSON):
```json
{
  "success": true,
  "operation": "auth_login",
  "data": {
    "message": "Authentication successful",
    "email": "user@example.com",
    "scopes": ["https://www.googleapis.com/auth/calendar"]
  }
}
```

#### events create

**Purpose**: Create a new calendar event

**Usage**:
```bash
gcal-cli events create \
  --title "Team Meeting" \
  --start "2024-01-15T10:00:00" \
  --end "2024-01-15T11:00:00" \
  [--description "Weekly sync"] \
  [--location "Conference Room A"] \
  [--attendees "user1@example.com,user2@example.com"] \
  [--recurrence "RRULE:FREQ=WEEKLY;COUNT=10"]
```

**Flags**:
- `--title` (required) - Event title/summary
- `--start` (required) - Start time (RFC3339 or common formats)
- `--end` (required) - End time (RFC3339 or common formats)
- `--description` - Event description
- `--location` - Event location
- `--attendees` - Comma-separated email addresses
- `--recurrence` - Recurrence rule (RFC5545 format)
- `--all-day` - Create all-day event

**Validation**:
- Start time must be before end time
- Email addresses must be valid format
- Recurrence rule must be valid RFC5545

**Output**: Event object with ID and all details

#### events list

**Purpose**: List events in a date range

**Usage**:
```bash
gcal-cli events list \
  --from "2024-01-15" \
  --to "2024-01-20" \
  [--max-results 50] \
  [--query "meeting"]
```

**Flags**:
- `--from` (required) - Start date (YYYY-MM-DD or RFC3339)
- `--to` (required) - End date (YYYY-MM-DD or RFC3339)
- `--max-results` - Maximum events to return (default: 250)
- `--query` - Search query string
- `--order-by` - Sort order (startTime|updated)

**Output**:
```json
{
  "success": true,
  "operation": "list",
  "data": {
    "events": [
      { /* event object */ },
      { /* event object */ }
    ],
    "count": 2,
    "nextPageToken": null
  },
  "metadata": {
    "calendarId": "primary",
    "from": "2024-01-15T00:00:00Z",
    "to": "2024-01-20T23:59:59Z"
  }
}
```

#### events get

**Purpose**: Retrieve single event by ID

**Usage**:
```bash
gcal-cli events get <event-id>
```

**Arguments**:
- `event-id` (required) - Google Calendar event ID

**Output**: Single event object

#### events update

**Purpose**: Update existing event

**Usage**:
```bash
gcal-cli events update <event-id> \
  [--title "Updated Title"] \
  [--start "2024-01-15T11:00:00"] \
  [--end "2024-01-15T12:00:00"] \
  [--description "Updated description"]
```

**Arguments**:
- `event-id` (required) - Event ID to update

**Flags**: Same as create (all optional)

**Behavior**:
- Only specified fields are updated (partial update)
- Unspecified fields remain unchanged
- Returns updated event object

#### events delete

**Purpose**: Delete event

**Usage**:
```bash
gcal-cli events delete <event-id> [--confirm]
```

**Arguments**:
- `event-id` (required) - Event ID to delete

**Flags**:
- `--confirm` - Skip confirmation prompt

**Output**:
```json
{
  "success": true,
  "operation": "delete",
  "data": {
    "eventId": "abc123xyz",
    "message": "Event deleted successfully"
  }
}
```

## Implementation Phases

### Phase 1: Foundation (Week 1)

**Goals**: Project structure, CLI framework, output formatting

**Tasks**:
1. Initialize Go module: `go mod init github.com/yourusername/gcal-cli`
2. Install dependencies:
   ```bash
   go get github.com/spf13/cobra@latest
   go get github.com/spf13/viper@latest
   go get google.golang.org/api/calendar/v3
   go get golang.org/x/oauth2/google
   ```
3. Create directory structure
4. Implement root command with Cobra
5. Set up Viper configuration loading
6. Implement output formatters (JSON, text, minimal)
7. Create error type system with codes
8. Write unit tests for output formatting

**Deliverables**:
- ✓ Project builds successfully
- ✓ Basic command structure in place
- ✓ Configuration loading works
- ✓ Output formatters produce correct schemas
- ✓ Error handling infrastructure ready

**Success Metrics**:
- `go build` completes without errors
- `gcal-cli --help` shows command structure
- Config file loads and merges with defaults
- All output formatters have >90% test coverage

### Phase 2: Authentication (Week 1-2)

**Goals**: OAuth2 flow, token management, secure storage

**Tasks**:
1. Implement OAuth2 configuration from credentials file
2. Create local HTTP server for OAuth callback
3. Implement authorization URL generation
4. Handle authorization code exchange
5. Implement token storage with file permissions
6. Add automatic token refresh logic
7. Create `auth login`, `auth status`, `auth logout` commands
8. Add token validation and expiry checking
9. Write integration tests for auth flow

**Deliverables**:
- ✓ Users can authenticate via browser
- ✓ Tokens stored securely and persist across sessions
- ✓ Tokens refresh automatically before expiry
- ✓ Clear error messages for auth failures

**Success Metrics**:
- Complete OAuth flow works end-to-end
- Token files have 0600 permissions
- Expired tokens refresh transparently
- Auth status command shows accurate state

### Phase 3: Core Calendar Operations (Week 2-3)

**Goals**: Event CRUD operations with Google Calendar API

**Tasks**:
1. Initialize Google Calendar API client
2. Implement event creation
3. Implement event listing with date range filtering
4. Implement event retrieval by ID
5. Add input validation for all operations
6. Implement retry logic with exponential backoff
7. Add rate limit handling
8. Create `events create`, `events list`, `events get` commands
9. Write unit tests with mocked API responses
10. Write integration tests with test calendar

**Deliverables**:
- ✓ Events can be created via CLI
- ✓ Events can be listed with filtering
- ✓ Single events can be retrieved
- ✓ API errors handled gracefully
- ✓ Retry logic works for transient failures

**Success Metrics**:
- Event creation returns valid event ID
- List operations return correct date ranges
- Get operations retrieve correct events
- API rate limits handled without crashes
- >80% code coverage for calendar package

### Phase 4: Advanced Operations (Week 3)

**Goals**: Event updates, deletion, attendees, recurrence

**Tasks**:
1. Implement event update (partial and full)
2. Implement event deletion
3. Add attendee management
4. Add recurrence rule support
5. Implement all-day event handling
6. Add timezone handling and conversion
7. Create `events update`, `events delete` commands
8. Add batch operation support
9. Write comprehensive tests

**Deliverables**:
- ✓ Events can be updated (partial and full)
- ✓ Events can be deleted
- ✓ Attendees can be added/removed
- ✓ Recurring events can be created
- ✓ Timezones handled correctly

**Success Metrics**:
- Update operations modify only specified fields
- Deletion returns proper confirmation
- Attendee email validation works
- Recurrence rules validate correctly
- Timezone conversions are accurate

### Phase 5: LLM Optimization (Week 4)

**Goals**: Refine for LLM agent consumption

**Tasks**:
1. Audit all JSON schemas for consistency
2. Ensure all error messages are machine-parseable
3. Add validation for all inputs with clear error messages
4. Implement idempotency where possible
5. Add comprehensive examples to help text
6. Create integration test suite simulating LLM usage
7. Add performance optimizations
8. Document JSON schemas

**Deliverables**:
- ✓ All responses follow consistent schema
- ✓ Error codes comprehensive and documented
- ✓ Input validation provides actionable feedback
- ✓ Performance meets targets (<2s for most operations)

**Success Metrics**:
- All JSON responses validate against schema
- Error messages include suggestedAction field
- Validation errors specify exact issue
- 95th percentile latency <2s

### Phase 6: Testing & Documentation (Week 4)

**Goals**: Comprehensive testing and documentation

**Tasks**:
1. Achieve >80% code coverage
2. Write integration tests for all commands
3. Create end-to-end test scenarios
4. Document all commands in README
5. Create usage examples for LLM agents
6. Write troubleshooting guide
7. Add inline code documentation
8. Create demo video/GIF

**Deliverables**:
- ✓ Comprehensive test suite
- ✓ Complete documentation
- ✓ Usage examples for common scenarios
- ✓ Troubleshooting guide

**Success Metrics**:
- >80% test coverage across all packages
- All commands documented with examples
- Zero critical bugs in issue tracker
- Documentation reviewed and approved

## Testing Strategy

### Unit Tests

**Scope**: Individual functions and methods

**Approach**:
- Mock external dependencies (Google Calendar API)
- Test all error paths
- Validate input/output schemas
- Use table-driven tests for multiple cases

**Example** (auth package):
```go
func TestTokenRefresh(t *testing.T) {
    tests := []struct {
        name        string
        token       *oauth2.Token
        expectError bool
    }{
        {
            name: "valid token needing refresh",
            token: &oauth2.Token{
                AccessToken:  "old-token",
                RefreshToken: "refresh-token",
                Expiry:       time.Now().Add(-1 * time.Hour),
            },
            expectError: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Integration Tests

**Scope**: Component interactions

**Approach**:
- Use test Google Calendar (or mock API server)
- Test full command execution paths
- Validate output formatting
- Test configuration loading

**Example**:
```go
func TestEventCreateCommand(t *testing.T) {
    // Set up test environment
    testConfig := setupTestConfig(t)
    mockAPI := setupMockCalendarAPI(t)

    // Execute command
    cmd := NewEventsCreateCommand()
    cmd.SetArgs([]string{
        "--title", "Test Event",
        "--start", "2024-01-15T10:00:00",
        "--end", "2024-01-15T11:00:00",
    })

    err := cmd.Execute()
    assert.NoError(t, err)

    // Validate output
    // Validate API calls
}
```

### End-to-End Tests

**Scope**: Full user workflows

**Scenarios**:
1. First-time setup and authentication
2. Create → List → Update → Delete workflow
3. Multiple calendar operations
4. Error recovery scenarios
5. LLM agent simulation

**Example** (LLM agent simulation):
```go
func TestLLMAgentWorkflow(t *testing.T) {
    // Simulate LLM agent creating and managing events

    // Step 1: Create event
    createCmd := exec.Command("gcal-cli", "events", "create",
        "--title", "Agent Meeting",
        "--start", "2024-01-15T10:00:00",
        "--end", "2024-01-15T11:00:00",
        "--format", "json")

    output, err := createCmd.Output()
    assert.NoError(t, err)

    var response map[string]interface{}
    json.Unmarshal(output, &response)

    assert.True(t, response["success"].(bool))
    eventID := response["data"].(map[string]interface{})["event"].(map[string]interface{})["id"].(string)

    // Step 2: List events
    // Step 3: Update event
    // Step 4: Delete event
}
```

## Dependencies

### Required Libraries

```go
require (
    github.com/spf13/cobra v1.9.1           // CLI framework
    github.com/spf13/viper v1.20.1          // Configuration
    google.golang.org/api v0.latest          // Google APIs
    golang.org/x/oauth2 v0.latest            // OAuth2
    github.com/fsnotify/fsnotify v1.latest   // File watching (Viper dependency)
)
```

### Development Dependencies

```go
require (
    github.com/stretchr/testify v1.latest   // Testing utilities
    github.com/golang/mock v1.latest        // Mocking
)
```

## Configuration Management

### Config File Location

Primary: `~/.config/gcal-cli/config.yaml`
Alternative (XDG): `$XDG_CONFIG_HOME/gcal-cli/config.yaml`

### Complete Config Schema

```yaml
# Calendar settings
calendar:
  default_calendar_id: "primary"
  default_timezone: "America/New_York"  # IANA timezone

# Output preferences
output:
  default_format: "json"  # json|text|minimal
  color_enabled: false    # for text output
  pretty_print: true      # format JSON with indentation

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
  rate_limit_buffer: 0.9  # Use 90% of rate limit

# Event defaults
events:
  default_duration_minutes: 60
  default_reminder_minutes: 10
  send_notifications: true
```

### Environment Variables

All config values can be overridden with environment variables:

```bash
GCAL_CALENDAR_DEFAULT_CALENDAR_ID="primary"
GCAL_OUTPUT_DEFAULT_FORMAT="json"
GCAL_AUTH_CREDENTIALS_PATH="~/credentials.json"
```

## Error Handling

### Error Code Definitions

```go
const (
    // Authentication errors
    ErrCodeAuthFailed       = "AUTH_FAILED"
    ErrCodeTokenExpired     = "TOKEN_EXPIRED"
    ErrCodeInvalidCreds     = "INVALID_CREDENTIALS"

    // Input validation errors
    ErrCodeInvalidInput     = "INVALID_INPUT"
    ErrCodeMissingRequired  = "MISSING_REQUIRED"
    ErrCodeInvalidFormat    = "INVALID_FORMAT"
    ErrCodeInvalidTimeRange = "INVALID_TIME_RANGE"

    // API errors
    ErrCodeNotFound         = "NOT_FOUND"
    ErrCodeRateLimit        = "RATE_LIMIT"
    ErrCodeAPIError         = "API_ERROR"
    ErrCodePermissionDenied = "PERMISSION_DENIED"

    // System errors
    ErrCodeConfigError      = "CONFIG_ERROR"
    ErrCodeNetworkError     = "NETWORK_ERROR"
    ErrCodeFileError        = "FILE_ERROR"
)
```

### Error Response Structure

```go
type ErrorResponse struct {
    Success bool   `json:"success"`
    Error   Error  `json:"error"`
}

type Error struct {
    Code            string `json:"code"`
    Message         string `json:"message"`
    Details         string `json:"details,omitempty"`
    Recoverable     bool   `json:"recoverable"`
    SuggestedAction string `json:"suggestedAction,omitempty"`
}
```

### Retry Strategy

**Transient Errors** (retry with exponential backoff):
- Network timeouts
- Rate limit errors (429)
- Server errors (500, 503)

**Permanent Errors** (fail immediately):
- Authentication failures (401)
- Permission denied (403)
- Not found (404)
- Invalid input (400)

**Retry Algorithm**:
```go
func retryWithBackoff(operation func() error, maxAttempts int) error {
    var err error
    delay := 1 * time.Second

    for attempt := 0; attempt < maxAttempts; attempt++ {
        err = operation()

        if err == nil {
            return nil
        }

        if !isRetryable(err) {
            return err
        }

        if attempt < maxAttempts-1 {
            time.Sleep(delay)
            delay *= 2  // Exponential backoff
        }
    }

    return fmt.Errorf("max retries exceeded: %w", err)
}
```

## Performance Targets

### Response Time Targets

- **Auth operations**: <5s (due to user interaction)
- **Event create**: <2s
- **Event list**: <3s (up to 250 events)
- **Event get**: <1s
- **Event update**: <2s
- **Event delete**: <1s

### Resource Limits

- **Memory**: <50MB during normal operation
- **Binary size**: <20MB
- **Config file**: <1KB
- **Token file**: <5KB

### Optimization Strategies

1. **Connection Reuse**: Reuse HTTP client and Calendar service
2. **Batch Operations**: Group API calls when possible
3. **Caching**: Cache calendar metadata, timezone data
4. **Lazy Loading**: Only load configuration when needed
5. **Efficient Parsing**: Use streaming JSON parser for large responses

## Security Considerations

### Token Storage

- Store in `~/.config/gcal-cli/tokens.json` with 0600 permissions
- Never log token values
- Clear tokens on logout
- Encrypt tokens at rest (future enhancement)

### Credentials Handling

- Credentials file (`credentials.json`) should be 0600
- Never include in version control
- Validate credentials before use
- Clear error messages without exposing sensitive data

### Input Validation

- Sanitize all user inputs
- Validate email addresses
- Validate date/time formats
- Prevent command injection in shell operations

### Network Security

- Always use HTTPS for API calls
- Validate SSL certificates
- Timeout for long-running operations
- Rate limit compliance to avoid abuse

## Future Enhancements

### Phase 7: Advanced Features

- **Natural Language Dates**: Parse "tomorrow at 2pm", "next Monday"
- **Recurring Event Management**: Better handling of recurring events
- **Calendar Sharing**: Manage calendar permissions
- **Attachment Support**: Add files to events
- **Free/Busy Queries**: Check availability
- **Multiple Calendar Support**: Operate on multiple calendars simultaneously
- **Event Templates**: Predefined event templates
- **Conflict Detection**: Warn about scheduling conflicts

## Project Completion

All 7 planned phases have been successfully implemented. The project is considered feature-complete and production-ready for Google Calendar integration.

### Future Enhancement Ideas (Not Planned)

Community contributions could potentially add:
- Multi-provider support (Outlook, Apple Calendar, CalDAV)
- Advanced webhook integrations
- Enhanced batch operations
- AI-powered scheduling suggestions

## Success Criteria

### Functional Requirements

- ✓ All CRUD operations work correctly
- ✓ OAuth2 authentication completes successfully
- ✓ Configuration persists across sessions
- ✓ All output formats (JSON, text, minimal) work
- ✓ Error handling provides clear, actionable messages
- ✓ All commands documented with examples

### Performance Requirements

- ✓ 95th percentile latency <2s for most operations
- ✓ Memory usage <50MB
- ✓ Binary size <20MB

### Quality Requirements

- ✓ >80% test coverage
- ✓ Zero critical bugs
- ✓ All edge cases handled
- ✓ Comprehensive documentation

### LLM-Specific Requirements

- ✓ Consistent JSON schema across all operations
- ✓ Machine-parseable error messages
- ✓ Idempotent operations where possible
- ✓ Clear error codes and suggested actions
- ✓ Self-contained commands (no session state)

## Getting Started (for Developers)

### Prerequisites

- Go 1.21 or later
- Google Cloud Platform account
- Google Calendar API enabled
- OAuth2 credentials (client ID and secret)

### Setup Steps

1. **Clone Repository**
   ```bash
   git clone https://github.com/yourusername/gcal-cli.git
   cd gcal-cli
   ```

2. **Install Dependencies**
   ```bash
   go mod download
   ```

3. **Build**
   ```bash
   go build -o gcal-cli ./cmd/gcal-cli
   ```

4. **Set Up Credentials**
   - Go to [Google Cloud Console](https://console.cloud.google.com)
   - Create new project or select existing
   - Enable Google Calendar API
   - Create OAuth2 credentials (Desktop application)
   - Download credentials JSON
   - Save to `~/.config/gcal-cli/credentials.json`

5. **Authenticate**
   ```bash
   ./gcal-cli auth login
   ```

6. **Test**
   ```bash
   ./gcal-cli events list --from "2024-01-15" --to "2024-01-20"
   ```

### Development Workflow

1. **Create Feature Branch**
   ```bash
   git checkout -b feature/your-feature
   ```

2. **Write Tests First** (TDD approach)
   ```bash
   go test ./pkg/yourpackage -v
   ```

3. **Implement Feature**

4. **Run All Tests**
   ```bash
   go test ./... -cover
   ```

5. **Build and Test Locally**
   ```bash
   go build -o gcal-cli ./cmd/gcal-cli
   ./gcal-cli your-command --flags
   ```

6. **Submit PR**

## References

### Documentation

- [Google Calendar API v3](https://developers.google.com/calendar/api/v3/reference)
- [Cobra Documentation](https://cobra.dev/)
- [Viper Documentation](https://github.com/spf13/viper)
- [OAuth2 for Go](https://pkg.go.dev/golang.org/x/oauth2)

### Related Projects

- [gcal-commander](https://github.com/buko106/gcal-commander) - Similar CLI tool (reference)
- [Google Calendar MCP](https://github.com/pashpashpash/google-calendar-mcp) - MCP server integration

### Code Examples

All code patterns in this plan are based on Context7 documentation for:
- `/googleapis/google-api-go-client` - Google Calendar API integration
- `/spf13/cobra` - CLI structure and commands
- `/spf13/viper` - Configuration management

---

**Document Version**: 1.0
**Last Updated**: 2025-11-10
**Status**: Ready for Implementation
