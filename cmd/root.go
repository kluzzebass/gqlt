package cmd

import (
	"github.com/spf13/cobra"
)

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
	// Add global persistent flags here if needed
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gqlt/config.json)")
}
