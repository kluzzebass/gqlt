package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// createTestCommand creates a properly configured command for testing
func createTestCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "gqlt"}
	cmd.AddCommand(configCmd)
	// Add global flags (same as in root.go)
	cmd.PersistentFlags().String("config-dir", "", "config directory (default is OS-specific)")
	cmd.PersistentFlags().String("use-config", "", "use specific configuration by name (overrides current selection)")
	cmd.PersistentFlags().String("format", "json", "Output format: json|table|yaml (default: json)")
	cmd.PersistentFlags().Bool("quiet", false, "Quiet mode - suppress non-essential output for automation")
	return cmd
}

// createFullTestCommand creates a command with all subcommands for testing
func createFullTestCommand() *cobra.Command {
	cmd := &cobra.Command{Use: "gqlt"}
	// Add global flags (same as in root.go)
	cmd.PersistentFlags().String("config-dir", "", "config directory (default is OS-specific)")
	cmd.PersistentFlags().String("use-config", "", "use specific configuration by name (overrides current selection)")
	cmd.PersistentFlags().String("format", "json", "Output format: json|table|yaml (default: json)")
	cmd.PersistentFlags().Bool("quiet", false, "Quiet mode - suppress non-essential output for automation")

	cmd.AddCommand(runCmd)
	cmd.AddCommand(configCmd)
	cmd.AddCommand(introspectCmd)
	cmd.AddCommand(describeCmd)
	cmd.AddCommand(validateCmd)
	cmd.AddCommand(docsCmd)
	return cmd
}

// setupTestEnvironment sets up a temporary test environment and returns cleanup function
func setupTestEnvironment(t *testing.T) (string, func()) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)

	cleanup := func() {
		os.Setenv("HOME", originalHome)
	}

	return tempDir, cleanup
}

// executeCommandWithOutput executes a command and captures its output
func executeCommandWithOutput(cmd *cobra.Command, args []string) (string, error) {
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

// getExpectedConfigPath returns the expected config path for the given temp directory
func getExpectedConfigPath(tempDir string) string {
	return filepath.Join(tempDir, "Library", "Application Support", "gqlt", "config.json")
}
