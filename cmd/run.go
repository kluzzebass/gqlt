package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
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
	url        string
	query      string
	queryFile  string
	operation  string
	vars       string
	varsFile   string
	headers    []string
	files      []string
	filesList  string
	configPath string
	outMode    string
	username   string
	password   string
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
	runCmd.Flags().StringVarP(&configPath, "config", "c", "", "Config file path (optional)")
	runCmd.Flags().StringVarP(&outMode, "out", "O", "json", "Output mode: json|pretty|raw")
	runCmd.Flags().StringVarP(&username, "username", "U", "", "Username for basic authentication")
	runCmd.Flags().StringVarP(&password, "password", "p", "", "Password for basic authentication")
}

func runGraphQL(cmd *cobra.Command, args []string) error {
	// Step 8: Input validation
	if query != "" && queryFile != "" {
		return fmt.Errorf("cannot specify both --query and --query-file")
	}
	if vars != "" && varsFile != "" {
		return fmt.Errorf("cannot specify both --vars and --vars-file")
	}

	// Step 9: Helper resolution
	queryStr, err := loadQuery()
	if err != nil {
		return fmt.Errorf("failed to load query: %w", err)
	}

	varsMap, err := loadVars()
	if err != nil {
		return fmt.Errorf("failed to load variables: %w", err)
	}

	headersMap := parseHeaders()
	_ = parseFiles() // File uploads not yet implemented

	// Step 10: Run GraphQL call
	// Create HTTP client with authentication
	var httpClient *http.Client

	if username != "" && password != "" {
		// Basic authentication
		httpClient = &http.Client{
			Transport: &basicAuthTransport{
				username: username,
				password: password,
			},
		}
	} else {
		// No authentication - bearer tokens handled via headers
		httpClient = &http.Client{}
	}

	// Build GraphQL request payload
	payload := map[string]interface{}{
		"query": queryStr,
	}

	if operation != "" {
		payload["operationName"] = operation
	}

	if len(varsMap) > 0 {
		payload["variables"] = varsMap
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal GraphQL request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headersMap {
		req.Header.Set(k, v)
	}

	// Execute request
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute GraphQL request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	// Step 11: Error handling
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Check for GraphQL errors
	if errors, ok := result["errors"]; ok && errors != nil {
		if errorList, ok := errors.([]interface{}); ok && len(errorList) > 0 {
			os.Exit(2)
		}
	}

	// Step 12: Output handler
	return handleOutput(result, outMode)
}

// Step 9: Helper functions
func loadQuery() (string, error) {
	if query != "" {
		return query, nil
	}
	if queryFile != "" {
		content, err := os.ReadFile(queryFile)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}
	// Read from stdin
	content, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func loadVars() (map[string]interface{}, error) {
	if vars != "" {
		var result map[string]interface{}
		err := json.Unmarshal([]byte(vars), &result)
		return result, err
	}
	if varsFile != "" {
		content, err := os.ReadFile(varsFile)
		if err != nil {
			return nil, err
		}
		var result map[string]interface{}
		err = json.Unmarshal(content, &result)
		return result, err
	}
	return make(map[string]interface{}), nil
}

func parseHeaders() map[string]string {
	result := make(map[string]string)
	for _, header := range headers {
		parts := strings.SplitN(header, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func parseFiles() map[string]string {
	result := make(map[string]string)
	for _, file := range files {
		parts := strings.SplitN(file, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func handleOutput(result map[string]interface{}, mode string) error {
	switch mode {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	case "pretty":
		return printPretty(result)
	case "raw":
		// Raw = unformatted JSON (no indentation)
		encoder := json.NewEncoder(os.Stdout)
		return encoder.Encode(result)
	default:
		return fmt.Errorf("invalid output mode: %s", mode)
	}
}

func printPretty(result map[string]interface{}) error {
	// Color scheme
	dataColor := color.New(color.FgGreen, color.Bold)
	errorColor := color.New(color.FgRed, color.Bold)
	keyColor := color.New(color.FgBlue, color.Bold)
	stringColor := color.New(color.FgYellow)
	numberColor := color.New(color.FgCyan)
	boolColor := color.New(color.FgMagenta)
	nullColor := color.New(color.FgHiBlack)

	// Print with colors
	fmt.Print("{\n")

	// Handle data
	if data, ok := result["data"]; ok {
		dataColor.Print("  \"data\": ")
		if err := printValue(data, "  ", keyColor, stringColor, numberColor, boolColor, nullColor); err != nil {
			return err
		}
		fmt.Print(",\n")
	}

	// Handle errors
	if errors, ok := result["errors"]; ok && errors != nil {
		errorColor.Print("  \"errors\": ")
		if err := printValue(errors, "  ", keyColor, stringColor, numberColor, boolColor, nullColor); err != nil {
			return err
		}
		fmt.Print(",\n")
	}

	// Handle extensions
	if extensions, ok := result["extensions"]; ok && extensions != nil {
		keyColor.Print("  \"extensions\": ")
		if err := printValue(extensions, "  ", keyColor, stringColor, numberColor, boolColor, nullColor); err != nil {
			return err
		}
		fmt.Print(",\n")
	}

	fmt.Print("}\n")
	return nil
}

func printValue(value interface{}, indent string, keyColor, stringColor, numberColor, boolColor, nullColor *color.Color) error {
	switch v := value.(type) {
	case map[string]interface{}:
		fmt.Print("{\n")
		first := true
		for k, val := range v {
			if !first {
				fmt.Print(",\n")
			}
			first = false
			keyColor.Printf("%s  \"%s\": ", indent, k)
			if err := printValue(val, indent+"  ", keyColor, stringColor, numberColor, boolColor, nullColor); err != nil {
				return err
			}
		}
		fmt.Printf("\n%s}", indent)
	case []interface{}:
		fmt.Print("[\n")
		for i, val := range v {
			if i > 0 {
				fmt.Print(",\n")
			}
			fmt.Printf("%s  ", indent)
			if err := printValue(val, indent+"  ", keyColor, stringColor, numberColor, boolColor, nullColor); err != nil {
				return err
			}
		}
		fmt.Printf("\n%s]", indent)
	case string:
		stringColor.Printf("\"%s\"", v)
	case float64:
		numberColor.Print(v)
	case bool:
		boolColor.Print(v)
	case nil:
		nullColor.Print("null")
	default:
		// Fallback to JSON encoding for unknown types
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return err
		}
		fmt.Print(string(jsonBytes))
	}
	return nil
}

// basicAuthTransport implements http.RoundTripper for basic authentication
type basicAuthTransport struct {
	username string
	password string
}

func (t *basicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	auth := base64.StdEncoding.EncodeToString([]byte(t.username + ":" + t.password))
	req.Header.Set("Authorization", "Basic "+auth)
	return http.DefaultTransport.RoundTrip(req)
}
