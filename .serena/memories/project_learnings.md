# gcal-cli Project Learnings

## Code Conventions

### Field Naming
- Client struct uses **capitalized** public fields: `Client.Service` not `Client.service`
- Always check existing struct definitions before using fields

### Function Naming
- Event conversion: `convertEvent()` not `convertGoogleEvent()`
- Pattern: Read existing code to find helper functions before creating new ones

### Error Handling
- Use error codes from `types` package
- Pattern: `types.NewAppError(types.ErrCodeX, "message", userFacing).WithDetails(details)`
- Never invent new error helper functions - use existing AppError creation

## Concurrency Patterns

### Parallel Calendar Operations
```go
var wg sync.WaitGroup
var mu sync.Mutex
errorsChan := make(chan error, len(calendarIDs))

for _, calID := range calendarIDs {
    wg.Add(1)
    go func(calendarID string) {
        defer wg.Done()
        // API call
        mu.Lock()
        // Update shared state
        mu.Unlock()
    }(calID)
}
wg.Wait()
close(errorsChan)
```

### Error Aggregation
- Use buffered error channel with size = number of goroutines
- Check channel after wg.Wait()
- Report first error or aggregate all errors

## Google Calendar API Patterns

### Time Handling
- Always use RFC3339 format: `time.Format(time.RFC3339)`
- Parse with: `time.Parse(time.RFC3339, dateString)`
- Include timezone in all time operations

### API Call Structure
```go
call := c.Service.Events.List(calendarID).
    TimeMin(start.Format(time.RFC3339)).
    TimeMax(end.Format(time.RFC3339)).
    SingleEvents(true).
    OrderBy("startTime")
    
if maxResults > 0 {
    call = call.MaxResults(int64(maxResults))
}

response, err := call.Context(ctx).Do()
```

## Natural Language Processing

### Date Parsing Strategy
1. Check for relative dates (today, tomorrow, yesterday)
2. Check for time offsets (in X hours/days/weeks/months)
3. Check for day of week (next Monday, this Friday)
4. Check for time of day combinations (tomorrow at 2pm)

### Performance
- Use regex for pattern matching (< 1ms per operation)
- Cache compiled regex patterns
- No external NLP dependencies needed for common patterns

## Testing Patterns

### Table-Driven Tests
```go
tests := []struct {
    name    string
    input   string
    want    string
    wantErr bool
}{
    {name: "today", input: "today", want: expectedDate, wantErr: false},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got, err := ParseDate(tt.input)
        if (err != nil) != tt.wantErr {
            t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
        }
        if got != tt.want {
            t.Errorf("got %v, want %v", got, tt.want)
        }
    })
}
```

## Configuration Management

### Template Storage
- Location: `~/.config/gcal-cli/templates.json`
- Use `config.GetConfigDir()` for cross-platform compatibility
- JSON marshaling with 2-space indent for readability

### Credentials
- OAuth2 credentials: `~/.config/gcal-cli/credentials.json`
- Token storage: `~/.config/gcal-cli/token.json`
- Never commit credentials to version control

## Common Pitfalls

### Field Access
❌ `c.service` - will fail, field is private
✅ `c.Service` - correct, field is public

### Function Names
❌ `convertGoogleEvent()` - doesn't exist
✅ `convertEvent()` - actual function name

### Error Creation
❌ `types.ErrInvalidTimeRange()` - doesn't exist
✅ `types.NewAppError(types.ErrCodeInvalidTimeRange, msg, true)` - correct pattern

### Time Validation
❌ Skip validation, assume times are valid
✅ Always validate timeMin < timeMax before API calls

## Architecture Decisions

### Multi-Calendar Design
- **Choice**: Parallel execution with goroutines
- **Rationale**: Process N calendars in ~time(1 calendar) + overhead
- **Trade-off**: More complex error handling but significant performance gain

### Natural Language Approach
- **Choice**: Regex-based parsing, no ML/AI
- **Rationale**: Common patterns are deterministic, <1ms performance
- **Trade-off**: Limited to English and predefined patterns, but zero dependencies

### Template System
- **Choice**: Local JSON storage
- **Rationale**: Simple, portable, no external dependencies
- **Trade-off**: Not synced across devices, but suitable for CLI tool
