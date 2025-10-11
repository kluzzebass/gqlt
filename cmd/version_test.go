package main

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestVersionCommandStructure(t *testing.T) {
	// Test that version command exists
	var versionCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "version" {
			versionCmd = cmd
			break
		}
	}
	if versionCmd == nil {
		t.Fatalf("Expected version command to be registered")
	}
}

func TestVersionCommand(t *testing.T) {
	// Test that version command executes without error
	cmd := createFullTestCommand()
	_, err := executeCommandWithOutput(cmd, []string{"version"})

	if err != nil {
		t.Errorf("version command failed: %v", err)
	}

	// NOTE: Output is suppressed to prevent test pollution.
	// Success is validated by command completing without error.
	// The version command outputs the version from VERSION file.
}

func TestVersionHelpCommand(t *testing.T) {
	// Test that version help command works
	cmd := createFullTestCommand()
	_, err := executeCommandWithOutput(cmd, []string{"version", "--help"})

	if err != nil {
		t.Errorf("version help failed: %v", err)
	}

	// NOTE: Output is suppressed. Success validated by no error.
}

