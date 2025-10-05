package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Config represents the main configuration structure
type Config struct {
	Current string                 `json:"current"` // active config name (defaults to "default")
	Configs map[string]ConfigEntry `json:"configs"` // named configurations
}

// ConfigEntry represents a single configuration
type ConfigEntry struct {
	Endpoint string            `json:"endpoint"`
	Headers  map[string]string `json:"headers"`
	Defaults struct {
		Out string `json:"out"`
	} `json:"defaults"`
	Comment string `json:"_comment,omitempty"` // AI-friendly documentation
}

// Schema represents the configuration schema for AI understanding
type Schema struct {
	Endpoint    string `json:"endpoint"`
	Headers     string `json:"headers"`
	DefaultsOut string `json:"defaults.out"`
}

// Load reads a configuration file from the specified path
// If path is empty, searches in standard locations
func Load(path string) (*Config, error) {
	if path == "" {
		// Search in standard locations based on OS
		var locations []string

		switch runtime.GOOS {
		case "darwin":
			// macOS: Application Support
			locations = []string{
				filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "gqlt", "config.json"),
			}
		case "windows":
			// Windows: AppData
			appData := os.Getenv("APPDATA")
			if appData == "" {
				appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
			}
			locations = []string{
				filepath.Join(appData, "gqlt", "config.json"),
			}
		default:
			// Linux/Unix: XDG Base Directory
			locations = []string{
				filepath.Join(os.Getenv("HOME"), ".config", "gqlt", "config.json"),
			}
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

// Save writes the configuration to the specified path
func (c *Config) Save(path string) error {
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

// GetCurrent returns the current active configuration
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
	case "headers.Authorization":
		if entry.Headers == nil {
			entry.Headers = make(map[string]string)
		}
		entry.Headers["Authorization"] = value
	case "headers.X-API-Key":
		if entry.Headers == nil {
			entry.Headers = make(map[string]string)
		}
		entry.Headers["X-API-Key"] = value
	case "defaults.out":
		entry.Defaults.Out = value
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
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
		Endpoint:    "GraphQL endpoint URL (required)",
		Headers:     "HTTP headers to include with requests",
		DefaultsOut: "Default output mode: json|pretty|raw",
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
		Defaults: struct {
			Out string `json:"out"`
		}{
			Out: "json",
		},
		Comment: "Default configuration - used when no specific config is active",
	}
}
