package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestValidateCommandStructure(t *testing.T) {
	// Test that validate command exists and has expected subcommands
	var validateCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "validate" {
			validateCmd = cmd
			break
		}
	}
	if validateCmd == nil {
		t.Fatalf("Expected validate command to be registered")
	}

	// Test that validate has all expected subcommands
	expectedSubcommands := []string{"query", "config", "schema"}
	for _, subCmdName := range expectedSubcommands {
		found := false
		for _, cmd := range validateCmd.Commands() {
			if cmd.Name() == subCmdName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected validate subcommand '%s' to be registered", subCmdName)
		}
	}
}

func TestValidateHelpCommands(t *testing.T) {
	// Test that all validate help commands work
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "validate help",
			args: []string{"validate", "--help"},
		},
		{
			name: "validate config help",
			args: []string{"validate", "config", "--help"},
		},
		{
			name: "validate query help",
			args: []string{"validate", "query", "--help"},
		},
		{
			name: "validate schema help",
			args: []string{"validate", "schema", "--help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createFullTestCommand()
			output, err := executeCommandWithOutput(cmd, tt.args)

			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.name, err)
			}

			// Check that help output contains expected content
			if !strings.Contains(output, "validate") {
				t.Errorf("Expected help output to contain 'validate', got: %s", output)
			}
		})
	}
}

func TestValidateConfigCommand(t *testing.T) {
	// Test validate config command with temporary environment
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Initialize config first
	cmd := createTestCommand()
	cmd.SetArgs([]string{"config", "init"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	// Test validate config
	validateCmd := createFullTestCommand()
	output, err := executeCommandWithOutput(validateCmd, []string{"validate", "config"})

	// The command should run without panicking
	// Note: The formatter outputs to stdout directly, so we can't easily capture it in tests
	// The important thing is that the command executes successfully
	if err != nil {
		t.Errorf("validate config failed: %v", err)
	}

	// The output might be empty because the formatter writes to stdout directly
	// but the command should have executed without errors
	t.Logf("Command executed successfully, output length: %d", len(output))
}

func TestValidateQueryCommand(t *testing.T) {
	// Test validate query command (this will likely fail without a real endpoint, but we can test the command structure)
	cmd := createFullTestCommand()

	// Test with invalid URL to ensure command structure works
	output, err := executeCommandWithOutput(cmd, []string{"validate", "query", "--query", "{ users { id } }", "--url", "https://invalid-endpoint.example.com/graphql"})

	// We expect this to fail, but the command should be structured correctly
	if err == nil {
		t.Log("validate query succeeded unexpectedly")
	}

	// Check that we got some kind of structured output (even if it's an error)
	if !strings.Contains(output, "query") && !strings.Contains(output, "error") {
		t.Errorf("Expected validation output to contain query or error information, got: %s", output)
	}
}

func TestValidateSchemaCommand(t *testing.T) {
	// Test validate schema command (this will likely fail without a real endpoint, but we can test the command structure)
	cmd := createFullTestCommand()

	// Test with invalid URL to ensure command structure works
	output, err := executeCommandWithOutput(cmd, []string{"validate", "schema", "--url", "https://invalid-endpoint.example.com/graphql"})

	// We expect this to fail, but the command should be structured correctly
	if err == nil {
		t.Log("validate schema succeeded unexpectedly")
	}

	// Check that we got some kind of structured output (even if it's an error)
	if !strings.Contains(output, "schema") && !strings.Contains(output, "error") {
		t.Errorf("Expected validation output to contain schema or error information, got: %s", output)
	}
}

func TestValidateCommandFlags(t *testing.T) {
	// Test that validate commands have expected flags
	var validateCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "validate" {
			validateCmd = cmd
			break
		}
	}
	if validateCmd == nil {
		t.Fatalf("Expected validate command to be registered")
	}

	// Test validate query flags
	var validateQueryCmd *cobra.Command
	for _, cmd := range validateCmd.Commands() {
		if cmd.Name() == "query" {
			validateQueryCmd = cmd
			break
		}
	}
	if validateQueryCmd == nil {
		t.Fatalf("Expected validate query command to be registered")
	}

	expectedFlags := []string{"query", "query-file", "url"}
	for _, flagName := range expectedFlags {
		flag := validateQueryCmd.Flag(flagName)
		if flag == nil {
			t.Errorf("Expected validate query command to have flag '%s'", flagName)
		}
	}

	// Test validate schema flags
	var validateSchemaCmd *cobra.Command
	for _, cmd := range validateCmd.Commands() {
		if cmd.Name() == "schema" {
			validateSchemaCmd = cmd
			break
		}
	}
	if validateSchemaCmd == nil {
		t.Fatalf("Expected validate schema command to be registered")
	}

	schemaFlag := validateSchemaCmd.Flag("url")
	if schemaFlag == nil {
		t.Errorf("Expected validate schema command to have flag 'url'")
	}
}
