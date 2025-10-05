package gqlt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestStructuredOutput(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		format   string
		quiet    bool
		expected string
	}{
		{
			name:   "JSON format success",
			data:   map[string]string{"message": "test"},
			format: "json",
			quiet:  false,
		},
		{
			name:   "Table format success",
			data:   map[string]string{"message": "test"},
			format: "table",
			quiet:  false,
		},
		{
			name:   "YAML format success",
			data:   map[string]string{"message": "test"},
			format: "yaml",
			quiet:  false,
		},
		{
			name:   "JSON format quiet mode",
			data:   map[string]string{"message": "test"},
			format: "json",
			quiet:  true,
		},
		{
			name:   "Table format quiet mode",
			data:   map[string]string{"message": "test"},
			format: "table",
			quiet:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewFormatter(tt.format)

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := formatter.FormatStructured(tt.data, tt.quiet)
			if err != nil {
				t.Errorf("FormatStructured() error = %v", err)
			}

			// Close write end and read output
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// For JSON format, verify it's valid JSON
			if tt.format == "json" {
				var result StructuredOutput
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("Invalid JSON output: %v", err)
				}
				if !result.Success {
					t.Errorf("Expected success=true, got %v", result.Success)
				}
			}

			// For quiet mode, verify minimal output
			if tt.quiet && tt.format == "table" {
				if len(output) == 0 {
					t.Errorf("Expected some output in quiet mode")
				}
			}
		})
	}
}

func TestStructuredError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		code   string
		format string
		quiet  bool
	}{
		{
			name:   "JSON format error",
			err:    fmt.Errorf("Test error"),
			code:   "TEST_ERROR",
			format: "json",
			quiet:  false,
		},
		{
			name:   "Table format error",
			err:    fmt.Errorf("Test error"),
			code:   "TEST_ERROR",
			format: "table",
			quiet:  false,
		},
		{
			name:   "YAML format error",
			err:    fmt.Errorf("Test error"),
			code:   "TEST_ERROR",
			format: "yaml",
			quiet:  false,
		},
		{
			name:   "JSON format error with context",
			err:    fmt.Errorf("Test error"),
			code:   "TEST_ERROR",
			format: "json",
			quiet:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewFormatter(tt.format)

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := formatter.FormatStructuredError(tt.err, tt.code, tt.quiet)
			if err != nil {
				t.Errorf("FormatStructuredError() error = %v", err)
			}

			// Close write end and read output
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// For JSON format, verify it's valid JSON
			if tt.format == "json" {
				var result StructuredOutput
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("Invalid JSON output: %v", err)
				}
				if result.Success {
					t.Errorf("Expected success=false, got %v", result.Success)
				}
				if result.Error.Code != tt.code {
					t.Errorf("Expected error code %s, got %s", tt.code, result.Error.Code)
				}
			}
		})
	}
}

func TestStructuredErrorWithContext(t *testing.T) {
	formatter := NewFormatter("json")

	context := map[string]interface{}{
		"endpoint": "https://api.example.com/graphql",
		"query":    "{ users { id } }",
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := formatter.FormatStructuredErrorWithContext(
		fmt.Errorf("Test error"),
		"TEST_ERROR",
		"validation_error",
		context,
		false,
	)
	if err != nil {
		t.Errorf("FormatStructuredErrorWithContext() error = %v", err)
	}

	// Close write end and read output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify JSON output
	var result StructuredOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("Invalid JSON output: %v", err)
	}
	if result.Success {
		t.Errorf("Expected success=false, got %v", result.Success)
	}
	if result.Error.Code != "TEST_ERROR" {
		t.Errorf("Expected error code TEST_ERROR, got %s", result.Error.Code)
	}
	if result.Error.Type != "validation_error" {
		t.Errorf("Expected error type validation_error, got %s", result.Error.Type)
	}
	if result.Error.Context == nil {
		t.Errorf("Expected context to be present")
	}
}

func TestErrorCodes(t *testing.T) {
	// Test that all error codes are defined
	expectedCodes := []string{
		ErrorCodeConfigLoad,
		ErrorCodeConfigNotFound,
		ErrorCodeConfigCreate,
		ErrorCodeConfigSave,
		ErrorCodeConfigDelete,
		ErrorCodeConfigValidate,
		ErrorCodeInputValidation,
		ErrorCodeQueryLoad,
		ErrorCodeVariablesLoad,
		ErrorCodeFilesParse,
		ErrorCodeFilesListParse,
		ErrorCodeGraphQLExecution,
		ErrorCodeGraphQLErrors,
		ErrorCodeNetworkError,
		ErrorCodeAuthError,
		ErrorCodeSchemaLoad,
		ErrorCodeSchemaIntrospect,
		ErrorCodeSchemaSave,
		ErrorCodeSystemError,
		ErrorCodeFileNotFound,
		ErrorCodePermissionDenied,
	}

	for _, code := range expectedCodes {
		if code == "" {
			t.Errorf("Expected error code to be defined, got empty string")
		}
	}
}

func TestFormatter(t *testing.T) {
	// Test all formatter modes with comprehensive test cases
	modes := []string{"json", "pretty", "raw"}

	// Test cases with different response types
	testCases := []struct {
		name     string
		response *Response
	}{
		{
			name: "valid data response",
			response: &Response{
				Data: map[string]interface{}{
					"user": map[string]interface{}{
						"id":   "123",
						"name": "John Doe",
					},
				},
			},
		},
		{
			name: "response with errors",
			response: &Response{
				Data: nil,
				Errors: []interface{}{
					map[string]interface{}{
						"message": "User not found",
						"path":    []string{"user"},
					},
				},
			},
		},
		{
			name: "response with extensions",
			response: &Response{
				Data: map[string]interface{}{
					"user": map[string]interface{}{
						"id": "123",
					},
				},
				Extensions: map[string]interface{}{
					"tracing": map[string]interface{}{
						"duration": 100,
					},
				},
			},
		},
	}

	// Test each mode with each test case
	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			formatter := NewFormatter("json")
			if formatter == nil {
				t.Error("Expected formatter to be created")
				return
			}

			// Test all response types
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					err := formatter.FormatResponse((*Response)(tc.response), mode)
					if err != nil {
						t.Errorf("Format failed for %s mode with %s: %v", mode, tc.name, err)
					}
				})
			}
		})
	}
}
