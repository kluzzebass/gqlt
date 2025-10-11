package gqlt

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Config represents the main configuration structure that manages multiple named configurations.
// It allows switching between different GraphQL endpoints and their associated settings.
type Config struct {
	Current string                 `json:"current"` // active config name (defaults to "default")
	Configs map[string]ConfigEntry `json:"configs"` // named configurations
}

// ConfigEntry represents a single configuration for a GraphQL endpoint.
// It contains the endpoint URL, headers, authentication credentials, default output format, and optional documentation.
type ConfigEntry struct {
	Endpoint string            `json:"endpoint"` // GraphQL endpoint URL
	Headers  map[string]string `json:"headers"`  // HTTP headers to send with requests
	Auth     struct {
		Token    string `json:"token,omitempty"`    // Bearer token for authentication
		Username string `json:"username,omitempty"` // Username for basic authentication
		Password string `json:"password,omitempty"` // Password for basic authentication
		APIKey   string `json:"api_key,omitempty"`  // API key for authentication
	} `json:"auth"`
	Comment string `json:"_comment,omitempty"` // AI-friendly documentation
}

// Schema represents the configuration schema for AI understanding
type Schema struct {
	Endpoint string `json:"endpoint"`
	Headers  string `json:"headers"`
}

// Path functions
func getDefaultPath() string {
	switch runtime.GOOS {
	case "darwin":
		// macOS: Application Support
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "gqlt")
	case "windows":
		// Windows: AppData
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		return filepath.Join(appData, "gqlt")
	default:
		// Linux/Unix: XDG Base Directory
		return filepath.Join(os.Getenv("HOME"), ".config", "gqlt")
	}
}

func getConfigPath() string {
	return filepath.Join(getDefaultPath(), "config.json")
}

func getSchemaPath() string {
	return filepath.Join(getDefaultPath(), "schema.json")
}

func getSchemaPathForConfig(configName string) string {
	return filepath.Join(getDefaultPath(), "schemas", configName+".json")
}

func getConfigPathForDir(configDir string) string {
	if configDir == "" {
		return getConfigPath()
	}
	return filepath.Join(configDir, "config.json")
}

func getSchemaPathForConfigInDir(configName, configDir string) string {
	return filepath.Join(configDir, "schemas", configName+".json")
}

func getJSONSchemaPathForConfigInDir(configName, configDir string) string {
	return filepath.Join(configDir, "schemas", configName+".json")
}

func getGraphQLSchemaPathForConfigInDir(configName, configDir string) string {
	return filepath.Join(configDir, "schemas", configName+".graphqls")
}

func getSchemasDir() string {
	return filepath.Join(getDefaultPath(), "schemas")
}

// Public path functions
// GetDefaultPath returns the default configuration directory path for the current OS.
// On Linux: ~/.config/gqlt
// On macOS: ~/Library/Application Support/gqlt
// On Windows: %APPDATA%/gqlt
func GetDefaultPath() string {
	return getDefaultPath()
}

// GetConfigPath returns the path to the main configuration file.
// This is typically config.json in the default configuration directory.
func GetConfigPath() string {
	return getConfigPath()
}

// GetSchemaPath returns the path to the default schema file.
// This is typically schema.json in the default configuration directory.
func GetSchemaPath() string {
	return getSchemaPath()
}

// GetSchemaPathForConfig returns the path to the schema file for a specific configuration.
// This is typically schemas/{configName}.json in the default configuration directory.
func GetSchemaPathForConfig(configName string) string {
	return getSchemaPathForConfig(configName)
}

// GetConfigPathForDir returns the path to the configuration file in a specific directory.
func GetConfigPathForDir(configDir string) string {
	return getConfigPathForDir(configDir)
}

// GetSchemaPathForConfigInDir returns the path to the schema file for a specific configuration
// in a specific directory.
func GetSchemaPathForConfigInDir(configName, configDir string) string {
	return getSchemaPathForConfigInDir(configName, configDir)
}

// GetJSONSchemaPathForConfigInDir returns the path to the JSON schema file for a specific
// configuration in a specific directory.
func GetJSONSchemaPathForConfigInDir(configName, configDir string) string {
	return getJSONSchemaPathForConfigInDir(configName, configDir)
}

// GetGraphQLSchemaPathForConfigInDir returns the path to the GraphQL SDL schema file for a
// specific configuration in a specific directory.
func GetGraphQLSchemaPathForConfigInDir(configName, configDir string) string {
	return getGraphQLSchemaPathForConfigInDir(configName, configDir)
}

// GetSchemasDir returns the path to the schemas directory.
// This is typically schemas/ in the default configuration directory.
func GetSchemasDir() string {
	return getSchemasDir()
}

