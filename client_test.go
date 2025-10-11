package gqlt

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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
		if !strings.Contains(contentType, "multipart/form-data") {
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

// Table-driven tests

func TestClient_Execute_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		variables      map[string]interface{}
		operationName  string
		responseData   interface{}
		responseErrors []interface{}
		wantErr        bool
		validateResp   func(*testing.T, *Response)
	}{
		{
			name:          "successful query",
			query:         `query GetUsers { users { id name } }`,
			variables:     nil,
			operationName: "GetUsers",
			responseData: map[string]interface{}{
				"users": []map[string]interface{}{
					{"id": "1", "name": "User 1"},
					{"id": "2", "name": "User 2"},
				},
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *Response) {
				if resp.Data == nil {
					t.Error("Expected data in response")
				}
				if len(resp.Errors) > 0 {
					t.Errorf("Unexpected errors: %v", resp.Errors)
				}
			},
		},
		{
			name:          "query with variables",
			query:         `query GetUser($id: ID!) { user(id: $id) { id name } }`,
			variables:     map[string]interface{}{"id": "123"},
			operationName: "GetUser",
			responseData: map[string]interface{}{
				"user": map[string]interface{}{
					"id":   "123",
					"name": "Test User",
				},
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *Response) {
				data, ok := resp.Data.(map[string]interface{})
				if !ok {
					t.Fatal("Expected data to be a map")
				}
				user, ok := data["user"].(map[string]interface{})
				if !ok {
					t.Fatal("Expected user in data")
				}
				if user["id"] != "123" {
					t.Errorf("Expected user id 123, got %v", user["id"])
				}
			},
		},
		{
			name:          "mutation",
			query:         `mutation CreateUser($input: CreateUserInput!) { createUser(input: $input) { id name } }`,
			variables:     map[string]interface{}{"input": map[string]interface{}{"name": "New User"}},
			operationName: "CreateUser",
			responseData: map[string]interface{}{
				"createUser": map[string]interface{}{
					"id":   "new-id",
					"name": "New User",
				},
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *Response) {
				if resp.Data == nil {
					t.Error("Expected data in response")
				}
			},
		},
		{
			name:          "query with errors",
			query:         `query { invalid }`,
			variables:     nil,
			operationName: "Invalid",
			responseData:  nil,
			responseErrors: []interface{}{
				map[string]interface{}{"message": "Invalid query"},
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *Response) {
				if len(resp.Errors) == 0 {
					t.Error("Expected errors in response")
				}
			},
		},
		{
			name:          "empty query",
			query:         "",
			variables:     nil,
			operationName: "",
			responseData:  nil,
			responseErrors: []interface{}{
				map[string]interface{}{"message": "Empty query"},
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *Response) {
				if len(resp.Errors) == 0 {
					t.Error("Expected error for empty query")
				}
			},
		},
		{
			name:          "query with special characters",
			query:         `query { user(name: "Test \"User\"") { id } }`,
			variables:     nil,
			operationName: "SpecialChars",
			responseData: map[string]interface{}{
				"user": map[string]interface{}{"id": "1"},
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *Response) {
				if resp.Data == nil {
					t.Error("Expected data in response")
				}
			},
		},
		{
			name:          "very long query",
			query:         strings.Repeat("{ user { id } } ", 100),
			variables:     nil,
			operationName: "LongQuery",
			responseData: map[string]interface{}{
				"result": "ok",
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *Response) {
				if resp.Data == nil {
					t.Error("Expected data in response")
				}
			},
		},
		{
			name:  "complex nested variables",
			query: `query ComplexQuery($input: ComplexInput!) { process(input: $input) { result } }`,
			variables: map[string]interface{}{
				"input": map[string]interface{}{
					"nested": map[string]interface{}{
						"array": []interface{}{1, 2, 3},
						"obj":   map[string]interface{}{"key": "value"},
					},
				},
			},
			operationName: "ComplexQuery",
			responseData: map[string]interface{}{
				"process": map[string]interface{}{"result": "processed"},
			},
			wantErr: false,
			validateResp: func(t *testing.T, resp *Response) {
				if resp.Data == nil {
					t.Error("Expected data in response")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := map[string]interface{}{
					"data": tt.responseData,
				}
				if tt.responseErrors != nil {
					response["errors"] = tt.responseErrors
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			// Create client
			client := NewClient(server.URL, nil)

			// Execute query
			resp, err := client.Execute(tt.query, tt.variables, tt.operationName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.validateResp != nil {
				tt.validateResp(t, resp)
			}
		})
	}
}

func TestClient_Introspect(t *testing.T) {
	// Test successful introspection
	t.Run("successful introspection", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"__schema": map[string]interface{}{
						"queryType": map[string]interface{}{
							"name": "Query",
						},
						"types": []interface{}{
							map[string]interface{}{
								"kind": "OBJECT",
								"name": "Query",
							},
							map[string]interface{}{
								"kind": "SCALAR",
								"name": "String",
							},
						},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL, nil)
		resp, err := client.Introspect()
		if err != nil {
			t.Fatalf("Introspect failed: %v", err)
		}

		if resp.Data == nil {
			t.Fatal("Expected data in introspection response")
		}

		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Expected data to be a map")
		}

		schema, ok := data["__schema"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected __schema in response")
		}

		if schema["queryType"] == nil {
			t.Error("Expected queryType in schema")
		}

		types, ok := schema["types"].([]interface{})
		if !ok || len(types) == 0 {
			t.Error("Expected non-empty types array")
		}
	})

	// Test introspection with custom schema
	t.Run("custom schema", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			customSchema := map[string]interface{}{
				"__schema": map[string]interface{}{
					"queryType": map[string]interface{}{
						"name": "CustomQuery",
					},
					"types": []interface{}{
						map[string]interface{}{
							"kind": "OBJECT",
							"name": "CustomType",
							"fields": []interface{}{
								map[string]interface{}{
									"name": "customField",
									"type": map[string]interface{}{
										"kind": "SCALAR",
										"name": "String",
									},
								},
							},
						},
					},
				},
			}
			response := map[string]interface{}{
				"data": customSchema,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL, nil)
		resp, err := client.Introspect()
		if err != nil {
			t.Fatalf("Introspect failed: %v", err)
		}

		data, ok := resp.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Expected data to be a map")
		}

		schema, ok := data["__schema"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected __schema in response")
		}

		queryType, ok := schema["queryType"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected queryType in schema")
		}

		if queryType["name"] != "CustomQuery" {
			t.Errorf("Expected CustomQuery, got %v", queryType["name"])
		}
	})
}

