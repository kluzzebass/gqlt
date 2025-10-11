package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/kluzzebass/gqlt"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute a GraphQL operation against an endpoint",
	Long: `Execute a GraphQL operation (query or mutation) against a GraphQL endpoint.
You can provide the query inline, from a file, or via stdin.`,
	Example: `# Basic query
gqlt run --url https://api.example.com/graphql --query "{ users { id name } }"

# Query with variables
gqlt run --url https://api.example.com/graphql --query "query($id: ID!) { user(id: $id) { name } }" --vars '{"id": "123"}'

# Query from stdin
echo "{ users { id name } }" | gqlt run --url https://api.example.com/graphql

# Mutation with file upload
gqlt run --url https://api.example.com/graphql --query "mutation($file: Upload!) { uploadFile(file: $file) }" --file avatar=./photo.jpg

# Using configuration
gqlt run --query "{ users { id name } }"  # Uses configured endpoint

# Authentication (precedence: Basic Auth > Bearer Token > API Key)
gqlt run --username user --password pass --query "{ me { id } }"  # Basic auth (highest precedence)
gqlt run --token "bearer-token" --query "{ me { id } }"          # Bearer token
gqlt run --api-key "api-key" --query "{ me { id } }"             # API key (lowest precedence)

# Structured output for AI agents
gqlt run --format json --quiet --query "{ users { id } }"

# Multiple file uploads
gqlt run --query "mutation($files: [Upload!]!) { uploadFiles(files: $files) }" --files-list files.txt`,
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
	token     string
	apiKey    string
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
	runCmd.Flags().StringArrayVarP(&files, "file", "f", []string{}, "File upload (name=path, repeatable, e.g. avatar=./photo.jpg)")
	runCmd.Flags().StringVarP(&filesList, "files-list", "F", "", "File containing list of files to upload (one per line, format: name=path, supports # comments, ~ expansion, and relative paths)")
	runCmd.Flags().StringVarP(&outMode, "out", "O", "json", "Output mode: json|raw")
	runCmd.Flags().StringVarP(&username, "username", "U", "", "Username for basic authentication")
	runCmd.Flags().StringVarP(&password, "password", "p", "", "Password for basic authentication")
	runCmd.Flags().StringVarP(&token, "token", "t", "", "Bearer token for authentication")
	runCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "API key for authentication (sets X-API-Key header)")
}

