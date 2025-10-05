package gqlt

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPaths(t *testing.T) {
	// Get all paths at the start
	basePath := GetDefaultPath()
	configPath := GetConfigPath()
	schemaPath := GetSchemaPath()

	// Test that all paths are absolute and non-empty
	if !filepath.IsAbs(basePath) || basePath == "" {
		t.Errorf("Base path should be absolute and non-empty, got: %s", basePath)
	}
	if !filepath.IsAbs(configPath) || configPath == "" {
		t.Errorf("Config path should be absolute and non-empty, got: %s", configPath)
	}
	if !filepath.IsAbs(schemaPath) || schemaPath == "" {
		t.Errorf("Schema path should be absolute and non-empty, got: %s", schemaPath)
	}

	// Test that all paths contain 'gqlt' directory
	if !strings.Contains(basePath, "gqlt") {
		t.Errorf("Base path should contain 'gqlt', got: %s", basePath)
	}
	if !strings.Contains(configPath, "gqlt") {
		t.Errorf("Config path should contain 'gqlt', got: %s", configPath)
	}
	if !strings.Contains(schemaPath, "gqlt") {
		t.Errorf("Schema path should contain 'gqlt', got: %s", schemaPath)
	}

	// Test OS-specific behavior
	switch runtime.GOOS {
	case "darwin":
		if !strings.Contains(basePath, "Library/Application Support") {
			t.Errorf("macOS path should contain 'Library/Application Support', got: %s", basePath)
		}
	case "windows":
		if !strings.Contains(basePath, "AppData") {
			t.Errorf("Windows path should contain 'AppData', got: %s", basePath)
		}
	default:
		if !strings.Contains(basePath, ".config") {
			t.Errorf("Linux/Unix path should contain '.config', got: %s", basePath)
		}
	}

	// Test that config and schema paths use the base path
	expectedConfigPath := filepath.Join(basePath, "config.json")
	expectedSchemaPath := filepath.Join(basePath, "schema.json")

	if configPath != expectedConfigPath {
		t.Errorf("Config path should be %s, got %s", expectedConfigPath, configPath)
	}
	if schemaPath != expectedSchemaPath {
		t.Errorf("Schema path should be %s, got %s", expectedSchemaPath, schemaPath)
	}

	// Test that paths have correct extensions
	if !strings.HasSuffix(configPath, "config.json") {
		t.Errorf("Config path should end with 'config.json', got: %s", configPath)
	}
	if !strings.HasSuffix(schemaPath, "schema.json") {
		t.Errorf("Schema path should end with 'schema.json', got: %s", schemaPath)
	}

	// Test that paths are unique
	if configPath == schemaPath {
		t.Error("Config and schema paths should be different")
	}

	// Test path stability (same result across multiple calls)
	if GetDefaultPath() != basePath {
		t.Error("GetDefaultPath should return the same path across calls")
	}
	if GetConfigPath() != configPath {
		t.Error("GetConfigPath should return the same path across calls")
	}
	if GetSchemaPath() != schemaPath {
		t.Error("GetSchemaPath should return the same path across calls")
	}
}

func TestPathsWithEnvVars(t *testing.T) {
	// Test with custom HOME environment variable
	originalHome := os.Getenv("HOME")
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	// Set custom HOME
	testHome := "/custom/home"
	os.Setenv("HOME", testHome)

	basePath := GetDefaultPath()
	configPath := GetConfigPath()
	schemaPath := GetSchemaPath()

	// All paths should contain custom HOME
	if !strings.Contains(basePath, testHome) {
		t.Errorf("Base path should contain custom HOME '%s', got: %s", testHome, basePath)
	}
	if !strings.Contains(configPath, testHome) {
		t.Errorf("Config path should contain custom HOME '%s', got: %s", testHome, configPath)
	}
	if !strings.Contains(schemaPath, testHome) {
		t.Errorf("Schema path should contain custom HOME '%s', got: %s", testHome, schemaPath)
	}
}

func TestDualFormatSchemaPaths(t *testing.T) {
	// Test dual format schema path functions
	basePath := "/test/config"
	configName := "production"

	// Test JSON schema path
	jsonPath := GetJSONSchemaPathForConfigInDir(configName, basePath)
	expectedJSONPath := filepath.Join(basePath, "schemas", configName+".json")
	if jsonPath != expectedJSONPath {
		t.Errorf("Expected JSON path %s, got %s", expectedJSONPath, jsonPath)
	}

	// Test GraphQL schema path
	graphqlPath := GetGraphQLSchemaPathForConfigInDir(configName, basePath)
	expectedGraphQLPath := filepath.Join(basePath, "schemas", configName+".graphqls")
	if graphqlPath != expectedGraphQLPath {
		t.Errorf("Expected GraphQL path %s, got %s", expectedGraphQLPath, graphqlPath)
	}

	// Test that paths are different
	if jsonPath == graphqlPath {
		t.Error("JSON and GraphQL schema paths should be different")
	}

	// Test that both paths are in the same directory
	jsonDir := filepath.Dir(jsonPath)
	graphqlDir := filepath.Dir(graphqlPath)
	if jsonDir != graphqlDir {
		t.Error("JSON and GraphQL schema paths should be in the same directory")
	}

	// Test with empty config directory
	emptyJSONPath := GetJSONSchemaPathForConfigInDir(configName, "")
	emptyGraphQLPath := GetGraphQLSchemaPathForConfigInDir(configName, "")

	// Should fall back to default paths
	if !strings.Contains(emptyJSONPath, "schemas") {
		t.Errorf("Empty config dir JSON path should contain 'schemas', got: %s", emptyJSONPath)
	}
	if !strings.Contains(emptyGraphQLPath, "schemas") {
		t.Errorf("Empty config dir GraphQL path should contain 'schemas', got: %s", emptyGraphQLPath)
	}

	// Test additional path functions for coverage
	schemasDir := GetSchemasDir()
	if !filepath.IsAbs(schemasDir) || schemasDir == "" {
		t.Errorf("Schemas directory should be absolute and non-empty, got: %s", schemasDir)
	}

	configSchemaPath := GetSchemaPathForConfig("test-config")
	if !filepath.IsAbs(configSchemaPath) || configSchemaPath == "" {
		t.Errorf("Config schema path should be absolute and non-empty, got: %s", configSchemaPath)
	}

	schemaPath := GetSchemaPath()
	if !filepath.IsAbs(schemaPath) || schemaPath == "" {
		t.Errorf("Schema path should be absolute and non-empty, got: %s", schemaPath)
	}
}