func TestClient_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name        string
		setupClient func() *Client
		query       string
		expectError bool
		errorCheck  func(*testing.T, error)
	}{
		{
			name: "invalid endpoint URL",
			setupClient: func() *Client {
				return NewClient("://invalid-url", nil)
			},
			query:       `query { test }`,
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Expected error for invalid URL")
				}
			},
		},
		{
			name: "unreachable endpoint",
			setupClient: func() *Client {
				return NewClient("http://localhost:1/nonexistent", nil)
			},
			query:       `query { test }`,
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Expected error for unreachable endpoint")
				}
			},
		},
		{
			name: "malformed GraphQL response",
			setupClient: func() *Client {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.Write([]byte(`{invalid json}`))
				}))
				// Note: server will be leaked in test, but it's acceptable for error testing
				return NewClient(server.URL, nil)
			},
			query:       `query { test }`,
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				if err == nil {
					t.Error("Expected error for malformed JSON")
				}
				if !strings.Contains(err.Error(), "parse") {
					t.Errorf("Expected parse error, got: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := tt.setupClient()
			_, err := client.Execute(tt.query, nil, "")

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if tt.errorCheck != nil {
				tt.errorCheck(t, err)
			}
		})
	}
}

func TestClient_SetHeadersEdgeCases(t *testing.T) {
	t.Run("set headers on nil headers", func(t *testing.T) {
		client := NewClient("http://example.com", nil)
		if client.headers != nil {
			t.Error("Expected nil headers initially")
		}

		client.SetHeaders(map[string]string{"Key": "Value"})
		if client.headers == nil {
			t.Fatal("Expected headers to be initialized")
		}

		if client.headers["Key"] != "Value" {
			t.Error("Expected header to be set")
		}
	})

	t.Run("overwrite existing headers", func(t *testing.T) {
		client := NewClient("http://example.com", map[string]string{"Key1": "Value1"})
		client.SetHeaders(map[string]string{"Key1": "NewValue1", "Key2": "Value2"})

		if client.headers["Key1"] != "NewValue1" {
			t.Error("Expected header to be overwritten")
		}

		if client.headers["Key2"] != "Value2" {
			t.Error("Expected new header to be added")
		}
	})

	t.Run("set empty headers map", func(t *testing.T) {
		client := NewClient("http://example.com", nil)
		client.SetHeaders(map[string]string{})

		if client.headers == nil {
			t.Error("Expected headers map to be initialized")
		}
	})
}

func TestClient_ExecuteWithFilesEdgeCases(t *testing.T) {
	t.Run("non-existent file", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"upload": map[string]interface{}{"success": true},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL, nil)
		files := map[string]string{
			"file": "/nonexistent/file.txt",
		}

		_, err := client.ExecuteWithFiles(`mutation { upload }`, nil, "Upload", files)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("empty files map", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := map[string]interface{}{
				"data": nil,
				"errors": []interface{}{
					map[string]interface{}{"message": "No files uploaded"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL, nil)
		resp, err := client.ExecuteWithFiles(`mutation { upload }`, nil, "Upload", map[string]string{})
		if err != nil {
			t.Fatalf("ExecuteWithFiles failed: %v", err)
		}

		if len(resp.Errors) == 0 {
			t.Error("Expected error for no files")
		}
	})
}

func TestClient_Coverage(t *testing.T) {
	// Additional tests to increase coverage

	t.Run("Execute with all parameters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Parse request to verify all parameters
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			if req["query"] == "" {
				t.Error("Expected query to be set")
			}
			if req["variables"] == nil {
				t.Error("Expected variables to be set")
			}
			if req["operationName"] == "" {
				t.Error("Expected operation name to be set")
			}

			response := map[string]interface{}{
				"data": map[string]interface{}{"result": "ok"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL, nil)
		resp, err := client.Execute(
			`query FullTest($id: ID!) { test(id: $id) }`,
			map[string]interface{}{"id": "123"},
			"FullTest",
		)

		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		if resp.Data == nil {
			t.Error("Expected data in response")
		}
	})

	t.Run("Response with extensions", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := map[string]interface{}{
				"data": map[string]interface{}{"result": "ok"},
				"extensions": map[string]interface{}{
					"tracing": map[string]interface{}{
						"duration": 123,
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewClient(server.URL, nil)
		resp, err := client.Execute(`query WithExtensions { test }`, nil, "WithExtensions")

		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}

		if resp.Extensions == nil {
			t.Error("Expected extensions in response")
		}

		if resp.Extensions["tracing"] == nil {
			t.Error("Expected tracing in extensions")
		}
	})
}
