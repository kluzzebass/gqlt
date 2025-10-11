package main

import (
	"fmt"

	"github.com/kluzzebass/gqlt"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display the current version of gqlt",
	Long:  `Display the current version of gqlt.`,
	Example: `# Show version
gqlt version

# Use in scripts
VERSION=$(gqlt version)
echo "Using gqlt version: $VERSION"`,
	RunE: versionCommand,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func versionCommand(cmd *cobra.Command, args []string) error {
	// Get version from library (uses VERSION file via go:embed)
	version := gqlt.Version()
	
	// Output just the version string for easy scripting
	fmt.Println(version)
	
	return nil
}

