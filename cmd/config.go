package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kluzzebass/gqlt/internal/config"
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

var configTemplateCmd = &cobra.Command{
	Use:   "template <name>",
	Short: "Create configuration from template",
	Long: `Create a new configuration from a predefined template.

Available templates:
  github     - GitHub GraphQL API
  localhost  - Local development server
  production - Production environment
  staging    - Staging environment
  apollo     - Apollo Server
  hasura     - Hasura GraphQL`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigTemplate,
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
	configCmd.AddCommand(configTemplateCmd)
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
	
	path := getConfigPath()
	if err := cfg.Save(path); err != nil {
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
	
	path := getConfigPath()
	if err := cfg.Save(path); err != nil {
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
	
	path := getConfigPath()
	if err := cfg.Save(path); err != nil {
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
	
	path := getConfigPath()
	if err := cfg.Save(path); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	
	fmt.Printf("Set %s.%s = %s\n", name, key, value)
	return nil
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	cfg := config.GetDefaultConfig()
	
	path := getConfigPath()
	if err := cfg.Save(path); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	
	fmt.Printf("Initialized configuration file at %s\n", path)
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
	schema := config.GetSchema()
	
	fmt.Println("Configuration Schema:")
	fmt.Println()
	fmt.Printf("endpoint: %s\n", schema.Endpoint)
	fmt.Printf("headers: %s\n", schema.Headers)
	fmt.Printf("defaults.out: %s\n", schema.DefaultsOut)
	fmt.Println()
	fmt.Println("Example configuration:")
	
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
	fmt.Println("Configuration Examples:")
	fmt.Println()
	fmt.Println("1. Initialize configuration:")
	fmt.Println("   gqlt config init")
	fmt.Println()
	fmt.Println("2. Create environment-specific configs:")
	fmt.Println("   gqlt config create production")
	fmt.Println("   gqlt config create staging")
	fmt.Println("   gqlt config create local")
	fmt.Println()
	fmt.Println("3. Configure endpoints:")
	fmt.Println("   gqlt config set production endpoint https://api.company.com/graphql")
	fmt.Println("   gqlt config set staging endpoint https://staging-api.company.com/graphql")
	fmt.Println("   gqlt config set local endpoint http://localhost:4000/graphql")
	fmt.Println()
	fmt.Println("4. Configure authentication:")
	fmt.Println("   gqlt config set production headers.Authorization 'Bearer prod-token'")
	fmt.Println("   gqlt config set staging headers.Authorization 'Bearer staging-token'")
	fmt.Println()
	fmt.Println("5. Configure output modes:")
	fmt.Println("   gqlt config set production defaults.out json")
	fmt.Println("   gqlt config set staging defaults.out pretty")
	fmt.Println()
	fmt.Println("6. Switch between configurations:")
	fmt.Println("   gqlt config use production")
	fmt.Println("   gqlt run -q '{ me { name } }'")
	fmt.Println()
	fmt.Println("7. Override specific values:")
	fmt.Println("   gqlt run -q '{ me { name } }' -u https://override.com/graphql")
	
	return nil
}

func runConfigTemplate(cmd *cobra.Command, args []string) error {
	templateName := args[0]
	
	templates := map[string]config.ConfigEntry{
		"github": {
			Endpoint: "https://api.github.com/graphql",
			Headers: map[string]string{
				"Authorization": "Bearer your-github-token",
			},
			Defaults: struct {
				Out string `json:"out"`
			}{
				Out: "pretty",
			},
			Comment: "GitHub GraphQL API configuration",
		},
		"localhost": {
			Endpoint: "http://localhost:4000/graphql",
			Headers:  make(map[string]string),
			Defaults: struct {
				Out string `json:"out"`
			}{
				Out: "pretty",
			},
			Comment: "Local development server",
		},
		"production": {
			Endpoint: "https://api.company.com/graphql",
			Headers: map[string]string{
				"Authorization": "Bearer prod-token",
			},
			Defaults: struct {
				Out string `json:"out"`
			}{
				Out: "json",
			},
			Comment: "Production environment",
		},
		"staging": {
			Endpoint: "https://staging-api.company.com/graphql",
			Headers: map[string]string{
				"Authorization": "Bearer staging-token",
			},
			Defaults: struct {
				Out string `json:"out"`
			}{
				Out: "pretty",
			},
			Comment: "Staging environment",
		},
		"apollo": {
			Endpoint: "http://localhost:4000/graphql",
			Headers:  make(map[string]string),
			Defaults: struct {
				Out string `json:"out"`
			}{
				Out: "pretty",
			},
			Comment: "Apollo Server configuration",
		},
		"hasura": {
			Endpoint: "http://localhost:8080/v1/graphql",
			Headers: map[string]string{
				"X-Hasura-Admin-Secret": "your-admin-secret",
			},
			Defaults: struct {
				Out string `json:"out"`
			}{
				Out: "pretty",
			},
			Comment: "Hasura GraphQL configuration",
		},
	}
	
	template, exists := templates[templateName]
	if !exists {
		return fmt.Errorf("template '%s' not found. Available templates: %s", 
			templateName, strings.Join(getTemplateNames(templates), ", "))
	}
	
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	
	cfg.Configs[templateName] = template
	
	path := getConfigPath()
	if err := cfg.Save(path); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	
	fmt.Printf("Created configuration '%s' from template\n", templateName)
	return nil
}

// Helper functions

func loadConfig() (*config.Config, error) {
	// Get config path from the run command's configPath variable
	// This is a bit of a hack, but it works for now
	return config.Load("")
}

func getConfigPath() string {
	// Default to ./.gqlt/config.json
	return "./.gqlt/config.json"
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

func printTable(entry config.ConfigEntry, name string) error {
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

func printConfigListTable(cfg *config.Config) error {
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

func getTemplateNames(templates map[string]config.ConfigEntry) []string {
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	return names
}
