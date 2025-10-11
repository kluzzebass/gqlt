package gqlt

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()

	if config == nil {
		t.Error("Expected default config to be non-nil")
		return
	}

	if config.Current != "default" {
		t.Errorf("Expected current to be 'default', got %s", config.Current)
	}

	if config.Configs == nil {
		t.Error("Expected Configs to be non-nil")
	}

	if len(config.Configs) != 1 {
		t.Errorf("Expected 1 config, got %d", len(config.Configs))
	}

	defaultEntry, exists := config.Configs["default"]
	if !exists {
		t.Error("Expected 'default' config to exist")
	}

	if defaultEntry.Endpoint != "" {
		t.Errorf("Expected default endpoint to be empty, got %s", defaultEntry.Endpoint)
	}
}

func TestConfigOperations(t *testing.T) {
	config := GetDefaultConfig()

	// Test Create
	err := config.Create("test")
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}

	if _, exists := config.Configs["test"]; !exists {
		t.Error("Expected 'test' config to exist after Create")
	}

	// Test Create duplicate
	err = config.Create("test")
	if err == nil {
		t.Error("Expected error when creating duplicate config")
	}

	// Test SetCurrent
	err = config.SetCurrent("test")
	if err != nil {
		t.Errorf("SetCurrent failed: %v", err)
	}

	if config.Current != "test" {
		t.Errorf("Expected current to be 'test', got %s", config.Current)
	}

	// Test SetCurrent non-existent
	err = config.SetCurrent("non-existent")
	if err == nil {
		t.Error("Expected error when setting current to non-existent config")
	}

	// Test SetValue
	err = config.SetValue("test", "endpoint", "https://api.example.com/graphql")
	if err != nil {
		t.Errorf("SetValue failed: %v", err)
	}

	if config.Configs["test"].Endpoint != "https://api.example.com/graphql" {
		t.Errorf("Expected endpoint to be set, got %s", config.Configs["test"].Endpoint)
	}

	// Test SetValue non-existent config
	err = config.SetValue("non-existent", "endpoint", "https://api.example.com/graphql")
	if err == nil {
		t.Error("Expected error when setting value on non-existent config")
	}

	// Test SetValue invalid key
	err = config.SetValue("test", "invalid-key", "value")
	if err == nil {
		t.Error("Expected error when setting invalid key")
	}

	// Test Delete
	err = config.Delete("test")
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	if _, exists := config.Configs["test"]; exists {
		t.Error("Expected 'test' config to be deleted")
	}

	// Test Delete default
	err = config.Delete("default")
	if err == nil {
		t.Error("Expected error when deleting default config")
	}

	// Test Delete non-existent
	err = config.Delete("non-existent")
	if err == nil {
		t.Error("Expected error when deleting non-existent config")
	}
}

func TestGetCurrent(t *testing.T) {
	config := GetDefaultConfig()

	// Test with default current
	current := config.GetCurrent()
	if current == nil {
		t.Error("Expected current config to be non-nil")
	}

	// Test with non-existent current (should fallback to default)
	config.Current = "non-existent"
	current = config.GetCurrent()
	if current == nil {
		t.Error("Expected fallback config to be non-nil")
	}

	// Test with empty configs (should return default entry)
	config.Configs = nil
	current = config.GetCurrent()
	if current == nil {
		t.Error("Expected default entry to be returned")
	}
}

func TestValidate(t *testing.T) {
	config := GetDefaultConfig()

	// Test valid config
	errors := config.Validate()
	if len(errors) > 0 {
		t.Errorf("Expected no validation errors, got: %v", errors)
	}

	// Test invalid current
	config.Current = ""
	errors = config.Validate()
	if len(errors) == 0 {
		t.Error("Expected validation error for empty current")
	}

	// Test non-existent current
	config.Current = "non-existent"
	errors = config.Validate()
	if len(errors) == 0 {
		t.Error("Expected validation error for non-existent current")
	}

	// Test config with empty endpoint
	config.Current = "default"
	config.Configs["test"] = ConfigEntry{
		Endpoint: "",
		Headers:  make(map[string]string),
	}
	errors = config.Validate()
	if len(errors) == 0 {
		t.Error("Expected validation error for config with empty endpoint")
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Create config
	config := GetDefaultConfig()
	config.Create("test")
	config.SetValue("test", "endpoint", "https://api.example.com/graphql")

	// Save config
	err := config.Save(configPath)
	if err != nil {
		t.Errorf("Save failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Load config
	loadedConfig, err := Load(configPath)
	if err != nil {
		t.Errorf("Load failed: %v", err)
	}

	if loadedConfig == nil {
		t.Error("Expected loaded config to be non-nil")
		return
	}

	// Verify loaded config
	if loadedConfig.Current != config.Current {
		t.Errorf("Expected current to be %s, got %s", config.Current, loadedConfig.Current)
	}

	if len(loadedConfig.Configs) != len(config.Configs) {
		t.Errorf("Expected %d configs, got %d", len(config.Configs), len(loadedConfig.Configs))
	}

	// Verify test config
	testConfig, exists := loadedConfig.Configs["test"]
	if !exists {
		t.Error("Expected 'test' config to exist in loaded config")
	}

	if testConfig.Endpoint != "https://api.example.com/graphql" {
		t.Errorf("Expected endpoint to be 'https://api.example.com/graphql', got %s", testConfig.Endpoint)
	}
}

func TestLoadNonExistent(t *testing.T) {
	// Test loading non-existent file (should return default config)
	config, err := Load("/non/existent/path.json")
	if err != nil {
		t.Errorf("Load failed: %v", err)
	}

	if config == nil {
		t.Error("Expected default config to be returned")
		return
	}

	// Should be default config
	if config.Current != "default" {
		t.Errorf("Expected current to be 'default', got %s", config.Current)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Create invalid JSON file
	err := os.WriteFile(configPath, []byte("invalid json"), 0644)
	if err != nil {
		t.Errorf("Failed to create invalid JSON file: %v", err)
	}

	// Test loading invalid JSON
	_, err = Load(tempDir) // This should try to load config.json from tempDir
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestGetSchema(t *testing.T) {
	schema := GetSchema()

	if schema == nil {
		t.Error("Expected schema to be non-nil")
		return
	}

	if schema.Endpoint == "" {
		t.Error("Expected schema endpoint description to be non-empty")
	}

	if schema.Headers == "" {
		t.Error("Expected schema headers description to be non-empty")
	}
}

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
