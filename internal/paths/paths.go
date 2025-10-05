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
