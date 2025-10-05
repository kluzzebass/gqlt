package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kluzzebass/gqlt"
)

// MockGraphQLServer provides a test GraphQL server
type MockGraphQLServer struct {
	server   *httptest.Server
	handlers map[string]func(*gqlt.Response) // operation name -> handler
}

// NewMockGraphQLServer creates a new mock GraphQL server
func NewMockGraphQLServer() *MockGraphQLServer {
	mock := &MockGraphQLServer{
		handlers: make(map[string]func(*gqlt.Response)),
	}

	mock.server = httptest.NewServer(http.HandlerFunc(mock.handleRequest))
	return mock
}

// Close shuts down the mock server
func (m *MockGraphQLServer) Close() {
	m.server.Close()
}

// URL returns the server URL
func (m *MockGraphQLServer) URL() string {
	return m.server.URL
}

// AddHandler adds a handler for a specific operation
func (m *MockGraphQLServer) AddHandler(operationName string, handler func(*gqlt.Response)) {
	m.handlers[operationName] = handler
}

// AddDefaultHandler adds a default handler for unmatched operations
func (m *MockGraphQLServer) AddDefaultHandler(handler func(*gqlt.Response)) {
	m.handlers[""] = handler
}

// handleRequest handles incoming GraphQL requests
func (m *MockGraphQLServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Parse GraphQL request
	var request struct {
		Query         string                 `json:"query"`
		Variables     map[string]interface{} `json:"variables"`
		OperationName string                 `json:"operationName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Find handler
	handler, exists := m.handlers[request.OperationName]
	if !exists {
		handler, exists = m.handlers[""] // Default handler
		if !exists {
			http.Error(w, "No handler for operation", http.StatusNotFound)
			return
		}
	}

	// Create response
	response := &gqlt.Response{}
	handler(response)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExampleMockServer demonstrates using the mock server
func ExampleMockServer(t *testing.T) {
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

	// Add handler for CreateUser operation
	mock.AddHandler("CreateUser", func(response *gqlt.Response) {
		response.Data = map[string]interface{}{
			"createUser": map[string]interface{}{
				"id":    "3",
				"name":  "New User",
				"email": "new@example.com",
			},
		}
	})

	// Add default handler for unmatched operations
	mock.AddDefaultHandler(func(response *gqlt.Response) {
		response.Errors = []interface{}{
			map[string]interface{}{"message": "Unknown operation"},
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

// GraphQLTestServer provides a more advanced test server with built-in responses
type GraphQLTestServer struct {
	server    *httptest.Server
	responses map[string]*gqlt.Response
	delay     time.Duration
}

// NewGraphQLTestServer creates a test server with predefined responses
func NewGraphQLTestServer() *GraphQLTestServer {
	server := &GraphQLTestServer{
		responses: make(map[string]*gqlt.Response),
	}

	server.server = httptest.NewServer(http.HandlerFunc(server.handleRequest))
	return server
}

// Close shuts down the test server
func (s *GraphQLTestServer) Close() {
	s.server.Close()
}

// URL returns the server URL
func (s *GraphQLTestServer) URL() string {
	return s.server.URL
}

// SetResponse sets a response for a specific operation
func (s *GraphQLTestServer) SetResponse(operationName string, response *gqlt.Response) {
	s.responses[operationName] = response
}

// SetDelay sets a delay for all responses (useful for testing timeouts)
func (s *GraphQLTestServer) SetDelay(delay time.Duration) {
	s.delay = delay
}

// handleRequest handles incoming requests
func (s *GraphQLTestServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Add delay if set
	if s.delay > 0 {
		time.Sleep(s.delay)
	}

	// Parse request
	var request struct {
		Query         string                 `json:"query"`
		Variables     map[string]interface{} `json:"variables"`
		OperationName string                 `json:"operationName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Find response
	response, exists := s.responses[request.OperationName]
	if !exists {
		response = &gqlt.Response{
			Errors: []interface{}{
				map[string]interface{}{"message": fmt.Sprintf("No response configured for operation: %s", request.OperationName)},
			},
		}
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ExampleTestServer demonstrates using the test server
func ExampleTestServer(t *testing.T) {
	// Create test server
	server := NewGraphQLTestServer()
	defer server.Close()

	// Configure responses
	server.SetResponse("GetUsers", &gqlt.Response{
		Data: map[string]interface{}{
			"users": []map[string]interface{}{
				{"id": "1", "name": "Test User"},
			},
		},
	})

	server.SetResponse("CreateUser", &gqlt.Response{
		Data: map[string]interface{}{
			"createUser": map[string]interface{}{
				"id":   "2",
				"name": "Created User",
			},
		},
	})

	// Test with client
	client := gqlt.NewClient(server.URL(), nil)

	// Test query
	response, err := client.Execute(
		`query GetUsers { users { id name } }`,
		nil,
		"GetUsers",
	)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Verify response
	if len(response.Errors) > 0 {
		t.Errorf("Unexpected errors: %v", response.Errors)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	users, exists := data["users"]
	if !exists {
		t.Error("Expected 'users' field")
	}

	userList, ok := users.([]interface{})
	if !ok || len(userList) == 0 {
		t.Error("Expected non-empty users array")
	}
}
