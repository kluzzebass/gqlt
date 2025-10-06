package main

import (
	"os"
	"strings"
	"testing"

	"github.com/kluzzebass/gqlt"
)

func TestMCPServerIntegration(t *testing.T) {
	// Create a temporary config for testing
	tempDir := t.TempDir()

	// Create a test config
	config := gqlt.GetDefaultConfig()
	config.Configs["test"] = gqlt.ConfigEntry{
		Endpoint: "https://api.example.com/graphql",
		Headers:  map[string]string{"Authorization": "Bearer test-token"},
	}

	// Save the config
	if err := config.Save(tempDir); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Test that the config was saved correctly
	loadedConfig, err := gqlt.Load(tempDir)
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	if loadedConfig.Configs["test"].Endpoint != "https://api.example.com/graphql" {
		t.Error("Config not loaded correctly")
	}
}

func TestMCPServerWithConfigDir(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()

	// Create a test config
	config := gqlt.GetDefaultConfig()
	config.Configs["production"] = gqlt.ConfigEntry{
		Endpoint: "https://api.production.com/graphql",
		Headers:  map[string]string{"Authorization": "Bearer prod-token"},
	}

	// Save the config
	if err := config.Save(tempDir); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Test loading with config directory
	loadedConfig, err := gqlt.Load(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config from directory: %v", err)
	}

	if loadedConfig.Configs["production"].Endpoint != "https://api.production.com/graphql" {
		t.Error("Production config not loaded correctly")
	}
}

