package graphql

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNewClient(t *testing.T) {
	endpoint := "https://api.example.com/graphql"
	headers := map[string]string{
		"Authorization": "Bearer token123",
		"X-API-Key":     "key456",
	}

	client := NewClient(endpoint, headers)

	if client.endpoint != endpoint {
		t.Errorf("Expected endpoint %s, got %s", endpoint, client.endpoint)
	}

	if len(client.headers) != len(headers) {
		t.Errorf("Expected %d headers, got %d", len(headers), len(client.headers))
	}

	for k, v := range headers {
		if client.headers[k] != v {
			t.Errorf("Expected header %s=%s, got %s=%s", k, v, k, client.headers[k])
		}
	}
}

func TestSetHeaders(t *testing.T) {
	client := NewClient("https://api.example.com/graphql", nil)

	headers := map[string]string{
		"Authorization": "Bearer token123",
		"X-API-Key":     "key456",
	}

	client.SetHeaders(headers)

	if len(client.headers) != len(headers) {
		t.Errorf("Expected %d headers, got %d", len(headers), len(client.headers))
	}

	for k, v := range headers {
		if client.headers[k] != v {
			t.Errorf("Expected header %s=%s, got %s=%s", k, v, k, client.headers[k])
		}
	}
}

func TestSetAuth(t *testing.T) {
	client := NewClient("https://api.example.com/graphql", nil)

	username := "testuser"
	password := "testpass"

	client.SetAuth(username, password)

	// Verify that the transport was set
	if client.httpClient.Transport == nil {
		t.Error("Expected transport to be set after SetAuth")
	}

	// Test that the transport is the right type
	transport := client.httpClient.Transport
	if transport == nil {
		t.Error("Expected transport to be set")
		return
	}

	// Test the transport type
	_, ok := transport.(*basicAuthTransport)
	if !ok {
		t.Error("Expected transport to be basicAuthTransport")
	}
}

func TestExecute(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Verify content type
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected application/json content type, got %s", contentType)
		}

		// Parse request body
		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Verify request structure
		if requestBody["query"] == nil {
			t.Error("Expected 'query' field in request body")
		}

		// Send mock response
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"user": map[string]interface{}{
					"id":   "123",
					"name": "John Doe",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client := NewClient(server.URL, nil)

	// Execute query
	query := `query { user { id name }`
	variables := map[string]interface{}{
		"id": "123",
	}
	operationName := "GetUser"

	result, err := client.Execute(query, variables, operationName)
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	// Verify response
	if result.Data == nil {
		t.Error("Expected data in response")
	}

	// Check that data contains user
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Error("Expected data to be map[string]interface{}")
	}

	user, ok := data["user"].(map[string]interface{})
	if !ok {
		t.Error("Expected user to be map[string]interface{}")
	}

	if user["id"] != "123" {
		t.Errorf("Expected user id '123', got %v", user["id"])
	}

	if user["name"] != "John Doe" {
		t.Errorf("Expected user name 'John Doe', got %v", user["name"])
	}

	// Test additional coverage - Execute with empty variables
	result2, err := client.Execute(query, nil, "")
	if err != nil {
		t.Errorf("Execute with nil variables failed: %v", err)
	}
	if result2 == nil {
		t.Error("Expected result2 to be non-nil")
	}
}

func TestExecuteWithErrors(t *testing.T) {
	// Create a mock server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": nil,
			"errors": []map[string]interface{}{
				{
					"message": "User not found",
					"path":    []string{"user"},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client := NewClient(server.URL, nil)

	// Execute query
	result, err := client.Execute("query { user { id } }", nil, "")
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	// Verify errors are present
	if len(result.Errors) == 0 {
		t.Error("Expected errors in response")
	}

	// Check error structure
	errorObj, ok := result.Errors[0].(map[string]interface{})
	if !ok {
		t.Error("Expected error to be map[string]interface{}")
	}

	if errorObj["message"] != "User not found" {
		t.Errorf("Expected error message 'User not found', got %v", errorObj["message"])
	}
}

func TestExecuteWithHeaders(t *testing.T) {
	// Create a mock server that checks headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that custom headers are present
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer token123" {
			t.Errorf("Expected Authorization header 'Bearer token123', got %s", authHeader)
		}

		apiKeyHeader := r.Header.Get("X-API-Key")
		if apiKeyHeader != "key456" {
			t.Errorf("Expected X-API-Key header 'key456', got %s", apiKeyHeader)
		}

		// Send success response
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"success": true,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with headers
	headers := map[string]string{
		"Authorization": "Bearer token123",
		"X-API-Key":     "key456",
	}
	client := NewClient(server.URL, headers)

	// Execute query
	result, err := client.Execute("query { success }", nil, "")
	if err != nil {
		t.Errorf("Execute failed: %v", err)
	}

	// Verify response
	if result.Data == nil {
		t.Error("Expected data in response")
	}
}

func TestExecuteErrorPaths(t *testing.T) {
	// Test with invalid endpoint
	client := NewClient("invalid-url", nil)
	_, err := client.Execute("query { user { id } }", nil, "")
	if err == nil {
		t.Error("Expected error for invalid endpoint")
	}

	// Test with server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client2 := NewClient(server.URL, nil)
	_, err = client2.Execute("query { user { id } }", nil, "")
	if err == nil {
		t.Error("Expected error for invalid JSON response")
	}
}

func TestExecuteWithFiles(t *testing.T) {
	// Create a mock server that handles multipart requests
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify content type is multipart
		contentType := r.Header.Get("Content-Type")
		if !contains(contentType, "multipart/form-data") {
			t.Errorf("Expected multipart content type, got %s", contentType)
		}

		// Send mock response
		response := map[string]interface{}{
			"data": map[string]interface{}{
				"upload": map[string]interface{}{
					"success": true,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client
	client := NewClient(server.URL, nil)

	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "test-file.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.WriteString("test content")
	tempFile.Close()

	// Execute with files
	files := map[string]string{
		"file": tempFile.Name(),
	}
	result, err := client.ExecuteWithFiles("mutation { upload(file: $file) { success } }", nil, "", files)
	if err != nil {
		t.Errorf("ExecuteWithFiles failed: %v", err)
	}

	// Verify response
	if result.Data == nil {
		t.Error("Expected data in response")
	}
}

func TestBasicAuthTransport(t *testing.T) {
	// Create a mock server that verifies basic auth
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		expected := "Basic dXNlcjpwYXNz" // base64("user:pass")
		if authHeader != expected {
			t.Errorf("Expected auth header %s, got %s", expected, authHeader)
		}

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"authenticated": true,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with basic auth
	client := NewClient(server.URL, nil)
	client.SetAuth("user", "pass")

	// Execute query
	result, err := client.Execute("query { authenticated }", nil, "")
	if err != nil {
		t.Errorf("Execute with basic auth failed: %v", err)
	}

	// Verify response
	if result.Data == nil {
		t.Error("Expected data in response")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && contains(s[1:], substr)
}
