package output

import (
	"fmt"
	"strings"

	"github.com/btafoya/gcal-cli/pkg/types"
)

// MinimalFormatter formats output as minimal (IDs only)
type MinimalFormatter struct{}

// Format formats a response as minimal output (IDs only)
func (f *MinimalFormatter) Format(response *types.Response) (string, error) {
	var builder strings.Builder

	if !response.Success {
		// For errors, just output the error code
		if response.Error != nil {
			builder.WriteString(fmt.Sprintf("ERROR: %s\n", response.Error.Code))
		} else {
			builder.WriteString("ERROR: UNKNOWN\n")
		}
		return builder.String(), nil
	}

	// Extract IDs from different data types
	switch data := response.Data.(type) {
	case *types.EventData:
		if data.Event != nil {
			builder.WriteString(fmt.Sprintf("%s\n", data.Event.ID))
		} else if data.EventID != "" {
			builder.WriteString(fmt.Sprintf("%s\n", data.EventID))
		}
	case *types.EventListData:
		for _, event := range data.Events {
			builder.WriteString(fmt.Sprintf("%s\n", event.ID))
		}
	case *types.AuthData:
		builder.WriteString(fmt.Sprintf("OK\n"))
	default:
		builder.WriteString("OK\n")
	}

	return builder.String(), nil
}
