# gqlt Library Usage Examples

This directory contains comprehensive examples showing how to use `gqlt` as a library in Go tests and applications.

## Files Overview

### `basic_test_example.go`
Basic examples of using gqlt in Go tests:
- Simple GraphQL queries
- Queries with variables
- Mutations
- Error handling

### `test_utilities.go`
Advanced test utilities for common GraphQL testing patterns:
- `GraphQLTestHelper` - Comprehensive test helper class
- Field assertions (`AssertFieldExists`, `AssertFieldValue`)
- Array validation (`AssertArrayLength`)
- Authentication helpers
- Convenience methods for queries and mutations

### `mock_server.go`
Mock server utilities for testing:
- `MockGraphQLServer` - Simple mock server with operation handlers
- `GraphQLTestServer` - Advanced test server with predefined responses
- Delay simulation for timeout testing
- Custom response configuration

### `integration_test_example.go`
Complete integration test examples:
- Real GraphQL API testing
- Schema introspection testing
- Configuration management
- Query validation
- File upload testing

## Quick Start

### Basic Usage

```go
package main

import (
    "testing"
    "github.com/kluzzebass/gqlt"
)

func TestBasicQuery(t *testing.T) {
    // Create client
    client := gqlt.NewClient("https://api.example.com/graphql", nil)
    
    // Execute query
    response, err := client.Execute("{ users { id name } }", nil, "", nil)
    if err != nil {
        t.Fatalf("Query failed: %v", err)
    }
    
    // Check for errors
    if len(response.Errors) > 0 {
        t.Errorf("GraphQL errors: %v", response.Errors)
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
}
```

### Using Test Utilities

```go
func TestWithHelper(t *testing.T) {
    helper := NewGraphQLTestHelper(t, "https://api.example.com/graphql")
    
    // Set authentication
    helper.SetAuth("bearer", map[string]string{
        "token": "your-token-here",
    })
    
    // Execute query
    response := helper.ExecuteQuery("{ users { id name } }", nil, "")
    
    // Use assertions
    helper.AssertNoErrors(response)
    helper.AssertFieldExists(response, "users")
    helper.AssertArrayLength(response, "users", 2)
}
```

### Using Mock Server

```go
func TestWithMockServer(t *testing.T) {
    // Create mock server
    mock := NewMockGraphQLServer()
    defer mock.Close()
    
    // Add handler
    mock.AddHandler("GetUsers", func(response *gqlt.Response) {
        response.Data = map[string]interface{}{
            "users": []map[string]interface{}{
                {"id": "1", "name": "John Doe"},
            },
        }
    })
    
    // Test with client
    client := gqlt.NewClient(mock.URL(), nil)
    response, err := client.Execute(
        `query GetUsers { users { id name } }`,
        nil,
        "GetUsers",
        nil,
    )
    
    // Verify response
    if err != nil {
        t.Fatalf("Query failed: %v", err)
    }
}
```

## Common Testing Patterns

### 1. Authentication Testing

```go
func TestWithAuth(t *testing.T) {
    helper := NewGraphQLTestHelper(t, "https://api.example.com/graphql")
    
    // Test different auth types
    testCases := []struct {
        name string
        authType string
        credentials map[string]string
    }{
        {"Bearer Token", "bearer", map[string]string{"token": "test-token"}},
        {"Basic Auth", "basic", map[string]string{"username": "user", "password": "pass"}},
        {"API Key", "apikey", map[string]string{"apikey": "test-key"}},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            helper.SetAuth(tc.authType, tc.credentials)
            response := helper.ExecuteQuery("{ viewer { id } }", nil, "")
            // Test response...
        })
    }
}
```

### 2. Error Testing

```go
func TestErrorHandling(t *testing.T) {
    helper := NewGraphQLTestHelper(t, "https://api.example.com/graphql")
    
    // Test invalid query
    response := helper.ExecuteQuery("{ invalidField }", nil, "")
    helper.AssertHasErrors(response)
    
    // Test authentication error
    response = helper.ExecuteQuery("{ viewer { id } }", nil, "")
    if len(response.Errors) > 0 {
        // Verify error message
        errorMsg := response.Errors[0].Message
        if !strings.Contains(errorMsg, "authentication") {
            t.Errorf("Expected authentication error, got: %s", errorMsg)
        }
    }
}
```

### 3. Schema Testing

```go
func TestSchemaIntrospection(t *testing.T) {
    client := gqlt.NewClient("https://api.example.com/graphql", nil)
    introspectClient := gqlt.NewIntrospect(client)
    
    // Introspect schema
    schema, err := introspectClient.IntrospectSchema()
    if err != nil {
        t.Fatalf("Schema introspection failed: %v", err)
    }
    
    // Analyze schema
    analyzer, err := gqlt.NewAnalyzer(schema)
    if err != nil {
        t.Fatalf("Failed to create analyzer: %v", err)
    }
    
    summary, err := analyzer.GetSummary()
    if err != nil {
        t.Fatalf("Failed to get summary: %v", err)
    }
    
    // Verify schema structure
    if summary.TotalTypes == 0 {
        t.Error("Expected schema to have types")
    }
}
```

## Best Practices

### 1. Test Organization
- Use table-driven tests for multiple scenarios
- Group related tests with subtests
- Use descriptive test names

### 2. Error Handling
- Always check for GraphQL errors
- Test both success and failure scenarios
- Verify error messages and codes

### 3. Response Validation
- Use type assertions carefully
- Validate required fields
- Check data types and structures

### 4. Mock Server Usage
- Use mock servers for unit tests
- Use real endpoints for integration tests
- Test both success and error responses

### 5. Configuration Management
- Use temporary directories for test configs
- Clean up after tests
- Test configuration loading and saving

## Running the Examples

To run the examples:

```bash
# Run all examples
go test ./examples/...

# Run specific example
go test ./examples/ -run TestBasicQuery

# Run with verbose output
go test -v ./examples/...
```

## Integration with CI/CD

These examples can be integrated into CI/CD pipelines:

```yaml
# GitHub Actions example
- name: Run GraphQL Tests
  run: |
    go test ./examples/... -v
    go test ./examples/... -race
    go test ./examples/... -cover
```

## Contributing

When adding new examples:
1. Follow the existing patterns
2. Include comprehensive error handling
3. Add documentation comments
4. Test your examples thoroughly
5. Update this README if needed
