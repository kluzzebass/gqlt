package main

import (
	"encoding/json"
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
- Quiet mode for automation

COMMON WORKFLOWS:
  # Initialize and setup
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

EXAMPLES:
  # Basic setup
  gqlt config init
  gqlt config create myapi
  gqlt config set myapi endpoint https://api.example.com/graphql
  gqlt config use myapi
  
  # With authentication
  gqlt config set-token myapi "your-bearer-token"
  gqlt config set-username myapi "username"
  gqlt config set-password myapi "password"
  gqlt config set myapi headers '{"X-API-Key": "api-key"}'
  
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
	Args:  cobra.ExactArgs(1),
	RunE:  configCreate,
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
  endpoint              - GraphQL endpoint URL
  headers.Authorization - Authorization header
  headers.X-API-Key     - API key header
  defaults.out          - Default output mode (json|pretty|raw)

Examples:
  gqlt config set production endpoint https://api.example.com/graphql
  gqlt config set production headers.Authorization "Bearer token123"
  gqlt config set production defaults.out pretty`,
	Args: cobra.ExactArgs(3),
	RunE: configSet,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long:  "Create a new configuration file with default settings.",
	RunE:  configInit,
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long:  "Check the configuration file for errors and provide suggestions.",
	RunE:  configValidate,
}

var configDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Show configuration schema",
	Long:  "Show the configuration schema and available options.",
	RunE:  configDescribe,
}

var configExamplesCmd = &cobra.Command{
	Use:   "examples",
	Short: "Show usage examples",
	Long:  "Show common usage examples and templates.",
	RunE:  configExamples,
}

var configCloneCmd = &cobra.Command{
	Use:   "clone <source> <target>",
	Short: "Clone an existing configuration",
	Long:  "Create a new configuration by copying an existing one.",
	Args:  cobra.ExactArgs(2),
	RunE:  configClone,
}

var configSetTokenCmd = &cobra.Command{
	Use:   "set-token <name> <token>",
	Short: "Set bearer token for a configuration",
	Long:  "Set the Authorization header with Bearer token for a named configuration.",
	Args:  cobra.ExactArgs(2),
	RunE:  configSetToken,
}

var configSetUsernameCmd = &cobra.Command{
	Use:   "set-username <name> <username>",
	Short: "Set username for basic authentication",
	Long:  "Set the username for basic authentication (requires password to be set separately).",
	Args:  cobra.ExactArgs(2),
	RunE:  configSetUsername,
}

var configSetPasswordCmd = &cobra.Command{
	Use:   "set-password <name> <password>",
	Short: "Set password for basic authentication",
	Long:  "Set the password for basic authentication (requires username to be set separately).",
	Args:  cobra.ExactArgs(2),
	RunE:  configSetPassword,
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configCreateCmd)
	configCmd.AddCommand(configDeleteCmd)
	configCmd.AddCommand(configUseCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configSetTokenCmd)
	configCmd.AddCommand(configSetUsernameCmd)
	configCmd.AddCommand(configSetPasswordCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configDescribeCmd)
	configCmd.AddCommand(configExamplesCmd)
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

func configDescribe(cmd *cobra.Command, args []string) error {
	schema := gqlt.GetSchema()

	description := fmt.Sprintf(`Configuration Schema:

endpoint: %s
headers: %s
defaults.out: %s

Example configuration:`, schema.Endpoint, schema.Headers, schema.DefaultsOut)

	fmt.Print(description)
	fmt.Println()

	example := map[string]interface{}{
		"current": "default",
		"configs": map[string]interface{}{
			"default": map[string]interface{}{
				"endpoint": "https://api.example.com/graphql",
				"headers": map[string]string{
					"Authorization": "Bearer your-token-here",
				},
				"defaults": map[string]string{
					"out": "pretty",
				},
			},
		},
	}

	jsonData, _ := json.MarshalIndent(example, "", "  ")
	fmt.Println(string(jsonData))

	return nil
}

func configExamples(cmd *cobra.Command, args []string) error {
	examples := `Configuration Examples:

1. Initialize configuration:
   gqlt config init

2. Create environment-specific configs:
   gqlt config create production
   gqlt config create staging
   gqlt config create local

3. Clone existing configurations:
   gqlt config clone production staging
   gqlt config clone default local

4. Configure endpoints:
   gqlt config set production endpoint https://api.company.com/graphql
   gqlt config set staging endpoint https://staging-api.company.com/graphql
   gqlt config set local endpoint http://localhost:4000/graphql

5. Configure authentication:
   gqlt config set-token production "prod-token"
   gqlt config set-token staging "staging-token"
   gqlt config set-username local "admin"
   gqlt config set-password local "secret"

6. Configure output modes:
   gqlt config set production defaults.out json
   gqlt config set staging defaults.out pretty

7. Switch between configurations:
   gqlt config use production
   gqlt run -q '{ me { name } }'

8. Override specific values:
   gqlt run -q '{ me { name } }' -u https://override.com/graphql`

	fmt.Print(examples)
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

func configSetToken(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		quietMode := cmd.Flag("quiet").Value.String() == "true"
		formatter := setupFormatter(cmd)
		return formatter.FormatStructuredError(err, "CONFIG_LOAD_ERROR", quietMode)
	}

	name := args[0]
	token := args[1]

	// Check if configuration exists
	entry, exists := cfg.Configs[name]
	if !exists {
		quietMode := cmd.Flag("quiet").Value.String() == "true"
		formatter := setupFormatter(cmd)
		return formatter.FormatStructuredError(fmt.Errorf("configuration '%s' does not exist", name), "CONFIG_NOT_FOUND", quietMode)
	}

	// Set the Authorization header with Bearer token
	if entry.Headers == nil {
		entry.Headers = make(map[string]string)
	}
	entry.Headers["Authorization"] = "Bearer " + token
	cfg.Configs[name] = entry

	if err := cfg.Save(configDir); err != nil {
		quietMode := cmd.Flag("quiet").Value.String() == "true"
		formatter := setupFormatter(cmd)
		return formatter.FormatStructuredError(fmt.Errorf("failed to save config: %w", err), "CONFIG_SAVE_ERROR", quietMode)
	}

	quietMode := cmd.Flag("quiet").Value.String() == "true"

	if !quietMode {
		fmt.Printf("Set bearer token for configuration '%s'\n", name)
	}

	formatter := setupFormatter(cmd)
	return formatter.FormatStructured(map[string]string{"message": "Bearer token set", "name": name}, quietMode)
}

