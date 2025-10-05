package cmd

import (
	"github.com/spf13/cobra"
)

var configPath string
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

func init() {
	// Add global persistent flags
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file (default is OS-specific)")
	rootCmd.PersistentFlags().StringVar(&configName, "use-config", "", "use specific configuration by name (overrides current selection)")
}
