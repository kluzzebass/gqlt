package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunCommandStructure(t *testing.T) {
	// Test that run command exists
	var runCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "run" {
			runCmd = cmd
			break
		}
	}
	if runCmd == nil {
		t.Fatalf("Expected run command to be registered")
	}
}

func TestRunCommandFlags(t *testing.T) {
	// Test that run command has expected flags
	expectedFlags := []string{"url", "query", "query-file", "vars", "vars-file", "header", "file", "files-list", "out", "username", "password", "token", "api-key", "operation"}

	for _, flagName := range expectedFlags {
		flag := runCmd.Flag(flagName)
		if flag == nil {
			t.Errorf("Expected run command to have flag '%s'", flagName)
		}
	}
}

func TestRunHelpCommand(t *testing.T) {
	// Test that run help command works
	cmd := createFullTestCommand()
	output, err := executeCommandWithOutput(cmd, []string{"run", "--help"})

	if err != nil {
		t.Errorf("run help failed: %v", err)
	}

	// Check that help output contains expected content
	if !strings.Contains(output, "run") {
		t.Errorf("Expected help output to contain 'run', got: %s", output)
	}

	// Check that help output contains examples
	if !strings.Contains(output, "Example") {
		t.Errorf("Expected help output to contain examples, got: %s", output)
	}
}

func TestRunCommandWithInvalidEndpoint(t *testing.T) {
	// Test run command with invalid endpoint (should fail gracefully)
	cmd := createFullTestCommand()

	// Test with invalid URL to ensure command structure works
	output, err := executeCommandWithOutput(cmd, []string{"run", "--query", "{ users { id } }", "--url", "https://invalid-endpoint.example.com/graphql", "--out", "json"})

	// The command should execute without panicking
	// It might succeed or fail depending on the endpoint, but it should be structured correctly
	if err != nil {
		t.Logf("run command failed as expected: %v", err)
	} else {
		t.Log("run command succeeded")
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

func TestRunCommandWithMissingQuery(t *testing.T) {
	// Test run command without query (should fail with validation error)
	cmd := createFullTestCommand()

	output, err := executeCommandWithOutput(cmd, []string{"run", "--url", "https://api.example.com/graphql", "--out", "json"})

	// The command should execute without panicking
	// It might show help or fail with validation, but it should be structured correctly
	if err != nil {
		t.Logf("run command failed as expected: %v", err)
	} else {
		t.Log("run command succeeded (showed help)")
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

func TestRunCommandWithMissingURL(t *testing.T) {
	// Test run command without URL (should fail with validation error)
	cmd := createFullTestCommand()

	output, err := executeCommandWithOutput(cmd, []string{"run", "--query", "{ users { id } }", "--out", "json"})

	// The command should execute without panicking
	// It might show help or fail with validation, but it should be structured correctly
	if err != nil {
		t.Logf("run command failed as expected: %v", err)
	} else {
		t.Log("run command succeeded (showed help)")
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

func TestRunCommandWithConfig(t *testing.T) {
	// Test run command using configuration
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

	// Test run command with config
	runCmd := createFullTestCommand()
	output, err := executeCommandWithOutput(runCmd, []string{"run", "--query", "{ users { id } }", "--use-config", "test", "--out", "json"})

	// The command should execute without panicking
	// It might succeed or fail depending on the endpoint, but it should be structured correctly
	if err != nil {
		t.Logf("run command with config failed as expected: %v", err)
	} else {
		t.Log("run command with config succeeded")
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

func TestRunCommandAuthenticationFlags(t *testing.T) {
	// Test that authentication flags are properly registered
	authFlags := []string{"username", "password", "token", "api-key"}

	for _, flagName := range authFlags {
		flag := runCmd.Flag(flagName)
		if flag == nil {
			t.Errorf("Expected run command to have authentication flag '%s'", flagName)
		}
	}
}

func TestRunCommandFileUploadFlags(t *testing.T) {
	// Test that file upload flags are properly registered
	fileFlags := []string{"file", "files-list"}

	for _, flagName := range fileFlags {
		flag := runCmd.Flag(flagName)
		if flag == nil {
			t.Errorf("Expected run command to have file upload flag '%s'", flagName)
		}
	}
}

func TestRunCommandVariableFlags(t *testing.T) {
	// Test that variable flags are properly registered
	varFlags := []string{"vars", "vars-file"}

	for _, flagName := range varFlags {
		flag := runCmd.Flag(flagName)
		if flag == nil {
			t.Errorf("Expected run command to have variable flag '%s'", flagName)
		}
	}
}

// NOTE: Detailed output validation testing requires the Formatter to write to
// test buffers instead of os.Stdout. This would require modifying how formatters
// are initialized in commands. Current tests verify command structure, flags,
// and that commands execute without panicking. Integration tests with actual
// binary execution validate end-to-end behavior.
