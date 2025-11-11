package output

import (
	"github.com/btafoya/gcal-cli/pkg/types"
)

// Format represents output format types
type Format string

const (
	// FormatJSON is the JSON output format (default for LLM agents)
	FormatJSON Format = "json"
	// FormatText is the human-readable text format
	FormatText Format = "text"
	// FormatMinimal is the minimal output format (IDs only)
	FormatMinimal Format = "minimal"
)

// Formatter is the interface for output formatting
type Formatter interface {
	// Format formats a response into the desired output format
	Format(response *types.Response) (string, error)
}

// NewFormatter creates a new formatter based on the specified format
func NewFormatter(format Format) Formatter {
	switch format {
	case FormatText:
		return &TextFormatter{}
	case FormatMinimal:
		return &MinimalFormatter{}
	case FormatJSON:
		fallthrough
	default:
		return &JSONFormatter{PrettyPrint: true}
	}
}

// ParseFormat converts a string to a Format type
func ParseFormat(s string) Format {
	switch s {
	case "text":
		return FormatText
	case "minimal":
		return FormatMinimal
	case "json":
		fallthrough
	default:
		return FormatJSON
	}
}
