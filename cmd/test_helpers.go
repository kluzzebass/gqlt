package main

import (
	"io"
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

// suppressOutput redirects stdout/stderr to prevent test pollution during a function call
func suppressOutput(fn func() error) error {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	
	// Open /dev/null for writing
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer devNull.Close()
	
	// Redirect stdout/stderr
	os.Stdout = devNull
	os.Stderr = devNull
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()
	
	return fn()
}

// executeCommand executes a command with output suppression
// Use this instead of cmd.Execute() to prevent test pollution
func executeCommand(cmd *cobra.Command) error {
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	
	return suppressOutput(func() error {
		return cmd.Execute()
	})
}

// executeCommandWithOutput executes a command and suppresses output to prevent pollution
// Returns empty string since output is discarded - tests should validate via error returns
func executeCommandWithOutput(cmd *cobra.Command, args []string) (string, error) {
	cmd.SetArgs(args)
	return "", executeCommand(cmd)
}

// getExpectedConfigPath returns the expected config path for the given temp directory
func getExpectedConfigPath(tempDir string) string {
	return filepath.Join(tempDir, "Library", "Application Support", "gqlt", "config.json")
}
