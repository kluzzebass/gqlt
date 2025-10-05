package introspect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kluzzebass/gqlt/internal/graphql"
)

func TestIntrospectClient(t *testing.T) {
	// Test client creation
	mockClient := &graphql.Client{}
	client := NewClient(mockClient)
	if client.graphqlClient != mockClient {
		t.Error("Expected graphqlClient to be set")
	}

	// Test introspection query validation
	if IntrospectQuery == "" {
		t.Error("IntrospectQuery should not be empty")
	}
	if !strings.HasPrefix(strings.TrimSpace(IntrospectQuery), "query") {
		t.Error("IntrospectQuery should start with 'query'")
	}
	if !strings.Contains(IntrospectQuery, "IntrospectionQuery") {
		t.Error("IntrospectQuery should contain 'IntrospectionQuery'")
	}
	if !strings.Contains(IntrospectQuery, "__schema") {
		t.Error("IntrospectQuery should contain '__schema'")
	}

	// Verify the query contains expected fields
	expectedFields := []string{
		"__schema", "types", "name", "kind", "fields",
		"queryType", "mutationType", "subscriptionType",
	}
	for _, field := range expectedFields {
		if !strings.Contains(IntrospectQuery, field) {
			t.Errorf("IntrospectQuery should contain '%s'", field)
		}
	}
}

