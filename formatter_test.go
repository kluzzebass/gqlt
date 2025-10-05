package gqlt

import (
	"testing"
)

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
			formatter := NewFormatter()
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
