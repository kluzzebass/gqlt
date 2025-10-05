package introspect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kluzzebass/gqlt/internal/graphql"
)

func TestNewClient(t *testing.T) {
	mockClient := &graphql.Client{}
	client := NewClient(mockClient)

	if client.graphqlClient != mockClient {
		t.Error("Expected graphqlClient to be set")
	}
}

func TestFetchSchema(t *testing.T) {
	// This test would require a mock GraphQL client
	// For now, we'll test the introspection query constant
	if IntrospectQuery == "" {
		t.Error("IntrospectQuery should not be empty")
	}

	// Verify the query contains expected fields
	expectedFields := []string{
		"__schema",
		"types",
		"name",
		"kind",
		"fields",
		"queryType",
		"mutationType",
		"subscriptionType",
	}

	for _, field := range expectedFields {
		if !strings.Contains(IntrospectQuery, field) {
			t.Errorf("IntrospectQuery should contain '%s'", field)
		}
	}
}

func TestSaveSchema(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	schemaPath := filepath.Join(tempDir, "schema.json")

	// Create mock schema response
	schemaData := map[string]interface{}{
		"__schema": map[string]interface{}{
			"types": []interface{}{
				map[string]interface{}{
					"name": "User",
					"kind": "OBJECT",
				},
			},
		},
	}

	response := &graphql.Response{
		Data: schemaData,
	}

	// Save schema
	err := SaveSchema(response, schemaPath)
	if err != nil {
		t.Errorf("SaveSchema failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		t.Error("Schema file was not created")
	}

	// Verify file content
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Errorf("Failed to read schema file: %v", err)
	}

	if len(data) == 0 {
		t.Error("Schema file is empty")
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Errorf("Schema file is not valid JSON: %v", err)
	}
}

func TestLoadSchema(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	schemaPath := filepath.Join(tempDir, "schema.json")

	// Create mock schema data
	schemaData := map[string]interface{}{
		"__schema": map[string]interface{}{
			"types": []interface{}{
				map[string]interface{}{
					"name": "User",
					"kind": "OBJECT",
				},
			},
		},
	}

	response := &graphql.Response{
		Data: schemaData,
	}

	// Save schema first
	err := SaveSchema(response, schemaPath)
	if err != nil {
		t.Errorf("SaveSchema failed: %v", err)
	}

	// Load schema
	loadedResponse, err := LoadSchema(schemaPath)
	if err != nil {
		t.Errorf("LoadSchema failed: %v", err)
	}

	if loadedResponse == nil {
		t.Error("Expected loaded response to be non-nil")
		return
	}

	// Verify data structure
	if loadedResponse.Data == nil {
		t.Error("Expected data in loaded response")
	}

	data, ok := loadedResponse.Data.(map[string]interface{})
	if !ok {
		t.Error("Expected data to be map[string]interface{}")
	}

	schema, ok := data["__schema"].(map[string]interface{})
	if !ok {
		t.Error("Expected __schema to be map[string]interface{}")
	}

	types, ok := schema["types"].([]interface{})
	if !ok {
		t.Error("Expected types to be []interface{}")
	}

	if len(types) != 1 {
		t.Errorf("Expected 1 type, got %d", len(types))
	}
}

func TestSchemaExists(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	existingPath := filepath.Join(tempDir, "existing.json")
	nonExistentPath := filepath.Join(tempDir, "non-existent.json")

	// Create a file
	err := os.WriteFile(existingPath, []byte("{}"), 0644)
	if err != nil {
		t.Errorf("Failed to create test file: %v", err)
	}

	// Test existing file
	if !SchemaExists(existingPath) {
		t.Error("Expected existing file to return true")
	}

	// Test non-existent file
	if SchemaExists(nonExistentPath) {
		t.Error("Expected non-existent file to return false")
	}
}

func TestLoadSchemaNonExistent(t *testing.T) {
	// Test loading non-existent file
	_, err := LoadSchema("/non/existent/path.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestLoadSchemaInvalidJSON(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	invalidPath := filepath.Join(tempDir, "invalid.json")

	// Create invalid JSON file
	err := os.WriteFile(invalidPath, []byte("invalid json"), 0644)
	if err != nil {
		t.Errorf("Failed to create invalid JSON file: %v", err)
	}

	// Test loading invalid JSON
	_, err = LoadSchema(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}
