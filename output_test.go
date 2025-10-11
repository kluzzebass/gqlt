package gqlt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestSetOutputAndSetErrorOutput(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"JSON formatter", "json"},
		{"Table formatter", "table"},
		{"YAML formatter", "yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewFormatter(tt.format)
			if formatter == nil {
				t.Fatal("Expected formatter to be created")
			}

			// Create custom output buffers
			outputBuf := &bytes.Buffer{}
			errorBuf := &bytes.Buffer{}

			// Set custom output writers
			formatter.SetOutput(outputBuf)
			formatter.SetErrorOutput(errorBuf)

			// Test normal output goes to custom output buffer
			testData := map[string]string{"message": "test output"}
			err := formatter.FormatStructured(testData, false)
			if err != nil {
				t.Errorf("FormatStructured() error = %v", err)
			}

			// Verify output went to custom buffer, not stdout
			output := outputBuf.String()
			if len(output) == 0 {
				t.Errorf("Expected output in custom buffer, got empty")
			}

			// Test error output goes to custom error buffer
			testError := fmt.Errorf("test error")
			err = formatter.FormatStructuredError(testError, "TEST_ERROR", false)
			if err != nil {
				t.Errorf("FormatStructuredError() error = %v", err)
			}

			// Verify error output went to custom error buffer
			errorOutput := errorBuf.String()
			if len(errorOutput) == 0 {
				t.Errorf("Expected error output in custom error buffer, got empty")
			}

			// Verify outputs are different (normal vs error)
			if output == errorOutput {
				t.Errorf("Expected different output for normal vs error, got same")
			}
		})
	}
}

func TestGetOutputAndGetErrorOutput(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"JSON formatter", "json"},
		{"Table formatter", "table"},
		{"YAML formatter", "yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewFormatter(tt.format)
			if formatter == nil {
				t.Fatal("Expected formatter to be created")
			}

			// Test default output (should be os.Stdout) - use interface methods
			outputBuf := &bytes.Buffer{}
			formatter.SetOutput(outputBuf)

			// Test that we can set and get output
			testData := map[string]string{"test": "data"}
			err := formatter.FormatStructured(testData, false)
			if err != nil {
				t.Errorf("FormatStructured failed: %v", err)
			}

			output := outputBuf.String()
			if len(output) == 0 {
				t.Errorf("Expected output in custom buffer")
			}

			// Test default error output (should be os.Stderr) - use interface methods
			errorBuf := &bytes.Buffer{}
			formatter.SetErrorOutput(errorBuf)

			// Test that we can set and get error output
			testError := fmt.Errorf("test error")
			err = formatter.FormatStructuredError(testError, "TEST_ERROR", false)
			if err != nil {
				t.Errorf("FormatStructuredError failed: %v", err)
			}

			errorOutput := errorBuf.String()
			if len(errorOutput) == 0 {
				t.Errorf("Expected error output in custom error buffer")
			}
		})
	}
}

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
			outputBuf := &bytes.Buffer{}
			formatter.SetOutput(outputBuf)

			err := formatter.FormatStructured(tt.data, tt.quiet)
			if err != nil {
				t.Errorf("FormatStructured() error = %v", err)
			}

			output := outputBuf.String()

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
			errorBuf := &bytes.Buffer{}
			formatter.SetErrorOutput(errorBuf)

			err := formatter.FormatStructuredError(tt.err, tt.code, tt.quiet)
			if err != nil {
				t.Errorf("FormatStructuredError() error = %v", err)
			}

			output := errorBuf.String()

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
	tests := []struct {
		name   string
		format string
	}{
		{"JSON formatter", "json"},
		{"Table formatter", "table"},
		{"YAML formatter", "yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewFormatter(tt.format)
			errorBuf := &bytes.Buffer{}
			formatter.SetErrorOutput(errorBuf)

			context := map[string]interface{}{
				"endpoint": "https://api.example.com/graphql",
				"query":    "{ users { id } }",
			}

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

			output := errorBuf.String()

			// For JSON format, verify it's valid JSON with context
			if tt.format == "json" {
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

			// For all formats, verify output is not empty
			if len(output) == 0 {
				t.Errorf("Expected non-empty output for %s format", tt.format)
			}
		})
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

func TestFormatResponse(t *testing.T) {
	// Test all formatter types with FormatResponse
	formats := []string{"json", "table", "yaml"}
	modes := []string{"compact"}

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

	// Test each format with each mode and test case
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			formatter := NewFormatter(format)
			if formatter == nil {
				t.Error("Expected formatter to be created")
				return
			}

			outputBuf := &bytes.Buffer{}
			formatter.SetOutput(outputBuf)

			// Test all response types with all modes
			for _, mode := range modes {
				t.Run(mode, func(t *testing.T) {
					for _, tc := range testCases {
						t.Run(tc.name, func(t *testing.T) {
							outputBuf.Reset() // Clear buffer for each test

							err := formatter.FormatResponse(tc.response, mode)
							if err != nil {
								t.Errorf("FormatResponse failed for %s format, %s mode with %s: %v", format, mode, tc.name, err)
							}

							// Verify output was written
							output := outputBuf.String()
							if len(output) == 0 {
								t.Errorf("Expected non-empty output for %s format, %s mode with %s", format, mode, tc.name)
							}

							// For JSON format and json mode, verify it's valid JSON
							if format == "json" && mode == "json" {
								var result map[string]interface{}
								if err := json.Unmarshal([]byte(output), &result); err != nil {
									t.Errorf("Invalid JSON output for %s: %v", tc.name, err)
								}
							}
						})
					}
				})
			}
		})
	}
}

