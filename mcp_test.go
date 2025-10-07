package gqlt

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestNewSDKServer(t *testing.T) {
	server, err := NewSDKServer()
	if err != nil {
		t.Fatalf("Failed to create SDK server: %v", err)
	}

	if server == nil {
		t.Fatal("Server should not be nil")
	}

	if server.server == nil {
		t.Fatal("MCP server should not be nil")
	}

	if server.client == nil {
		t.Fatal("Client should not be nil")
	}

	if server.schemaCache == nil {
		t.Fatal("Schema cache should not be nil")
	}
}

func TestSDKServer_StartStop(t *testing.T) {
	server, err := NewSDKServer()
	if err != nil {
		t.Fatalf("Failed to create SDK server: %v", err)
	}

	// Test Stop without starting (should not error)
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer stopCancel()

	err = server.Stop(stopCtx)
	if err != nil {
		t.Errorf("Stop should not return error: %v", err)
	}

	// Test that we can create multiple servers
	server2, err := NewSDKServer()
	if err != nil {
		t.Fatalf("Failed to create second SDK server: %v", err)
	}

	if server2 == nil {
		t.Fatal("Second server should not be nil")
	}
}

func TestSDKServer_handleExecuteQuery(t *testing.T) {
	server, err := NewSDKServer()
	if err != nil {
		t.Fatalf("Failed to create SDK server: %v", err)
	}

	// Test with a valid GraphQL endpoint
	input := ExecuteQueryInput{
		Query:    `query { __schema { types { name } } }`,
		Endpoint: "https://countries.trevorblades.com/graphql",
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	result, output, err := server.handleExecuteQuery(ctx, req, input)
	if err != nil {
		t.Fatalf("handleExecuteQuery failed: %v", err)
	}

	if result != nil {
		t.Error("Result should be nil for successful execution")
	}

	if output.Data == nil {
		t.Error("Output data should not be nil")
	}

	if output.ElapsedMs <= 0 {
		t.Error("Elapsed time should be positive")
	}
}

func TestSDKServer_handleExecuteQuery_InvalidEndpoint(t *testing.T) {
	server, err := NewSDKServer()
	if err != nil {
		t.Fatalf("Failed to create SDK server: %v", err)
	}

	input := ExecuteQueryInput{
		Query:    `query { __schema { types { name } } }`,
		Endpoint: "https://invalid-endpoint.com/graphql",
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	result, _, err := server.handleExecuteQuery(ctx, req, input)
	if err != nil {
		t.Fatalf("handleExecuteQuery failed: %v", err)
	}

	if result == nil {
		t.Error("Result should not be nil for failed execution")
		return
	}

	if !result.IsError {
		t.Error("Result should indicate error")
	}

	if len(result.Content) == 0 {
		t.Error("Result should have error content")
	}
}

func TestSDKServer_handleDescribeType(t *testing.T) {
	server, err := NewSDKServer()
	if err != nil {
		t.Fatalf("Failed to create SDK server: %v", err)
	}

	input := DescribeTypeInput{
		TypeName: "Country",
		Endpoint: "https://countries.trevorblades.com/graphql",
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	result, output, err := server.handleDescribeType(ctx, req, input)
	if err != nil {
		t.Fatalf("handleDescribeType failed: %v", err)
	}

	if result != nil {
		t.Error("Result should be nil for successful execution")
	}

	if output.TypeInfo == "" {
		t.Error("Type info should not be empty")
	}

	// Check that the type info contains expected content
	if !strings.Contains(output.TypeInfo, "Country") {
		t.Error("Type info should contain 'Country'")
	}

	if !strings.Contains(output.TypeInfo, "Fields:") {
		t.Error("Type info should contain 'Fields:'")
	}
}

func TestSDKServer_handleDescribeType_InvalidType(t *testing.T) {
	server, err := NewSDKServer()
	if err != nil {
		t.Fatalf("Failed to create SDK server: %v", err)
	}

	input := DescribeTypeInput{
		TypeName: "NonExistentType",
		Endpoint: "https://countries.trevorblades.com/graphql",
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	result, _, err := server.handleDescribeType(ctx, req, input)
	if err != nil {
		t.Fatalf("handleDescribeType failed: %v", err)
	}

	if result == nil {
		t.Error("Result should not be nil for failed execution")
		return
	}

	if !result.IsError {
		t.Error("Result should indicate error")
	}
}

func TestSDKServer_handleListTypes(t *testing.T) {
	server, err := NewSDKServer()
	if err != nil {
		t.Fatalf("Failed to create SDK server: %v", err)
	}

	input := ListTypesInput{
		Endpoint: "https://countries.trevorblades.com/graphql",
		Kind:     "OBJECT",
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	result, output, err := server.handleListTypes(ctx, req, input)
	if err != nil {
		t.Fatalf("handleListTypes failed: %v", err)
	}

	if result != nil {
		t.Error("Result should be nil for successful execution")
	}

	if len(output.TypeNames) == 0 {
		t.Error("Type names should not be empty")
	}

	if output.Count != len(output.TypeNames) {
		t.Error("Count should match number of type names")
	}

	// Check that we got some expected types
	expectedTypes := []string{"Country", "Continent", "Query"}
	for _, expected := range expectedTypes {
		if !slices.Contains(output.TypeNames, expected) {
			t.Errorf("Expected type '%s' not found in results", expected)
		}
	}
}

func TestSDKServer_handleListTypes_WithFilter(t *testing.T) {
	server, err := NewSDKServer()
	if err != nil {
		t.Fatalf("Failed to create SDK server: %v", err)
	}

	input := ListTypesInput{
		Endpoint: "https://countries.trevorblades.com/graphql",
		Filter:   "^C.*", // Types starting with 'C'
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	result, output, err := server.handleListTypes(ctx, req, input)
	if err != nil {
		t.Fatalf("handleListTypes failed: %v", err)
	}

	if result != nil {
		t.Error("Result should be nil for successful execution")
	}

	// All returned types should start with 'C'
	for _, typeName := range output.TypeNames {
		if len(typeName) == 0 || typeName[0] != 'C' {
			t.Errorf("Type '%s' should start with 'C'", typeName)
		}
	}
}

func TestSDKServer_matchesRegex(t *testing.T) {
	server, err := NewSDKServer()
	if err != nil {
		t.Fatalf("Failed to create SDK server: %v", err)
	}

	tests := []struct {
		name     string
		pattern  string
		expected bool
	}{
		{"Country", "^C.*", true},
		{"Country", ".*try$", true},
		{"Country", "^A.*", false},
		{"Country", "Country", true},
		{"Country", "Invalid", false},
		{"Country", ".*", true},
		{"Country", "^$", false},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+tt.pattern, func(t *testing.T) {
			result := server.matchesRegex(tt.name, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchesRegex(%q, %q) = %v, want %v", tt.name, tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestSDKServer_matchesRegex_InvalidPattern(t *testing.T) {
	server, err := NewSDKServer()
	if err != nil {
		t.Fatalf("Failed to create SDK server: %v", err)
	}

	// Test with invalid regex pattern - should fall back to exact match
	result := server.matchesRegex("Country", "[invalid")
	if result {
		t.Error("Invalid regex should fall back to exact match and return false")
	}

	// Test exact match fallback
	result = server.matchesRegex("Country", "Country")
	if !result {
		t.Error("Exact match should work as fallback")
	}
}

func TestSDKServer_formatType(t *testing.T) {
	server, err := NewSDKServer()
	if err != nil {
		t.Fatalf("Failed to create SDK server: %v", err)
	}

	tests := []struct {
		name     string
		typeInfo map[string]interface{}
		expected string
	}{
		{
			"Simple type",
			map[string]interface{}{"name": "String", "kind": "SCALAR"},
			"String",
		},
		{
			"Non-null type",
			map[string]interface{}{
				"kind":   "NON_NULL",
				"ofType": map[string]interface{}{"name": "String", "kind": "SCALAR"},
			},
			"String!",
		},
		{
			"List type",
			map[string]interface{}{
				"kind":   "LIST",
				"ofType": map[string]interface{}{"name": "String", "kind": "SCALAR"},
			},
			"[String]",
		},
		{
			"List of non-null",
			map[string]interface{}{
				"kind": "LIST",
				"ofType": map[string]interface{}{
					"kind":   "NON_NULL",
					"ofType": map[string]interface{}{"name": "String", "kind": "SCALAR"},
				},
			},
			"[String!]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := server.formatType(tt.typeInfo)
			if result != tt.expected {
				t.Errorf("formatType() = %v, want %v", result, tt.expected)
			}
		})
	}
}
