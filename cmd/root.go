package main

import (
	"github.com/spf13/cobra"
)

var configDir string
var configName string
var outputFormat string
var quietMode bool

var rootCmd = &cobra.Command{
	Use:   "gqlt",
	Short: "A minimal, composable command-line client for running GraphQL operations",
	Long: `gqlt is a minimal, composable command-line client for running GraphQL operations.
It supports queries, mutations, subscriptions, introspection, and more.

AI-FRIENDLY FEATURES:
- Structured JSON output with --format json
- Machine-readable error codes for automation
- Quiet mode (--quiet) for script integration
- Comprehensive help with examples

COMMON PATTERNS:
  # Basic query execution
  gqlt run --url https://api.example.com/graphql --query "{ users { id name } }"
  
  # Using configuration
  gqlt config create production
  gqlt config set production endpoint https://api.example.com/graphql
  gqlt run --query "{ users { id name } }"
  
  # File uploads
  gqlt run --query "mutation($file: Upload!) { uploadFile(file: $file) }" --file avatar=./photo.jpg
  
  # Introspection
  gqlt introspect --url https://api.example.com/graphql
  
  # Schema analysis
  gqlt describe User --url https://api.example.com/graphql

AUTHENTICATION:
  # Bearer token
  gqlt run --token "your-token" --query "{ me { id } }"
  
  # Basic auth
  gqlt run --username user --password pass --query "{ me { id } }"
  
  # API key
  gqlt run --api-key "your-api-key" --query "{ me { id } }"

OUTPUT FORMATS:
  # JSON (default, structured)
  gqlt run --format json --query "{ users { id } }"
  
  # Table format
  gqlt config list --format table
  
  # YAML format
  gqlt config show --format yaml`,
	Version: "0.1.0",
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

// Main is the entry point for the CLI
func main() {
	Execute()
}

func init() {
	// Add global persistent flags
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", "", "config directory (default is OS-specific)")
	rootCmd.PersistentFlags().StringVar(&configName, "use-config", "", "use specific configuration by name (overrides current selection)")
	rootCmd.PersistentFlags().StringVar(&outputFormat, "format", "json", "Output format: json|table|yaml (default: json)")
	rootCmd.PersistentFlags().BoolVarP(&quietMode, "quiet", "", false, "Quiet mode - suppress non-essential output for automation")
}
