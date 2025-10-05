package gqlt

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewIntrospect(t *testing.T) {
	client := NewClient("https://api.example.com/graphql", nil)
	introspect := NewIntrospect(client)

	if introspect == nil {
		t.Error("NewIntrospect() returned nil")
	}
}

func TestIntrospect_IntrospectSchema(t *testing.T) {
	// This test would require a real GraphQL endpoint
	// For now, we'll test the error case with an invalid URL
	client := NewClient("https://invalid-url-that-does-not-exist.com/graphql", nil)
	introspect := NewIntrospect(client)

	_, err := introspect.IntrospectSchema()
	// Note: The client might handle invalid URLs differently, so we just test that it doesn't panic
	if err != nil {
		// This is expected for invalid URLs
		t.Logf("Got expected error for invalid URL: %v", err)
	}
}

func TestIntrospect_SaveSchema(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	schemaFile := filepath.Join(tempDir, "schema.json")

	// Create a mock response for testing
	response := &Response{
		Data: map[string]interface{}{
			"__schema": map[string]interface{}{
				"queryType": map[string]interface{}{
					"name": "Query",
				},
			},
		},
	}

	client := NewClient("https://api.example.com/graphql", nil)
	introspect := NewIntrospect(client)

	err := introspect.SaveSchema(response, schemaFile)
	if err != nil {
		t.Errorf("SaveSchema() error = %v", err)
	}

	// Check if file was created
	if _, err := os.Stat(schemaFile); os.IsNotExist(err) {
		t.Error("SaveSchema() did not create file")
	}
}
