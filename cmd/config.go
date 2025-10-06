package main

import (
	"fmt"

	"github.com/kluzzebass/gqlt"
	"github.com/spf13/cobra"
)

// setupFormatter creates a formatter with the command's output writers
func setupFormatter(cmd *cobra.Command) gqlt.Formatter {
	outputFormat := cmd.Flag("format").Value.String()
	formatter := gqlt.NewFormatter(outputFormat)
	formatter.SetOutput(cmd.OutOrStdout())
	formatter.SetErrorOutput(cmd.ErrOrStderr())
	return formatter
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage gqlt configuration files",
	Long: `Manage gqlt configuration files with support for multiple named configurations.
This allows you to store different settings for different environments (production, staging, local, etc.).

AI-FRIENDLY FEATURES:
- Structured output with --format json|table|yaml
- Machine-readable error codes
- Quiet mode for automation`,
	Example: `# Initialize and setup
gqlt config init
gqlt config create production
gqlt config set production endpoint https://api.prod.com/graphql
gqlt config set production headers '{"Authorization": "Bearer token"}'
gqlt config use production

# Environment management
gqlt config create staging
gqlt config set staging endpoint https://api.staging.com/graphql
gqlt config create local
gqlt config set local endpoint http://localhost:4000/graphql

# Configuration inspection
gqlt config list --format table
gqlt config show production --format json
gqlt config validate

# With authentication
gqlt config set myapi auth.token "your-bearer-token"
gqlt config set myapi auth.username "username"
gqlt config set myapi auth.password "password"
gqlt config set myapi auth.api_key "api-key"

# Clone configuration
gqlt config clone production staging

# Structured output for AI agents
gqlt config list --format json --quiet
gqlt config show --format yaml`,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

var configShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show current or named configuration",
	Long:  "Show the current configuration or a specific named configuration.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  configShow,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configurations",
	Long:  "List all available configurations with their current status.",
	RunE:  configList,
}

var configCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new configuration",
	Long:  "Create a new named configuration with default values.",
	Example: `gqlt config create production
gqlt config create staging
gqlt config create local`,
	Args: cobra.ExactArgs(1),
	RunE: configCreate,
}

var configDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a configuration",
	Long:  "Delete a named configuration (cannot delete default).",
	Args:  cobra.ExactArgs(1),
	RunE:  configDelete,
}

var configUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch to a configuration",
	Long:  "Switch the current active configuration to the specified name.",
	Args:  cobra.ExactArgs(1),
	RunE:  configUse,
}

