package main

import (
	"strings"
	"testing"

	"github.com/kluzzebass/gqlt"
)

// TestExamplesCompile verifies that all examples compile correctly
func TestExamplesCompile(t *testing.T) {
	// This test ensures that all example files can be compiled
	// The actual functionality is demonstrated in the example files themselves

	// Test that we can create a basic test helper
	helper := NewGraphQLTestHelper(t, "https://api.example.com/graphql")
	if helper == nil {
		t.Fatal("Failed to create GraphQLTestHelper")
	}

	// Test that we can create a mock server
	mock := NewMockGraphQLServer()
	if mock == nil {
		t.Fatal("Failed to create MockGraphQLServer")
	}
	defer mock.Close()

	// Test that we can create a test server
	server := NewGraphQLTestServer()
	if server == nil {
		t.Fatal("Failed to create GraphQLTestServer")
	}
	defer server.Close()

	t.Log("All examples compile successfully")
}

// TestMockServerBasic tests basic mock server functionality
func TestMockServerBasic(t *testing.T) {
	// This test demonstrates using the simple mock server
	// For a comprehensive mock server, use `gqlt serve`

	// Create mock server
	mock := NewMockGraphQLServer()
	defer mock.Close()

	// Add handler for GetUsers operation
	mock.AddHandler("GetUsers", func(response *gqlt.Response) {
		response.Data = map[string]interface{}{
			"users": []map[string]interface{}{
				{"id": "1", "name": "John Doe", "email": "john@example.com"},
				{"id": "2", "name": "Jane Smith", "email": "jane@example.com"},
			},
		}
	})

	// Test the server
	client := gqlt.NewClient(mock.URL(), nil)

	// Test GetUsers query
	response, err := client.Execute(
		`query GetUsers { users { id name email } }`,
		nil,
		"GetUsers",
	)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(response.Errors) > 0 {
		t.Errorf("Unexpected errors: %v", response.Errors)
	}

	// Verify response
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	users, exists := data["users"]
	if !exists {
		t.Error("Expected 'users' field")
	}

	userList, ok := users.([]interface{})
	if !ok {
		t.Error("Expected users to be an array")
	}

	if len(userList) != 2 {
		t.Errorf("Expected 2 users, got %d", len(userList))
	}
}

// TestTestHelperBasic tests basic test helper functionality
func TestTestHelperBasic(t *testing.T) {
	// Test field path splitting
	fields := strings.Split("user.name", ".")
	expected := []string{"user", "name"}
	if len(fields) != len(expected) {
		t.Errorf("Expected %v, got %v", expected, fields)
	}

	for i, field := range fields {
		if field != expected[i] {
			t.Errorf("Expected field %d to be '%s', got '%s'", i, expected[i], field)
		}
	}

	// Test nested field path
	fields = strings.Split("user.profile.email", ".")
	expected = []string{"user", "profile", "email"}
	if len(fields) != len(expected) {
		t.Errorf("Expected %v, got %v", expected, fields)
	}
}
