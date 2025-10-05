package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestBasicCommandStructure(t *testing.T) {
	// Test that all main commands exist
	mainCommands := []string{"run", "config", "introspect", "describe"}

	for _, cmdName := range mainCommands {
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == cmdName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected main command '%s' to be registered", cmdName)
		}
	}
}

func TestConfigSubcommands(t *testing.T) {
	// Test that config has all expected subcommands
	expectedSubcommands := []string{"show", "list", "create", "delete", "use", "set", "init", "validate", "describe", "examples", "clone"}

	for _, subCmdName := range expectedSubcommands {
		found := false
		for _, cmd := range configCmd.Commands() {
			if cmd.Name() == subCmdName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected config subcommand '%s' to be registered", subCmdName)
		}
	}
}

func TestCommandHelp(t *testing.T) {
	// Test that commands have proper help text
	commands := []*cobra.Command{runCmd, configCmd, introspectCmd, describeCmd}

	for _, cmd := range commands {
		if cmd.Short == "" {
			t.Errorf("Command '%s' should have a short description", cmd.Name())
		}
		if cmd.Long == "" {
			t.Errorf("Command '%s' should have a long description", cmd.Name())
		}
	}
}

func TestCommandFlags(t *testing.T) {
	// Test that run command has expected flags
	expectedFlags := []string{"url", "query", "query-file", "vars", "vars-file", "header", "file", "files-list", "out", "username", "password", "operation"}

	for _, flagName := range expectedFlags {
		flag := runCmd.Flag(flagName)
		if flag == nil {
			t.Errorf("Expected run command to have flag '%s'", flagName)
		}
	}

	// Test config command persistent flags
	formatFlag := configCmd.PersistentFlags().Lookup("format")
	if formatFlag == nil {
		t.Error("Expected config command to have 'format' persistent flag")
	}
}

func TestCommandExecution(t *testing.T) {
	// Test basic command execution
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "help command",
			args:    []string{"--help"},
			wantErr: false,
		},
		{
			name:    "run help",
			args:    []string{"run", "--help"},
			wantErr: false,
		},
		{
			name:    "config help",
			args:    []string{"config", "--help"},
			wantErr: false,
		},
		{
			name:    "invalid command",
			args:    []string{"invalid-command"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh command for each test
			cmd := &cobra.Command{Use: "gqlt"}
			cmd.AddCommand(runCmd)
			cmd.AddCommand(configCmd)
			cmd.AddCommand(introspectCmd)
			cmd.AddCommand(describeCmd)

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Set args and execute
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			// Check error expectation
			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestConfigInit(t *testing.T) {
	// Test config init with temporary directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	// Set temporary home directory
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create a fresh command
	cmd := &cobra.Command{Use: "gqlt"}
	cmd.AddCommand(configCmd)

	// Execute config init
	cmd.SetArgs([]string{"config", "init"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	// Check that config file was created
	// On macOS, it should be in ~/Library/Application Support/gqlt/
	expectedPath := filepath.Join(tempDir, "Library", "Application Support", "gqlt", "config.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected config file to be created at %s", expectedPath)
	}
}

func TestConfigList(t *testing.T) {
	// Test config list after init
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Initialize config first
	cmd := &cobra.Command{Use: "gqlt"}
	cmd.AddCommand(configCmd)

	cmd.SetArgs([]string{"config", "init"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	// Now test config list
	cmd.SetArgs([]string{"config", "list"})
	err = cmd.Execute()

	if err != nil {
		t.Fatalf("config list failed: %v", err)
	}
}

func TestConfigCreate(t *testing.T) {
	// Test config create
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Initialize config first
	cmd := &cobra.Command{Use: "gqlt"}
	cmd.AddCommand(configCmd)

	cmd.SetArgs([]string{"config", "init"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	// Test config create
	cmd.SetArgs([]string{"config", "create", "test"})
	err = cmd.Execute()

	if err != nil {
		t.Fatalf("config create failed: %v", err)
	}
}

func TestCommandArgsValidation(t *testing.T) {
	// Test command argument validation
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "config create without name",
			args:    []string{"config", "create"},
			wantErr: true,
		},
		{
			name:    "config delete without name",
			args:    []string{"config", "delete"},
			wantErr: true,
		},
		{
			name:    "config use without name",
			args:    []string{"config", "use"},
			wantErr: true,
		},
		{
			name:    "config set without args",
			args:    []string{"config", "set"},
			wantErr: true,
		},
		{
			name:    "config clone without args",
			args:    []string{"config", "clone"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "gqlt"}
			cmd.AddCommand(configCmd)

			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if tt.wantErr && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