// Load reads a configuration file from the specified config directory.
// If configDir is empty, it searches in standard locations (current directory, then default path).
// Returns a Config struct with the loaded configuration or an error if loading fails.
//
// Example:
//
//	config, err := gqlt.Load("/path/to/config")
//	if err != nil {
//	    log.Fatal(err)
//	}
func Load(configDir string) (*Config, error) {
	var path string
	if configDir == "" {
		// Search in standard locations based on OS
		locations := []string{
			getConfigPath(),
		}

		for _, loc := range locations {
			if _, err := os.Stat(loc); err == nil {
				path = loc
				break
			}
		}

		if path == "" {
			// No config file found, return default config
			return GetDefaultConfig(), nil
		}
	} else {
		// Use config directory
		path = getConfigPathForDir(configDir)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return GetDefaultConfig(), nil
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Ensure default config exists
	if config.Configs == nil {
		config.Configs = make(map[string]ConfigEntry)
	}
	if _, exists := config.Configs["default"]; !exists {
		config.Configs["default"] = getDefaultConfigEntry()
	}
	if config.Current == "" {
		config.Current = "default"
	}

	return &config, nil
}

// Save writes the configuration to the specified config directory.
// Creates the directory if it doesn't exist and writes the configuration as JSON.
//
// Example:
//
//	err := config.Save("/path/to/config")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (c *Config) Save(configDir string) error {
	// Use config directory
	path := getConfigPathForDir(configDir)

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

// GetCurrent returns the current active configuration entry.
// If the current configuration doesn't exist, it falls back to the "default" configuration,
// or creates a default entry if no configurations exist.
//
// Example:
//
//	current := config.GetCurrent()
//	fmt.Printf("Current endpoint: %s\n", current.Endpoint)
func (c *Config) GetCurrent() *ConfigEntry {
	if entry, exists := c.Configs[c.Current]; exists {
		return &entry
	}
	// Fallback to default
	if entry, exists := c.Configs["default"]; exists {
		return &entry
	}
	// Last resort
	defaultEntry := getDefaultConfigEntry()
	return &defaultEntry
}

// GetHeaders returns the HTTP headers for this configuration entry,
// including computed authentication headers based on stored credentials.
func (e *ConfigEntry) GetHeaders() map[string]string {
	headers := make(map[string]string)

	// Copy existing headers
	for k, v := range e.Headers {
		headers[k] = v
	}

	// Compute authentication headers based on stored credentials
	// Priority: Basic Auth > Bearer Token > API Key

	// Basic Authentication (requires both username and password)
	if e.Auth.Username != "" && e.Auth.Password != "" {
		// Only set if Authorization header isn't already set
		if _, exists := headers["Authorization"]; !exists {
			headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(e.Auth.Username+":"+e.Auth.Password))
		}
	}

	// Bearer Token (if no basic auth and token is set)
	if e.Auth.Token != "" {
		if _, exists := headers["Authorization"]; !exists {
			headers["Authorization"] = "Bearer " + e.Auth.Token
		}
	}

	// API Key (if no other auth and API key is set)
	if e.Auth.APIKey != "" {
		if _, exists := headers["X-API-Key"]; !exists {
			headers["X-API-Key"] = e.Auth.APIKey
		}
	}

	return headers
}

// SetCurrent sets the current active configuration
func (c *Config) SetCurrent(name string) error {
	if _, exists := c.Configs[name]; !exists {
		return fmt.Errorf("configuration '%s' does not exist", name)
	}
	c.Current = name
	return nil
}

// Create creates a new configuration entry
func (c *Config) Create(name string) error {
	if _, exists := c.Configs[name]; exists {
		return fmt.Errorf("configuration '%s' already exists", name)
	}
	c.Configs[name] = getDefaultConfigEntry()
	return nil
}

// Delete removes a configuration entry
func (c *Config) Delete(name string) error {
	if name == "default" {
		return fmt.Errorf("cannot delete default configuration")
	}
	if _, exists := c.Configs[name]; !exists {
		return fmt.Errorf("configuration '%s' does not exist", name)
	}
	delete(c.Configs, name)

	// If we deleted the current config, switch to default
	if c.Current == name {
		c.Current = "default"
	}
	return nil
}

// SetValue sets a value in a configuration entry
func (c *Config) SetValue(name, key, value string) error {
	entry, exists := c.Configs[name]
	if !exists {
		return fmt.Errorf("configuration '%s' does not exist", name)
	}

	switch key {
	case "endpoint":
		entry.Endpoint = value
	case "auth.token":
		entry.Auth.Token = value
	case "auth.username":
		entry.Auth.Username = value
	case "auth.password":
		entry.Auth.Password = value
	case "auth.api_key":
		entry.Auth.APIKey = value
	default:
		// Handle headers.<name> pattern
		if strings.HasPrefix(key, "headers.") {
			headerName := strings.TrimPrefix(key, "headers.")
			if entry.Headers == nil {
				entry.Headers = make(map[string]string)
			}
			entry.Headers[headerName] = value
		} else {
			return fmt.Errorf("unknown configuration key: %s", key)
		}
	}

	c.Configs[name] = entry
	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() []string {
	var errors []string

	if c.Current == "" {
		errors = append(errors, "current configuration is not set")
	}

	if _, exists := c.Configs[c.Current]; !exists {
		errors = append(errors, fmt.Sprintf("current configuration '%s' does not exist", c.Current))
	}

	for name, entry := range c.Configs {
		if entry.Endpoint == "" && name != "default" {
			errors = append(errors, fmt.Sprintf("configuration '%s' has no endpoint", name))
		}
	}

	return errors
}

// GetSchema returns the configuration schema for AI understanding
func GetSchema() *Schema {
	return &Schema{
		Endpoint: "GraphQL endpoint URL (required)",
		Headers:  "HTTP headers to include with requests",
	}
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *Config {
	return &Config{
		Current: "default",
		Configs: map[string]ConfigEntry{
			"default": getDefaultConfigEntry(),
		},
	}
}

// getDefaultConfigEntry returns a default configuration entry
func getDefaultConfigEntry() ConfigEntry {
	return ConfigEntry{
		Endpoint: "",
		Headers:  make(map[string]string),
		Comment:  "Default configuration - used when no specific config is active",
	}
}