func configSetUsername(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		quietMode := cmd.Flag("quiet").Value.String() == "true"
		formatter := setupFormatter(cmd)
		return formatter.FormatStructuredError(err, "CONFIG_LOAD_ERROR", quietMode)
	}

	name := args[0]
	username := args[1]

	// Check if configuration exists
	entry, exists := cfg.Configs[name]
	if !exists {
		quietMode := cmd.Flag("quiet").Value.String() == "true"
		formatter := setupFormatter(cmd)
		return formatter.FormatStructuredError(fmt.Errorf("configuration '%s' does not exist", name), "CONFIG_NOT_FOUND", quietMode)
	}

	// Set the username in headers (for basic auth, we'll store both username and password)
	if entry.Headers == nil {
		entry.Headers = make(map[string]string)
	}
	entry.Headers["X-Username"] = username
	cfg.Configs[name] = entry

	if err := cfg.Save(configDir); err != nil {
		quietMode := cmd.Flag("quiet").Value.String() == "true"
		formatter := setupFormatter(cmd)
		return formatter.FormatStructuredError(fmt.Errorf("failed to save config: %w", err), "CONFIG_SAVE_ERROR", quietMode)
	}

	quietMode := cmd.Flag("quiet").Value.String() == "true"

	if !quietMode {
		fmt.Printf("Set username for configuration '%s'\n", name)
	}

	formatter := setupFormatter(cmd)
	return formatter.FormatStructured(map[string]string{"message": "Username set", "name": name}, quietMode)
}

func configSetPassword(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		quietMode := cmd.Flag("quiet").Value.String() == "true"
		formatter := setupFormatter(cmd)
		return formatter.FormatStructuredError(err, "CONFIG_LOAD_ERROR", quietMode)
	}

	name := args[0]
	password := args[1]

	// Check if configuration exists
	entry, exists := cfg.Configs[name]
	if !exists {
		quietMode := cmd.Flag("quiet").Value.String() == "true"
		formatter := setupFormatter(cmd)
		return formatter.FormatStructuredError(fmt.Errorf("configuration '%s' does not exist", name), "CONFIG_NOT_FOUND", quietMode)
	}

	// Set the password in headers (for basic auth, we'll store both username and password)
	if entry.Headers == nil {
		entry.Headers = make(map[string]string)
	}
	entry.Headers["X-Password"] = password
	cfg.Configs[name] = entry

	if err := cfg.Save(configDir); err != nil {
		quietMode := cmd.Flag("quiet").Value.String() == "true"
		formatter := setupFormatter(cmd)
		return formatter.FormatStructuredError(fmt.Errorf("failed to save config: %w", err), "CONFIG_SAVE_ERROR", quietMode)
	}

	quietMode := cmd.Flag("quiet").Value.String() == "true"

	if !quietMode {
		fmt.Printf("Set password for configuration '%s'\n", name)
	}

	formatter := setupFormatter(cmd)
	return formatter.FormatStructured(map[string]string{"message": "Password set", "name": name}, quietMode)
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