func TestFormatter(t *testing.T) {
	// Test all formatter modes with comprehensive test cases
	modes := []string{"compact"}

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

			outputBuf := &bytes.Buffer{}
			formatter.SetOutput(outputBuf)

			// Test all response types
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					outputBuf.Reset() // Clear buffer for each test

					err := formatter.FormatResponse((*Response)(tc.response), mode)
					if err != nil {
						t.Errorf("Format failed for %s mode with %s: %v", mode, tc.name, err)
					}

					// Verify output was written
					output := outputBuf.String()
					if len(output) == 0 {
						t.Errorf("Expected non-empty output for %s mode with %s", mode, tc.name)
					}
				})
			}
		})
	}
}

func TestFormatterEdgeCases(t *testing.T) {
	t.Run("nil formatter", func(t *testing.T) {
		formatter := NewFormatter("invalid")
		if formatter != nil {
			t.Errorf("Expected nil formatter for invalid format, got %T", formatter)
		}
	})

	t.Run("nil data handling", func(t *testing.T) {
		formatter := NewFormatter("json")
		outputBuf := &bytes.Buffer{}
		formatter.SetOutput(outputBuf)

		err := formatter.FormatStructured(nil, false)
		if err != nil {
			t.Errorf("FormatStructured with nil data should not error: %v", err)
		}

		output := outputBuf.String()
		if len(output) == 0 {
			t.Errorf("Expected output even with nil data")
		}
	})

	t.Run("empty string data", func(t *testing.T) {
		formatter := NewFormatter("json")
		outputBuf := &bytes.Buffer{}
		formatter.SetOutput(outputBuf)

		err := formatter.FormatStructured("", false)
		if err != nil {
			t.Errorf("FormatStructured with empty string should not error: %v", err)
		}

		output := outputBuf.String()
		if len(output) == 0 {
			t.Errorf("Expected output even with empty string data")
		}
	})

	t.Run("large data handling", func(t *testing.T) {
		formatter := NewFormatter("json")
		outputBuf := &bytes.Buffer{}
		formatter.SetOutput(outputBuf)

		// Create a large data structure
		largeData := make(map[string]interface{})
		for i := 0; i < 1000; i++ {
			largeData[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
		}

		err := formatter.FormatStructured(largeData, false)
		if err != nil {
			t.Errorf("FormatStructured with large data should not error: %v", err)
		}

		output := outputBuf.String()
		if len(output) == 0 {
			t.Errorf("Expected output for large data")
		}

		// Verify it's valid JSON
		var result StructuredOutput
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Errorf("Invalid JSON output for large data: %v", err)
		}
	})
}

func TestFormatterErrorHandling(t *testing.T) {
	t.Run("nil error handling", func(t *testing.T) {
		formatter := NewFormatter("json")
		errorBuf := &bytes.Buffer{}
		formatter.SetErrorOutput(errorBuf)

		err := formatter.FormatStructuredError(nil, "TEST_ERROR", false)
		if err != nil {
			t.Errorf("FormatStructuredError with nil error should not return error: %v", err)
		}

		output := errorBuf.String()
		if len(output) == 0 {
			t.Errorf("Expected output even with nil error")
		}
	})

	t.Run("empty error code", func(t *testing.T) {
		formatter := NewFormatter("json")
		errorBuf := &bytes.Buffer{}
		formatter.SetErrorOutput(errorBuf)

		err := formatter.FormatStructuredError(fmt.Errorf("test error"), "", false)
		if err != nil {
			t.Errorf("FormatStructuredError with empty code should not return error: %v", err)
		}

		output := errorBuf.String()
		if len(output) == 0 {
			t.Errorf("Expected output even with empty error code")
		}
	})

	t.Run("context with nil values", func(t *testing.T) {
		formatter := NewFormatter("json")
		errorBuf := &bytes.Buffer{}
		formatter.SetErrorOutput(errorBuf)

		context := map[string]interface{}{
			"nil_value": nil,
			"empty":     "",
			"zero":      0,
		}

		err := formatter.FormatStructuredErrorWithContext(
			fmt.Errorf("test error"),
			"TEST_ERROR",
			"test_type",
			context,
			false,
		)
		if err != nil {
			t.Errorf("FormatStructuredErrorWithContext with nil context values should not error: %v", err)
		}

		output := errorBuf.String()
		if len(output) == 0 {
			t.Errorf("Expected output even with nil context values")
		}
	})
}

func TestFormatterPerformance(t *testing.T) {
	t.Run("concurrent formatting", func(t *testing.T) {
		// Test concurrent access to formatters - each goroutine gets its own formatter
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				defer func() { done <- true }()

				// Each goroutine gets its own formatter instance
				formatter := NewFormatter("json")
				outputBuf := &bytes.Buffer{}
				formatter.SetOutput(outputBuf)

				data := map[string]interface{}{
					"id":   id,
					"data": fmt.Sprintf("concurrent test %d", id),
				}

				err := formatter.FormatStructured(data, false)
				if err != nil {
					t.Errorf("Concurrent FormatStructured failed for goroutine %d: %v", id, err)
				}

				output := outputBuf.String()
				if len(output) == 0 {
					t.Errorf("Expected output for goroutine %d", id)
				}
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
