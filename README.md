# gcal-cli

<div align="center">

**A powerful Google Calendar command-line tool with natural language support**

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

[Features](#features) • [Installation](#installation) • [Quick Start](#quick-start) • [Documentation](#documentation) • [Contributing](#contributing)

</div>

---

## Overview

gcal-cli is a production-ready Google Calendar CLI tool written in Go, designed for both human users and LLM agent integration. It combines natural language processing with powerful scheduling features to make calendar management effortless from the command line.

## Features

### 🗣️ Natural Language Support
Parse human-friendly date and time expressions:
- `"tomorrow at 2pm"` → 2024-01-16T14:00:00
- `"next Monday"` → Next Monday at midnight
- `"in 2 hours"` → 2 hours from now
- `"Friday at 3:30pm"` → This Friday at 15:30

### 📅 Smart Scheduling
- **Free/Busy Queries** - Check availability across calendars
- **Conflict Detection** - Prevent double-booking before creating events
- **Available Time Slots** - Find open times for meetings
- **Multi-Calendar Support** - Manage multiple calendars simultaneously

### ⚡ Event Management
- Complete CRUD operations (Create, Read, Update, Delete)
- Event templates for common event types
- Recurring events and all-day events
- Attendee management
- Attachment support

### 🤖 LLM-Optimized
- **JSON Output** - Structured, machine-readable responses
- **Consistent Schemas** - Predictable API responses
- **Error Codes** - Structured error handling with recovery suggestions
- **Idempotency** - Safe retry operations

### 🔐 Authentication
- OAuth2 with Google Calendar API
- Automatic token refresh
- Secure credential storage

## Installation

### Pre-built Binaries (Recommended)

Download the latest release for your platform from the [Releases page](https://github.com/btafoya/gcal-cli/releases/latest).

**Linux:**
```bash
# Download and extract
curl -LO https://github.com/btafoya/gcal-cli/releases/latest/download/gcal-cli-1.0.0-ubuntu.tar.gz
tar -xzf gcal-cli-1.0.0-ubuntu.tar.gz

# Move to PATH
sudo mv gcal-cli /usr/local/bin/
```

**macOS:**
```bash
# Download and extract
curl -LO https://github.com/btafoya/gcal-cli/releases/latest/download/gcal-cli-1.0.0-macos.tar.gz
tar -xzf gcal-cli-1.0.0-macos.tar.gz

# Move to PATH
sudo mv gcal-cli /usr/local/bin/
```

**Windows:**
```powershell
# Download from releases page and extract gcal-cli.exe
# Add to PATH or run from download location
```

### From Source

```bash
git clone https://github.com/btafoya/gcal-cli.git
cd gcal-cli
go build -o gcal-cli ./cmd/gcal-cli
```

### Requirements
- Google Cloud Platform account
- Google Calendar API enabled
- OAuth2 credentials ([setup guide](./USER-INSTRUCTIONS.md#authentication))
- *For building from source: Go 1.23 or later*

## Quick Start

### 1. Authentication

```bash
# First-time setup (opens browser for OAuth2)
./gcal-cli auth login

# Check authentication status
./gcal-cli auth status
```

See [USER-INSTRUCTIONS.md](./USER-INSTRUCTIONS.md) for detailed authentication setup.

### 2. Create Events

```bash
# Using natural language
./gcal-cli events create \
  --title "Team Standup" \
  --start "tomorrow at 9am" \
  --end "tomorrow at 9:15am"

# Using templates
./gcal-cli events create \
  --template meeting \
  --start "next Monday at 2pm"

# With attendees
./gcal-cli events create \
  --title "Project Review" \
  --start "Friday at 3pm" \
  --end "Friday at 4pm" \
  --attendees "alice@example.com,bob@example.com"
```

### 3. List Events

```bash
# Upcoming events
./gcal-cli events list --from "today" --to "next week"

# Multiple calendars
./gcal-cli events list \
  --calendars "primary,work@example.com" \
  --from "today"

# JSON output (for automation)
./gcal-cli events list --format json
```

### 4. Smart Scheduling

```bash
# Check for conflicts
./gcal-cli events check-conflicts \
  --start "tomorrow at 2pm" \
  --end "tomorrow at 3pm"

# Find available time slots
./gcal-cli events find-free \
  --from "today" \
  --to "next week" \
  --duration 60
```

## Usage Examples

### Event Templates

Built-in templates for common event types:

```bash
# Available templates: meeting, 1on1, lunch, focus, standup, interview
./gcal-cli events create --template standup --start "tomorrow at 9am"
./gcal-cli events create --template 1on1 --start "Friday at 2pm"
```

### Natural Language Dates

Supported patterns:

| Pattern | Example | Result |
|---------|---------|--------|
| Relative | `today`, `tomorrow`, `yesterday` | Specific date |
| Time offsets | `in 2 hours`, `in 30 minutes` | Time from now |
| Day of week | `next Monday`, `this Friday` | Next occurrence |
| Combined | `tomorrow at 2pm`, `Monday at 9am` | Date + time |

### Output Formats

```bash
# JSON (default - machine-readable)
./gcal-cli events list --format json

# Text (human-readable)
./gcal-cli events list --format text

# Minimal (IDs only - for piping)
./gcal-cli events list --format minimal | xargs -I {} ./gcal-cli events delete {}
```

## Configuration

Configuration file: `~/.config/gcal-cli/config.yaml`

```yaml
calendar:
  default_calendar_id: "primary"
  default_timezone: ""

output:
  default_format: "json"
  pretty_print: true

api:
  retry_attempts: 3
  timeout_seconds: 30
```

Environment variables (override config):
```bash
export GCAL_OUTPUT_DEFAULT_FORMAT=json
export GCAL_CALENDAR_DEFAULT_CALENDAR_ID=primary
```

## Documentation

- **[USER-INSTRUCTIONS.md](./USER-INSTRUCTIONS.md)** - Complete user guide with examples
- **[SCHEMAS.md](./SCHEMAS.md)** - JSON API schemas and response formats
- **[TROUBLESHOOTING.md](./TROUBLESHOOTING.md)** - Common issues and solutions
- **[CONTRIBUTING.md](./CONTRIBUTING.md)** - Contribution guidelines

## Project Structure

```
gcal-cli/
├── cmd/gcal-cli/       # Main application entry
├── pkg/
│   ├── auth/          # OAuth2 authentication
│   ├── calendar/      # Calendar operations
│   ├── config/        # Configuration management
│   ├── output/        # Output formatters
│   └── types/         # Shared types and errors
├── internal/
│   └── commands/      # CLI command implementations
└── docs/              # Implementation documentation
```

## Error Handling

Structured error codes for reliable automation:

| Code | Description | Recoverable |
|------|-------------|-------------|
| `AUTH_FAILED` | Authentication failure | Yes - run `auth login` |
| `TOKEN_EXPIRED` | Token expired | Yes - automatic refresh |
| `RATE_LIMIT` | API rate limit exceeded | Yes - retry with backoff |
| `INVALID_INPUT` | Invalid input value | No - fix input |
| `NOT_FOUND` | Resource not found | No |

Example error response:
```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT",
    "message": "API rate limit exceeded",
    "recoverable": true,
    "suggestedAction": "Wait 1 minute and retry"
  }
}
```

## Performance

- **Natural language parsing**: <1ms latency
- **Multi-calendar operations**: Parallel execution with goroutines
- **Free/busy queries**: ~500ms for 5 calendars
- **Automatic retry**: Exponential backoff for transient errors

## Development

### Building

```bash
# Development build
go build -o gcal-cli ./cmd/gcal-cli

# Build with automatic version info
VERSION=$(./scripts/version.sh)
go build -ldflags "-X github.com/btafoya/gcal-cli/internal/commands.Version=$VERSION" \
         -o gcal-cli ./cmd/gcal-cli

# Build release binaries for all platforms
./scripts/build-release.sh
```

### Versioning

```bash
# Show current version
./scripts/version.sh current

# Show next version
./scripts/version.sh next

# Create a new release tag
./scripts/version.sh tag 1.0.0

# Push tag to trigger GitHub release
git push origin v1.0.0
```

Version format: `MAJOR.MINOR.PATCH[-dev.N+HASH]`
- Released: `1.0.0`
- Development: `1.0.0-dev.5+a1b2c3d`

### Testing

```bash
# Run all tests
go test ./...

# With coverage
go test ./... -cover

# Specific package
go test ./pkg/calendar -v
```

### Dependencies

```
github.com/spf13/cobra v1.10.1      # CLI framework
github.com/spf13/viper v1.21.0      # Configuration
google.golang.org/api v0.255.0      # Google Calendar API
golang.org/x/oauth2 v0.33.0         # OAuth2 authentication
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

### Areas for Contribution
- Increase test coverage (currently ~80%)
- Add more natural language patterns
- Improve documentation and examples
- Performance optimizations
- New event template types

## License

MIT License - See [LICENSE](LICENSE) for details.

## Author

Developed by [btafoya](https://github.com/btafoya) with focus on LLM agent integration and natural language interaction.

## Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Google Calendar API](https://developers.google.com/calendar) - Calendar integration

---

<div align="center">

**[⬆ back to top](#gcal-cli)**

Made with ❤️ using Go

</div>
