package input

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadQuery(t *testing.T) {
	// Test with direct query
	query, err := LoadQuery("query { user { id } }", "")
	if err != nil {
		t.Errorf("LoadQuery failed: %v", err)
	}
	if query != "query { user { id } }" {
		t.Errorf("Expected query 'query { user { id } }', got %s", query)
	}

	// Test with query file
	tempDir := t.TempDir()
	queryFile := filepath.Join(tempDir, "query.graphql")
	err = os.WriteFile(queryFile, []byte("query { user { id name } }"), 0644)
	if err != nil {
		t.Errorf("Failed to create query file: %v", err)
	}

	query, err = LoadQuery("", queryFile)
	if err != nil {
		t.Errorf("LoadQuery from file failed: %v", err)
	}
	if query != "query { user { id name } }" {
		t.Errorf("Expected query from file, got %s", query)
	}

	// Test with non-existent file
	_, err = LoadQuery("", "/non/existent/file.graphql")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestLoadVariables(t *testing.T) {
	// Test with direct variables
	variables, err := LoadVariables(`{"id": "123"}`, "")
	if err != nil {
		t.Errorf("LoadVariables failed: %v", err)
	}
	if variables["id"] != "123" {
		t.Errorf("Expected id '123', got %v", variables["id"])
	}

	// Test with variables file
	tempDir := t.TempDir()
	varsFile := filepath.Join(tempDir, "vars.json")
	err = os.WriteFile(varsFile, []byte(`{"id": "456", "name": "test"}`), 0644)
	if err != nil {
		t.Errorf("Failed to create vars file: %v", err)
	}

	variables, err = LoadVariables("", varsFile)
	if err != nil {
		t.Errorf("LoadVariables from file failed: %v", err)
	}
	if variables["id"] != "456" {
		t.Errorf("Expected id '456', got %v", variables["id"])
	}
	if variables["name"] != "test" {
		t.Errorf("Expected name 'test', got %v", variables["name"])
	}

	// Test with invalid JSON
	_, err = LoadVariables(`{"invalid": json}`, "")
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}

	// Test with non-existent file
	_, err = LoadVariables("", "/non/existent/vars.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestLoadHeaders(t *testing.T) {
	// Test with valid headers
	headers := []string{"Authorization: Bearer token123", "X-API-Key: key456"}
	result := LoadHeaders(headers)
	if result["Authorization"] != "Bearer token123" {
		t.Errorf("Expected Authorization header, got %s", result["Authorization"])
	}
	if result["X-API-Key"] != "key456" {
		t.Errorf("Expected X-API-Key header, got %s", result["X-API-Key"])
	}

	// Test with empty headers
	result = LoadHeaders([]string{})
	if len(result) != 0 {
		t.Errorf("Expected empty headers, got %v", result)
	}

	// Test with invalid header format (should be ignored)
	result = LoadHeaders([]string{"InvalidHeader"})
	if len(result) != 0 {
		t.Errorf("Expected empty headers for invalid format, got %v", result)
	}
}

func TestParseFiles(t *testing.T) {
	// Test with empty files list
	result, err := ParseFiles([]string{})
	if err != nil {
		t.Errorf("ParseFiles failed: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %v", result)
	}

	// Test with files list (currently returns empty map as not implemented)
	files := []string{"file1:file1.txt", "file2:file2.txt"}
	result, err = ParseFiles(files)
	if err != nil {
		t.Errorf("ParseFiles failed: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty result (not implemented), got %v", result)
	}
}