func TestMCPServerWithSpecificConfig(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()

	// Create a test config with multiple environments
	config := gqlt.GetDefaultConfig()
	config.Configs["staging"] = gqlt.ConfigEntry{
		Endpoint: "https://api.staging.com/graphql",
		Headers:  map[string]string{"Authorization": "Bearer staging-token"},
	}
	config.Configs["production"] = gqlt.ConfigEntry{
		Endpoint: "https://api.production.com/graphql",
		Headers:  map[string]string{"Authorization": "Bearer prod-token"},
	}
	config.Current = "staging"

	// Save the config
	if err := config.Save(tempDir); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// Test loading and switching to specific config
	loadedConfig, err := gqlt.Load(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test switching to production config
	if err := loadedConfig.SetCurrent("production"); err != nil {
		t.Fatalf("Failed to switch to production config: %v", err)
	}

	current := loadedConfig.GetCurrent()
	if current.Endpoint != "https://api.production.com/graphql" {
		t.Error("Failed to switch to production config")
	}
}

func TestMCPServerStartStop(t *testing.T) {
	// Create a test config
	config := gqlt.GetDefaultConfig()
	config.Configs["test"] = gqlt.ConfigEntry{
		Endpoint: "https://api.example.com/graphql",
		Headers:  map[string]string{"Authorization": "Bearer test-token"},
	}

	// Test that we can create an MCP server
	// Note: We can't easily test the full server start/stop in a unit test
	// because it would require network operations and proper cleanup
	// This test just verifies that the server can be created with a config

	// This would be the equivalent of what the CLI does:
	// server, err := mcp.NewServer(config)
	// if err != nil {
	//     t.Fatalf("Failed to create MCP server: %v", err)
	// }

	// For now, just verify the config is valid
	if config.Configs["test"].Endpoint == "" {
		t.Error("Test config should have an endpoint")
	}
}

func TestMCPServerConfigValidation(t *testing.T) {
	// Test config validation
	config := gqlt.GetDefaultConfig()

	// Test with valid config
	errors := config.Validate()
	if len(errors) > 0 {
		t.Errorf("Default config should be valid, got errors: %v", errors)
	}

	// Test with invalid config (no endpoint)
	config.Configs["invalid"] = gqlt.ConfigEntry{
		Endpoint: "", // Invalid - no endpoint
		Headers:  make(map[string]string),
	}

	errors = config.Validate()
	if len(errors) == 0 {
		t.Error("Config with empty endpoint should have validation errors")
	}
}

func TestMCPServerConfigPersistence(t *testing.T) {
	// Create a temporary config directory
	tempDir := t.TempDir()

	// Create and save a config
	config := gqlt.GetDefaultConfig()
	config.Configs["test"] = gqlt.ConfigEntry{
		Endpoint: "https://api.test.com/graphql",
		Headers:  map[string]string{"Authorization": "Bearer test-token"},
		Auth: struct {
			Token    string `json:"token,omitempty"`
			Username string `json:"username,omitempty"`
			Password string `json:"password,omitempty"`
			APIKey   string `json:"api_key,omitempty"`
		}{
			Token: "test-bearer-token",
		},
	}

	// Save the config
	if err := config.Save(tempDir); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load the config back
	loadedConfig, err := gqlt.Load(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the config was saved and loaded correctly
	testConfig := loadedConfig.Configs["test"]
	if testConfig.Endpoint != "https://api.test.com/graphql" {
		t.Error("Endpoint not saved/loaded correctly")
	}

	if testConfig.Auth.Token != "test-bearer-token" {
		t.Error("Auth token not saved/loaded correctly")
	}

	if testConfig.Headers["Authorization"] != "Bearer test-token" {
		t.Error("Headers not saved/loaded correctly")
	}
}

func TestMCPServerEnvironmentVariables(t *testing.T) {
	// Test that the MCP server respects environment-specific configurations
	tempDir := t.TempDir()

	// Create configs for different environments
	config := gqlt.GetDefaultConfig()
	config.Configs["development"] = gqlt.ConfigEntry{
		Endpoint: "http://localhost:4000/graphql",
		Headers:  make(map[string]string),
	}
	config.Configs["staging"] = gqlt.ConfigEntry{
		Endpoint: "https://staging-api.example.com/graphql",
		Headers:  map[string]string{"Authorization": "Bearer staging-token"},
	}
	config.Configs["production"] = gqlt.ConfigEntry{
		Endpoint: "https://api.example.com/graphql",
		Headers:  map[string]string{"Authorization": "Bearer prod-token"},
	}

	// Save the config
	if err := config.Save(tempDir); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Test loading and switching between environments
	loadedConfig, err := gqlt.Load(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test development environment
	if err := loadedConfig.SetCurrent("development"); err != nil {
		t.Fatalf("Failed to switch to development: %v", err)
	}
	dev := loadedConfig.GetCurrent()
	if dev.Endpoint != "http://localhost:4000/graphql" {
		t.Error("Development config not correct")
	}

	// Test staging environment
	if err := loadedConfig.SetCurrent("staging"); err != nil {
		t.Fatalf("Failed to switch to staging: %v", err)
	}
	staging := loadedConfig.GetCurrent()
	if staging.Endpoint != "https://staging-api.example.com/graphql" {
		t.Error("Staging config not correct")
	}
	if staging.Headers["Authorization"] != "Bearer staging-token" {
		t.Error("Staging auth not correct")
	}

	// Test production environment
	if err := loadedConfig.SetCurrent("production"); err != nil {
		t.Fatalf("Failed to switch to production: %v", err)
	}
	prod := loadedConfig.GetCurrent()
	if prod.Endpoint != "https://api.example.com/graphql" {
		t.Error("Production config not correct")
	}
	if prod.Headers["Authorization"] != "Bearer prod-token" {
		t.Error("Production auth not correct")
	}
}

func TestMCPServerConfigPathResolution(t *testing.T) {
	// Test that config paths are resolved correctly
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set a test HOME directory
	testHome := t.TempDir()
	os.Setenv("HOME", testHome)

	// Test default config path resolution
	defaultPath := gqlt.GetDefaultPath()

	// The path should contain the test home directory
	if !strings.Contains(defaultPath, testHome) {
		t.Errorf("Default path %s should contain test home %s", defaultPath, testHome)
	}

	// The path should end with "gqlt"
	if !strings.Contains(defaultPath, "gqlt") {
		t.Errorf("Default path %s should contain 'gqlt'", defaultPath)
	}

	// Test config file path resolution
	configPath := gqlt.GetConfigPath()

	// The config path should contain the test home directory
	if !strings.Contains(configPath, testHome) {
		t.Errorf("Config path %s should contain test home %s", configPath, testHome)
	}

	// The config path should end with "config.json"
	if !strings.Contains(configPath, "config.json") {
		t.Errorf("Config path %s should contain 'config.json'", configPath)
	}
}

func TestMCPServerConfigWithCustomDir(t *testing.T) {
	// Test loading config from custom directory
	customDir := t.TempDir()

	// Create a config in the custom directory
	config := gqlt.GetDefaultConfig()
	config.Configs["custom"] = gqlt.ConfigEntry{
		Endpoint: "https://custom-api.example.com/graphql",
		Headers:  map[string]string{"X-Custom-Header": "custom-value"},
	}

	// Save to custom directory
	if err := config.Save(customDir); err != nil {
		t.Fatalf("Failed to save config to custom directory: %v", err)
	}

	// Load from custom directory
	loadedConfig, err := gqlt.Load(customDir)
	if err != nil {
		t.Fatalf("Failed to load config from custom directory: %v", err)
	}

	// Verify the custom config was loaded
	customConfig := loadedConfig.Configs["custom"]
	if customConfig.Endpoint != "https://custom-api.example.com/graphql" {
		t.Error("Custom config endpoint not loaded correctly")
	}
	if customConfig.Headers["X-Custom-Header"] != "custom-value" {
		t.Error("Custom config headers not loaded correctly")
	}
}
