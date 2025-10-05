package config

import (
	"os"
	"path/filepath"
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

	if defaultEntry.Defaults.Out != "json" {
		t.Errorf("Expected default output to be 'json', got %s", defaultEntry.Defaults.Out)
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
		Defaults: struct {
			Out string `json:"out"`
		}{
			Out: "json",
		},
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
	invalidPath := filepath.Join(tempDir, "invalid.json")

	// Create invalid JSON file
	err := os.WriteFile(invalidPath, []byte("invalid json"), 0644)
	if err != nil {
		t.Errorf("Failed to create invalid JSON file: %v", err)
	}

	// Test loading invalid JSON
	_, err = Load(invalidPath)
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

	if schema.DefaultsOut == "" {
		t.Error("Expected schema defaults.out description to be non-empty")
	}
}
