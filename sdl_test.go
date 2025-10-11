package gqlt

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClient_FetchSDL(t *testing.T) {
	testSDL := `schema {
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
}`

	tests := []struct {
		name         string
		path         string
		expectError  bool
		expectSDL    bool
	}{
		{
			name:        "SDL at /schema.graphql",
			path:        "/schema.graphql",
			expectError: false,
			expectSDL:   true,
		},
		{
			name:        "SDL at /graphql/schema.graphql",
			path:        "/graphql/schema.graphql",
			expectError: false,
			expectSDL:   true,
		},
		{
			name:        "SDL at /sdl",
			path:        "/sdl",
			expectError: false,
			expectSDL:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server that only responds to the specific path
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == tt.path {
					w.Header().Set("Content-Type", "text/plain")
					w.Write([]byte(testSDL))
				} else {
					http.NotFound(w, r)
				}
			}))
			defer server.Close()

			client := NewClient(server.URL, nil)
			sdl, err := client.FetchSDL()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if tt.expectSDL {
				if !strings.Contains(sdl, "type Query") {
					t.Error("Expected SDL to contain 'type Query'")
				}
				if !strings.Contains(sdl, "schema {") {
					t.Error("Expected SDL to contain 'schema {'")
				}
			}
		})
	}
}

func TestClient_FetchSDL_NoEndpoint(t *testing.T) {
	// Server that doesn't serve SDL at any path
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	_, err := client.FetchSDL()

	if err == nil {
		t.Error("Expected error when SDL is not available")
	}

	if !strings.Contains(err.Error(), "could not fetch SDL") {
		t.Errorf("Expected 'could not fetch SDL' error, got: %v", err)
	}
}

