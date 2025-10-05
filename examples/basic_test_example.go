package main

import (
	"testing"

	"github.com/kluzzebass/gqlt"
)

// ExampleTestBasicQuery demonstrates basic GraphQL query testing
func ExampleTestBasicQuery(t *testing.T) {
	// Create a GraphQL client
	client := gqlt.NewClient("https://api.example.com/graphql", nil)

	// Execute a simple query
	response, err := client.Execute("{ users { id name } }", nil, "")
	if err != nil {
		t.Fatalf("Query execution failed: %v", err)
	}

	// Check for GraphQL errors
	if len(response.Errors) > 0 {
		t.Errorf("GraphQL errors: %v", response.Errors)
	}

	// Verify response structure
	if response.Data == nil {
		t.Error("Expected data in response")
	}

	// Access response data
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	users, exists := data["users"]
	if !exists {
		t.Error("Expected 'users' field in response")
	}

	// Type assertion for further validation
	userList, ok := users.([]interface{})
	if !ok {
		t.Error("Expected users to be an array")
	}

	t.Logf("Found %d users", len(userList))
}

// ExampleTestQueryWithVariables demonstrates testing with variables
func ExampleTestQueryWithVariables(t *testing.T) {
	client := gqlt.NewClient("https://api.example.com/graphql", nil)

	// Define query with variables
	query := `
		query GetUser($id: ID!) {
			user(id: $id) {
				id
				name
				email
			}
		}
	`

	// Set variables
	variables := map[string]interface{}{
		"id": "123",
	}

	// Execute query
	response, err := client.Execute(query, variables, "GetUser")
	if err != nil {
		t.Fatalf("Query execution failed: %v", err)
	}

	// Validate response
	if len(response.Errors) > 0 {
		t.Errorf("GraphQL errors: %v", response.Errors)
	}

	// Check specific fields
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	user, exists := data["user"]
	if !exists {
		t.Error("Expected 'user' field in response")
	}

	userMap, ok := user.(map[string]interface{})
	if !ok {
		t.Fatal("Expected user to be a map")
	}

	// Validate user fields
	requiredFields := []string{"id", "name", "email"}
	for _, field := range requiredFields {
		if _, exists := userMap[field]; !exists {
			t.Errorf("Expected field '%s' in user object", field)
		}
	}
}

// ExampleTestMutation demonstrates testing mutations
func ExampleTestMutation(t *testing.T) {
	client := gqlt.NewClient("https://api.example.com/graphql", nil)

	// Define mutation
	mutation := `
		mutation CreateUser($input: UserInput!) {
			createUser(input: $input) {
				id
				name
				email
			}
		}
	`

	// Set input variables
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
		},
	}

	// Execute mutation
	response, err := client.Execute(mutation, variables, "CreateUser")
	if err != nil {
		t.Fatalf("Mutation execution failed: %v", err)
	}

	// Check for errors
	if len(response.Errors) > 0 {
		t.Errorf("GraphQL errors: %v", response.Errors)
	}

	// Validate response
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	createdUser, exists := data["createUser"]
	if !exists {
		t.Error("Expected 'createUser' field in response")
	}

	userMap, ok := createdUser.(map[string]interface{})
	if !ok {
		t.Fatal("Expected createUser to be a map")
	}

	// Validate the created user has an ID
	if id, exists := userMap["id"]; !exists || id == nil {
		t.Error("Expected created user to have an ID")
	}
}
