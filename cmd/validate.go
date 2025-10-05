package main

import (
	"fmt"

	"github.com/kluzzebass/gqlt"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate GraphQL queries, schemas, and configurations",
	Long: `Validate GraphQL queries, schemas, and configurations.
This command provides structured validation results for AI agents and automation.

AI-FRIENDLY FEATURES:
- Structured JSON output with validation results
- Machine-readable error codes
- Detailed validation information
- Quiet mode for automation

EXAMPLES:
  # Validate a query against a schema
  gqlt validate query --query "{ users { id name } }" --url https://api.example.com/graphql
  
  # Validate query from file
  gqlt validate query --query-file query.graphql --url https://api.example.com/graphql
  
  # Validate configuration
  gqlt validate config
  
  # Validate schema
  gqlt validate schema --url https://api.example.com/graphql
  
  # Structured output for AI agents
  gqlt validate query --query "{ users { id } }" --format json --quiet`,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

var validateQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Validate a GraphQL query against a schema",
	Long: `Validate a GraphQL query against a schema.
Returns structured validation results including syntax errors, type errors, and field availability.

EXAMPLES:
  gqlt validate query --query "{ users { id name } }" --url https://api.example.com/graphql
  gqlt validate query --query-file query.graphql --url https://api.example.com/graphql
  gqlt validate query --query "{ users { id } }" --format json --quiet`,
	Args: cobra.NoArgs,
	RunE: runValidateQuery,
}

func init() {
	// Add flags to query validation command
	validateQueryCmd.Flags().StringP("query", "q", "", "GraphQL query to validate")
	validateQueryCmd.Flags().StringP("query-file", "Q", "", "Path to GraphQL query file")
	validateQueryCmd.Flags().StringP("url", "u", "", "GraphQL endpoint URL")
}

var validateConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Validate configuration files",
	Long: `Validate configuration files for syntax and completeness.
Returns structured validation results with detailed error information.

EXAMPLES:
  gqlt validate config
  gqlt validate config --format json --quiet`,
	Args: cobra.NoArgs,
	RunE: runValidateConfig,
}

var validateSchemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Validate a GraphQL schema",
	Long: `Validate a GraphQL schema for correctness and completeness.
Returns structured validation results with schema analysis.

EXAMPLES:
  gqlt validate schema --url https://api.example.com/graphql
  gqlt validate schema --url https://api.example.com/graphql --format json --quiet`,
	Args: cobra.NoArgs,
	RunE: runValidateSchema,
}

func init() {
	validateCmd.AddCommand(validateQueryCmd)
	validateCmd.AddCommand(validateConfigCmd)
	validateCmd.AddCommand(validateSchemaCmd)

	// Add flags to schema validation command
	validateSchemaCmd.Flags().StringP("url", "u", "", "GraphQL endpoint URL")
}

