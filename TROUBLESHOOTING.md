# Troubleshooting Guide

Common issues and solutions for gcal-cli.

## Table of Contents

- [Authentication Issues](#authentication-issues)
- [API Errors](#api-errors)
- [Configuration Problems](#configuration-problems)
- [Input Validation Errors](#input-validation-errors)
- [Network Issues](#network-issues)
- [Build and Installation](#build-and-installation)

## Authentication Issues

### Error: "AUTH_FAILED: authentication failed or not configured"

**Cause**: No authentication credentials or invalid credentials file.

**Solution**:
```bash
# 1. Verify credentials file exists
ls -la ~/.config/gcal-cli/credentials.json

# 2. If missing, download from Google Cloud Console
# Go to https://console.cloud.google.com
# Enable Google Calendar API
# Create OAuth2 credentials (Desktop application)
# Download credentials.json

# 3. Place in correct location
mkdir -p ~/.config/gcal-cli
mv ~/Downloads/credentials.json ~/.config/gcal-cli/

# 4. Run authentication
gcal-cli auth login
```

### Error: "TOKEN_EXPIRED: authentication token has expired"

**Cause**: OAuth token has expired.

**Solution**:
```bash
# Re-authenticate (tokens will refresh automatically)
gcal-cli auth login
```

**Prevention**: Enable automatic token refresh in config:
```yaml
# ~/.config/gcal-cli/config.yaml
auth:
  auto_refresh: true
```

### Browser doesn't open during authentication

**Cause**: System can't detect default browser or running headless.

**Solution**:
```bash
# The CLI will print a URL - manually copy and paste into browser
gcal-cli auth login

# Look for output like:
# Go to the following link in your browser:
# https://accounts.google.com/o/oauth2/auth?...
```

### Error: "Permission denied" when reading credentials

**Cause**: Incorrect file permissions.

**Solution**:
```bash
# Set proper permissions (owner read/write only)
chmod 600 ~/.config/gcal-cli/credentials.json
chmod 600 ~/.config/gcal-cli/tokens.json

# Verify
ls -la ~/.config/gcal-cli/
```

## API Errors

### Error: "RATE_LIMIT: API rate limit exceeded"

**Cause**: Too many requests to Google Calendar API.

**Solution**:
```bash
# Wait 5-10 seconds and retry
sleep 5
gcal-cli events list --from "2024-01-15" --to "2024-01-20"
```

**LLM Agent Pattern**:
```bash
perform_with_retry() {
  RESULT=$(gcal-cli events create --title "Test" \
    --start "2024-01-15T10:00:00" --end "2024-01-15T11:00:00" \
    --format json 2>&1)

  ERROR_CODE=$(echo $RESULT | jq -r '.error.code // empty')

  if [ "$ERROR_CODE" = "RATE_LIMIT" ]; then
    echo "Rate limited, waiting..."
    sleep 5
    # Retry once
    gcal-cli events create --title "Test" \
      --start "2024-01-15T10:00:00" --end "2024-01-15T11:00:00"
  else
    echo $RESULT
  fi
}
```

**Prevention**: Adjust rate limit buffer in config:
```yaml
# ~/.config/gcal-cli/config.yaml
api:
  rate_limit_buffer: 0.8  # Use only 80% of rate limit
```

### Error: "API_ERROR: API operation failed"

**Cause**: Google Calendar API service issue or network problem.

**Solution**:
```bash
# 1. Check Google Calendar API status
# Visit: https://status.cloud.google.com/

# 2. Verify network connectivity
ping google.com

# 3. Check for API quota limits
# Go to: https://console.cloud.google.com/apis/api/calendar-json.googleapis.com/quotas

# 4. Retry with increased timeout
export GCAL_API_TIMEOUT_SECONDS=60
gcal-cli events list --from "2024-01-15" --to "2024-01-20"
```

### Error: "NOT_FOUND: Event not found"

**Cause**: Event ID doesn't exist or was deleted.

**Solution**:
```bash
# 1. List events to find correct ID
gcal-cli events list --from "2024-01-01" --to "2024-12-31" --format json | \
  jq -r '.data.events[] | "\(.id) - \(.summary)"'

# 2. Verify event exists before operations
EVENT_ID="abc123xyz"
RESULT=$(gcal-cli events get "$EVENT_ID" --format json 2>&1)
if [ "$(echo $RESULT | jq -r '.success')" = "true" ]; then
  # Event exists, proceed with update/delete
  gcal-cli events update "$EVENT_ID" --title "Updated"
else
  echo "Event not found: $EVENT_ID"
fi
```

### Error: "PERMISSION_DENIED: insufficient permissions for calendar"

**Cause**: User doesn't have access to the specified calendar.

**Solution**:
```bash
# 1. List accessible calendars
gcal-cli calendars list --format json

# 2. Use primary calendar
gcal-cli events list --calendar-id primary --from "2024-01-15" --to "2024-01-20"

# 3. Request calendar access from owner if needed
```

## Configuration Problems

### Error: "CONFIG_ERROR: could not determine home directory"

**Cause**: $HOME environment variable not set.

**Solution**:
```bash
# Set HOME environment variable
export HOME=/home/$(whoami)

# Or use XDG_CONFIG_HOME
export XDG_CONFIG_HOME=/path/to/config
```

### Error: "CONFIG_ERROR: error reading config file"

**Cause**: Malformed YAML in config file.

**Solution**:
```bash
# 1. Validate YAML syntax
cat ~/.config/gcal-cli/config.yaml

# 2. Check for common YAML errors:
# - Incorrect indentation (use spaces, not tabs)
# - Missing quotes around special characters
# - Invalid boolean values (use true/false, not yes/no)

# 3. Reset to defaults
gcal-cli config init

# 4. Verify configuration
gcal-cli config show
```

### Config file not found

**Cause**: Config file doesn't exist (this is okay, defaults will be used).

**Solution**:
```bash
# Create config file with defaults
gcal-cli config init

# Or manually create
mkdir -p ~/.config/gcal-cli
cat > ~/.config/gcal-cli/config.yaml << 'EOF'
calendar:
  default_calendar_id: "primary"
  default_timezone: ""

output:
  default_format: "json"
  color_enabled: false
  pretty_print: true
EOF
```

### Environment variables not working

**Cause**: Incorrect variable names or not exported.

**Solution**:
```bash
# Environment variables must be prefixed with GCAL_
# and use underscores to separate nested keys

# ✓ Correct
export GCAL_OUTPUT_DEFAULT_FORMAT=json
export GCAL_CALENDAR_DEFAULT_CALENDAR_ID=primary

# ✗ Incorrect
export OUTPUT_DEFAULT_FORMAT=json  # Missing GCAL_ prefix
export GCAL_OUTPUT.DEFAULT_FORMAT=json  # Use underscores, not dots

# Verify variable is set
env | grep GCAL_
```

## Input Validation Errors

### Error: "INVALID_INPUT: invalid time format"

**Cause**: Time string doesn't match expected format.

**Solution**:
```bash
# Supported formats:
# - RFC3339: 2024-01-15T10:00:00Z or 2024-01-15T10:00:00-05:00
# - ISO-like: 2024-01-15T10:00:00 or 2024-01-15T10:00
# - Space-separated: 2024-01-15 10:00:00 or 2024-01-15 10:00

# ✓ Valid examples
gcal-cli events create --title "Test" \
  --start "2024-01-15T10:00:00" \
  --end "2024-01-15T11:00:00"

gcal-cli events create --title "Test" \
  --start "2024-01-15 10:00" \
  --end "2024-01-15 11:00"

# ✗ Invalid examples
# Missing time: --start "2024-01-15"
# Wrong separator: --start "2024/01/15 10:00"
# Invalid format: --start "Jan 15 2024 10:00 AM"
```

### Error: "MISSING_REQUIRED: required field 'title' is missing"

**Cause**: Required flag not provided.

**Solution**:
```bash
# Check which flags are required
gcal-cli events create --help

# Event creation requires: --title, --start, --end
gcal-cli events create \
  --title "Required Title" \
  --start "2024-01-15T10:00:00" \
  --end "2024-01-15T11:00:00"
```

### Error: "INVALID_TIME_RANGE: end time must be after start time"

**Cause**: End time is before or equal to start time.

**Solution**:
```bash
# ✓ Correct: end after start
gcal-cli events create --title "Test" \
  --start "2024-01-15T10:00:00" \
  --end "2024-01-15T11:00:00"

# ✗ Incorrect: end before start
# --start "2024-01-15T14:00:00" --end "2024-01-15T13:00:00"

# For all-day events, end should be next day
gcal-cli events create --title "All Day" \
  --start "2024-01-15" \
  --end "2024-01-16" \
  --all-day
```

### Invalid email format for attendees

**Cause**: Malformed email addresses.

**Solution**:
```bash
# ✓ Correct format
gcal-cli events create --title "Meeting" \
  --start "2024-01-15T10:00:00" \
  --end "2024-01-15T11:00:00" \
  --attendees "user1@example.com,user2@example.com"

# ✗ Incorrect
# Missing @: --attendees "user1,user2"
# Spaces: --attendees "user1@example.com, user2@example.com" (remove spaces)
```

## Network Issues

### Error: "NETWORK_ERROR: network communication failed"

**Cause**: Can't reach Google Calendar API servers.

**Solution**:
```bash
# 1. Check internet connectivity
ping google.com

# 2. Check if firewall is blocking
curl -I https://www.googleapis.com/calendar/v3/users/me/calendarList

# 3. Try with increased timeout
export GCAL_API_TIMEOUT_SECONDS=60
gcal-cli events list --from "2024-01-15" --to "2024-01-20"

# 4. Check proxy settings if behind corporate proxy
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=http://proxy.example.com:8080
```

### SSL/TLS certificate errors

**Cause**: System doesn't trust Google's certificates or using corporate proxy.

**Solution**:
```bash
# 1. Update CA certificates (Linux)
sudo update-ca-certificates

# 2. Update CA certificates (macOS)
# System certificates are auto-updated

# 3. If behind corporate proxy with MITM certificate
# Add corporate certificate to system trust store

# 4. Verify TLS connection
openssl s_client -connect www.googleapis.com:443
```

### Timeout errors

**Cause**: Slow network or large result sets.

**Solution**:
```bash
# Increase timeout in config
cat > ~/.config/gcal-cli/config.yaml << 'EOF'
api:
  timeout_seconds: 60  # Increase from default 30
EOF

# Or via environment variable
export GCAL_API_TIMEOUT_SECONDS=60

# Reduce result set size
gcal-cli events list \
  --from "2024-01-15" \
  --to "2024-01-20" \
  --max-results 50  # Instead of default 250
```

## Build and Installation

### Build fails with "command not found: go"

**Cause**: Go not installed or not in PATH.

**Solution**:
```bash
# Install Go
# Visit: https://golang.org/dl/

# Or using package manager
# Ubuntu/Debian:
sudo apt-get install golang-go

# macOS:
brew install go

# Verify installation
go version
```

### Build fails with "package not found"

**Cause**: Dependencies not downloaded.

**Solution**:
```bash
# Download dependencies
go mod download

# Verify go.mod and go.sum are present
ls -la go.mod go.sum

# Clean module cache if corrupted
go clean -modcache
go mod download
```

### Permission denied when running binary

**Cause**: Binary not executable.

**Solution**:
```bash
# Make binary executable
chmod +x gcal-cli

# Verify permissions
ls -la gcal-cli

# Should show: -rwxr-xr-x
```

### Command not found after installation

**Cause**: Binary not in PATH.

**Solution**:
```bash
# Option 1: Add current directory to PATH
export PATH=$PATH:$(pwd)

# Option 2: Move to directory in PATH
sudo mv gcal-cli /usr/local/bin/

# Option 3: Create symlink
sudo ln -s $(pwd)/gcal-cli /usr/local/bin/gcal-cli

# Verify
which gcal-cli
```

## Debugging Tips

### Enable Verbose Logging

```bash
# Set log level via environment
export GCAL_LOG_LEVEL=debug

# Run command
gcal-cli events list --from "2024-01-15" --to "2024-01-20"
```

### Inspect JSON Responses

```bash
# Pretty-print JSON output
gcal-cli events list --from "2024-01-15" --to "2024-01-20" --format json | jq '.'

# Extract specific fields
gcal-cli events list --from "2024-01-15" --to "2024-01-20" --format json | \
  jq '.data.events[] | {id, summary, start}'

# Check for errors
gcal-cli events create --title "Test" --start "invalid" --end "invalid" \
  --format json 2>&1 | jq '.error'
```

### Verify Configuration

```bash
# Show current configuration
gcal-cli config show

# Test configuration loading
gcal-cli config show --format json | jq '.data.config'
```

### Test Authentication

```bash
# Check authentication status
gcal-cli auth status --format json | jq '.'

# Expected output when authenticated:
# {
#   "success": true,
#   "data": {
#     "authenticated": true,
#     "email": "user@example.com",
#     ...
#   }
# }
```

## Common LLM Agent Issues

### JSON Parsing Errors

**Issue**: Can't parse JSON output.

**Solution**:
```bash
# Always use --format json flag
gcal-cli events list --from "2024-01-15" --to "2024-01-20" --format json

# Validate JSON before parsing
RESULT=$(gcal-cli events list --from "2024-01-15" --to "2024-01-20" \
  --format json 2>&1)

if echo "$RESULT" | jq -e . > /dev/null 2>&1; then
  # Valid JSON
  echo "$RESULT" | jq '.data.events'
else
  # Invalid JSON or error
  echo "Error: Invalid JSON output"
  echo "$RESULT"
fi
```

### Handling Errors in Scripts

**Issue**: Script fails on errors.

**Solution**:
```bash
# Don't use set -e for gcal-cli commands
# Handle errors explicitly

RESULT=$(gcal-cli events create --title "Test" \
  --start "2024-01-15T10:00:00" --end "2024-01-15T11:00:00" \
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

### Idempotency Issues

**Issue**: Duplicate events created.

**Solution**: Implement idempotency check:
```bash
create_or_update_event() {
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
    gcal-cli events create --title "$TITLE" --start "$START" --end "$END"
  else
    # Update existing event
    gcal-cli events update "$EXISTING" --start "$START" --end "$END"
  fi
}
```

## Getting Help

### Check Version

```bash
gcal-cli version
```

### View Help Text

```bash
# General help
gcal-cli --help

# Command-specific help
gcal-cli events --help
gcal-cli events create --help

# All commands have examples in help text
```

### Report Issues

If you encounter a bug:

1. **Gather Information**:
   ```bash
   # Version info
   gcal-cli version

   # Configuration
   gcal-cli config show

   # Error output with JSON
   gcal-cli <command> --format json 2>&1
   ```

2. **Check Existing Issues**:
   - Visit: https://github.com/btafoya/gcal-cli/issues

3. **Create New Issue**:
   - Include version information
   - Include complete error output
   - Include steps to reproduce
   - Include expected vs actual behavior

## Additional Resources

- [README.md](./README.md) - Complete usage guide
- [SCHEMAS.md](./SCHEMAS.md) - JSON schema documentation
- [pkg/examples/examples.go](./pkg/examples/examples.go) - Usage examples
- [PLAN.md](./PLAN.md) - Implementation plan and architecture
