package calendar

import (
	"context"
	"fmt"
	"sync"

	"github.com/btafoya/gcal-cli/pkg/types"
)

// BatchResult represents the result of a batch operation
type BatchResult struct {
	Success bool        `json:"success"`
	Index   int         `json:"index"`
	EventID string      `json:"eventId,omitempty"`
	Event   *types.Event `json:"event,omitempty"`
	Error   *types.AppError `json:"error,omitempty"`
}

// BatchCreateParams contains parameters for batch event creation
type BatchCreateParams struct {
	Events      []CreateEventParams
	ContinueOnError bool
	MaxConcurrent   int
}

// BatchCreateEvents creates multiple events concurrently
func (c *Client) BatchCreateEvents(ctx context.Context, params BatchCreateParams) ([]*BatchResult, error) {
	if len(params.Events) == 0 {
		return nil, types.ErrInvalidInput("events", "no events provided")
	}

	// Set default concurrency
	if params.MaxConcurrent <= 0 {
		params.MaxConcurrent = 5
	}

	results := make([]*BatchResult, len(params.Events))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, params.MaxConcurrent)
	var mu sync.Mutex

	for i, eventParams := range params.Events {
		wg.Add(1)
		go func(index int, ep CreateEventParams) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Create event
			event, err := c.CreateEvent(ctx, ep)

			// Store result
			mu.Lock()
			if err != nil {
				appErr, ok := err.(*types.AppError)
				if !ok {
					appErr = types.ErrAPIError.
						WithDetails(fmt.Sprintf("failed to create event at index %d", index)).
						WithWrappedError(err)
				}
				results[index] = &BatchResult{
					Success: false,
					Index:   index,
					Error:   appErr,
				}
			} else {
				results[index] = &BatchResult{
					Success: true,
					Index:   index,
					EventID: event.ID,
					Event:   event,
				}
			}
			mu.Unlock()
		}(i, eventParams)
	}

	wg.Wait()

	// Check if we should fail on any error
	if !params.ContinueOnError {
		for _, result := range results {
			if !result.Success {
				return results, types.ErrAPIError.
					WithDetails("batch create failed: one or more events failed to create")
			}
		}
	}

	return results, nil
}

// BatchUpdateParams contains parameters for batch event updates
type BatchUpdateParams struct {
	Updates         map[string]CreateEventParams // eventID -> update params
	ContinueOnError bool
	MaxConcurrent   int
}

// BatchUpdateEvents updates multiple events concurrently
func (c *Client) BatchUpdateEvents(ctx context.Context, params BatchUpdateParams) ([]*BatchResult, error) {
	if len(params.Updates) == 0 {
		return nil, types.ErrInvalidInput("updates", "no updates provided")
	}

	// Set default concurrency
	if params.MaxConcurrent <= 0 {
		params.MaxConcurrent = 5
	}

	results := make([]*BatchResult, 0, len(params.Updates))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, params.MaxConcurrent)
	var mu sync.Mutex

	index := 0
	for eventID, updateParams := range params.Updates {
		wg.Add(1)
		go func(idx int, id string, up CreateEventParams) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Update event
			event, err := c.UpdateEvent(ctx, id, up)

			// Store result
			mu.Lock()
			if err != nil {
				appErr, ok := err.(*types.AppError)
				if !ok {
					appErr = types.ErrAPIError.
						WithDetails(fmt.Sprintf("failed to update event %s", id)).
						WithWrappedError(err)
				}
				results = append(results, &BatchResult{
					Success: false,
					Index:   idx,
					EventID: id,
					Error:   appErr,
				})
			} else {
				results = append(results, &BatchResult{
					Success: true,
					Index:   idx,
					EventID: id,
					Event:   event,
				})
			}
			mu.Unlock()
		}(index, eventID, updateParams)
		index++
	}

	wg.Wait()

	// Check if we should fail on any error
	if !params.ContinueOnError {
		for _, result := range results {
			if !result.Success {
				return results, types.ErrAPIError.
					WithDetails("batch update failed: one or more events failed to update")
			}
		}
	}

	return results, nil
}

// BatchDeleteParams contains parameters for batch event deletion
type BatchDeleteParams struct {
	EventIDs        []string
	ContinueOnError bool
	MaxConcurrent   int
}

// BatchDeleteEvents deletes multiple events concurrently
func (c *Client) BatchDeleteEvents(ctx context.Context, params BatchDeleteParams) ([]*BatchResult, error) {
	if len(params.EventIDs) == 0 {
		return nil, types.ErrInvalidInput("eventIds", "no event IDs provided")
	}

	// Set default concurrency
	if params.MaxConcurrent <= 0 {
		params.MaxConcurrent = 5
	}

	results := make([]*BatchResult, len(params.EventIDs))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, params.MaxConcurrent)
	var mu sync.Mutex

	for i, eventID := range params.EventIDs {
		wg.Add(1)
		go func(index int, id string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Delete event
			err := c.DeleteEvent(ctx, id)

			// Store result
			mu.Lock()
			if err != nil {
				appErr, ok := err.(*types.AppError)
				if !ok {
					appErr = types.ErrAPIError.
						WithDetails(fmt.Sprintf("failed to delete event %s", id)).
						WithWrappedError(err)
				}
				results[index] = &BatchResult{
					Success: false,
					Index:   index,
					EventID: id,
					Error:   appErr,
				}
			} else {
				results[index] = &BatchResult{
					Success: true,
					Index:   index,
					EventID: id,
				}
			}
			mu.Unlock()
		}(i, eventID)
	}

	wg.Wait()

	// Check if we should fail on any error
	if !params.ContinueOnError {
		for _, result := range results {
			if !result.Success {
				return results, types.ErrAPIError.
					WithDetails("batch delete failed: one or more events failed to delete")
			}
		}
	}

	return results, nil
}

// GetBatchSummary returns a summary of batch operation results
func GetBatchSummary(results []*BatchResult) map[string]int {
	summary := map[string]int{
		"total":   len(results),
		"success": 0,
		"failed":  0,
	}

	for _, result := range results {
		if result.Success {
			summary["success"]++
		} else {
			summary["failed"]++
		}
	}

	return summary
}
