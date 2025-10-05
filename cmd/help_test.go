package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestHelpTextContainsAIFeatures(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "root help contains AI features",
			args:     []string{"--help"},
			expected: []string{"AI-FRIENDLY FEATURES", "COMMON PATTERNS", "AUTHENTICATION", "OUTPUT FORMATS"},
		},
		{
			name:     "run help contains examples",
			args:     []string{"run", "--help"},
			expected: []string{"EXAMPLES", "QUERY SOURCES", "VARIABLES", "AUTHENTICATION", "OUTPUT MODES"},
		},
		{
			name:     "config help contains workflows",
			args:     []string{"config", "--help"},
			expected: []string{"AI-FRIENDLY FEATURES", "COMMON WORKFLOWS", "EXAMPLES"},
		},
		{
			name:     "validate help contains AI features",
			args:     []string{"validate", "--help"},
			expected: []string{"AI-FRIENDLY FEATURES", "EXAMPLES"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh command for each test
			cmd := rootCmd
			
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute command
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			// Close write end and read output
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if err != nil {
				t.Errorf("Execute() error = %v", err)
			}

			// Check that all expected sections are present
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected help text to contain '%s', but it didn't", expected)
				}
			}
		})
	}
}

func TestHelpTextContainsExamples(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "root help contains basic examples",
			args:    []string{"--help"},
			expected: []string{"gqlt run --url", "gqlt config create", "gqlt introspect", "gqlt describe"},
		},
		{
			name:     "run help contains query examples",
			args:     []string{"run", "--help"},
			expected: []string{"gqlt run --url", "gqlt run --query", "gqlt run --token", "gqlt run --format json"},
		},
		{
			name:     "config help contains workflow examples",
			args:     []string{"config", "--help"},
			expected: []string{"gqlt config init", "gqlt config create", "gqlt config set", "gqlt config use"},
		},
		{
			name:     "validate help contains validation examples",
			args:     []string{"validate", "--help"},
			expected: []string{"gqlt validate query", "gqlt validate config", "gqlt validate schema"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh command for each test
			cmd := rootCmd
			
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute command
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			// Close write end and read output
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if err != nil {
				t.Errorf("Execute() error = %v", err)
			}

			// Check that all expected examples are present
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected help text to contain '%s', but it didn't", expected)
				}
			}
		})
	}
}

func TestHelpTextContainsStructuredOutputInfo(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "root help mentions structured output",
			args:     []string{"--help"},
			expected: []string{"--format json", "Machine-readable error codes", "Quiet mode"},
		},
		{
			name:     "run help mentions output formats",
			args:     []string{"run", "--help"},
			expected: []string{"json: Structured JSON", "pretty: Colorized formatted JSON", "raw: Unformatted JSON"},
		},
		{
			name:     "config help mentions output formats",
			args:     []string{"config", "--help"},
			expected: []string{"--format json|table|yaml", "Machine-readable error codes"},
		},
		{
			name:     "validate help mentions structured output",
			args:     []string{"validate", "--help"},
			expected: []string{"Structured JSON output", "Machine-readable error codes", "--format json --quiet"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh command for each test
			cmd := rootCmd
			
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute command
			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			// Close write end and read output
			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if err != nil {
				t.Errorf("Execute() error = %v", err)
			}

			// Check that all expected structured output info is present
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected help text to contain '%s', but it didn't", expected)
				}
			}
		})
	}
}
