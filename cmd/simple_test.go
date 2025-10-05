package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
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

func TestBasicCommandStructure(t *testing.T) {
	// Test that all main commands exist
	mainCommands := []string{"run", "config", "introspect", "describe", "validate"}

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

func TestValidateSubcommands(t *testing.T) {
	// Test that validate has all expected subcommands
	expectedSubcommands := []string{"query", "config", "schema"}

	// Find validate command
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

func TestConfigSubcommands(t *testing.T) {
	// Test that config has all expected subcommands
	expectedSubcommands := []string{"show", "list", "create", "delete", "use", "set", "set-token", "set-username", "set-password", "init", "validate", "describe", "examples", "clone"}

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

func TestCommandFlags(t *testing.T) {
	// Test that run command has expected flags
	expectedFlags := []string{"url", "query", "query-file", "vars", "vars-file", "header", "file", "files-list", "out", "username", "password", "token", "api-key", "operation"}

	for _, flagName := range expectedFlags {
		flag := runCmd.Flag(flagName)
		if flag == nil {
			t.Errorf("Expected run command to have flag '%s'", flagName)
		}
	}

	// Test that config command no longer has format flag (it's now global)
	formatFlag := configCmd.PersistentFlags().Lookup("format")
	if formatFlag != nil {
		t.Error("Expected config command to NOT have 'format' persistent flag (it's now global)")
	}
}

func TestGlobalFlags(t *testing.T) {
	// Test that root command has expected global flags
	expectedFlags := []string{"config-dir", "use-config", "format", "quiet"}

	for _, flagName := range expectedFlags {
		flag := rootCmd.PersistentFlags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected global flag '%s' to be registered", flagName)
		}
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
			name:    "validate help",
			args:    []string{"validate", "--help"},
			wantErr: false,
		},
		{
			name:    "validate config help",
			args:    []string{"validate", "config", "--help"},
			wantErr: false,
		},
		{
			name:    "validate query help",
			args:    []string{"validate", "query", "--help"},
			wantErr: false,
		},
		{
			name:    "validate schema help",
			args:    []string{"validate", "schema", "--help"},
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
			cmd.AddCommand(validateCmd)

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
	cmd := createTestCommand()

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
	cmd := createTestCommand()

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
		{
			name:    "config set-token without args",
			args:    []string{"config", "set-token"},
			wantErr: true,
		},
		{
			name:    "config set-username without args",
			args:    []string{"config", "set-username"},
			wantErr: true,
		},
		{
			name:    "config set-password without args",
			args:    []string{"config", "set-password"},
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

func TestConfigSetToken(t *testing.T) {
	// Test config set-token functionality
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Initialize config first
	cmd := createTestCommand()

	cmd.SetArgs([]string{"config", "init"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	// Create a test configuration
	cmd.SetArgs([]string{"config", "create", "test"})
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("config create failed: %v", err)
	}

	// Test setting token
	cmd.SetArgs([]string{"config", "set-token", "test", "test-token-123"})
	err = cmd.Execute()

	if err != nil {
		t.Fatalf("config set-token failed: %v", err)
	}

	// Verify the token was set by showing the config with a fresh command
	showCmd := createTestCommand()
	var buf bytes.Buffer
	showCmd.SetOut(&buf)
	showCmd.SetErr(&buf)
	showCmd.SetArgs([]string{"config", "show", "test", "--format", "json"})
	err = showCmd.Execute()
	if err != nil {
		t.Fatalf("config show failed: %v", err)
	}

	// Check that the output contains the Authorization header
	output := buf.String()
	if !strings.Contains(output, "Bearer test-token-123") {
		t.Errorf("Expected config to contain 'Bearer test-token-123', got: %s", output)
	}
}

func TestConfigSetUsername(t *testing.T) {
	// Test config set-username functionality
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Initialize config first
	cmd := createTestCommand()

	cmd.SetArgs([]string{"config", "init"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	// Create a test configuration
	cmd.SetArgs([]string{"config", "create", "test"})
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("config create failed: %v", err)
	}

	// Test setting username
	cmd.SetArgs([]string{"config", "set-username", "test", "testuser"})
	err = cmd.Execute()

	if err != nil {
		t.Fatalf("config set-username failed: %v", err)
	}

	// Verify the username was set by showing the config with a fresh command
	showCmd := createTestCommand()
	var buf bytes.Buffer
	showCmd.SetOut(&buf)
	showCmd.SetErr(&buf)
	showCmd.SetArgs([]string{"config", "show", "test", "--format", "json"})
	err = showCmd.Execute()
	if err != nil {
		t.Fatalf("config show failed: %v", err)
	}

	// Check that the output contains the username
	output := buf.String()
	if !strings.Contains(output, "testuser") {
		t.Errorf("Expected config to contain 'testuser', got: %s", output)
	}
}

func TestConfigSetPassword(t *testing.T) {
	// Test config set-password functionality
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Initialize config first
	cmd := createTestCommand()

	cmd.SetArgs([]string{"config", "init"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	// Create a test configuration
	cmd.SetArgs([]string{"config", "create", "test"})
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("config create failed: %v", err)
	}

	// Test setting password
	cmd.SetArgs([]string{"config", "set-password", "test", "testpass"})
	err = cmd.Execute()

	if err != nil {
		t.Fatalf("config set-password failed: %v", err)
	}

	// Verify the password was set by showing the config with a fresh command
	showCmd := createTestCommand()
	var buf bytes.Buffer
	showCmd.SetOut(&buf)
	showCmd.SetErr(&buf)
	showCmd.SetArgs([]string{"config", "show", "test", "--format", "json"})
	err = showCmd.Execute()
	if err != nil {
		t.Fatalf("config show failed: %v", err)
	}

	// Check that the output contains the password
	output := buf.String()
	if !strings.Contains(output, "testpass") {
		t.Errorf("Expected config to contain 'testpass', got: %s", output)
	}
}

func TestConfigSetTokenNonExistentConfig(t *testing.T) {
	// Test that set-token fails for non-existent configuration
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Initialize config first
	cmd := createTestCommand()

	cmd.SetArgs([]string{"config", "init"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	// Test setting token on non-existent config with a fresh command
	errorCmd := createTestCommand()
	var buf bytes.Buffer
	errorCmd.SetOut(&buf)
	errorCmd.SetErr(&buf)
	errorCmd.SetArgs([]string{"config", "set-token", "nonexistent", "test-token", "--format", "json"})
	err = errorCmd.Execute()

	// The command should return structured error output, not a Go error
	output := buf.String()
	if !strings.Contains(output, "CONFIG_NOT_FOUND") {
		t.Errorf("Expected error output to contain 'CONFIG_NOT_FOUND', got: %s", output)
	}
}
