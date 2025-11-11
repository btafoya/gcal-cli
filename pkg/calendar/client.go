package calendar

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	"github.com/btafoya/gcal-cli/pkg/types"
)

// Client wraps the Google Calendar API client with retry logic
type Client struct {
	Service    *calendar.Service
	CalendarID string
	MaxRetries int
	RetryDelay time.Duration
}

// NewClient creates a new calendar client
func NewClient(service *calendar.Service, calendarID string) *Client {
	return &Client{
		Service:    service,
		CalendarID: calendarID,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}
}

// withRetry executes a function with exponential backoff retry logic
func (c *Client) withRetry(ctx context.Context, operation string, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			delay := c.RetryDelay * time.Duration(1<<uint(attempt-1))
			select {
			case <-ctx.Done():
				return types.ErrNetworkError("operation cancelled")
			case <-time.After(delay):
			}
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryable(err) {
			break
		}
	}

	// Wrap the final error
	return types.ErrAPIError.
		WithDetails(fmt.Sprintf("%s failed after %d attempts", operation, c.MaxRetries+1)).
		WithWrappedError(lastErr)
}

// isRetryable determines if an error should be retried
func isRetryable(err error) bool {
	if apiErr, ok := err.(*googleapi.Error); ok {
		// Retry on rate limit, server errors, and timeouts
		switch apiErr.Code {
		case 429: // Too Many Requests
			return true
		case 500, 502, 503, 504: // Server errors
			return true
		}
	}
	return false
}

// handleAPIError converts Google API errors to application errors
func handleAPIError(err error, operation string) error {
	if err == nil {
		return nil
	}

	apiErr, ok := err.(*googleapi.Error)
	if !ok {
		return types.ErrAPIError.
			WithDetails(fmt.Sprintf("%s failed", operation)).
			WithWrappedError(err)
	}

	switch apiErr.Code {
	case 400:
		return types.ErrInvalidInput("request", apiErr.Message).
			WithWrappedError(err)
	case 401:
		return types.ErrAuthFailed("invalid or expired credentials").
			WithWrappedError(err).
			WithSuggestedAction("Run 'gcal-cli auth login' to re-authenticate")
	case 403:
		return types.NewAppError(types.ErrCodePermissionDenied,
			"insufficient permissions", true).
			WithDetails(apiErr.Message).
			WithWrappedError(err).
			WithSuggestedAction("Check calendar sharing settings")
	case 404:
		return types.ErrNotFound("event", "").
			WithDetails(apiErr.Message).
			WithWrappedError(err)
	case 409:
		return types.NewAppError(types.ErrCodeInvalidInput,
			"conflict with existing event", true).
			WithDetails(apiErr.Message).
			WithWrappedError(err)
	case 429:
		return types.ErrRateLimit().
			WithWrappedError(err)
	case 500, 502, 503, 504:
		return types.NewAppError(types.ErrCodeAPIError,
			"Google Calendar service error", true).
			WithDetails(apiErr.Message).
			WithWrappedError(err).
			WithSuggestedAction("Try again in a few moments")
	default:
		return types.ErrAPIError.
			WithDetails(fmt.Sprintf("API error %d: %s", apiErr.Code, apiErr.Message)).
			WithWrappedError(err)
	}
}
