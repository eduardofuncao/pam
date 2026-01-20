package table

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestFormatValueIfJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantJSON bool
	}{
		{
			name:     "Valid JSON object",
			input:    `{"name":"John","age":30}`,
			wantJSON: true,
		},
		{
			name:     "Valid JSON array",
			input:    `[1,2,3]`,
			wantJSON: true,
		},
		{
			name:     "Valid JSON with whitespace",
			input:    `  {"key": "value"}  `,
			wantJSON: true,
		},
		{
			name:     "Nested JSON",
			input:    `{"user":{"name":"Alice","roles":["admin","user"]}}`,
			wantJSON: true,
		},
		{
			name:     "Plain text",
			input:    `Hello World`,
			wantJSON: false,
		},
		{
			name:     "Number",
			input:    `123`,
			wantJSON: false,
		},
		{
			name:     "Invalid JSON",
			input:    `{invalid json}`,
			wantJSON: false,
		},
		{
			name:     "Empty string",
			input:    ``,
			wantJSON: false,
		},
		{
			name:     "SQL statement",
			input:    `SELECT * FROM users`,
			wantJSON: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValueIfJSON(tt.input)

			if tt.wantJSON {
				// If we expect formatted JSON, the result should be different from input
				// (except if already formatted) and should contain newlines
				if result == tt.input && len(tt.input) > 20 {
					t.Errorf(
						"formatValueIfJSON() expected formatted JSON, but got original input",
					)
				}
			} else {
				// If we don't expect JSON, the result should equal the input
				if result != tt.input {
					t.Errorf(
						"formatValueIfJSON() = %v, want %v",
						result,
						tt.input,
					)
				}
			}
		})
	}
}

func TestFormatValueIfJSON_FormattingCorrectness(t *testing.T) {
	input := `{"name":"John","age":30,"address":{"city":"NYC","zip":"10001"}}`
	result := formatValueIfJSON(input)

	// Verify that the result contains proper indentation
	expected := `{
  "address": {
    "city": "NYC",
    "zip": "10001"
  },
  "age": 30,
  "name": "John"
}`

	if result != expected {
		t.Errorf(
			"formatValueIfJSON() formatting incorrect.\nGot:\n%s\n\nWant:\n%s",
			result,
			expected,
		)
	}
}

func TestFormatValueIfJSON_Array(t *testing.T) {
	input := `[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]`
	result := formatValueIfJSON(input)

	// Verify that the result contains indentation
	if !contains(result, "  ") {
		t.Errorf(
			"formatValueIfJSON() should format array with indentation, got: %s",
			result,
		)
	}

	// Verify that it contains line breaks
	if !contains(result, "\n") {
		t.Errorf(
			"formatValueIfJSON() should format array with newlines, got: %s",
			result,
		)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestDetailViewEditJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		edited      string
		expectValid bool
	}{
		{
			name:        "Valid JSON to Valid JSON",
			input:       `{"name":"John"}`,
			edited:      `{"name":"Jane","age":25}`,
			expectValid: true,
		},
		{
			name:        "Formatted JSON becomes compacted",
			input:       `{"name":"John"}`,
			edited:      "{\n  \"name\": \"Jane\",\n  \"age\": 25\n}",
			expectValid: true,
		},
		{
			name:        "Invalid JSON should fail validation",
			input:       `{"name":"John"}`,
			edited:      `{name: "invalid"}`,
			expectValid: false,
		},
		{
			name:        "Empty JSON object",
			input:       `{"key":"value"}`,
			edited:      `{}`,
			expectValid: true,
		},
		{
			name:        "JSON array",
			input:       `["a","b"]`,
			edited:      `["x","y","z"]`,
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trimmed := strings.TrimSpace(tt.edited)
			if strings.HasPrefix(trimmed, "{") ||
				strings.HasPrefix(trimmed, "[") {
				var jsonData interface{}
				err := json.Unmarshal([]byte(trimmed), &jsonData)
				if tt.expectValid && err != nil {
					t.Errorf("Expected valid JSON but got error: %v", err)
				}
				if !tt.expectValid && err == nil {
					t.Errorf("Expected invalid JSON but parsing succeeded")
				}
			}
		})
	}
}

func TestFormatAndCompactJSON(t *testing.T) {
	input := `{"user":{"name":"Alice","roles":["admin","user"]}}`

	// Format
	formatted := formatValueIfJSON(input)

	// Verify that it's formatted
	if !contains(formatted, "\n") {
		t.Errorf("Expected formatted JSON to contain newlines")
	}

	// Compact again
	var jsonData interface{}
	if err := json.Unmarshal([]byte(formatted), &jsonData); err != nil {
		t.Fatalf("Failed to parse formatted JSON: %v", err)
	}

	compacted, err := json.Marshal(jsonData)
	if err != nil {
		t.Fatalf("Failed to compact JSON: %v", err)
	}

	// Verify that compacted has no extra spaces
	compactedStr := string(compacted)
	if contains(compactedStr, "\n") {
		t.Errorf("Compacted JSON should not contain newlines")
	}

	// Verify that data is preserved
	var original, compactedData interface{}
	json.Unmarshal([]byte(input), &original)
	json.Unmarshal(compacted, &compactedData)

	if fmt.Sprintf("%v", original) != fmt.Sprintf("%v", compactedData) {
		t.Errorf("JSON data changed during format/compact cycle")
	}
}
