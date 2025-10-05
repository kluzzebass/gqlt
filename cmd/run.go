package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/kluzzebass/gqlt/internal/config"
	"github.com/kluzzebass/gqlt/internal/graphql"
	"github.com/kluzzebass/gqlt/internal/input"
	"github.com/kluzzebass/gqlt/internal/output"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute a GraphQL operation against an endpoint",
	Long: `Execute a GraphQL operation (query or mutation) against a GraphQL endpoint.
You can provide the query inline, from a file, or via stdin.`,
	RunE: runGraphQL,
}

var (
	url       string
	query     string
	queryFile string
	operation string
	vars      string
	varsFile  string
	headers   []string
	files     []string
	filesList string
	outMode   string
	username  string
	password  string
)

func init() {
	rootCmd.AddCommand(runCmd)

	// Define flags with short options
	runCmd.Flags().StringVarP(&url, "url", "u", "", "GraphQL endpoint URL (required if not in config)")
	runCmd.Flags().StringVarP(&query, "query", "q", "", "Inline GraphQL document")
	runCmd.Flags().StringVarP(&queryFile, "query-file", "Q", "", "Path to .graphql file")
	runCmd.Flags().StringVarP(&operation, "operation", "o", "", "Operation name")
	runCmd.Flags().StringVarP(&vars, "vars", "v", "", "JSON object with variables")
	runCmd.Flags().StringVarP(&varsFile, "vars-file", "V", "", "Path to JSON file with variables")
	runCmd.Flags().StringArrayVarP(&headers, "header", "H", []string{}, "HTTP header (key=value, repeatable)")
	runCmd.Flags().StringArrayVarP(&files, "file", "f", []string{}, "File upload (name=path, repeatable)")
	runCmd.Flags().StringVarP(&filesList, "files-list", "F", "", "File containing list of files to upload")
	runCmd.Flags().StringVarP(&outMode, "out", "O", "json", "Output mode: json|pretty|raw")
	runCmd.Flags().StringVarP(&username, "username", "U", "", "Username for basic authentication")
	runCmd.Flags().StringVarP(&password, "password", "p", "", "Password for basic authentication")
}

func runGraphQL(cmd *cobra.Command, args []string) error {
	// Step 7.5: Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Merge config with CLI flags
	mergeConfigWithFlags(cfg)

	// Step 8: Input validation
	if query != "" && queryFile != "" {
		return fmt.Errorf("cannot specify both --query and --query-file")
	}
	if vars != "" && varsFile != "" {
		return fmt.Errorf("cannot specify both --vars and --vars-file")
	}

	// Step 9: Helper resolution
	queryStr, err := input.LoadQuery(query, queryFile)
	if err != nil {
		return fmt.Errorf("failed to load query: %w", err)
	}

	varsMap, err := input.LoadVariables(vars, varsFile)
	if err != nil {
		return fmt.Errorf("failed to load variables: %w", err)
	}

	headersMap := input.LoadHeaders(headers)
	_, _ = input.ParseFiles(files) // File uploads not yet implemented

	// Step 10: Run GraphQL call
	// Create GraphQL client
	client := graphql.NewClient(url, headersMap)

	// Set authentication if provided
	if username != "" && password != "" {
		client.SetAuth(username, password)
	}

	// Execute GraphQL operation
	result, err := client.Execute(queryStr, varsMap, operation)
	if err != nil {
		return fmt.Errorf("failed to execute GraphQL operation: %w", err)
	}

	// Step 11: Check for GraphQL errors
	if len(result.Errors) > 0 {
		os.Exit(2)
	}

	// Step 12: Output formatting
	formatter := output.NewFormatter(outMode)
	return formatter.Format(result)
}

// mergeConfigWithFlags merges configuration values with CLI flags
// CLI flags take precedence over config values
func mergeConfigWithFlags(cfg *config.Config) {
	var current *config.ConfigEntry

	// Use specific config name if provided, otherwise use current
	if configName != "" {
		if entry, exists := cfg.Configs[configName]; exists {
			current = &entry
		} else {
			// Config name not found, fall back to current
			current = cfg.GetCurrent()
		}
	} else {
		current = cfg.GetCurrent()
	}

	// Only set values from config if CLI flags are not provided
	if url == "" && current.Endpoint != "" {
		url = current.Endpoint
	}

	if outMode == "json" && current.Defaults.Out != "" {
		outMode = current.Defaults.Out
	}

	// Merge headers from config
	for k, v := range current.Headers {
		// Only add if not already specified via CLI
		found := false
		for _, h := range headers {
			if strings.HasPrefix(h, k+"=") {
				found = true
				break
			}
		}
		if !found {
			headers = append(headers, k+"="+v)
		}
	}
}