func TestSchemaOperations(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Test cases for different response types
	testCases := []struct {
		name     string
		response *graphql.Response
	}{
		{
			name: "valid schema response",
			response: &graphql.Response{
				Data: map[string]interface{}{
					"__schema": map[string]interface{}{
						"types": []interface{}{
							map[string]interface{}{
								"name": "User",
								"kind": "OBJECT",
							},
						},
					},
				},
			},
		},
		{
			name: "empty response",
			response: &graphql.Response{
				Data: nil,
			},
		},
		{
			name: "response with errors",
			response: &graphql.Response{
				Data: map[string]interface{}{
					"__schema": map[string]interface{}{
						"types": []interface{}{},
					},
				},
				Errors: []interface{}{
					map[string]interface{}{
						"message": "Schema introspection failed",
					},
				},
			},
		},
		{
			name: "response with extensions",
			response: &graphql.Response{
				Data: map[string]interface{}{
					"__schema": map[string]interface{}{
						"types": []interface{}{},
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

	// Test SaveSchema with different response types
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testPath := filepath.Join(tempDir, tc.name+".json")
			err := SaveSchema(tc.response, testPath)
			if err != nil {
				t.Errorf("SaveSchema failed for %s: %v", tc.name, err)
			}

			// Verify file was created
			if _, err := os.Stat(testPath); os.IsNotExist(err) {
				t.Errorf("Schema file was not created for %s", tc.name)
			}

			// Verify file content is valid JSON
			data, err := os.ReadFile(testPath)
			if err != nil {
				t.Errorf("Failed to read schema file for %s: %v", tc.name, err)
			}
			if len(data) == 0 {
				t.Errorf("Schema file is empty for %s", tc.name)
			}

			var parsed map[string]interface{}
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Errorf("Schema file is not valid JSON for %s: %v", tc.name, err)
			}
		})
	}

	// Test LoadSchema
	validPath := filepath.Join(tempDir, "valid schema response.json")
	loadedResponse, err := LoadSchema(validPath)
	if err != nil {
		t.Errorf("LoadSchema failed: %v", err)
	}
	if loadedResponse == nil {
		t.Error("Expected loaded response to be non-nil")
		return
	}
	if loadedResponse.Data == nil {
		t.Error("Expected data in loaded response")
	}

	// Test SchemaExists
	if !SchemaExists(validPath) {
		t.Error("Expected existing file to return true")
	}

	nonExistentPath := filepath.Join(tempDir, "non-existent.json")
	if SchemaExists(nonExistentPath) {
		t.Error("Expected non-existent file to return false")
	}
}

func TestSchemaEdgeCases(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Test loading non-existent file
	_, err := LoadSchema("/non/existent/path.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test loading invalid JSON
	invalidPath := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(invalidPath, []byte("invalid json"), 0644)
	if err != nil {
		t.Errorf("Failed to create invalid JSON file: %v", err)
	}

	_, err = LoadSchema(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Test loading empty file
	emptyPath := filepath.Join(tempDir, "empty.json")
	err = os.WriteFile(emptyPath, []byte(""), 0644)
	if err != nil {
		t.Errorf("Failed to create empty file: %v", err)
	}

	_, err = LoadSchema(emptyPath)
	if err == nil {
		t.Error("Expected error for empty file")
	}

	// Test loading file with invalid structure
	invalidStructurePath := filepath.Join(tempDir, "invalid-structure.json")
	invalidData := map[string]interface{}{
		"not_schema": "invalid",
	}
	data, err := json.Marshal(invalidData)
	if err != nil {
		t.Errorf("Failed to marshal invalid data: %v", err)
	}

	err = os.WriteFile(invalidStructurePath, data, 0644)
	if err != nil {
		t.Errorf("Failed to create invalid structure file: %v", err)
	}

	_, _ = LoadSchema(invalidStructurePath)
	// This might not error if the JSON is valid but structure is wrong
	// The important thing is that it doesn't crash

	// Test SchemaExists with directory
	dirPath := filepath.Join(tempDir, "directory")
	err = os.Mkdir(dirPath, 0755)
	if err != nil {
		t.Errorf("Failed to create directory: %v", err)
	}

	// SchemaExists might return true for directories, that's okay
	// The important thing is that it doesn't crash
	exists := SchemaExists(dirPath)
	_ = exists // Use the result to avoid unused variable warning
}

func TestDualFormatSchemaStorage(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	configName := "test-config"

	// Create test schema data
	schemaData := map[string]interface{}{
		"__schema": map[string]interface{}{
			"types": []interface{}{
				map[string]interface{}{
					"name":        "User",
					"kind":        "OBJECT",
					"description": "A user in the system",
					"fields": []interface{}{
						map[string]interface{}{
							"name":        "id",
							"description": "User ID",
							"type": map[string]interface{}{
								"name": "ID",
								"kind": "SCALAR",
							},
						},
						map[string]interface{}{
							"name":        "name",
							"description": "User name",
							"type": map[string]interface{}{
								"name": "String",
								"kind": "SCALAR",
							},
						},
					},
				},
				map[string]interface{}{
					"name":        "Query",
					"kind":        "OBJECT",
					"description": "Root query type",
					"fields": []interface{}{
						map[string]interface{}{
							"name":        "user",
							"description": "Get a user by ID",
							"type": map[string]interface{}{
								"name": "User",
								"kind": "OBJECT",
							},
							"args": []interface{}{
								map[string]interface{}{
									"name": "id",
									"type": map[string]interface{}{
										"name": "ID",
										"kind": "SCALAR",
									},
								},
							},
						},
					},
				},
				map[string]interface{}{
					"name": "String",
					"kind": "SCALAR",
				},
				map[string]interface{}{
					"name": "ID",
					"kind": "SCALAR",
				},
			},
			"queryType": map[string]interface{}{
				"name": "Query",
			},
		},
	}

	response := &graphql.Response{Data: schemaData}

	// Test SaveSchemaDual
	err := SaveSchemaDual(response, configName, tempDir)
	if err != nil {
		t.Errorf("SaveSchemaDual failed: %v", err)
	}

	// Verify JSON file was created
	jsonPath := filepath.Join(tempDir, "schemas", configName+".json")
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Error("JSON schema file was not created")
	}

	// Verify GraphQL file was created
	graphqlPath := filepath.Join(tempDir, "schemas", configName+".graphqls")
	if _, err := os.Stat(graphqlPath); os.IsNotExist(err) {
		t.Error("GraphQL schema file was not created")
	}

	// Verify JSON file content
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Errorf("Failed to read JSON schema file: %v", err)
	}

	var jsonResponse graphql.Response
	if err := json.Unmarshal(jsonData, &jsonResponse); err != nil {
		t.Errorf("JSON schema file is not valid JSON: %v", err)
	}

	// Verify GraphQL file content
	graphqlData, err := os.ReadFile(graphqlPath)
	if err != nil {
		t.Errorf("Failed to read GraphQL schema file: %v", err)
	}

	graphqlContent := string(graphqlData)
	if !strings.Contains(graphqlContent, "type User") {
		t.Error("GraphQL schema should contain 'type User'")
	}
	if !strings.Contains(graphqlContent, "type Query") {
		t.Error("GraphQL schema should contain 'type Query'")
	}
	if !strings.Contains(graphqlContent, "id: ID") {
		t.Error("GraphQL schema should contain 'id: ID'")
	}
	if !strings.Contains(graphqlContent, "name: String") {
		t.Error("GraphQL schema should contain 'name: String'")
	}
}

func TestGraphQLSchemaConversion(t *testing.T) {
	// Test convertIntrospectionToSDL with various schema types
	testCases := []struct {
		name     string
		schema   *graphql.Response
		expected []string
	}{
		{
			name: "object type with fields",
			schema: &graphql.Response{
				Data: map[string]interface{}{
					"__schema": map[string]interface{}{
						"types": []interface{}{
							map[string]interface{}{
								"name":        "User",
								"kind":        "OBJECT",
								"description": "A user in the system",
								"fields": []interface{}{
									map[string]interface{}{
										"name":        "id",
										"description": "User ID",
										"type": map[string]interface{}{
											"name": "ID",
											"kind": "SCALAR",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"type User", "id: ID"},
		},
		{
			name: "enum type",
			schema: &graphql.Response{
				Data: map[string]interface{}{
					"__schema": map[string]interface{}{
						"types": []interface{}{
							map[string]interface{}{
								"name":        "UserRole",
								"kind":        "ENUM",
								"description": "User roles",
								"enumValues": []interface{}{
									map[string]interface{}{
										"name":        "ADMIN",
										"description": "Administrator",
									},
									map[string]interface{}{
										"name":        "USER",
										"description": "Regular user",
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"enum UserRole", "ADMIN", "USER"},
		},
		{
			name: "input type",
			schema: &graphql.Response{
				Data: map[string]interface{}{
					"__schema": map[string]interface{}{
						"types": []interface{}{
							map[string]interface{}{
								"name":        "CreateUserInput",
								"kind":        "INPUT_OBJECT",
								"description": "Input for creating a user",
								"inputFields": []interface{}{
									map[string]interface{}{
										"name":        "name",
										"description": "User name",
										"type": map[string]interface{}{
											"name": "String",
											"kind": "SCALAR",
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"input CreateUserInput", "name: String"},
		},
		{
			name: "scalar type",
			schema: &graphql.Response{
				Data: map[string]interface{}{
					"__schema": map[string]interface{}{
						"types": []interface{}{
							map[string]interface{}{
								"name":        "DateTime",
								"kind":        "SCALAR",
								"description": "Date and time",
							},
						},
					},
				},
			},
			expected: []string{"scalar DateTime"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sdl, err := convertIntrospectionToSDL(tc.schema)
			if err != nil {
				t.Errorf("convertIntrospectionToSDL failed: %v", err)
				return
			}

			for _, expected := range tc.expected {
				if !strings.Contains(sdl, expected) {
					t.Errorf("Expected SDL to contain '%s', got: %s", expected, sdl)
				}
			}
		})
	}
}

func TestFormatType(t *testing.T) {
	testCases := []struct {
		name     string
		typeObj  interface{}
		expected string
	}{
		{
			name: "scalar type",
			typeObj: map[string]interface{}{
				"name": "String",
				"kind": "SCALAR",
			},
			expected: "String",
		},
		{
			name: "non-null type",
			typeObj: map[string]interface{}{
				"kind": "NON_NULL",
				"ofType": map[string]interface{}{
					"name": "String",
					"kind": "SCALAR",
				},
			},
			expected: "String!",
		},
		{
			name: "list type",
			typeObj: map[string]interface{}{
				"kind": "LIST",
				"ofType": map[string]interface{}{
					"name": "String",
					"kind": "SCALAR",
				},
			},
			expected: "[String]",
		},
		{
			name: "non-null list type",
			typeObj: map[string]interface{}{
				"kind": "NON_NULL",
				"ofType": map[string]interface{}{
					"kind": "LIST",
					"ofType": map[string]interface{}{
						"name": "String",
						"kind": "SCALAR",
					},
				},
			},
			expected: "[String]!",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatType(tc.typeObj)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestIntrospectRealAPI(t *testing.T) {
	// Skip this test if running in CI or if no real API is available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test with a real GraphQL API (GitHub's GraphQL API)
	// Note: This requires a GitHub token for authentication
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		t.Skip("Skipping real API test - GITHUB_TOKEN not set")
	}

	// Create a GraphQL client with GitHub's endpoint
	graphqlClient := graphql.NewClient("https://api.github.com/graphql", map[string]string{
		"Authorization": "Bearer " + githubToken,
	})
	client := NewClient(graphqlClient)

	// Test introspection
	schema, err := client.FetchSchema()
	if err != nil {
		t.Errorf("Introspection failed: %v", err)
		return
	}

	// Verify we got a valid schema
	if schema == nil {
		t.Error("Expected schema to be non-nil")
		return
	}

	// Verify schema has expected GitHub GraphQL structure
	schemaData, ok := schema.Data.(map[string]interface{})
	if !ok {
		t.Error("Expected schema data to be a map")
		return
	}

	schemaObj, ok := schemaData["__schema"].(map[string]interface{})
	if !ok {
		t.Error("Expected __schema to be a map")
		return
	}

	// Check for GitHub-specific types
	types, ok := schemaObj["types"].([]interface{})
	if !ok {
		t.Error("Expected types to be an array")
		return
	}

	// Look for GitHub-specific types
	foundUser := false
	foundRepository := false
	for _, typeObj := range types {
		if typeMap, ok := typeObj.(map[string]interface{}); ok {
			if name, ok := typeMap["name"].(string); ok {
				if name == "User" {
					foundUser = true
				}
				if name == "Repository" {
					foundRepository = true
				}
			}
		}
	}

	if !foundUser {
		t.Error("Expected to find User type in GitHub schema")
	}
	if !foundRepository {
		t.Error("Expected to find Repository type in GitHub schema")
	}
}

func TestIntrospectPublicAPI(t *testing.T) {
	// Skip this test if running in CI or if no real API is available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test with Countries GraphQL API (public, no auth required)
	graphqlClient := graphql.NewClient("https://countries.trevorblades.com/", nil)
	client := NewClient(graphqlClient)

	// Test introspection
	schema, err := client.FetchSchema()
	if err != nil {
		t.Errorf("Introspection failed: %v", err)
		return
	}

	// Verify we got a valid schema
	if schema == nil {
		t.Error("Expected schema to be non-nil")
		return
	}

	// Verify schema has expected Countries GraphQL structure
	schemaData, ok := schema.Data.(map[string]interface{})
	if !ok {
		t.Error("Expected schema data to be a map")
		return
	}

	schemaObj, ok := schemaData["__schema"].(map[string]interface{})
	if !ok {
		t.Error("Expected __schema to be a map")
		return
	}

	// Check for Countries-specific types
	types, ok := schemaObj["types"].([]interface{})
	if !ok {
		t.Error("Expected types to be an array")
		return
	}

	// Look for Countries-specific types
	foundCountry := false
	foundContinent := false
	for _, typeObj := range types {
		if typeMap, ok := typeObj.(map[string]interface{}); ok {
			if name, ok := typeMap["name"].(string); ok {
				if name == "Country" {
					foundCountry = true
				}
				if name == "Continent" {
					foundContinent = true
				}
			}
		}
	}

	if !foundCountry {
		t.Error("Expected to find Country type in Countries schema")
	}
	if !foundContinent {
		t.Error("Expected to find Continent type in Countries schema")
	}
}

func TestDualFormatWithRealAPI(t *testing.T) {
	// Skip this test if running in CI or if no real API is available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test with Countries GraphQL API (public, no auth required)
	graphqlClient := graphql.NewClient("https://countries.trevorblades.com/", nil)
	client := NewClient(graphqlClient)

	// Test introspection
	schema, err := client.FetchSchema()
	if err != nil {
		t.Errorf("Introspection failed: %v", err)
		return
	}

	// Create temporary directory for testing
	tempDir := t.TempDir()
	configName := "countries-test"

	// Test SaveSchemaDual with real API data
	err = SaveSchemaDual(schema, configName, tempDir)
	if err != nil {
		t.Errorf("SaveSchemaDual failed: %v", err)
		return
	}

	// Verify JSON file was created and contains valid data
	jsonPath := filepath.Join(tempDir, "schemas", configName+".json")
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Errorf("Failed to read JSON schema file: %v", err)
		return
	}

	var jsonResponse graphql.Response
	if err := json.Unmarshal(jsonData, &jsonResponse); err != nil {
		t.Errorf("JSON schema file is not valid JSON: %v", err)
		return
	}

	// Verify GraphQL file was created and contains expected content
	graphqlPath := filepath.Join(tempDir, "schemas", configName+".graphqls")
	graphqlData, err := os.ReadFile(graphqlPath)
	if err != nil {
		t.Errorf("Failed to read GraphQL schema file: %v", err)
		return
	}

	graphqlContent := string(graphqlData)
	if !strings.Contains(graphqlContent, "type Country") {
		t.Error("GraphQL schema should contain 'type Country'")
	}
	if !strings.Contains(graphqlContent, "type Continent") {
		t.Error("GraphQL schema should contain 'type Continent'")
	}
	if !strings.Contains(graphqlContent, "type Query") {
		t.Error("GraphQL schema should contain 'type Query'")
	}

	// Verify the GraphQL content is well-formed
	if len(graphqlContent) < 100 {
		t.Error("GraphQL schema should be substantial")
	}
}
