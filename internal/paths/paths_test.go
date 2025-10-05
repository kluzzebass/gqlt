package paths

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetDefaultPath(t *testing.T) {
	path := GetDefaultPath()

	// Verify path is not empty
	if path == "" {
		t.Error("GetDefaultPath() returned empty string")
	}

	// Verify path contains expected directory structure
	expectedDir := "gqlt"
	if !strings.Contains(path, expectedDir) {
		t.Errorf("GetDefaultPath() should contain 'gqlt' directory, got: %s", path)
	}
}

func TestGetConfigPath(t *testing.T) {
	configPath := GetConfigPath()

	// Verify path is not empty
	if configPath == "" {
		t.Error("GetConfigPath() returned empty string")
	}

	// Verify path ends with config.json
	expectedFile := "config.json"
	if !strings.Contains(configPath, expectedFile) {
		t.Errorf("GetConfigPath() should end with 'config.json', got: %s", configPath)
	}

	// Verify path contains gqlt directory
	if !strings.Contains(configPath, "gqlt") {
		t.Errorf("GetConfigPath() should contain 'gqlt' directory, got: %s", configPath)
	}
}

func TestGetSchemaPath(t *testing.T) {
	schemaPath := GetSchemaPath()

	// Verify path is not empty
	if schemaPath == "" {
		t.Error("GetSchemaPath() returned empty string")
	}

	// Verify path ends with schema.json
	expectedFile := "schema.json"
	if !strings.Contains(schemaPath, expectedFile) {
		t.Errorf("GetSchemaPath() should end with 'schema.json', got: %s", schemaPath)
	}

	// Verify path contains gqlt directory
	if !strings.Contains(schemaPath, "gqlt") {
		t.Errorf("GetSchemaPath() should contain 'gqlt' directory, got: %s", schemaPath)
	}
}

func TestPathConsistency(t *testing.T) {
	// Test that all paths use the same base directory
	basePath := GetDefaultPath()
	configPath := GetConfigPath()
	schemaPath := GetSchemaPath()

	// Config path should be base + config.json
	expectedConfigPath := filepath.Join(basePath, "config.json")
	if configPath != expectedConfigPath {
		t.Errorf("Config path mismatch: expected %s, got %s", expectedConfigPath, configPath)
	}

	// Schema path should be base + schema.json
	expectedSchemaPath := filepath.Join(basePath, "schema.json")
	if schemaPath != expectedSchemaPath {
		t.Errorf("Schema path mismatch: expected %s, got %s", expectedSchemaPath, schemaPath)
	}
}

func TestOSSpecificPaths(t *testing.T) {
	// Test that paths are OS-specific
	path := GetDefaultPath()

	switch runtime.GOOS {
	case "darwin":
		// macOS should use Library/Application Support
		if !strings.Contains(path, "Library") || !strings.Contains(path, "Application Support") {
			t.Errorf("macOS path should contain 'Library/Application Support', got: %s", path)
		}
	case "windows":
		// Windows should use AppData
		if !strings.Contains(path, "AppData") {
			t.Errorf("Windows path should contain 'AppData', got: %s", path)
		}
	default:
		// Linux/Unix should use .config
		if !strings.Contains(path, ".config") {
			t.Errorf("Linux/Unix path should contain '.config', got: %s", path)
		}
	}
}

func TestPathStructure(t *testing.T) {
	// Test that paths have proper structure
	configPath := GetConfigPath()
	schemaPath := GetSchemaPath()

	// Both should be absolute paths
	if !filepath.IsAbs(configPath) {
		t.Errorf("Config path should be absolute, got: %s", configPath)
	}

	if !filepath.IsAbs(schemaPath) {
		t.Errorf("Schema path should be absolute, got: %s", schemaPath)
	}

	// Both should have proper file extensions
	if filepath.Ext(configPath) != ".json" {
		t.Errorf("Config path should have .json extension, got: %s", configPath)
	}

	if filepath.Ext(schemaPath) != ".json" {
		t.Errorf("Schema path should have .json extension, got: %s", schemaPath)
	}
}
