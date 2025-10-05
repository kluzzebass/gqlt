package input

import (
	"os"
	"path/filepath"
	"strings"
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

	// Test with valid file specifications
	files := []string{"file1=test1.txt", "file2=test2.txt"}

	// Create temporary files for testing
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "test1.txt")
	file2 := filepath.Join(tempDir, "test2.txt")

	err = os.WriteFile(file1, []byte("content1"), 0644)
	if err != nil {
		t.Errorf("Failed to create test file 1: %v", err)
	}

	err = os.WriteFile(file2, []byte("content2"), 0644)
	if err != nil {
		t.Errorf("Failed to create test file 2: %v", err)
	}

	// Update file paths to use temp files
	files[0] = "file1=" + file1
	files[1] = "file2=" + file2

	result, err = ParseFiles(files)
	if err != nil {
		t.Errorf("ParseFiles failed: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 files, got %d", len(result))
	}
	if result["file1"] != file1 {
		t.Errorf("Expected file1 path %s, got %s", file1, result["file1"])
	}
	if result["file2"] != file2 {
		t.Errorf("Expected file2 path %s, got %s", file2, result["file2"])
	}

	// Test with invalid format
	invalidFiles := []string{"invalid-format"}
	_, err = ParseFiles(invalidFiles)
	if err == nil {
		t.Error("Expected error for invalid file format")
	}

	// Test with non-existent file
	nonExistentFiles := []string{"file1=non-existent.txt"}
	_, err = ParseFiles(nonExistentFiles)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestParseFilesFromList(t *testing.T) {
	// Test with empty files list path
	result, err := ParseFilesFromList("")
	if err != nil {
		t.Errorf("ParseFilesFromList failed: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %v", result)
	}

	// Test with files list file
	tempDir := t.TempDir()
	filesListPath := filepath.Join(tempDir, "files.txt")

	// Create files list content
	filesListContent := `# This is a comment
file1=test1.txt
file2=./test2.txt
file3=../test3.txt

# Another comment
file4=test4.txt`

	err = os.WriteFile(filesListPath, []byte(filesListContent), 0644)
	if err != nil {
		t.Errorf("Failed to create files list: %v", err)
	}

	result, err = ParseFilesFromList(filesListPath)
	if err != nil {
		t.Errorf("ParseFilesFromList failed: %v", err)
	}
	if len(result) != 4 {
		t.Errorf("Expected 4 files, got %d", len(result))
	}

	// Verify paths were resolved to absolute paths
	for i, file := range result {
		if !strings.Contains(file, "=") {
			t.Errorf("Expected file %d to contain '=', got '%s'", i, file)
		}
		parts := strings.SplitN(file, "=", 2)
		if len(parts) != 2 {
			t.Errorf("Expected file %d to be in 'name=path' format, got '%s'", i, file)
		}
		// Check that the path is absolute
		if !filepath.IsAbs(parts[1]) {
			t.Errorf("Expected file %d path to be absolute, got '%s'", i, parts[1])
		}
	}

	// Test with invalid format in list
	invalidListContent := `file1=test1.txt
invalid-format
file2=test2.txt`

	invalidListPath := filepath.Join(tempDir, "invalid.txt")
	err = os.WriteFile(invalidListPath, []byte(invalidListContent), 0644)
	if err != nil {
		t.Errorf("Failed to create invalid list: %v", err)
	}

	_, err = ParseFilesFromList(invalidListPath)
	if err == nil {
		t.Error("Expected error for invalid format in list")
	}
}

func TestResolveFilePath(t *testing.T) {
	// Test basic path resolution
	line := "file1=./test.txt"
	result, err := resolveFilePath(line)
	if err != nil {
		t.Errorf("resolveFilePath failed: %v", err)
	}
	if !strings.HasPrefix(result, "file1=") {
		t.Errorf("Expected result to start with 'file1=', got '%s'", result)
	}
	path := strings.TrimPrefix(result, "file1=")
	if !filepath.IsAbs(path) {
		t.Errorf("Expected resolved path to be absolute, got '%s'", path)
	}

	// Test ~ expansion
	line = "file2=~/test.txt"
	result, err = resolveFilePath(line)
	if err != nil {
		t.Errorf("resolveFilePath with ~ failed: %v", err)
	}
	if !strings.HasPrefix(result, "file2=") {
		t.Errorf("Expected result to start with 'file2=', got '%s'", result)
	}
	path = strings.TrimPrefix(result, "file2=")
	if !filepath.IsAbs(path) {
		t.Errorf("Expected resolved path to be absolute, got '%s'", path)
	}
	if !strings.Contains(path, "test.txt") {
		t.Errorf("Expected path to contain 'test.txt', got '%s'", path)
	}

	// Test invalid line (no =)
	line = "invalid-line"
	result, err = resolveFilePath(line)
	if err != nil {
		t.Errorf("resolveFilePath with invalid line should not error: %v", err)
	}
	if result != line {
		t.Errorf("Expected invalid line to be returned as-is, got '%s'", result)
	}
}
