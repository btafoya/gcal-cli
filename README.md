# gcal-cli - Google Calendar CLI Tool

A powerful Google Calendar command-line interface written in Go, designed for both human users and LLM agent integration. Features natural language date parsing, intelligent scheduling, multi-calendar support, and comprehensive event management.

## Status: Phase 7 Complete ✅

**Current Version**: v0.7.0-dev

### 🎯 Key Features

- ✅ **Natural Language Dates** - "tomorrow at 2pm", "next Monday", "in 2 hours"
- ✅ **Smart Scheduling** - Free/busy queries, conflict detection, available time slots
- ✅ **Multi-Calendar Support** - Manage multiple calendars simultaneously with parallel operations
- ✅ **Event Templates** - Pre-configured templates for common event types
- ✅ **LLM-Optimized Output** - JSON, text, and minimal formats for automation
- ✅ **Complete CRUD Operations** - Create, read, update, delete events and calendars
- ✅ **OAuth2 Authentication** - Secure Google Calendar API integration
- ✅ **Comprehensive Error Handling** - Structured error codes and retry logic

## Quick Start

### Build

```bash
go build -o gcal-cli ./cmd/gcal-cli
```

### Basic Usage

```bash
# Show help
./gcal-cli --help

# Show version
./gcal-cli version

# Initialize configuration
./gcal-cli config init

# Show configuration
./gcal-cli config show

# Set configuration value
./gcal-cli config set output.default_format text
```

### Authentication

Authenticate with Google Calendar using OAuth2:

```bash
# First-time authentication (opens browser)
./gcal-cli auth login

# Check authentication status
./gcal-cli auth status

# Logout (removes stored credentials)
./gcal-cli auth logout
```

**Prerequisites**: Download OAuth2 credentials from Google Cloud Console and save to `~/.config/gcal-cli/credentials.json`

See [USER-INSTRUCTIONS.md](./USER-INSTRUCTIONS.md) for detailed authentication setup.

### Calendar Events

#### Basic Operations

```bash
# Create event with natural language dates
./gcal-cli events create \
  --title "Team Meeting" \
  --start "tomorrow at 2pm" \
  --end "tomorrow at 3pm" \
  --description "Weekly sync"

# List upcoming events
./gcal-cli events list --from "today" --to "next Friday"

# Get event details
./gcal-cli events get <event-id>

# Update event
./gcal-cli events update <event-id> --title "Updated Title"

# Delete event
./gcal-cli events delete <event-id>
```

#### Advanced Features

```bash
# Create event from template
./gcal-cli events create --template meeting --start "next Monday at 9am"

# Check for scheduling conflicts
./gcal-cli events check-conflicts --start "tomorrow at 2pm" --end "tomorrow at 3pm"

# Find available time slots
./gcal-cli events find-free --from "today" --to "next week" --duration 60

# List events from multiple calendars
./gcal-cli events list --calendars "primary,work@example.com" --from "today"
```

### Output Formats

The CLI supports three output formats for LLM-friendly interactions:

#### JSON (Default - LLM Optimized)
```bash
./gcal-cli version --format json
```

Output:
```json
{
  "success": true,
  "operation": "version",
  "data": {
    "version": "dev",
    "commit": "unknown",
    "buildDate": "unknown"
  },
  "metadata": {
    "timestamp": "2024-01-15T10:00:00Z"
  }
}
```

#### Text (Human-Readable)
```bash
./gcal-cli config show --format text
```

#### Minimal (IDs Only - for Piping)
```bash
./gcal-cli version --format minimal
```

## Project Structure

```
gcal-cli/
├── cmd/
│   └── gcal-cli/          # Main application entry point
│       ├── main.go
│       └── root.go         # Root Cobra command
├── pkg/
│   ├── auth/              # Authentication (Phase 2)
│   ├── calendar/          # Calendar operations (Phase 3)
│   ├── config/            # Viper configuration ✓
│   ├── output/            # Output formatters ✓
│   │   ├── formatter.go
│   │   ├── json.go
│   │   ├── text.go
│   │   └── minimal.go
│   └── types/             # Shared types ✓
│       ├── errors.go
│       ├── event.go
│       └── response.go
├── internal/
│   └── commands/          # Command implementations ✓
│       ├── config.go
│       └── version.go
└── test/                  # Tests
    ├── integration/
    └── fixtures/
```

## Configuration

Configuration file: `~/.config/gcal-cli/config.yaml`

### Default Configuration

```yaml
calendar:
  default_calendar_id: "primary"
  default_timezone: ""

output:
  default_format: "json"
  color_enabled: false
  pretty_print: true

auth:
  credentials_path: "~/.config/gcal-cli/credentials.json"
  tokens_path: "~/.config/gcal-cli/tokens.json"
  auto_refresh: true

api:
  retry_attempts: 3
  retry_delay_ms: 1000
  retry_max_delay_ms: 10000
  timeout_seconds: 30
  rate_limit_buffer: 0.9

events:
  default_duration_minutes: 60
  default_reminder_minutes: 10
  send_notifications: true
```

