package types

import "time"

// Response represents a standard API response structure
type Response struct {
	Success   bool                   `json:"success"`
	Operation string                 `json:"operation,omitempty"`
	Data      interface{}            `json:"data,omitempty"`
	Error     *AppError              `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// SuccessResponse creates a successful response
func SuccessResponse(operation string, data interface{}) *Response {
	return &Response{
		Success:   true,
		Operation: operation,
		Data:      data,
		Metadata: map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// ErrorResponse creates an error response
func ErrorResponse(err *AppError) *Response {
	return &Response{
		Success: false,
		Error:   err,
		Metadata: map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// WithMetadata adds metadata to the response
func (r *Response) WithMetadata(key string, value interface{}) *Response {
	if r.Metadata == nil {
		r.Metadata = make(map[string]interface{})
	}
	r.Metadata[key] = value
	return r
}

// EventData wraps event data in responses
type EventData struct {
	Event   *Event `json:"event,omitempty"`
	EventID string `json:"eventId,omitempty"`
	Message string `json:"message,omitempty"`
}

// EventListData wraps event list data in responses
type EventListData struct {
	Events        []*Event `json:"events"`
	Count         int      `json:"count"`
	NextPageToken string   `json:"nextPageToken,omitempty"`
}

// AuthData wraps authentication data in responses
type AuthData struct {
	Message string   `json:"message"`
	Email   string   `json:"email,omitempty"`
	Scopes  []string `json:"scopes,omitempty"`
}