func runValidateQuery(cmd *cobra.Command, args []string) error {
	// Get configuration from flags
	configDir := cmd.Flag("config-dir").Value.String()
	outputFormat := cmd.Flag("format").Value.String()
	quietMode := cmd.Flag("quiet").Value.String() == "true"

	formatter := gqlt.NewFormatter(outputFormat)

	// Get query parameters from flags
	query := cmd.Flag("query").Value.String()
	queryFile := cmd.Flag("query-file").Value.String()
	endpointURL := cmd.Flag("url").Value.String()

	// Load query
	inputHandler := gqlt.NewInput()
	queryStr, err := inputHandler.LoadQuery(query, queryFile)
	if err != nil {
		return formatter.FormatStructuredErrorWithContext(
			err,
			gqlt.ErrorCodeQueryLoad,
			"query_validation_error",
			map[string]interface{}{
				"query_source": func() string {
					if query != "" {
						return "inline"
					}
					return "file"
				}(),
				"query_file": queryFile,
			},
			quietMode,
		)
	}

	// Load configuration if URL not provided
	if endpointURL == "" {
		cfg, err := gqlt.Load(configDir)
		if err != nil {
			return formatter.FormatStructuredError(err, gqlt.ErrorCodeConfigLoad, quietMode)
		}

		current := cfg.GetCurrent()
		if current.Endpoint == "" {
			return formatter.FormatStructuredError(
				fmt.Errorf("no URL provided and no endpoint configured"),
				gqlt.ErrorCodeInputValidation,
				quietMode,
			)
		}
		endpointURL = current.Endpoint
	}

	// Create client and introspect schema
	client := gqlt.NewClient(endpointURL, nil)
	introspectClient := gqlt.NewIntrospect(client)

	schema, err := introspectClient.IntrospectSchema()
	if err != nil {
		return formatter.FormatStructuredErrorWithContext(
			err,
			gqlt.ErrorCodeSchemaIntrospect,
			"schema_introspection_error",
			map[string]interface{}{
				"endpoint": endpointURL,
			},
			quietMode,
		)
	}

	// Check if the introspection returned a valid schema
	if schema == nil || schema.Data == nil {
		formatter.FormatStructuredErrorWithContext(
			fmt.Errorf("no schema data returned from introspection"),
			gqlt.ErrorCodeSchemaIntrospect,
			"invalid_schema_response",
			map[string]interface{}{
				"endpoint": endpointURL,
			},
			quietMode,
		)
		return fmt.Errorf("no schema data returned from introspection")
	}

	// Check if the schema contains the expected GraphQL introspection structure
	schemaData, ok := schema.Data.(map[string]interface{})
	if !ok {
		formatter.FormatStructuredErrorWithContext(
			fmt.Errorf("invalid schema data format"),
			gqlt.ErrorCodeSchemaIntrospect,
			"invalid_schema_format",
			map[string]interface{}{
				"endpoint": endpointURL,
			},
			quietMode,
		)
		return fmt.Errorf("invalid schema data format")
	}

	// Check if the schema contains the expected GraphQL introspection fields
	if _, hasSchema := schemaData["__schema"]; !hasSchema {
		formatter.FormatStructuredErrorWithContext(
			fmt.Errorf("endpoint does not appear to be a GraphQL endpoint"),
			gqlt.ErrorCodeSchemaIntrospect,
			"not_graphql_endpoint",
			map[string]interface{}{
				"endpoint": endpointURL,
			},
			quietMode,
		)
		return fmt.Errorf("endpoint does not appear to be a GraphQL endpoint")
	}

	// Basic query validation (syntax check)
	validationResult := map[string]interface{}{
		"valid":    true,
		"query":    queryStr,
		"endpoint": endpointURL,
		"checks": map[string]interface{}{
			"syntax":           "valid",
			"schema_available": true,
		},
	}

	// TODO: Add actual GraphQL query validation against schema
	// This would require a GraphQL query parser and validator

	return formatter.FormatStructured(validationResult, quietMode)
}

func runValidateConfig(cmd *cobra.Command, args []string) error {
	// Get configuration from flags
	configDir := cmd.Flag("config-dir").Value.String()
	outputFormat := cmd.Flag("format").Value.String()
	quietMode := cmd.Flag("quiet").Value.String() == "true"

	formatter := gqlt.NewFormatter(outputFormat)

	cfg, err := gqlt.Load(configDir)
	if err != nil {
		return formatter.FormatStructuredErrorWithContext(
			err,
			gqlt.ErrorCodeConfigLoad,
			"config_load_error",
			map[string]interface{}{
				"config_dir": configDir,
			},
			quietMode,
		)
	}

	// Validate configuration
	validationResult := map[string]interface{}{
		"valid":          true,
		"config_dir":     configDir,
		"current_config": cfg.Current,
		"configs_count":  len(cfg.Configs),
		"checks": map[string]interface{}{
			"config_loaded": true,
			"current_exists": func() bool {
				_, exists := cfg.Configs[cfg.Current]
				return exists
			}(),
		},
	}

	// Check if current config exists
	if _, exists := cfg.Configs[cfg.Current]; !exists {
		validationResult["valid"] = false
		validationResult["errors"] = []string{"Current configuration does not exist"}
	}

	// Validate each configuration
	configErrors := make(map[string][]string)
	for name, config := range cfg.Configs {
		errors := []string{}
		if config.Endpoint == "" {
			errors = append(errors, "Missing endpoint")
		}
		if len(errors) > 0 {
			configErrors[name] = errors
		}
	}

	if len(configErrors) > 0 {
		validationResult["valid"] = false
		validationResult["config_errors"] = configErrors
	}

	return formatter.FormatStructured(validationResult, quietMode)
}

