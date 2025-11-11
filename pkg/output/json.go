package output

import (
	"encoding/json"

	"github.com/btafoya/gcal-cli/pkg/types"
)

// JSONFormatter formats output as JSON
type JSONFormatter struct {
	PrettyPrint bool
}

// Format formats a response as JSON
func (f *JSONFormatter) Format(response *types.Response) (string, error) {
	var data []byte
	var err error

	if f.PrettyPrint {
		data, err = json.MarshalIndent(response, "", "  ")
	} else {
		data, err = json.Marshal(response)
	}

	if err != nil {
		return "", err
	}

	return string(data), nil
}