func TestSDLToIntrospection(t *testing.T) {
	sdl := `schema {
  query: Query
  mutation: Mutation
}

type Query {
  hello: String
  user(id: ID!): User
}

type Mutation {
  createUser(name: String!, email: String!): User
}

type User {
  id: ID!
  name: String!
  email: String!
}

enum Role {
  ADMIN
  USER
}

input UserInput {
  name: String!
  email: String!
}

union SearchResult = User

scalar DateTime
`

	result, err := SDLToIntrospection(sdl)
	if err != nil {
		t.Fatalf("Failed to convert SDL to introspection: %v", err)
	}

	// Check that result has the expected structure
	data, ok := result.(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	schema, ok := data["__schema"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected __schema field")
	}

	// Check query type
	queryType, ok := schema["queryType"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected queryType field")
	}
	if queryType["name"] != "Query" {
		t.Errorf("Expected queryType name to be 'Query', got %v", queryType["name"])
	}

	// Check mutation type
	mutationType, ok := schema["mutationType"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected mutationType field")
	}
	if mutationType["name"] != "Mutation" {
		t.Errorf("Expected mutationType name to be 'Mutation', got %v", mutationType["name"])
	}

	// Check types array exists
	types, ok := schema["types"].([]interface{})
	if !ok {
		t.Fatal("Expected types array")
	}
	if len(types) == 0 {
		t.Error("Expected non-empty types array")
	}

	// Check directives array exists
	directives, ok := schema["directives"].([]interface{})
	if !ok {
		t.Fatal("Expected directives array")
	}
	if directives == nil {
		t.Error("Expected directives to not be nil")
	}
}

func TestSDLToIntrospection_InvalidSDL(t *testing.T) {
	invalidSDL := `this is not valid SDL syntax {{{`

	_, err := SDLToIntrospection(invalidSDL)
	if err == nil {
		t.Error("Expected error for invalid SDL")
	}

	if !strings.Contains(err.Error(), "failed to parse SDL") {
		t.Errorf("Expected parse error, got: %v", err)
	}
}

func TestSDLToIntrospection_TypeKinds(t *testing.T) {
	sdl := `schema {
  query: Query
}

type Query {
  hello: String
}

enum Status {
  ACTIVE
  INACTIVE
}

interface Node {
  id: ID!
}

input CreateInput {
  name: String!
}

union Result = Query

scalar CustomScalar
`

	result, err := SDLToIntrospection(sdl)
	if err != nil {
		t.Fatalf("Failed to convert SDL: %v", err)
	}

	data := result.(map[string]interface{})
	schema := data["__schema"].(map[string]interface{})
	types := schema["types"].([]interface{})

	// Find and verify each type kind
	foundKinds := make(map[string]bool)
	for _, typeItem := range types {
		typeDef := typeItem.(map[string]interface{})
		kind := typeDef["kind"].(string)
		name := typeDef["name"].(string)

		switch name {
		case "Query":
			if kind != "OBJECT" {
				t.Errorf("Expected Query to be OBJECT, got %s", kind)
			}
			foundKinds["OBJECT"] = true
		case "Status":
			if kind != "ENUM" {
				t.Errorf("Expected Status to be ENUM, got %s", kind)
			}
			foundKinds["ENUM"] = true
		case "Node":
			if kind != "INTERFACE" {
				t.Errorf("Expected Node to be INTERFACE, got %s", kind)
			}
			foundKinds["INTERFACE"] = true
		case "CreateInput":
			if kind != "INPUT_OBJECT" {
				t.Errorf("Expected CreateInput to be INPUT_OBJECT, got %s", kind)
			}
			foundKinds["INPUT_OBJECT"] = true
		case "Result":
			if kind != "UNION" {
				t.Errorf("Expected Result to be UNION, got %s", kind)
			}
			foundKinds["UNION"] = true
		case "CustomScalar":
			if kind != "SCALAR" {
				t.Errorf("Expected CustomScalar to be SCALAR, got %s", kind)
			}
			foundKinds["SCALAR"] = true
		}
	}

	expectedKinds := []string{"OBJECT", "ENUM", "INTERFACE", "INPUT_OBJECT", "UNION", "SCALAR"}
	for _, kind := range expectedKinds {
		if !foundKinds[kind] {
			t.Errorf("Did not find type with kind %s", kind)
		}
	}
}

func TestClient_Introspect_SDLFallback(t *testing.T) {
	testSDL := `schema {
  query: Query
}

type Query {
  hello: String
}`

	// Create server that returns error for introspection but serves SDL
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// SDL endpoint (GET)
		if r.Method == "GET" && r.URL.Path == "/schema.graphql" {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(testSDL))
			return
		}

		// POST GraphQL endpoint - return introspection error
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"errors":[{"message":"Introspection is not allowed"}]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	result, err := client.Introspect()

	if err != nil {
		t.Fatalf("Expected SDL fallback to work, got error: %v", err)
	}

	if result.Data == nil {
		t.Fatal("Expected data from SDL fallback")
	}

	data, ok := result.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	schema, ok := data["__schema"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected __schema field from SDL fallback")
	}

	queryType, ok := schema["queryType"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected queryType from SDL fallback")
	}

	if queryType["name"] != "Query" {
		t.Errorf("Expected queryType name 'Query', got %v", queryType["name"])
	}
}

func TestClient_Introspect_PreferIntrospection(t *testing.T) {
	// Server that returns both introspection and SDL, but with different data
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// SDL endpoint (GET)
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(`schema { query: Query } type Query { fromSDL: String }`))
			return
		}

		// POST GraphQL endpoint - return successful introspection
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"data": {
				"__schema": {
					"queryType": {"name": "QueryFromIntrospection"},
					"types": []
				}
			}
		}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	result, err := client.Introspect()

	if err != nil {
		t.Fatalf("Introspection failed: %v", err)
	}

	data := result.Data.(map[string]interface{})
	schema := data["__schema"].(map[string]interface{})
	queryType := schema["queryType"].(map[string]interface{})

	// Should use introspection result, not SDL
	if queryType["name"] != "QueryFromIntrospection" {
		t.Error("Expected introspection to be preferred over SDL")
	}
}

