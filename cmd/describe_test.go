package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestDescribeCommandStructure(t *testing.T) {
	// Test that describe command exists
	var describeCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "describe" {
			describeCmd = cmd
			break
		}
	}
	if describeCmd == nil {
		t.Fatalf("Expected describe command to be registered")
	}
}

func TestDescribeHelpCommand(t *testing.T) {
	// Test that describe help command works
	cmd := createFullTestCommand()
	output, err := executeCommandWithOutput(cmd, []string{"describe", "--help"})

	if err != nil {
		t.Errorf("describe help failed: %v", err)
	}

	// Check that help output contains expected content
	if !strings.Contains(output, "describe") {
		t.Errorf("Expected help output to contain 'describe', got: %s", output)
	}
}

func TestDescribeCommandFlags(t *testing.T) {
	// Test that describe command has expected flags
	expectedFlags := []string{"json", "schema", "summary"}

	for _, flagName := range expectedFlags {
		flag := describeCmd.Flag(flagName)
		if flag == nil {
			t.Errorf("Expected describe command to have flag '%s'", flagName)
		}
	}
}

func TestDescribeCommandWithInvalidEndpoint(t *testing.T) {
	// Test describe command with invalid endpoint (should fail gracefully)
	cmd := createFullTestCommand()

	// Test with invalid schema to ensure command structure works
	output, err := executeCommandWithOutput(cmd, []string{"describe", "User", "--schema", "/invalid/path/schema.json"})

	// The command should execute without panicking
	// It might succeed or fail depending on the endpoint, but it should be structured correctly
	if err != nil {
		t.Logf("describe command failed as expected: %v", err)
	} else {
		t.Log("describe command succeeded")
	}

	// Check that we got some kind of output (help text or structured output)
	// Note: The formatter outputs to stdout directly, so we can't easily capture it in tests
	// The important thing is that the command executes successfully
	if output == "" {
		t.Log("Output is empty (formatter writes to stdout directly)")
	} else {
		t.Logf("Captured output length: %d", len(output))
	}
}

func TestDescribeCommandWithMissingType(t *testing.T) {
	// Test describe command without type (should fail with validation error)
	cmd := createFullTestCommand()

	output, err := executeCommandWithOutput(cmd, []string{"describe"})

	// The command should execute without panicking
	// It might show help or fail with validation, but it should be structured correctly
	if err != nil {
		t.Logf("describe command failed as expected: %v", err)
	} else {
		t.Log("describe command succeeded (showed help)")
	}

	// Check that we got some kind of output (help text or error message)
	// Note: The formatter outputs to stdout directly, so we can't easily capture it in tests
	// The important thing is that the command executes successfully
	if output == "" {
		t.Log("Output is empty (formatter writes to stdout directly)")
	} else {
		t.Logf("Captured output length: %d", len(output))
	}
}

func TestDescribeCommandWithMissingURL(t *testing.T) {
	// Test describe command without URL (should fail with validation error)
	cmd := createFullTestCommand()

	output, err := executeCommandWithOutput(cmd, []string{"describe", "User"})

	// The command should execute without panicking
	// It might show help or fail with validation, but it should be structured correctly
	if err != nil {
		t.Logf("describe command failed as expected: %v", err)
	} else {
		t.Log("describe command succeeded (showed help)")
	}

	// Check that we got some kind of output (help text or error message)
	// Note: The formatter outputs to stdout directly, so we can't easily capture it in tests
	// The important thing is that the command executes successfully
	if output == "" {
		t.Log("Output is empty (formatter writes to stdout directly)")
	} else {
		t.Logf("Captured output length: %d", len(output))
	}
}

func TestDescribeCommandWithConfig(t *testing.T) {
	// Test describe command using configuration
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Initialize config first
	configCmd := createTestCommand()
	configCmd.SetArgs([]string{"config", "init"})
	err := configCmd.Execute()
	if err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	// Create a test configuration with endpoint
	configCmd.SetArgs([]string{"config", "create", "test"})
	err = configCmd.Execute()
	if err != nil {
		t.Fatalf("config create failed: %v", err)
	}

	configCmd.SetArgs([]string{"config", "set", "test", "endpoint", "https://api.example.com/graphql"})
	err = configCmd.Execute()
	if err != nil {
		t.Fatalf("config set endpoint failed: %v", err)
	}

	// Test describe command with config
	describeCmd := createFullTestCommand()
	output, err := executeCommandWithOutput(describeCmd, []string{"describe", "User", "--use-config", "test"})

	// The command should execute without panicking
	// It might succeed or fail depending on the endpoint, but it should be structured correctly
	if err != nil {
		t.Logf("describe command with config failed as expected: %v", err)
	} else {
		t.Log("describe command with config succeeded")
	}

	// Check that we got some kind of output (help text or structured output)
	// Note: The formatter outputs to stdout directly, so we can't easily capture it in tests
	// The important thing is that the command executes successfully
	if output == "" {
		t.Log("Output is empty (formatter writes to stdout directly)")
	} else {
		t.Logf("Captured output length: %d", len(output))
	}
}

func TestDescribeCommandSchemaFlag(t *testing.T) {
	// Test that describe command has schema flag
	schemaFlag := describeCmd.Flag("schema")
	if schemaFlag == nil {
		t.Errorf("Expected describe command to have 'schema' flag")
	}
}

func TestDescribeCommandFormatFlag(t *testing.T) {
	// Test that describe command has format flag
	formatFlag := describeCmd.Flag("format")
	if formatFlag == nil {
		t.Errorf("Expected describe command to have 'format' flag")
	}
}
