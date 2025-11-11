package types

import "fmt"

// Error codes for structured error handling
const (
	// Authentication errors
	ErrCodeAuthFailed   = "AUTH_FAILED"
	ErrCodeTokenExpired = "TOKEN_EXPIRED"
	ErrCodeInvalidCreds = "INVALID_CREDENTIALS"

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
	ErrCodeConfigError  = "CONFIG_ERROR"
	ErrCodeNetworkError = "NETWORK_ERROR"
	ErrCodeFileError    = "FILE_ERROR"
)

// AppError represents a structured error with code and recovery information
type AppError struct {
	Code            string `json:"code"`
	Message         string `json:"message"`
	Details         string `json:"details,omitempty"`
	Recoverable     bool   `json:"recoverable"`
	SuggestedAction string `json:"suggestedAction,omitempty"`
	wrapped         error  // Internal error for debugging
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped error for error chain support
func (e *AppError) Unwrap() error {
	return e.wrapped
}

// NewAppError creates a new AppError
func NewAppError(code, message string, recoverable bool) *AppError {
	return &AppError{
		Code:        code,
		Message:     message,
		Recoverable: recoverable,
	}
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithSuggestedAction adds a suggested action to the error
func (e *AppError) WithSuggestedAction(action string) *AppError {
	e.SuggestedAction = action
	return e
}

// WithWrappedError wraps an underlying error
func (e *AppError) WithWrappedError(err error) *AppError {
	e.wrapped = err
	return e
}

// Common error constructors

// ErrAuthFailed creates an authentication failure error
func ErrAuthFailed(message string) *AppError {
	return NewAppError(ErrCodeAuthFailed, message, true).
		WithSuggestedAction("Run 'gcal-cli auth login' to authenticate")
}

// ErrTokenExpired creates a token expiration error
func ErrTokenExpired() *AppError {
	return NewAppError(ErrCodeTokenExpired, "Authentication token has expired", true).
		WithSuggestedAction("Run 'gcal-cli auth login' to re-authenticate")
}

// ErrInvalidInput creates an invalid input error
func ErrInvalidInput(field, reason string) *AppError {
	return NewAppError(ErrCodeInvalidInput,
		fmt.Sprintf("Invalid value for %s", field), true).
		WithDetails(reason)
}

// ErrMissingRequired creates a missing required field error
func ErrMissingRequired(field string) *AppError {
	return NewAppError(ErrCodeMissingRequired,
		fmt.Sprintf("Required field '%s' is missing", field), true).
		WithSuggestedAction(fmt.Sprintf("Provide the --%s flag", field))
}

// ErrNotFound creates a not found error
func ErrNotFound(resource, id string) *AppError {
	return NewAppError(ErrCodeNotFound,
		fmt.Sprintf("%s not found", resource), false).
		WithDetails(fmt.Sprintf("ID: %s", id))
}

// ErrRateLimit creates a rate limit error
func ErrRateLimit() *AppError {
	return NewAppError(ErrCodeRateLimit,
		"API rate limit exceeded", true).
		WithSuggestedAction("Wait a moment and try again")
}

// ErrConfigError creates a configuration error
func ErrConfigError(message string) *AppError {
	return NewAppError(ErrCodeConfigError, message, true).
		WithSuggestedAction("Run 'gcal-cli config init' to reset configuration")
}

// ErrNetworkError creates a network error
func ErrNetworkError(message string) *AppError {
	return NewAppError(ErrCodeNetworkError, message, true).
		WithSuggestedAction("Check your internet connection and try again")
}

// ErrFileError creates a file operation error
var ErrFileError = NewAppError(ErrCodeFileError, "File operation failed", true)

// ErrAPIError creates an API error
var ErrAPIError = NewAppError(ErrCodeAPIError, "API operation failed", true)

// ErrInvalidCreds creates an invalid credentials error
var ErrInvalidCreds = NewAppError(ErrCodeInvalidCreds, "Invalid credentials", true).
	WithSuggestedAction("Download OAuth2 credentials from Google Cloud Console")
