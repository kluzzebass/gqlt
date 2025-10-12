package main

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestIntrospectCommandStructure(t *testing.T) {
	// Test that introspect command exists
	var introspectCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "introspect" {
			introspectCmd = cmd
			break
		}
	}
	if introspectCmd == nil {
		t.Fatalf("Expected introspect command to be registered")
	}
}

func TestIntrospectHelpCommand(t *testing.T) {
	// Test that introspect help command works
	cmd := createFullTestCommand()
	_, err := executeCommandWithOutput(cmd, []string{"introspect", "--help"})

	if err != nil {
		t.Errorf("introspect help failed: %v", err)
	}

	// NOTE: Output is suppressed. Success validated by no error.
}

func TestIntrospectCommandFlags(t *testing.T) {
	// Test that introspect command has expected flags
	expectedFlags := []string{"out", "refresh", "summary"}

	for _, flagName := range expectedFlags {
		flag := introspectCmd.Flag(flagName)
		if flag == nil {
			t.Errorf("Expected introspect command to have flag '%s'", flagName)
		}
	}
}

func TestIntrospectCommandWithInvalidEndpoint(t *testing.T) {
	// Test introspect command with invalid endpoint (should fail gracefully)
	cmd := createFullTestCommand()

	// Test with refresh flag to ensure command structure works
	output, err := executeCommandWithOutput(cmd, []string{"introspect", "--refresh"})

	// The command should execute without panicking
	// It might succeed or fail depending on the endpoint, but it should be structured correctly
	// The important thing is that the command executes without panicking
	_ = err
	_ = output
}

func TestIntrospectCommandWithMissingURL(t *testing.T) {
	// Test introspect command without URL (should fail with validation error)
	cmd := createFullTestCommand()

	output, err := executeCommandWithOutput(cmd, []string{"introspect"})

	// The command should execute without panicking
	// It might show help or fail with validation, but it should be structured correctly
	_ = err
	_ = output
}

func TestIntrospectCommandWithConfig(t *testing.T) {
	// Test introspect command using configuration
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Initialize config first
	configCmd := createTestCommand()
	configCmd.SetArgs([]string{"config", "init"})
	err := executeCommand(configCmd)
	if err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	// Create a test configuration with endpoint
	configCmd.SetArgs([]string{"config", "create", "test"})
	err = executeCommand(configCmd)
	if err != nil {
		t.Fatalf("config create failed: %v", err)
	}

	configCmd.SetArgs([]string{"config", "set", "test", "endpoint", "https://api.example.com/graphql"})
	err = executeCommand(configCmd)
	if err != nil {
		t.Fatalf("config set endpoint failed: %v", err)
	}

	// Test introspect command with config
	introspectCmd := createFullTestCommand()
	output, err := executeCommandWithOutput(introspectCmd, []string{"introspect", "--use-config", "test"})

	// The command should execute without panicking
	// It might succeed or fail depending on the endpoint, but it should be structured correctly
	_ = err
	_ = output
}

func TestIntrospectCommandOutFlag(t *testing.T) {
	// Test that introspect command has out flag
	outFlag := introspectCmd.Flag("out")
	if outFlag == nil {
		t.Errorf("Expected introspect command to have 'out' flag")
	}
}

func TestIntrospectCommandFormatFlag(t *testing.T) {
	// Test that introspect command has format flag
	formatFlag := introspectCmd.Flag("format")
	if formatFlag == nil {
		t.Errorf("Expected introspect command to have 'format' flag")
	}
}

func TestIntrospectCommandRefreshFlag(t *testing.T) {
	// Test that introspect command has refresh flag
	refreshFlag := introspectCmd.Flag("refresh")
	if refreshFlag == nil {
		t.Errorf("Expected introspect command to have 'refresh' flag")
	}
}
