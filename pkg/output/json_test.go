package output

import (
	"encoding/json"
	"testing"

	"github.com/btafoya/gcal-cli/pkg/types"
)

func TestJSONFormatter_Format_Success(t *testing.T) {
	formatter := &JSONFormatter{PrettyPrint: false}

	data := map[string]interface{}{
		"message": "test message",
		"value":   123,
	}

	response := types.SuccessResponse("test_operation", data)

	output, err := formatter.Format(response)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Parse the output to verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// Verify structure
	if success, ok := result["success"].(bool); !ok || !success {
		t.Error("Expected success to be true")
	}

	if operation, ok := result["operation"].(string); !ok || operation != "test_operation" {
		t.Errorf("Expected operation to be 'test_operation', got %v", operation)
	}
}

func TestJSONFormatter_Format_Error(t *testing.T) {
	formatter := &JSONFormatter{PrettyPrint: false}

	appErr := types.ErrInvalidInput("testField", "test reason")
	response := types.ErrorResponse(appErr)

	output, err := formatter.Format(response)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Parse the output
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}

	// Verify error structure
	if success, ok := result["success"].(bool); !ok || success {
		t.Error("Expected success to be false")
	}

	errorObj, ok := result["error"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected error object")
	}

	if code, ok := errorObj["code"].(string); !ok || code != types.ErrCodeInvalidInput {
		t.Errorf("Expected error code %s, got %v", types.ErrCodeInvalidInput, code)
	}
}

func TestJSONFormatter_Format_PrettyPrint(t *testing.T) {
	formatter := &JSONFormatter{PrettyPrint: true}

	response := types.SuccessResponse("test", map[string]interface{}{"key": "value"})

	output, err := formatter.Format(response)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Check for indentation (pretty print indicator)
	if len(output) < 10 {
		t.Error("Pretty printed output should be longer")
	}

	// Should contain newlines and spaces for indentation
	if !containsChar(output, '\n') {
		t.Error("Pretty printed output should contain newlines")
	}
}

func containsChar(s string, ch rune) bool {
	for _, c := range s {
		if c == ch {
			return true
		}
	}
	return false
}
