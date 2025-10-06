package main

import (
	"os"
	"strings"
	"testing"
)

func TestConfigInit(t *testing.T) {
	// Test config init with temporary directory
	tempDir, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create a fresh command
	cmd := createTestCommand()

	// Execute config init
	cmd.SetArgs([]string{"config", "init"})
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	// Check that config file was created
	expectedPath := getExpectedConfigPath(tempDir)
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected config file to be created at %s", expectedPath)
	}
}

func TestConfigList(t *testing.T) {
	// Test config list after init
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

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
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

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

func TestConfigArgsValidation(t *testing.T) {
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
			cmd := createTestCommand()

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

func TestConfigSet(t *testing.T) {
	// Test config set functionality with the new unified set command
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

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

	// Test setting endpoint
	cmd.SetArgs([]string{"config", "set", "test", "endpoint", "https://api.example.com/graphql"})
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("config set endpoint failed: %v", err)
	}

	// Test setting auth token
	cmd.SetArgs([]string{"config", "set", "test", "auth.token", "test-token-123"})
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("config set auth.token failed: %v", err)
	}

	// Test setting username
	cmd.SetArgs([]string{"config", "set", "test", "auth.username", "testuser"})
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("config set auth.username failed: %v", err)
	}

	// Test setting password
	cmd.SetArgs([]string{"config", "set", "test", "auth.password", "testpass"})
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("config set auth.password failed: %v", err)
	}

	// Verify the settings were applied by showing the config
	showCmd := createTestCommand()
	output, err := executeCommandWithOutput(showCmd, []string{"config", "show", "test", "--format", "json"})
	if err != nil {
		t.Fatalf("config show failed: %v", err)
	}

	// Check that the output contains the expected values
	if !strings.Contains(output, "https://api.example.com/graphql") {
		t.Errorf("Expected config to contain endpoint, got: %s", output)
	}
	if !strings.Contains(output, "test-token-123") {
		t.Errorf("Expected config to contain token, got: %s", output)
	}
	if !strings.Contains(output, "testuser") {
		t.Errorf("Expected config to contain username, got: %s", output)
	}
	if !strings.Contains(output, "testpass") {
		t.Errorf("Expected config to contain password, got: %s", output)
	}
}

func TestConfigSetNonExistentConfig(t *testing.T) {
	// Test that set fails for non-existent configuration
	_, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Initialize config first
	cmd := createTestCommand()

	cmd.SetArgs([]string{"config", "init"})
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("config init failed: %v", err)
	}

	// Test setting value on non-existent config
	errorCmd := createTestCommand()
	output, err := executeCommandWithOutput(errorCmd, []string{"config", "set", "nonexistent", "endpoint", "https://api.example.com/graphql", "--format", "json"})

	// The command should return an error
	if err == nil {
		t.Error("Expected error but got none")
	}

	// Check that the error output contains expected error information
	if !strings.Contains(output, "does not exist") {
		t.Errorf("Expected error output to contain 'does not exist', got: %s", output)
	}
}