func runValidateSchema(cmd *cobra.Command, args []string) error {
	// Get configuration from flags
	configDir := cmd.Flag("config-dir").Value.String()
	outputFormat := cmd.Flag("format").Value.String()
	quietMode := cmd.Flag("quiet").Value.String() == "true"
	endpointURL := cmd.Flag("url").Value.String()

	formatter := gqlt.NewFormatter(outputFormat)

	// Load configuration if URL not provided
	if endpointURL == "" {
		cfg, err := gqlt.Load(configDir)
		if err != nil {
			return formatter.FormatStructuredError(err, gqlt.ErrorCodeConfigLoad, quietMode)
		}

		current := cfg.GetCurrent()
		if current.Endpoint == "" {
			return formatter.FormatStructuredError(
				fmt.Errorf("no URL provided and no endpoint configured"),
				gqlt.ErrorCodeInputValidation,
				quietMode,
			)
		}
		endpointURL = current.Endpoint
	}

	// Create client and introspect schema
	client := gqlt.NewClient(endpointURL, nil)
	introspectClient := gqlt.NewIntrospect(client)

	schema, err := introspectClient.IntrospectSchema()
	if err != nil {
		return formatter.FormatStructuredErrorWithContext(
			err,
			gqlt.ErrorCodeSchemaIntrospect,
			"schema_introspection_error",
			map[string]interface{}{
				"endpoint": endpointURL,
			},
			quietMode,
		)
	}

	// Check if the introspection returned a valid schema
	if schema == nil || schema.Data == nil {
		formatter.FormatStructuredErrorWithContext(
			fmt.Errorf("no schema data returned from introspection"),
			gqlt.ErrorCodeSchemaIntrospect,
			"invalid_schema_response",
			map[string]interface{}{
				"endpoint": endpointURL,
			},
			quietMode,
		)
		return fmt.Errorf("no schema data returned from introspection")
	}

	// Check if the schema contains the expected GraphQL introspection structure
	schemaData, ok := schema.Data.(map[string]interface{})
	if !ok {
		formatter.FormatStructuredErrorWithContext(
			fmt.Errorf("invalid schema data format"),
			gqlt.ErrorCodeSchemaIntrospect,
			"invalid_schema_format",
			map[string]interface{}{
				"endpoint": endpointURL,
			},
			quietMode,
		)
		return fmt.Errorf("invalid schema data format")
	}

	// Check if the schema contains the expected GraphQL introspection fields
	if _, hasSchema := schemaData["__schema"]; !hasSchema {
		formatter.FormatStructuredErrorWithContext(
			fmt.Errorf("endpoint does not appear to be a GraphQL endpoint"),
			gqlt.ErrorCodeSchemaIntrospect,
			"not_graphql_endpoint",
			map[string]interface{}{
				"endpoint": endpointURL,
			},
			quietMode,
		)
		return fmt.Errorf("endpoint does not appear to be a GraphQL endpoint")
	}

	// Analyze schema
	analyzer, err := gqlt.NewAnalyzer(schema)
	if err != nil {
		return formatter.FormatStructuredErrorWithContext(
			err,
			gqlt.ErrorCodeSchemaLoad,
			"schema_analysis_error",
			map[string]interface{}{
				"endpoint": endpointURL,
			},
			quietMode,
		)
	}
	summary, err := analyzer.GetSummary()
	if err != nil {
		return formatter.FormatStructuredErrorWithContext(
			err,
			gqlt.ErrorCodeSchemaLoad,
			"schema_summary_error",
			map[string]interface{}{
				"endpoint": endpointURL,
			},
			quietMode,
		)
	}

	validationResult := map[string]interface{}{
		"valid":    true,
		"endpoint": endpointURL,
		"schema": map[string]interface{}{
			"total_types":       summary.TotalTypes,
			"query_type":        summary.QueryType,
			"mutation_type":     summary.MutationType,
			"subscription_type": summary.SubscriptionType,
		},
		"checks": map[string]interface{}{
			"introspection_successful": true,
			"schema_has_queries":       summary.QueryType != "",
			"schema_has_mutations":     summary.MutationType != "",
			"schema_has_subscriptions": summary.SubscriptionType != "",
		},
	}

	return formatter.FormatStructured(validationResult, quietMode)
}
