package main

import (
	"testing"

	"github.com/kluzzebass/gqlt"
)

// GraphQLTestHelper provides utilities for common GraphQL testing patterns
type GraphQLTestHelper struct {
	client *gqlt.Client
	t      *testing.T
}

// NewGraphQLTestHelper creates a new test helper
func NewGraphQLTestHelper(t *testing.T, endpoint string) *GraphQLTestHelper {
	client := gqlt.NewClient(endpoint, nil)
	return &GraphQLTestHelper{
		client: client,
		t:      t,
	}
}

// AssertNoErrors checks that a response has no GraphQL errors
func (h *GraphQLTestHelper) AssertNoErrors(response *gqlt.Response) {
	if len(response.Errors) > 0 {
		h.t.Errorf("Expected no GraphQL errors, got: %v", response.Errors)
	}
}

// AssertHasErrors checks that a response has GraphQL errors
func (h *GraphQLTestHelper) AssertHasErrors(response *gqlt.Response) {
	if len(response.Errors) == 0 {
		h.t.Error("Expected GraphQL errors, but got none")
	}
}

// AssertFieldExists checks that a specific field exists in the response data
func (h *GraphQLTestHelper) AssertFieldExists(response *gqlt.Response, fieldPath string) {
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		h.t.Fatal("Expected response data to be a map")
	}
	
	// Simple field path support (e.g., "user.name")
	fields := splitFieldPath(fieldPath)
	current := data
	
	for i, field := range fields {
		value, exists := current[field]
		if !exists {
			h.t.Errorf("Expected field '%s' at path '%s'", field, fieldPath)
			return
		}
		
		if i < len(fields)-1 {
			// Not the last field, should be a map
			current, ok = value.(map[string]interface{})
			if !ok {
				h.t.Errorf("Expected field '%s' to be a map at path '%s'", field, fieldPath)
				return
			}
		}
	}
}

// AssertFieldValue checks that a field has a specific value
func (h *GraphQLTestHelper) AssertFieldValue(response *gqlt.Response, fieldPath string, expected interface{}) {
	value := h.GetFieldValue(response, fieldPath)
	if value != expected {
		h.t.Errorf("Expected field '%s' to be %v, got %v", fieldPath, expected, value)
	}
}

// GetFieldValue retrieves a field value from response data
func (h *GraphQLTestHelper) GetFieldValue(response *gqlt.Response, fieldPath string) interface{} {
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		h.t.Fatal("Expected response data to be a map")
	}
	
	fields := splitFieldPath(fieldPath)
	current := data
	
	for i, field := range fields {
		value, exists := current[field]
		if !exists {
			h.t.Errorf("Field '%s' not found at path '%s'", field, fieldPath)
			return nil
		}
		
		if i < len(fields)-1 {
			current, ok = value.(map[string]interface{})
			if !ok {
				h.t.Errorf("Expected field '%s' to be a map at path '%s'", field, fieldPath)
				return nil
			}
		} else {
			return value
		}
	}
	
	return nil
}

// AssertArrayLength checks that an array field has a specific length
func (h *GraphQLTestHelper) AssertArrayLength(response *gqlt.Response, fieldPath string, expectedLength int) {
	value := h.GetFieldValue(response, fieldPath)
	array, ok := value.([]interface{})
	if !ok {
		h.t.Errorf("Expected field '%s' to be an array", fieldPath)
		return
	}
	
	if len(array) != expectedLength {
		h.t.Errorf("Expected array '%s' to have length %d, got %d", fieldPath, expectedLength, len(array))
	}
}

// ExecuteQuery is a convenience method for executing queries
func (h *GraphQLTestHelper) ExecuteQuery(query string, variables map[string]interface{}, operationName string) *gqlt.Response {
	response, err := h.client.Execute(query, variables, operationName)
	if err != nil {
		h.t.Fatalf("Query execution failed: %v", err)
	}
	return response
}

// ExecuteMutation is a convenience method for executing mutations
func (h *GraphQLTestHelper) ExecuteMutation(mutation string, variables map[string]interface{}, operationName string) *gqlt.Response {
	response, err := h.client.Execute(mutation, variables, operationName)
	if err != nil {
		h.t.Fatalf("Mutation execution failed: %v", err)
	}
	return response
}

// SetAuth configures authentication for the client
func (h *GraphQLTestHelper) SetAuth(authType string, credentials map[string]string) {
	switch authType {
	case "bearer":
		if token, exists := credentials["token"]; exists {
			// Set bearer token via headers
			h.client.SetHeaders(map[string]string{
				"Authorization": "Bearer " + token,
			})
		}
	case "basic":
		if username, exists := credentials["username"]; exists {
			if password, exists := credentials["password"]; exists {
				h.client.SetAuth(username, password)
			}
		}
	case "apikey":
		if apiKey, exists := credentials["apikey"]; exists {
			// Set API key via headers
			h.client.SetHeaders(map[string]string{
				"X-API-Key": apiKey,
			})
		}
	}
}

// SetHeaders sets custom headers for the client
func (h *GraphQLTestHelper) SetHeaders(headers map[string]string) {
	h.client.SetHeaders(headers)
}

// Helper function to split field paths like "user.name" into ["user", "name"]
func splitFieldPath(path string) []string {
	// Simple implementation - can be enhanced for more complex paths
	var fields []string
	var current string
	
	for _, char := range path {
		if char == '.' {
			if current != "" {
				fields = append(fields, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	
	if current != "" {
		fields = append(fields, current)
	}
	
	return fields
}

// ExampleTestWithHelper demonstrates using the test helper
func ExampleTestWithHelper(t *testing.T) {
	helper := NewGraphQLTestHelper(t, "https://api.example.com/graphql")
	
	// Set authentication if needed
	helper.SetAuth("bearer", map[string]string{
		"token": "your-token-here",
	})
	
	// Execute a query
	response := helper.ExecuteQuery("{ users { id name } }", nil, "")
	
	// Use assertions
	helper.AssertNoErrors(response)
	helper.AssertFieldExists(response, "users")
	helper.AssertArrayLength(response, "users", 2) // Expect 2 users
	
	// Check specific values
	firstUser := helper.GetFieldValue(response, "users.0")
	if firstUser == nil {
		t.Error("Expected first user to exist")
	}
}
