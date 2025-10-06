package mcp

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kluzzebass/gqlt"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SDKServer wraps the official MCP SDK server with gqlt functionality
type SDKServer struct {
	server *mcp.Server
	config *gqlt.Config
	client *gqlt.Client
}

// NewSDKServer creates a new MCP server using the official SDK
func NewSDKServer(config *gqlt.Config) (*SDKServer, error) {
	// Create the official MCP server following the SDK pattern
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "gqlt-mcp-server",
		Version: gqlt.Version(),
	}, nil)

	// Create gqlt client
	client := gqlt.NewClient("", nil)

	sdkServer := &SDKServer{
		server: server,
		config: config,
		client: client,
	}

	// Register all tools using the official SDK pattern
	if err := sdkServer.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	// TODO: Add resources and prompts later

	return sdkServer, nil
}

// Start starts the MCP server using stdin/stdout
func (s *SDKServer) Start(ctx context.Context, address string) error {
	log.Printf("Starting MCP server using stdin/stdout")

	// Use the SDK's built-in stdin/stdout transport following the official pattern
	return s.server.Run(ctx, &mcp.StdioTransport{})
}

// Stop stops the MCP server
func (s *SDKServer) Stop(ctx context.Context) error {
	// The SDK handles graceful shutdown
	return nil
}

// registerTools registers all MCP tools with the server
func (s *SDKServer) registerTools() error {
	// Add a simple test tool using the official SDK pattern
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "test_tool",
		Description: "A simple test tool to verify the MCP server is working",
	}, s.handleTestTool)

	// Add GraphQL execution tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "execute_query",
		Description: "Execute a GraphQL query, mutation, or subscription",
	}, s.handleExecuteQuery)

	// Add schema introspection tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "introspect_schema",
		Description: "Get complete GraphQL schema information via introspection",
	}, s.handleIntrospectSchema)

	// Add query validation tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "validate_query",
		Description: "Check GraphQL query validity against schema",
	}, s.handleValidateQuery)

	// Add type description tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "describe_type",
		Description: "Analyze specific GraphQL types and fields",
	}, s.handleDescribeType)

	return nil
}

// TODO: Add resource and prompt registration later

// Tool handlers

// TestToolInput defines the input schema for the test tool
type TestToolInput struct {
	Message string `json:"message" jsonschema:"A test message to echo back"`
}

// TestToolOutput defines the output schema for the test tool
type TestToolOutput struct {
	Response string `json:"response" jsonschema:"The response from the test tool"`
}

// ExecuteQueryInput defines the input schema for the execute_query tool
type ExecuteQueryInput struct {
	Query         string                 `json:"query" jsonschema:"The GraphQL query string"`
	Variables     map[string]interface{} `json:"variables,omitempty" jsonschema:"Variables to pass to the query"`
	OperationName string                 `json:"operationName,omitempty" jsonschema:"The operation name to execute"`
	Endpoint      string                 `json:"endpoint" jsonschema:"GraphQL endpoint URL"`
	Headers       map[string]string      `json:"headers,omitempty" jsonschema:"HTTP headers to include"`
}

// ExecuteQueryOutput defines the output schema for the execute_query tool
type ExecuteQueryOutput struct {
	Data      interface{} `json:"data" jsonschema:"The GraphQL response data"`
	Errors    interface{} `json:"errors,omitempty" jsonschema:"Any GraphQL errors"`
	ElapsedMs int64       `json:"elapsed_ms" jsonschema:"Query execution time in milliseconds"`
}

// IntrospectSchemaInput defines the input schema for the introspect_schema tool
type IntrospectSchemaInput struct {
	Endpoint string            `json:"endpoint" jsonschema:"GraphQL endpoint URL"`
	Headers  map[string]string `json:"headers,omitempty" jsonschema:"HTTP headers to include"`
}

// IntrospectSchemaOutput defines the output schema for the introspect_schema tool
type IntrospectSchemaOutput struct {
	Schema interface{} `json:"schema" jsonschema:"The complete GraphQL schema"`
}

// ValidateQueryInput defines the input schema for the validate_query tool
type ValidateQueryInput struct {
	Query    string            `json:"query" jsonschema:"The GraphQL query to validate"`
	Endpoint string            `json:"endpoint" jsonschema:"GraphQL endpoint URL"`
	Headers  map[string]string `json:"headers,omitempty" jsonschema:"HTTP headers to include"`
}

// ValidateQueryOutput defines the output schema for the validate_query tool
type ValidateQueryOutput struct {
	Valid   bool   `json:"valid" jsonschema:"Whether the query is valid"`
	Message string `json:"message" jsonschema:"Validation result message"`
}

