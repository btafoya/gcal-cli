# Contributing to gcal-cli

Thank you for your interest in contributing to gcal-cli! This document provides guidelines for contributing to the project.

## Development Status

The project is currently at **Phase 7 of 9** in the development roadmap. See [PLAN.md](./PLAN.md) for the complete implementation plan.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Google Cloud Platform account (for testing)
- Google Calendar API credentials

### Setup

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/gcal-cli.git
   cd gcal-cli
   ```
3. Add upstream remote:
   ```bash
   git remote add upstream https://github.com/btafoya/gcal-cli.git
   ```
4. Install dependencies:
   ```bash
   go mod download
   ```

## Development Workflow

### Branch Strategy

- `master` - Stable releases
- Feature branches - `feature/your-feature-name`
- Bug fixes - `fix/issue-description`

### Making Changes

1. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following the code style guidelines

3. Write or update tests:
   ```bash
   go test ./...
   ```

4. Run the linter (if available):
   ```bash
   go vet ./...
   go fmt ./...
   ```

5. Commit your changes:
   ```bash
   git add .
   git commit -m "Brief description of changes"
   ```

6. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

7. Create a Pull Request

## Code Style Guidelines

### Go Conventions

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting
- Keep functions focused and single-purpose
- Write descriptive variable and function names
- Add comments for exported functions and types

### Project Patterns

- **Error Handling**: Use the structured error system in `pkg/types/errors.go`
- **Configuration**: Use Viper for configuration management
- **CLI Commands**: Use Cobra framework patterns
- **Output**: Support all three output formats (JSON, Text, Minimal)
- **Testing**: Table-driven tests for comprehensive coverage

### Example Code Style

```go
// Good: Clear function with proper error handling
func ParseNaturalLanguageDate(input string, timezone *time.Location) (string, error) {
    if input == "" {
        return "", types.ErrInvalidInput("date", "date string cannot be empty")
    }

    // Parse logic here
    result := parseDate(input, timezone)

    return result.Format(time.RFC3339), nil
}

// Good: Table-driven test
func TestParseDate(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {name: "today", input: "today", want: "2024-01-15T00:00:00Z", wantErr: false},
        {name: "invalid", input: "", want: "", wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseDate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseDate() error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("ParseDate() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run tests for specific package
go test ./pkg/calendar -v

# Run specific test
go test ./pkg/calendar -run TestParseNaturalLanguageDate
```

### Writing Tests

- Write table-driven tests when possible
- Test both success and error cases
- Use descriptive test names
- Aim for >80% coverage for new code
- Include integration tests for API interactions

## Documentation

### Code Documentation

- Add godoc comments for all exported functions and types
- Include examples in documentation when helpful
- Keep comments up-to-date with code changes

### User Documentation

- Update [USER-INSTRUCTIONS.md](./USER-INSTRUCTIONS.md) for new features
- Add examples to demonstrate usage
- Update [SCHEMAS.md](./SCHEMAS.md) for API changes
- Add troubleshooting entries to [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) as needed

## Pull Request Process

1. **Before Submitting**:
   - Ensure all tests pass
   - Run `go fmt` and `go vet`
   - Update documentation
   - Add/update tests for your changes

2. **PR Description**:
   - Describe what changes you made and why
   - Reference any related issues
   - Include examples of new functionality
   - Note any breaking changes

3. **Review Process**:
   - Respond to review comments
   - Make requested changes
   - Keep the PR focused on a single feature/fix

4. **Merging**:
   - PRs require approval before merging
   - Squash commits for cleaner history (when appropriate)
   - Delete branch after merge

## Areas for Contribution

### High Priority

- **Phase 8**: Multi-provider support (Outlook, Apple Calendar)
- **Phase 9**: Advanced LLM features (batch operations, webhooks)
- **Testing**: Increase test coverage
- **Documentation**: Improve examples and guides

### Good First Issues

- Add more natural language date patterns
- Improve error messages
- Add new event templates
- Write additional tests
- Fix typos in documentation

### Feature Requests

Check the [Issues](https://github.com/btafoya/gcal-cli/issues) page for requested features or create a new issue to discuss your idea before starting work.

## Code of Conduct

### Our Standards

- Be respectful and inclusive
- Provide constructive feedback
- Focus on what is best for the project
- Show empathy towards other contributors

### Reporting Issues

If you encounter bugs or have feature requests:

1. Check existing issues first
2. Create a new issue with:
   - Clear title and description
   - Steps to reproduce (for bugs)
   - Expected vs actual behavior
   - Your environment (OS, Go version, etc.)

## Questions?

- Open an issue for general questions
- Review [USER-INSTRUCTIONS.md](./USER-INSTRUCTIONS.md) for usage help
- Check [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) for common issues

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to gcal-cli! 🎉
