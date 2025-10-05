package output

import (
	"testing"

	"github.com/kluzzebass/gqlt/internal/graphql"
)

func TestFormatter(t *testing.T) {
	// Test all formatter modes with comprehensive test cases
	modes := []string{"json", "pretty", "raw"}

	// Test cases with different response types
	testCases := []struct {
		name     string
		response *graphql.Response
	}{
		{
			name: "valid data response",
			response: &graphql.Response{
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
			response: &graphql.Response{
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
			response: &graphql.Response{
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
			formatter := NewFormatter(mode)
			if formatter == nil {
				t.Error("Expected formatter to be created")
			}
			if formatter.mode != mode {
				t.Errorf("Expected mode '%s', got %s", mode, formatter.mode)
			}

			// Test all response types
			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					err := formatter.Format(tc.response)
					if err != nil {
						t.Errorf("Format failed for %s mode with %s: %v", mode, tc.name, err)
					}
				})
			}
		})
	}
}
