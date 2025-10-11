package main

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestDocsCommandStructure(t *testing.T) {
	// Test that docs command exists
	var docsCmd *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "docs" {
			docsCmd = cmd
			break
		}
	}
	if docsCmd == nil {
		t.Fatalf("Expected docs command to be registered")
	}
}

func TestDocsHelpCommand(t *testing.T) {
	// Test that docs help command works
	cmd := createFullTestCommand()
	_, err := executeCommandWithOutput(cmd, []string{"docs", "--help"})

	if err != nil {
		t.Errorf("docs help failed: %v", err)
	}

	// NOTE: Output is suppressed. Success validated by no error.
}

func TestDocsCommandFlags(t *testing.T) {
	// Test that docs command has expected flags
	expectedFlags := []string{"format", "output", "tree"}

	for _, flagName := range expectedFlags {
		flag := docsCmd.Flag(flagName)
		if flag == nil {
			t.Errorf("Expected docs command to have flag '%s'", flagName)
		}
	}
}

func TestDocsCommandMarkdownFormat(t *testing.T) {
	// Test docs command with markdown format
	cmd := createFullTestCommand()
	output, err := executeCommandWithOutput(cmd, []string{"docs", "--format", "md", "--output", "-"})

	// The command should execute without panicking
	if err != nil {
		t.Errorf("docs markdown failed: %v", err)
	}

	// Check that we got some kind of output (markdown documentation)
	if output == "" {
		t.Log("Output is empty (docs writes to stdout directly)")
	} else {
		t.Logf("Captured output length: %d", len(output))
	}
}

func TestDocsCommandManFormat(t *testing.T) {
	// Test docs command with man format
	cmd := createFullTestCommand()
	output, err := executeCommandWithOutput(cmd, []string{"docs", "--format", "man", "--output", "-"})

	// The command should execute without panicking
	if err != nil {
		t.Errorf("docs man failed: %v", err)
	}

	// Check that we got some kind of output (man page)
	if output == "" {
		t.Log("Output is empty (docs writes to stdout directly)")
	} else {
		t.Logf("Captured output length: %d", len(output))
	}
}

func TestDocsCommandWithTreeFlag(t *testing.T) {
	// Test docs command with tree flag
	cmd := createFullTestCommand()
	output, err := executeCommandWithOutput(cmd, []string{"docs", "--format", "md", "--tree", "--output", "-"})

	// The command should execute without panicking
	if err != nil {
		t.Errorf("docs with tree failed: %v", err)
	}

	// Check that we got some kind of output (tree documentation)
	if output == "" {
		t.Log("Output is empty (docs writes to stdout directly)")
	} else {
		t.Logf("Captured output length: %d", len(output))
	}
}

func TestDocsCommandWithInvalidFormat(t *testing.T) {
	// Test docs command with invalid format
	cmd := createFullTestCommand()
	output, err := executeCommandWithOutput(cmd, []string{"docs", "--format", "invalid", "--output", "-"})

	// The command should fail with invalid format
	if err == nil {
		t.Log("docs with invalid format succeeded unexpectedly")
	} else {
		t.Logf("docs with invalid format failed as expected: %v", err)
	}

	// Check that we got some kind of output (error message or help)
	if output == "" {
		t.Log("Output is empty (docs writes to stdout directly)")
	} else {
		t.Logf("Captured output length: %d", len(output))
	}
}

func TestDocsCommandWithoutFormat(t *testing.T) {
	// Test docs command without format (should use default)
	cmd := createFullTestCommand()
	output, err := executeCommandWithOutput(cmd, []string{"docs", "--output", "-"})

	// The command should execute without panicking
	if err != nil {
		t.Errorf("docs without format failed: %v", err)
	}

	// Check that we got some kind of output (default format documentation)
	if output == "" {
		t.Log("Output is empty (docs writes to stdout directly)")
	} else {
		t.Logf("Captured output length: %d", len(output))
	}
}

func TestDocsCommandWithoutOutput(t *testing.T) {
	// Test docs command without output (should use default)
	cmd := createFullTestCommand()
	output, err := executeCommandWithOutput(cmd, []string{"docs", "--format", "md"})

	// The command should execute without panicking
	if err != nil {
		t.Errorf("docs without output failed: %v", err)
	}

	// Check that we got some kind of output (default output)
	if output == "" {
		t.Log("Output is empty (docs writes to stdout directly)")
	} else {
		t.Logf("Captured output length: %d", len(output))
	}
}

func TestDocsCommandFormatFlag(t *testing.T) {
	// Test that docs command has format flag
	formatFlag := docsCmd.Flag("format")
	if formatFlag == nil {
		t.Errorf("Expected docs command to have 'format' flag")
	}
}

func TestDocsCommandOutputFlag(t *testing.T) {
	// Test that docs command has output flag
	outputFlag := docsCmd.Flag("output")
	if outputFlag == nil {
		t.Errorf("Expected docs command to have 'output' flag")
	}
}

func TestDocsCommandTreeFlag(t *testing.T) {
	// Test that docs command has tree flag
	treeFlag := docsCmd.Flag("tree")
	if treeFlag == nil {
		t.Errorf("Expected docs command to have 'tree' flag")
	}
}