var configSetCmd = &cobra.Command{
	Use:   "set <name> <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value for a named configuration.

Available keys:
  endpoint                    - GraphQL endpoint URL (required)
  headers.<name>              - Custom HTTP header (e.g., headers.X-Custom "value")
  headers.Authorization       - Authorization header (e.g., "Bearer token")
  headers.X-API-Key           - API key header
  auth.token                  - Bearer token for authentication
  auth.username               - Username for basic authentication
  auth.password               - Password for basic authentication
  auth.api_key                - API key for authentication
  defaults.out                - Default output mode (json|pretty|raw)

Authentication precedence:
  1. Basic auth (auth.username + auth.password)
  2. Bearer token (auth.token)
  3. API key (auth.api_key)
  4. Custom headers (headers.Authorization, headers.X-API-Key)`,
	Example: `# Basic configuration
gqlt config set production endpoint https://api.example.com/graphql
gqlt config set production defaults.out pretty

# Authentication methods
gqlt config set production auth.token "your-bearer-token"
gqlt config set production auth.username "admin"
gqlt config set production auth.password "secret"
gqlt config set production auth.api_key "api-key-123"

# Custom headers
gqlt config set production headers.X-Custom "custom-value"
gqlt config set production headers.Authorization "Bearer manual-token"
gqlt config set production headers.X-API-Key "manual-api-key"`,
	Args: cobra.ExactArgs(3),
	RunE: configSet,
}

var configInitCmd = &cobra.Command{
	Use:     "init",
	Short:   "Initialize configuration file",
	Long:    "Create a new configuration file with default settings.",
	Example: `gqlt config init`,
	RunE:    configInit,
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long:  "Check the configuration file for errors and provide suggestions.",
	RunE:  configValidate,
}

var configCloneCmd = &cobra.Command{
	Use:   "clone <source> <target>",
	Short: "Clone an existing configuration",
	Long:  "Create a new configuration by copying an existing one.",
	Args:  cobra.ExactArgs(2),
	RunE:  configClone,
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configCreateCmd)
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configUseCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configCloneCmd)
}

func configShow(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		quietMode := cmd.Flag("quiet").Value.String() == "true"
		formatter := setupFormatter(cmd)
		return formatter.FormatStructuredError(err, "CONFIG_LOAD_ERROR", quietMode)
	}

	name := "current"
	if len(args) > 0 {
		name = args[0]
	}

	if name == "current" {
		name = cfg.Current
	}

	entry, exists := cfg.Configs[name]
	if !exists {
		quietMode := cmd.Flag("quiet").Value.String() == "true"
		formatter := setupFormatter(cmd)
		return formatter.FormatStructuredError(fmt.Errorf("configuration '%s' does not exist", name), "CONFIG_NOT_FOUND", quietMode)
	}

	quietMode := cmd.Flag("quiet").Value.String() == "true"
	formatter := setupFormatter(cmd)
	return formatter.FormatStructured(entry, quietMode)
}

func configList(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		quietMode := cmd.Flag("quiet").Value.String() == "true"
		formatter := setupFormatter(cmd)
		return formatter.FormatStructuredError(err, "CONFIG_LOAD_ERROR", quietMode)
	}

	quietMode := cmd.Flag("quiet").Value.String() == "true"
	formatter := setupFormatter(cmd)
	return formatter.FormatStructured(cfg.Configs, quietMode)
}

func configCreate(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		quietMode := cmd.Flag("quiet").Value.String() == "true"
		formatter := setupFormatter(cmd)
		return formatter.FormatStructuredError(err, "CONFIG_LOAD_ERROR", quietMode)
	}

	name := args[0]
	if err := cfg.Create(name); err != nil {
		quietMode := cmd.Flag("quiet").Value.String() == "true"
		formatter := setupFormatter(cmd)
		return formatter.FormatStructuredError(err, "CONFIG_CREATE_ERROR", quietMode)
	}

	if err := cfg.Save(configDir); err != nil {
		quietMode := cmd.Flag("quiet").Value.String() == "true"
		formatter := setupFormatter(cmd)
		return formatter.FormatStructuredError(fmt.Errorf("failed to save config: %w", err), "CONFIG_SAVE_ERROR", quietMode)
	}

	quietMode := cmd.Flag("quiet").Value.String() == "true"

	if !quietMode {
		fmt.Printf("Created configuration '%s'\n", name)
	}

	formatter := setupFormatter(cmd)
	return formatter.FormatStructured(map[string]string{"message": "Configuration created", "name": name}, quietMode)
}

func configDelete(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	name := args[0]
	if err := cfg.Delete(name); err != nil {
		return err
	}

	if err := cfg.Save(configDir); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Deleted configuration '%s'\n", name)
	return nil
}

func configUse(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	name := args[0]
	if err := cfg.SetCurrent(name); err != nil {
		return err
	}

	if err := cfg.Save(configDir); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Switched to configuration '%s'\n", name)
	return nil
}

func configSet(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	name := args[0]
	key := args[1]
	value := args[2]

	if err := cfg.SetValue(name, key, value); err != nil {
		return err
	}

	if err := cfg.Save(configDir); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Set %s.%s = %s\n", name, key, value)
	return nil
}

func configInit(cmd *cobra.Command, args []string) error {
	cfg := gqlt.GetDefaultConfig()

	if err := cfg.Save(configDir); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	configPath := getConfigPath()
	fmt.Printf("Initialized configuration file at %s\n", configPath)
	return nil
}

func configValidate(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	errors := cfg.Validate()
	if len(errors) == 0 {
		fmt.Println("Configuration is valid")
		return nil
	}

	fmt.Println("Configuration has errors:")
	for _, err := range errors {
		fmt.Printf("  - %s\n", err)
	}
	return nil
}

func configClone(cmd *cobra.Command, args []string) error {
	sourceName := args[0]
	targetName := args[1]

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Check if source exists
	sourceConfig, exists := cfg.Configs[sourceName]
	if !exists {
		return fmt.Errorf("source configuration '%s' does not exist", sourceName)
	}

	// Check if target already exists
	if _, exists := cfg.Configs[targetName]; exists {
		return fmt.Errorf("target configuration '%s' already exists", targetName)
	}

	// Clone the configuration
	cfg.Configs[targetName] = sourceConfig

	// Save the updated config
	if err := cfg.Save(configDir); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Cloned configuration '%s' to '%s'\n", sourceName, targetName)
	return nil
}

// Helper functions

func loadConfig() (*gqlt.Config, error) {
	// Use the global configDir variable
	return gqlt.Load(configDir)
}

func getConfigPath() string {
	// Use global configDir if set, otherwise default to OS-specific path
	if configDir != "" {
		return gqlt.GetConfigPathForDir(configDir)
	}

	// Use the shared default path function
	return gqlt.GetConfigPath()
}
