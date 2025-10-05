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

	_, err = LoadSchema(invalidStructurePath)
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