### Environment Variables

All configuration values can be overridden with environment variables prefixed with `GCAL_`:

```bash
export GCAL_OUTPUT_DEFAULT_FORMAT=json
export GCAL_CALENDAR_DEFAULT_CALENDAR_ID=primary
```

## Development

### Prerequisites

- Go 1.21 or later
- Google Cloud Platform account (for Phase 2+)
- Google Calendar API enabled (for Phase 2+)

### Dependencies

```
github.com/spf13/cobra v1.10.1      # CLI framework
github.com/spf13/viper v1.21.0      # Configuration
google.golang.org/api v0.255.0      # Google APIs
golang.org/x/oauth2 v0.33.0         # OAuth2
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run specific package tests
go test ./pkg/output -v
```

### Building

```bash
# Development build
go build -o gcal-cli ./cmd/gcal-cli

# Production build with version info
go build -ldflags "-X github.com/btafoya/gcal-cli/internal/commands.Version=1.0.0 \
                    -X github.com/btafoya/gcal-cli/internal/commands.Commit=$(git rev-parse --short HEAD) \
                    -X github.com/btafoya/gcal-cli/internal/commands.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
         -o gcal-cli ./cmd/gcal-cli
```

## Error Handling

The CLI uses structured error codes for machine-parseable error handling:

### Error Codes

- `AUTH_FAILED` - Authentication failure
- `TOKEN_EXPIRED` - Token expiration
- `INVALID_INPUT` - Input validation failure
- `MISSING_REQUIRED` - Required field missing
- `NOT_FOUND` - Resource not found
- `RATE_LIMIT` - API rate limit exceeded
- `API_ERROR` - Google Calendar API error
- `CONFIG_ERROR` - Configuration error
- `NETWORK_ERROR` - Network connectivity issue

### Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "INVALID_INPUT",
    "message": "Invalid value for field",
    "details": "Additional context",
    "recoverable": true,
    "suggestedAction": "Corrective action"
  }
}
```

## Roadmap

### Phase 1: Foundation ✅
- [x] Project structure and CLI framework
- [x] Configuration management
- [x] Output formatters (JSON, Text, Minimal)
- [x] Error handling infrastructure

### Phase 2: Authentication ✅
- [x] OAuth2 flow implementation
- [x] Token storage and automatic refresh
- [x] Auth commands (login, logout, status)

### Phase 3: Event Operations ✅
- [x] Complete event CRUD operations
- [x] All-day event support
- [x] Attendee management
- [x] Recurring events

### Phase 4: Calendar Operations ✅
- [x] List and get calendars
- [x] Calendar metadata management

### Phase 5: LLM Optimization ✅
- [x] Machine-readable JSON output
- [x] Structured error responses
- [x] Idempotency support

### Phase 6: Testing & Documentation ✅
- [x] Comprehensive test coverage
- [x] Complete documentation
- [x] User guide and troubleshooting

### Phase 7: Advanced Features ✅
- [x] Natural language date parsing
- [x] Free/busy queries and conflict detection
- [x] Multi-calendar support
- [x] Event templates
- [x] Calendar sharing

**Project Status**: All planned phases complete. The tool is production-ready with comprehensive features for Google Calendar management.

## Documentation

- **[USER-INSTRUCTIONS.md](./USER-INSTRUCTIONS.md)** - Complete user guide with examples
- **[SCHEMAS.md](./SCHEMAS.md)** - JSON schema documentation
- **[TROUBLESHOOTING.md](./TROUBLESHOOTING.md)** - Common issues and solutions
- **[docs/PLAN.md](./docs/PLAN.md)** - Detailed implementation plan
- **[docs/PHASE7_COMPLETE.md](./docs/PHASE7_COMPLETE.md)** - Phase 7 implementation details
- **[CONTRIBUTING.md](./CONTRIBUTING.md)** - Contribution guidelines

## Performance

- Natural language parsing: <1ms latency
- Multi-calendar operations: Parallel execution with goroutines
- Free/busy queries: ~500ms for 5 calendars
- Automatic retry with exponential backoff

## Requirements

- Go 1.21 or later
- Google Cloud Platform account
- Google Calendar API enabled
- OAuth2 credentials from Google Cloud Console

## Installation

```bash
# Clone repository
git clone https://github.com/btafoya/gcal-cli.git
cd gcal-cli

# Build
go build -o gcal-cli ./cmd/gcal-cli

# Install globally (optional)
go install ./cmd/gcal-cli
```

## License

MIT License - See LICENSE file for details

## Contributing

Contributions welcome! This project follows Go best practices and includes comprehensive tests. See [PLAN.md](./PLAN.md) for the development roadmap.

## Author

Developed by btafoya with focus on LLM agent integration and natural language interaction.
