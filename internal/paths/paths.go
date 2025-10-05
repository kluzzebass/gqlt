package paths

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetDefaultPath returns the OS-specific default directory for gqlt
func GetDefaultPath() string {
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

// GetConfigPath returns the default config file path
func GetConfigPath() string {
	return filepath.Join(GetDefaultPath(), "config.json")
}

// GetSchemaPath returns the default schema file path
func GetSchemaPath() string {
	return filepath.Join(GetDefaultPath(), "schema.json")
}

// GetSchemaPathForConfig returns the schema file path for a specific configuration
func GetSchemaPathForConfig(configName string) string {
	return filepath.Join(GetDefaultPath(), "schemas", configName+".json")
}

// GetConfigPathForDir returns the config file path for a specific config directory
func GetConfigPathForDir(configDir string) string {
	if configDir == "" {
		return GetConfigPath()
	}
	return filepath.Join(configDir, "config.json")
}

// GetSchemaPathForConfigInDir returns the schema file path for a specific configuration in a config directory
func GetSchemaPathForConfigInDir(configName, configDir string) string {
	return filepath.Join(configDir, "schemas", configName+".json")
}

// GetJSONSchemaPathForConfigInDir returns the JSON schema file path for a specific configuration in a config directory
func GetJSONSchemaPathForConfigInDir(configName, configDir string) string {
	return filepath.Join(configDir, "schemas", configName+".json")
}

// GetGraphQLSchemaPathForConfigInDir returns the GraphQL schema file path for a specific configuration in a config directory
func GetGraphQLSchemaPathForConfigInDir(configName, configDir string) string {
	return filepath.Join(configDir, "schemas", configName+".graphqls")
}

// GetSchemasDir returns the schemas directory
func GetSchemasDir() string {
	return filepath.Join(GetDefaultPath(), "schemas")
}