func runGraphQL(cmd *cobra.Command, args []string) error {
	// Step 7.5: Load configuration
	cfg, err := gqlt.Load(configDir)
	if err != nil {
		formatter := gqlt.NewFormatter(outputFormat)
		return formatter.FormatStructuredError(fmt.Errorf("failed to load config: %w", err), "CONFIG_LOAD_ERROR", quietMode)
	}

	// Merge config with CLI flags
	mergeConfigWithFlags(cfg)

	// Step 8: Input validation
	if query != "" && queryFile != "" {
		formatter := gqlt.NewFormatter(outputFormat)
		return formatter.FormatStructuredError(fmt.Errorf("cannot specify both --query and --query-file"), "INPUT_VALIDATION_ERROR", quietMode)
	}
	if vars != "" && varsFile != "" {
		formatter := gqlt.NewFormatter(outputFormat)
		return formatter.FormatStructuredError(fmt.Errorf("cannot specify both --vars and --vars-file"), "INPUT_VALIDATION_ERROR", quietMode)
	}

	// Step 9: Helper resolution
	inputHandler := gqlt.NewInput()
	queryStr, err := inputHandler.LoadQuery(query, queryFile)
	if err != nil {
		formatter := gqlt.NewFormatter(outputFormat)
		return formatter.FormatStructuredError(fmt.Errorf("failed to load query: %w", err), "QUERY_LOAD_ERROR", quietMode)
	}

	varsMap, err := inputHandler.LoadVariables(vars, varsFile)
	if err != nil {
		formatter := gqlt.NewFormatter(outputFormat)
		return formatter.FormatStructuredError(fmt.Errorf("failed to load variables: %w", err), "VARIABLES_LOAD_ERROR", quietMode)
	}

	headersMap := inputHandler.LoadHeaders(headers)

	// Parse file uploads
	filesMap, err := inputHandler.ParseFiles(files)
	if err != nil {
		formatter := gqlt.NewFormatter(outputFormat)
		return formatter.FormatStructuredError(fmt.Errorf("failed to parse files: %w", err), "FILES_PARSE_ERROR", quietMode)
	}

	// Parse files from list if provided
	if filesList != "" {
		filesFromList, err := inputHandler.ParseFilesFromList(filesList)
		if err != nil {
			formatter := gqlt.NewFormatter(outputFormat)
			return formatter.FormatStructuredError(fmt.Errorf("failed to parse files list: %w", err), "FILES_LIST_PARSE_ERROR", quietMode)
		}

		// Parse the files from list
		filesFromListMap, err := inputHandler.ParseFiles(filesFromList)
		if err != nil {
			formatter := gqlt.NewFormatter(outputFormat)
			return formatter.FormatStructuredError(fmt.Errorf("failed to parse files from list: %w", err), "FILES_LIST_PARSE_ERROR", quietMode)
		}

		// Merge with existing files
		for name, path := range filesFromListMap {
			filesMap[name] = path
		}
	}

	// Step 10: Run GraphQL call
	// Create GraphQL client
	client := gqlt.NewClient(url, headersMap)

	// Set authentication if provided
	if username != "" && password != "" {
		client.SetAuth(username, password)
		if token != "" {
			// Warn that token is being ignored in favor of basic auth
			fmt.Fprintf(os.Stderr, "Warning: Both basic auth and token provided. Using basic auth (token ignored).\n")
		}
		if apiKey != "" {
			// Warn that API key is being ignored in favor of basic auth
			fmt.Fprintf(os.Stderr, "Warning: Both basic auth and API key provided. Using basic auth (API key ignored).\n")
		}
	} else if token != "" {
		// Set Bearer token authentication
		client.SetHeaders(map[string]string{
			"Authorization": "Bearer " + token,
		})
		if apiKey != "" {
			// Warn that API key is being ignored in favor of token auth
			fmt.Fprintf(os.Stderr, "Warning: Both token and API key provided. Using token auth (API key ignored).\n")
		}
	} else if apiKey != "" {
		// Set API key authentication
		client.SetHeaders(map[string]string{
			"X-API-Key": apiKey,
		})
	}

	// Execute GraphQL operation (with or without files)
	var result *gqlt.Response
	if len(filesMap) > 0 {
		// Use multipart/form-data for file uploads
		result, err = client.ExecuteWithFiles(queryStr, varsMap, operation, filesMap)
		if err != nil {
			formatter := gqlt.NewFormatter(outputFormat)
			return formatter.FormatStructuredError(fmt.Errorf("failed to execute GraphQL operation with files: %w", err), "GRAPHQL_EXECUTION_ERROR", quietMode)
		}
	} else {
		// Use regular JSON for operations without files
		result, err = client.Execute(queryStr, varsMap, operation)
		if err != nil {
			formatter := gqlt.NewFormatter(outputFormat)
			return formatter.FormatStructuredError(fmt.Errorf("failed to execute GraphQL operation: %w", err), "GRAPHQL_EXECUTION_ERROR", quietMode)
		}
	}

	// Step 11: Check for GraphQL errors
	if len(result.Errors) > 0 {
		formatter := gqlt.NewFormatter(outputFormat)
		formatter.FormatStructuredError(fmt.Errorf("GraphQL errors occurred"), "GRAPHQL_ERRORS", quietMode)
		os.Exit(2)
	}

	// Step 12: Output formatting
	formatter := gqlt.NewFormatter(outputFormat)

	// Use structured output for non-json formats (table, yaml)
	if outputFormat != "json" {
		// For structured output, include the full response
		responseData := map[string]interface{}{
			"data":   result.Data,
			"errors": result.Errors,
		}
		if result.Extensions != nil {
			responseData["extensions"] = result.Extensions
		}
		return formatter.FormatStructured(responseData, quietMode)
	}

	// Use GraphQL formatting for json format with output modes (json, raw)
	return formatter.FormatResponse(result, outMode)
}

// mergeConfigWithFlags merges configuration values with CLI flags
// CLI flags take precedence over config values
func mergeConfigWithFlags(cfg *gqlt.Config) {
	var current *gqlt.ConfigEntry

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
