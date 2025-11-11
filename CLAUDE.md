# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A Google Calendar CLI tool written in Go, designed specifically for LLM agent integration. The tool enables programmatic creation, modification, and removal of calendar events through a command-line interface optimized for agent workflows.

## Development Commands

### Building
```bash
go build -o gcal-cli ./cmd/gcal-cli
```

### Running
```bash
./gcal-cli [command] [flags]
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./pkg/calendar

# Run a specific test
go test -run TestCreateEvent ./pkg/calendar
```

### Linting
```bash
# Install golangci-lint if not already installed
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```

### Dependencies
```bash
# Add a new dependency
go get <package>

# Update dependencies
go mod tidy

# Verify dependencies
go mod verify
```

## Architecture

### Project Structure Philosophy

This CLI follows the standard Go project layout:
- `cmd/` - Main application entry points
- `pkg/` - Public library code that can be imported by other projects
- `internal/` - Private application code that cannot be imported externally
- `api/` - API definitions and protobuf files (if needed)

### Core Components

**Authentication Layer** (`pkg/auth/` or `internal/auth/`)
- Handles OAuth2 flow with Google Calendar API
- Manages token storage and refresh logic
- Credentials should be stored in `~/.config/gcal-cli/` or similar user config directory

**Calendar Operations** (`pkg/calendar/`)
- Core CRUD operations for calendar events
- Abstraction layer over Google Calendar API
- Should expose a clean interface for LLM agents

**CLI Interface** (`cmd/gcal-cli/`)
- Command parsing and validation
- Structured output formatting (JSON/plain text for agent consumption)
- Error handling with clear, parseable messages

**Agent Integration** (`pkg/agent/` or similar)
- LLM-friendly input/output formatting
- Structured responses (JSON schema compliance)
- Context management for multi-step agent workflows

### Design Considerations for LLM Agent Use

**Output Format**: All responses should be structured (JSON by default) with consistent schemas that LLM agents can reliably parse. Include flags like `--format json|text` for flexibility.

**Error Messages**: Errors must be machine-parseable with clear error codes and actionable messages. Use structured error responses rather than free-form text.

**Idempotency**: Operations should be idempotent where possible. Creating an event with the same parameters twice should handle gracefully (either return existing event or clear error).

**State Management**: Since LLMs are stateless, avoid requiring session state. Each command should be self-contained with all necessary context in the command itself.

## Google Calendar API Integration

### Authentication Setup
- Uses OAuth2 with Google Calendar API
- Requires credentials JSON file from Google Cloud Console
- Token refresh should be automatic and transparent
- Store tokens securely in user config directory with appropriate file permissions (0600)

### API Scopes
The application will need:
- `https://www.googleapis.com/auth/calendar` - Full calendar access
- Or more restrictive scopes based on requirements

### Rate Limiting
Google Calendar API has rate limits. Implement:
- Exponential backoff for rate limit errors
- Request queuing if needed for batch operations
- Clear error messages when rate limits are hit

## Command Design Patterns

Commands should follow this general pattern:
```bash
gcal-cli <action> <resource> [flags]

Examples:
gcal-cli create event --title "Meeting" --start "2024-01-15T10:00:00" --end "2024-01-15T11:00:00"
gcal-cli list events --from "2024-01-15" --to "2024-01-20" --format json
gcal-cli update event <event-id> --title "Updated Meeting"
gcal-cli delete event <event-id>
```

### Required Flags for Agent Clarity
- `--calendar-id` - Specify which calendar (default to "primary")
- `--format` - Output format (json, text, minimal)
- `--timezone` - Explicit timezone handling (default to system timezone)

## Testing Strategy

**Unit Tests**: Mock Google Calendar API responses for calendar operations testing

**Integration Tests**: Use Google Calendar API test environment or separate test calendar

**Agent Simulation Tests**: Test command parsing and output formatting with LLM-like inputs

## Configuration Management

Configuration file location: `~/.config/gcal-cli/config.yaml`

Should include:
- Default calendar ID
- Preferred timezone
- Output format preferences
- Token storage location
- API credentials reference

## Dependencies to Consider

- `google.golang.org/api/calendar/v3` - Official Google Calendar API client
- `github.com/spf13/cobra` - CLI framework (recommended for Go CLIs)
- `github.com/spf13/viper` - Configuration management
- `golang.org/x/oauth2` - OAuth2 authentication

## Future Extensibility

Design the codebase to support:
- Multiple calendar providers (not just Google)
- Webhook/event subscription for real-time updates
- Batch operations for efficiency
- Natural language date/time parsing for better LLM integration
