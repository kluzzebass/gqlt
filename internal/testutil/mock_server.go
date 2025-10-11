package testutil

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/kluzzebass/gqlt"
)

// MockGraphQLServer provides a flexible test GraphQL server with support for
// queries, mutations, introspection, and file uploads
type MockGraphQLServer struct {
	server          *httptest.Server
	handlers        map[string]ResponseHandler
	defaultHandler  ResponseHandler
	delay           time.Duration
	requestLog      []Request
	introspectionOn bool
	schema          map[string]interface{}
}

// ResponseHandler is a function that generates a GraphQL response
type ResponseHandler func(req Request) *gqlt.Response

// Request represents a parsed GraphQL request
type Request struct {
	Query         string
	Variables     map[string]interface{}
	OperationName string
	Headers       http.Header
	Files         map[string][]byte // For file uploads
}

// NewMockGraphQLServer creates a new mock GraphQL server with default settings
func NewMockGraphQLServer() *MockGraphQLServer {
	mock := &MockGraphQLServer{
		handlers:        make(map[string]ResponseHandler),
		requestLog:      make([]Request, 0),
		introspectionOn: true,
		schema:          createDefaultSchema(),
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

// AddHandler adds a handler for a specific operation name
func (m *MockGraphQLServer) AddHandler(operationName string, handler ResponseHandler) {
	m.handlers[operationName] = handler
}

// SetDefaultHandler sets a handler for unmatched operations
func (m *MockGraphQLServer) SetDefaultHandler(handler ResponseHandler) {
	m.defaultHandler = handler
}

// SetDelay sets a response delay (useful for testing timeouts)
func (m *MockGraphQLServer) SetDelay(delay time.Duration) {
	m.delay = delay
}

// EnableIntrospection enables GraphQL introspection queries
func (m *MockGraphQLServer) EnableIntrospection(enable bool) {
	m.introspectionOn = enable
}

// SetSchema sets the schema returned by introspection queries
func (m *MockGraphQLServer) SetSchema(schema map[string]interface{}) {
	m.schema = schema
}

// GetRequestLog returns all requests received by the server
func (m *MockGraphQLServer) GetRequestLog() []Request {
	return m.requestLog
}

// ClearRequestLog clears the request log
func (m *MockGraphQLServer) ClearRequestLog() {
	m.requestLog = make([]Request, 0)
}

// handleRequest handles incoming HTTP requests
func (m *MockGraphQLServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Add delay if configured
	if m.delay > 0 {
		time.Sleep(m.delay)
	}

	// Handle SDL schema endpoints (GET requests)
	if r.Method == "GET" && (r.URL.Path == "/schema.graphql" || r.URL.Path == "/graphql/schema.graphql" || r.URL.Path == "/sdl") {
		m.handleSDLRequest(w, r)
		return
	}

	// Parse request based on content type
	var req Request
	var err error

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		req, err = m.parseMultipartRequest(r)
	} else {
		req, err = m.parseJSONRequest(r)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	req.Headers = r.Header

	// Log request
	m.requestLog = append(m.requestLog, req)

	// Handle introspection query
	if m.introspectionOn && isIntrospectionQuery(req.Query) {
		m.handleIntrospectionQuery(w, req)
		return
	}

	// Find and execute handler
	var response *gqlt.Response
	if handler, exists := m.handlers[req.OperationName]; exists {
		response = handler(req)
	} else if m.defaultHandler != nil {
		response = m.defaultHandler(req)
	} else {
		response = &gqlt.Response{
			Errors: []interface{}{
				map[string]interface{}{
					"message": fmt.Sprintf("No handler configured for operation: %s", req.OperationName),
				},
			},
		}
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// parseJSONRequest parses a JSON GraphQL request
func (m *MockGraphQLServer) parseJSONRequest(r *http.Request) (Request, error) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, fmt.Errorf("invalid JSON: %w", err)
	}
	return req, nil
}

// parseMultipartRequest parses a multipart GraphQL request (with file uploads)
func (m *MockGraphQLServer) parseMultipartRequest(r *http.Request) (Request, error) {
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32 MB max
		return Request{}, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	var req Request

	// Parse operations
	operations := r.FormValue("operations")
	if operations == "" {
		return Request{}, fmt.Errorf("missing operations field")
	}

	if err := json.Unmarshal([]byte(operations), &req); err != nil {
		return Request{}, fmt.Errorf("invalid operations JSON: %w", err)
	}

	// Parse file map
	mapData := r.FormValue("map")
	if mapData != "" {
		req.Files = make(map[string][]byte)

		var fileMap map[string][]string
		if err := json.Unmarshal([]byte(mapData), &fileMap); err != nil {
			return Request{}, fmt.Errorf("invalid map JSON: %w", err)
		}

		// Read uploaded files
		for name := range fileMap {
			file, _, err := r.FormFile(name)
			if err != nil {
				continue
			}
			defer file.Close()

			data, err := io.ReadAll(file)
			if err != nil {
				continue
			}

			req.Files[name] = data
		}
	}

	return req, nil
}

// handleIntrospectionQuery handles GraphQL introspection queries
func (m *MockGraphQLServer) handleIntrospectionQuery(w http.ResponseWriter, req Request) {
	response := &gqlt.Response{
		Data: m.schema,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSDLRequest handles requests for SDL schema
func (m *MockGraphQLServer) handleSDLRequest(w http.ResponseWriter, r *http.Request) {
	sdl := `schema {
  query: Query
}

type Query {
  hello: String
  user(id: ID!): User
}

type User {
  id: ID!
  name: String!
  email: String!
}

scalar String
scalar ID
`
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(sdl))
}

// isIntrospectionQuery checks if a query is an introspection query
func isIntrospectionQuery(query string) bool {
	return strings.Contains(query, "__schema") ||
		strings.Contains(query, "__type") ||
		strings.Contains(query, "IntrospectionQuery")
}

// createDefaultSchema creates a basic GraphQL schema for introspection
func createDefaultSchema() map[string]interface{} {
	return map[string]interface{}{
		"__schema": map[string]interface{}{
			"queryType": map[string]interface{}{
				"name": "Query",
			},
			"mutationType": map[string]interface{}{
				"name": "Mutation",
			},
			"subscriptionType": nil,
			"types": []interface{}{
				map[string]interface{}{
					"kind":        "OBJECT",
					"name":        "Query",
					"description": "The root query type",
					"fields": []interface{}{
						map[string]interface{}{
							"name":        "user",
							"description": "Get a user by ID",
							"args": []interface{}{
								map[string]interface{}{
									"name": "id",
									"type": map[string]interface{}{
										"kind": "NON_NULL",
										"ofType": map[string]interface{}{
											"kind": "SCALAR",
											"name": "ID",
										},
									},
								},
							},
							"type": map[string]interface{}{
								"kind": "OBJECT",
								"name": "User",
							},
						},
						map[string]interface{}{
							"name":        "users",
							"description": "Get all users",
							"args":        []interface{}{},
							"type": map[string]interface{}{
								"kind": "LIST",
								"ofType": map[string]interface{}{
									"kind": "OBJECT",
									"name": "User",
								},
							},
						},
					},
				},
				map[string]interface{}{
					"kind":        "OBJECT",
					"name":        "Mutation",
					"description": "The root mutation type",
					"fields": []interface{}{
						map[string]interface{}{
							"name":        "createUser",
							"description": "Create a new user",
							"args": []interface{}{
								map[string]interface{}{
									"name": "input",
									"type": map[string]interface{}{
										"kind": "NON_NULL",
										"ofType": map[string]interface{}{
											"kind": "INPUT_OBJECT",
											"name": "CreateUserInput",
										},
									},
								},
							},
							"type": map[string]interface{}{
								"kind": "OBJECT",
								"name": "User",
							},
						},
					},
				},
				map[string]interface{}{
					"kind":        "OBJECT",
					"name":        "User",
					"description": "A user in the system",
					"fields": []interface{}{
						map[string]interface{}{
							"name": "id",
							"type": map[string]interface{}{
								"kind": "NON_NULL",
								"ofType": map[string]interface{}{
									"kind": "SCALAR",
									"name": "ID",
								},
							},
						},
						map[string]interface{}{
							"name": "name",
							"type": map[string]interface{}{
								"kind": "NON_NULL",
								"ofType": map[string]interface{}{
									"kind": "SCALAR",
									"name": "String",
								},
							},
						},
						map[string]interface{}{
							"name": "email",
							"type": map[string]interface{}{
								"kind": "SCALAR",
								"name": "String",
							},
						},
					},
				},
				map[string]interface{}{
					"kind":        "INPUT_OBJECT",
					"name":        "CreateUserInput",
					"description": "Input for creating a user",
					"inputFields": []interface{}{
						map[string]interface{}{
							"name": "name",
							"type": map[string]interface{}{
								"kind": "NON_NULL",
								"ofType": map[string]interface{}{
									"kind": "SCALAR",
									"name": "String",
								},
							},
						},
						map[string]interface{}{
							"name": "email",
							"type": map[string]interface{}{
								"kind": "SCALAR",
								"name": "String",
							},
						},
					},
				},
				map[string]interface{}{
					"kind": "SCALAR",
					"name": "ID",
				},
				map[string]interface{}{
					"kind": "SCALAR",
					"name": "String",
				},
				map[string]interface{}{
					"kind": "SCALAR",
					"name": "Int",
				},
				map[string]interface{}{
					"kind": "SCALAR",
					"name": "Boolean",
				},
			},
			"directives": []interface{}{},
		},
	}
}

// Helper functions for creating common responses

// SuccessResponse creates a successful response with data
func SuccessResponse(data interface{}) *gqlt.Response {
	return &gqlt.Response{
		Data: data,
	}
}

// ErrorResponse creates an error response
func ErrorResponse(message string) *gqlt.Response {
	return &gqlt.Response{
		Errors: []interface{}{
			map[string]interface{}{
				"message": message,
			},
		},
	}
}

// DataWithErrors creates a response with both data and errors
func DataWithErrors(data interface{}, errors []string) *gqlt.Response {
	errorList := make([]interface{}, len(errors))
	for i, msg := range errors {
		errorList[i] = map[string]interface{}{
			"message": msg,
		}
	}
	return &gqlt.Response{
		Data:   data,
		Errors: errorList,
	}
}