// DescribeTypeInput defines the input schema for the describe_type tool
type DescribeTypeInput struct {
	TypeName string            `json:"typeName" jsonschema:"The GraphQL type name to describe"`
	Endpoint string            `json:"endpoint" jsonschema:"GraphQL endpoint URL"`
	Headers  map[string]string `json:"headers,omitempty" jsonschema:"HTTP headers to include"`
}

// DescribeTypeOutput defines the output schema for the describe_type tool
type DescribeTypeOutput struct {
	TypeInfo string `json:"type_info" jsonschema:"Information about the GraphQL type"`
}

func (s *SDKServer) handleTestTool(ctx context.Context, req *mcp.CallToolRequest, input TestToolInput) (
	*mcp.CallToolResult,
	TestToolOutput,
	error,
) {
	message := "Hello from gqlt MCP server!"
	if input.Message != "" {
		message = fmt.Sprintf("Echo: %s", input.Message)
	}

	return nil, TestToolOutput{Response: message}, nil
}

func (s *SDKServer) handleExecuteQuery(ctx context.Context, req *mcp.CallToolRequest, input ExecuteQueryInput) (
	*mcp.CallToolResult,
	ExecuteQueryOutput,
	error,
) {
	// Create a new client for this specific endpoint
	client := gqlt.NewClient(input.Endpoint, nil)

	// Set headers if provided
	if len(input.Headers) > 0 {
		client.SetHeaders(input.Headers)
	}

	// Execute the query
	start := time.Now()
	result, err := client.Execute(input.Query, input.Variables, input.OperationName)
	elapsed := time.Since(start)

	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Query execution failed: %v", err),
				},
			},
			IsError: true,
		}, ExecuteQueryOutput{}, nil
	}

	return nil, ExecuteQueryOutput{
		Data:      result.Data,
		Errors:    result.Errors,
		ElapsedMs: elapsed.Milliseconds(),
	}, nil
}

func (s *SDKServer) handleIntrospectSchema(ctx context.Context, req *mcp.CallToolRequest, input IntrospectSchemaInput) (
	*mcp.CallToolResult,
	IntrospectSchemaOutput,
	error,
) {
	// Create a new client for this specific endpoint
	client := gqlt.NewClient(input.Endpoint, nil)

	// Set headers if provided
	if len(input.Headers) > 0 {
		client.SetHeaders(input.Headers)
	}

	// Introspect the schema
	result, err := client.Introspect()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Schema introspection failed: %v", err),
				},
			},
			IsError: true,
		}, IntrospectSchemaOutput{}, nil
	}

	return nil, IntrospectSchemaOutput{
		Schema: result.Data,
	}, nil
}

func (s *SDKServer) handleValidateQuery(ctx context.Context, req *mcp.CallToolRequest, input ValidateQueryInput) (
	*mcp.CallToolResult,
	ValidateQueryOutput,
	error,
) {
	// Create a new client for this specific endpoint
	client := gqlt.NewClient(input.Endpoint, nil)

	// Set headers if provided
	if len(input.Headers) > 0 {
		client.SetHeaders(input.Headers)
	}

	// Validate by attempting to execute (simplified approach)
	// In a real implementation, you'd want proper query validation
	_, err := client.Execute(input.Query, nil, "")
	if err != nil {
		return nil, ValidateQueryOutput{
			Valid:   false,
			Message: fmt.Sprintf("Query validation failed: %v", err),
		}, nil
	}

	return nil, ValidateQueryOutput{
		Valid:   true,
		Message: "Query is valid",
	}, nil
}

func (s *SDKServer) handleDescribeType(ctx context.Context, req *mcp.CallToolRequest, input DescribeTypeInput) (
	*mcp.CallToolResult,
	DescribeTypeOutput,
	error,
) {
	// Create a new client for this specific endpoint
	client := gqlt.NewClient(input.Endpoint, nil)

	// Set headers if provided
	if len(input.Headers) > 0 {
		client.SetHeaders(input.Headers)
	}

	// Get the schema first
	result, err := client.Introspect()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Schema introspection failed: %v", err),
				},
			},
			IsError: true,
		}, DescribeTypeOutput{}, nil
	}

	// For now, just return a placeholder - we can improve this later
	typeInfo := fmt.Sprintf("Type analysis for '%s' on endpoint '%s':\n", input.TypeName, input.Endpoint)
	typeInfo += "Schema introspection completed. Type details would be extracted from the schema here."
	typeInfo += fmt.Sprintf("\n\nSchema data available: %v", result.Data != nil)

	return nil, DescribeTypeOutput{
		TypeInfo: typeInfo,
	}, nil
}

// TODO: Add resource and prompt handlers later
