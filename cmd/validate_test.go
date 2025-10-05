package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kluzzebass/gqlt"
)

func TestValidateConfigLogic(t *testing.T) {
	tests := []struct {
		name         string
		setupConfig  func(tempDir string) error
		expectValid  bool
		expectErrors bool
	}{
		{
			name: "default config with no endpoint",
			setupConfig: func(tempDir string) error {
				// Default config is created automatically
				return nil
			},
			expectValid:  false,
			expectErrors: true,
		},
		{
			name: "config with valid endpoint",
			setupConfig: func(tempDir string) error {
				cfg, err := gqlt.Load(tempDir)
				if err != nil {
					return err
				}
				entry := cfg.Configs["default"]
				entry.Endpoint = "https://api.example.com/graphql"
				cfg.Configs["default"] = entry
				return cfg.Save(tempDir)
			},
			expectValid:  true,
			expectErrors: false,
		},
		{
			name: "multiple configs with mixed validity",
			setupConfig: func(tempDir string) error {
				cfg, err := gqlt.Load(tempDir)
				if err != nil {
					return err
				}
				entry := cfg.Configs["default"]
				entry.Endpoint = "https://api.example.com/graphql"
				cfg.Configs["default"] = entry

				testEntry := gqlt.ConfigEntry{
					Endpoint: "",
					Headers:  map[string]string{},
				}
				testEntry.Defaults.Out = "json"
				cfg.Configs["test"] = testEntry
				return cfg.Save(tempDir)
			},
			expectValid:  false,
			expectErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			if err := tt.setupConfig(tempDir); err != nil {
				t.Fatalf("Failed to setup config: %v", err)
			}

			cfg, err := gqlt.Load(tempDir)
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			// Validate configuration
			validationResult := map[string]interface{}{
				"valid":          true,
				"config_dir":     tempDir,
				"current_config": cfg.Current,
				"configs_count":  len(cfg.Configs),
			}

			// Validate each configuration
			configErrors := make(map[string][]string)
			for name, config := range cfg.Configs {
				errors := []string{}
				if config.Endpoint == "" {
					errors = append(errors, "Missing endpoint")
				}
				if len(errors) > 0 {
					configErrors[name] = errors
				}
			}

			if len(configErrors) > 0 {
				validationResult["valid"] = false
				validationResult["config_errors"] = configErrors
			}

			isValid := validationResult["valid"].(bool)
			if isValid != tt.expectValid {
				t.Errorf("Expected valid=%v, got %v", tt.expectValid, isValid)
			}

			hasErrors := len(configErrors) > 0
			if hasErrors != tt.expectErrors {
				t.Errorf("Expected errors=%v, got %v (errors: %v)", tt.expectErrors, hasErrors, configErrors)
			}
		})
	}
}

func TestValidateQueryLogic(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		queryFile   string
		setupFile   func(dir string) (string, error)
		expectError bool
	}{
		{
			name:        "inline query",
			query:       "{ users { id } }",
			queryFile:   "",
			expectError: false,
		},
		{
			name:      "query from file",
			query:     "",
			queryFile: "query.graphql",
			setupFile: func(dir string) (string, error) {
				path := filepath.Join(dir, "query.graphql")
				err := os.WriteFile(path, []byte("{ users { id name } }"), 0644)
				return path, err
			},
			expectError: false,
		},
		{
			name:        "no query provided",
			query:       "",
			queryFile:   "",
			expectError: true,
		},
		{
			name:        "query file not found",
			query:       "",
			queryFile:   "nonexistent.graphql",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			inputHandler := gqlt.NewInput()

			queryFile := tt.queryFile
			if tt.setupFile != nil {
				var err error
				queryFile, err = tt.setupFile(tempDir)
				if err != nil {
					t.Fatalf("Failed to setup test file: %v", err)
				}
			} else if tt.queryFile != "" {
				queryFile = filepath.Join(tempDir, tt.queryFile)
			}

			queryStr, err := inputHandler.LoadQuery(tt.query, queryFile)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if queryStr == "" {
					t.Errorf("Expected non-empty query string")
				}
			}
		})
	}
}

func TestValidateSchemaLogic(t *testing.T) {
	tests := []struct {
		name        string
		endpoint    string
		expectError bool
	}{
		{
			name:        "valid client creation",
			endpoint:    "https://api.example.com/graphql",
			expectError: false,
		},
		{
			name:        "empty endpoint",
			endpoint:    "",
			expectError: false, // Client creation doesn't fail for empty endpoint
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := gqlt.NewClient(tt.endpoint, nil)
			if client == nil {
				t.Fatal("Failed to create client")
			}

			introspectClient := gqlt.NewIntrospect(client)
			if introspectClient == nil {
				t.Fatal("Failed to create introspect client")
			}
		})
	}
}

func TestValidateOutputFormats(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"json formatter", "json"},
		{"table formatter", "table"},
		{"yaml formatter", "yaml"},
		{"unknown formatter fallback", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := gqlt.NewFormatter(tt.format)
			if tt.format == "unknown" {
				// Unknown formats should return nil
				if formatter != nil {
					t.Errorf("Expected nil formatter for unknown format %s", tt.format)
				}
				return
			}
			if formatter == nil {
				t.Errorf("Expected non-nil formatter for format %s", tt.format)
			}

			// Test structured output
			data := map[string]interface{}{
				"test": "value",
			}

			// Redirect stdout to capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := formatter.FormatStructured(data, false)

			w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Errorf("FormatStructured failed: %v", err)
			}

			// Read captured output
			var buf []byte
			buf = make([]byte, 4096)
			n, _ := r.Read(buf)
			output := string(buf[:n])

			if output == "" {
				t.Errorf("Expected non-empty output")
			}

			// For JSON format, verify it's valid JSON
			if tt.format == "json" || tt.format == "unknown" {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("Invalid JSON output: %v", err)
				}
			}
		})
	}
}

func TestValidateErrorCodes(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"config load error", gqlt.ErrorCodeConfigLoad},
		{"config not found", gqlt.ErrorCodeConfigNotFound},
		{"query load error", gqlt.ErrorCodeQueryLoad},
		{"schema introspect error", gqlt.ErrorCodeSchemaIntrospect},
		{"graphql execution error", gqlt.ErrorCodeGraphQLExecution},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := gqlt.NewFormatter("json")
			errorBuf := &bytes.Buffer{}
			formatter.SetErrorOutput(errorBuf)

			err := formatter.FormatStructuredError(
				os.ErrNotExist,
				tt.code,
				false,
			)

			if err != nil {
				t.Errorf("FormatStructuredError failed: %v", err)
			}

			output := errorBuf.String()

			// Verify JSON output contains error code
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Errorf("Invalid JSON output: %v", err)
			}

			if result["success"] != false {
				t.Errorf("Expected success=false, got %v", result["success"])
			}

			errorInfo, ok := result["error"].(map[string]interface{})
			if !ok {
				t.Errorf("Expected error field in output")
			}

			if errorInfo["code"] != tt.code {
				t.Errorf("Expected error code %s, got %v", tt.code, errorInfo["code"])
			}
		})
	}
}
