package main

import (
	"fmt"

	"github.com/kluzzebass/gqlt"
	"github.com/spf13/cobra"
)

var introspectCmd = &cobra.Command{
	Use:   "introspect",
	Short: "Fetch and cache GraphQL schema via introspection",
	Long: `Fetch the GraphQL schema from an endpoint using introspection
and save it to a local cache file for use with other commands.`,
	Example: `# Fetch schema from URL
gqlt introspect --url https://api.example.com/graphql

# Fetch schema with authentication
gqlt introspect --url https://api.example.com/graphql --token "bearer-token"

# Force refresh cached schema
gqlt introspect --refresh

# Show schema summary
gqlt introspect --summary

# Save to specific file
gqlt introspect --output schema.json`,
	RunE: introspect,
}

var (
	introspectRefresh bool
	introspectOut     string
	introspectSummary bool
)

func init() {
	rootCmd.AddCommand(introspectCmd)

	// Define flags
	introspectCmd.Flags().BoolVar(&introspectRefresh, "refresh", false, "ignore cache and fetch fresh schema")
	introspectCmd.Flags().StringVar(&introspectOut, "out", "", "output file path (default is OS-specific)")
	introspectCmd.Flags().BoolVar(&introspectSummary, "summary", false, "show summary instead of saving to file")
}

func introspect(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := gqlt.Load(configDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Merge config with flags
	mergeConfigWithFlags(cfg)

	// Determine output path
	outputPath := introspectOut
	if outputPath == "" {
		// Use config-specific schema path
		if configDir != "" {
			outputPath = gqlt.GetSchemaPathForConfigInDir(cfg.Current, configDir)
		} else {
			outputPath = gqlt.GetSchemaPathForConfig(cfg.Current)
		}
	}

	// Check if cache exists and refresh is not requested
	if !introspectRefresh {
		if gqlt.SchemaExists(outputPath) {
			if introspectSummary {
				return showSchemaSummary(outputPath)
			}
			fmt.Printf("Schema already cached at %s (use --refresh to update)\n", outputPath)
			return nil
		}
	}

	// Get endpoint from config or flag
	endpoint := url
	if endpoint == "" {
		current := cfg.GetCurrent()
		if current.Endpoint == "" {
			return fmt.Errorf("no endpoint specified. Use --url flag or set endpoint in config")
		}
		endpoint = current.Endpoint
	}

	// Create GraphQL client
	client := gqlt.NewClient(endpoint, make(map[string]string))

	// Set authentication if provided
	if username != "" && password != "" {
		client.SetAuth(username, password)
	}

	// Add headers from config
	if current := cfg.GetCurrent(); current.Headers != nil {
		client.SetHeaders(current.Headers)
	}

	// Create introspection client
	introspectClient := gqlt.NewIntrospect(client)

	// Execute introspection query
	result, err := introspectClient.IntrospectSchema()
	if err != nil {
		return fmt.Errorf("failed to fetch schema: %w", err)
	}

	// Show summary if requested
	if introspectSummary {
		return showSchemaSummaryFromResult(result)
	}

	// Save schema to file(s)
	if configDir != "" {
		// Use dual format saving (JSON + GraphQL)
		if err := gqlt.SaveSchemaDual(result, cfg.Current, configDir); err != nil {
			return fmt.Errorf("failed to save schema: %w", err)
		}
		fmt.Printf("Schema saved to %s and %s\n",
			gqlt.GetJSONSchemaPathForConfigInDir(cfg.Current, configDir),
			gqlt.GetGraphQLSchemaPathForConfigInDir(cfg.Current, configDir))
	} else {
		// Use single JSON format for backward compatibility
		if err := gqlt.SaveSchema(result, outputPath); err != nil {
			return fmt.Errorf("failed to save schema: %w", err)
		}
		fmt.Printf("Schema saved to %s\n", outputPath)
	}
	return nil
}

func showSchemaSummary(filePath string) error {
	analyzer, err := gqlt.LoadAnalyzerFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load schema analyzer: %w", err)
	}

	summary, err := analyzer.GetSummary()
	if err != nil {
		return fmt.Errorf("failed to get schema summary: %w", err)
	}

	fmt.Printf("GraphQL Schema Summary:\n")
	fmt.Printf("  Total Types: %d\n", summary.TotalTypes)

	if summary.QueryType != "" {
		fmt.Printf("  Query Type: %s\n", summary.QueryType)
	}

	if summary.MutationType != "" {
		fmt.Printf("  Mutation Type: %s\n", summary.MutationType)
	}

	if summary.SubscriptionType != "" {
		fmt.Printf("  Subscription Type: %s\n", summary.SubscriptionType)
	}

	return nil
}

func showSchemaSummaryFromResult(result *gqlt.Response) error {
	analyzer, err := gqlt.NewAnalyzer(result)
	if err != nil {
		return fmt.Errorf("failed to create schema analyzer: %w", err)
	}

	summary, err := analyzer.GetSummary()
	if err != nil {
		return fmt.Errorf("failed to get schema summary: %w", err)
	}

	fmt.Printf("GraphQL Schema Summary:\n")
	fmt.Printf("  Total Types: %d\n", summary.TotalTypes)

	if summary.QueryType != "" {
		fmt.Printf("  Query Type: %s\n", summary.QueryType)
	}

	if summary.MutationType != "" {
		fmt.Printf("  Mutation Type: %s\n", summary.MutationType)
	}

	if summary.SubscriptionType != "" {
		fmt.Printf("  Subscription Type: %s\n", summary.SubscriptionType)
	}

	return nil
}
