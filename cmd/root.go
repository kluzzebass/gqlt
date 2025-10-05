package main

import (
	"github.com/spf13/cobra"
)

var configDir string
var configName string

var rootCmd = &cobra.Command{
	Use:   "gqlt",
	Short: "A minimal, composable command-line client for running GraphQL operations",
	Long: `gqlt is a minimal, composable command-line client for running GraphQL operations.
It supports queries, mutations, subscriptions, introspection, and more.`,
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
}
