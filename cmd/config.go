package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kluzzebass/gqlt"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage gqlt configuration files",
	Long: `Manage gqlt configuration files with support for multiple named configurations.
This allows you to store different settings for different environments (production, staging, local, etc.).

Examples:
  gqlt config init                    # Initialize config file
  gqlt config list                   # List all configurations
  gqlt config show                   # Show current configuration
  gqlt config create production      # Create new configuration
  gqlt config use production         # Switch to production config
  gqlt config set production endpoint https://api.example.com/graphql
  gqlt config validate               # Validate configuration`,
}

var (
	format string
)

func init() {
	rootCmd.AddCommand(configCmd)

	// Global flags
	configCmd.PersistentFlags().StringVarP(&format, "format", "f", "json", "Output format: json|table|yaml")
}

var configShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show current or named configuration",
	Long:  "Show the current configuration or a specific named configuration.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runConfigShow,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configurations",
	Long:  "List all available configurations with their current status.",
	RunE:  runConfigList,
}

var configCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new configuration",
	Long:  "Create a new named configuration with default values.",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigCreate,
}

var configDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a configuration",
	Long:  "Delete a named configuration (cannot delete default).",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigDelete,
}

var configUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Switch to a configuration",
	Long:  "Switch the current active configuration to the specified name.",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigUse,
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
	RunE: runConfigSet,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long:  "Create a new configuration file with default settings.",
	RunE:  runConfigInit,
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long:  "Check the configuration file for errors and provide suggestions.",
	RunE:  runConfigValidate,
}

var configDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Show configuration schema",
	Long:  "Show the configuration schema and available options.",
	RunE:  runConfigDescribe,
}

var configExamplesCmd = &cobra.Command{
	Use:   "examples",
	Short: "Show usage examples",
	Long:  "Show common usage examples and templates.",
	RunE:  runConfigExamples,
}

var configCloneCmd = &cobra.Command{
	Use:   "clone <source> <target>",
	Short: "Clone an existing configuration",
	Long:  "Create a new configuration by copying an existing one.",
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigClone,
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
	configCmd.AddCommand(configDescribeCmd)
	configCmd.AddCommand(configExamplesCmd)
	configCmd.AddCommand(configCloneCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
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
		return fmt.Errorf("configuration '%s' does not exist", name)
	}

	switch format {
	case "json":
		return printJSON(entry)
	case "table":
		return printTable(entry, name)
	case "yaml":
		return printYAML(entry)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func runConfigList(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	switch format {
	case "json":
		return printJSON(cfg.Configs)
	case "table":
		return printConfigListTable(cfg)
	case "yaml":
		return printYAML(cfg.Configs)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func runConfigCreate(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	name := args[0]
	if err := cfg.Create(name); err != nil {
		return err
	}

	if err := cfg.Save(configDir); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Created configuration '%s'\n", name)
	return nil
}

func runConfigDelete(cmd *cobra.Command, args []string) error {
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

func runConfigUse(cmd *cobra.Command, args []string) error {
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

func runConfigSet(cmd *cobra.Command, args []string) error {
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

func runConfigInit(cmd *cobra.Command, args []string) error {
	cfg := gqlt.GetDefaultConfig()

	if err := cfg.Save(configDir); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	configPath := getConfigPath()
	fmt.Printf("Initialized configuration file at %s\n", configPath)
	return nil
}

func runConfigValidate(cmd *cobra.Command, args []string) error {
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

func runConfigDescribe(cmd *cobra.Command, args []string) error {
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

func runConfigExamples(cmd *cobra.Command, args []string) error {
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
   gqlt config set production headers.Authorization 'Bearer prod-token'
   gqlt config set staging headers.Authorization 'Bearer staging-token'

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

func runConfigClone(cmd *cobra.Command, args []string) error {
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

func printJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func printYAML(v interface{}) error {
	// For now, just print JSON (YAML support can be added later)
	return printJSON(v)
}

func printTable(entry gqlt.ConfigEntry, name string) error {
	fmt.Printf("Configuration: %s\n", name)
	fmt.Printf("  Endpoint: %s\n", entry.Endpoint)
	fmt.Printf("  Headers:\n")
	for k, v := range entry.Headers {
		fmt.Printf("    %s: %s\n", k, v)
	}
	fmt.Printf("  Default Output: %s\n", entry.Defaults.Out)
	if entry.Comment != "" {
		fmt.Printf("  Comment: %s\n", entry.Comment)
	}
	return nil
}

func printConfigListTable(cfg *gqlt.Config) error {
	fmt.Printf("Current: %s\n", cfg.Current)
	fmt.Println("Configurations:")
	for name, entry := range cfg.Configs {
		status := ""
		if name == cfg.Current {
			status = " (current)"
		}
		fmt.Printf("  %s%s\n", name, status)
		fmt.Printf("    Endpoint: %s\n", entry.Endpoint)
		fmt.Printf("    Output: %s\n", entry.Defaults.Out)
	}
	return nil
}
