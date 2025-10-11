package gqlt

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SDKServer wraps the official MCP SDK server with gqlt functionality
type SDKServer struct {
	server      *mcp.Server
	client      *Client
	schemaCache map[string]interface{} // endpoint -> schema data
	cacheMutex  sync.RWMutex
}

// NewSDKServer creates a new MCP server using the official SDK
func NewSDKServer() (*SDKServer, error) {
	// Create the official MCP server following the SDK pattern
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "gqlt-mcp-server",
		Version: Version(),
	}, nil)

	// Create gqlt client
	client := NewClient("", nil)

	sdkServer := &SDKServer{
		server:      server,
		client:      client,
		schemaCache: make(map[string]interface{}),
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
	// Add GraphQL execution tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "execute_query",
		Description: "Execute a GraphQL query, mutation, or subscription",
	}, s.handleExecuteQuery)

	// Add type description tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "describe_type",
		Description: "Analyze specific GraphQL types and fields",
	}, s.handleDescribeType)

	// Add list types tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "list_types",
		Description: "List GraphQL type names with optional filtering",
	}, s.handleListTypes)

	// Add version tool
	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "version",
		Description: "Get the current version of gqlt",
	}, s.handleVersion)

	return nil
}

// TODO: Add resource and prompt registration later

// Tool handlers

// ExecuteQueryInput defines the input schema for the execute_query tool
type ExecuteQueryInput struct {
	Query         string                 `json:"query" jsonschema:"The GraphQL query string"`
	Variables     map[string]interface{} `json:"variables,omitempty" jsonschema:"Variables to pass to the query"`
	OperationName string                 `json:"operationName,omitempty" jsonschema:"The operation name to execute"`
	Endpoint      string                 `json:"endpoint" jsonschema:"GraphQL endpoint URL"`
	Headers       map[string]string      `json:"headers,omitempty" jsonschema:"HTTP headers to include"`
	Files         map[string]string      `json:"files,omitempty" jsonschema:"File uploads (variable name to local file path mapping, e.g. {'avatar': '/path/to/photo.jpg'})"`
}

// ExecuteQueryOutput defines the output schema for the execute_query tool
type ExecuteQueryOutput struct {
	Data      interface{} `json:"data" jsonschema:"The GraphQL response data"`
	Errors    interface{} `json:"errors,omitempty" jsonschema:"Any GraphQL errors"`
	ElapsedMs int64       `json:"elapsed_ms" jsonschema:"Query execution time in milliseconds"`
}

// DescribeTypeInput defines the input schema for the describe_type tool
type DescribeTypeInput struct {
	TypeName string            `json:"typeName" jsonschema:"The GraphQL type name to describe"`
	Endpoint string            `json:"endpoint" jsonschema:"GraphQL endpoint URL"`
	Headers  map[string]string `json:"headers,omitempty" jsonschema:"HTTP headers to include"`
	NoCache  bool              `json:"noCache,omitempty" jsonschema:"Skip cache and force fresh schema introspection"`
}

// DescribeTypeOutput defines the output schema for the describe_type tool
type DescribeTypeOutput struct {
	TypeInfo string `json:"type_info" jsonschema:"Information about the GraphQL type"`
}

// ListTypesInput defines the input schema for the list_types tool
type ListTypesInput struct {
	Endpoint string            `json:"endpoint" jsonschema:"GraphQL endpoint URL"`
	Filter   string            `json:"filter,omitempty" jsonschema:"Optional regex pattern to filter type names (e.g., 'Input.*', '.*Type', 'User.*')"`
	Kind     string            `json:"kind,omitempty" jsonschema:"Optional type kind filter (OBJECT, ENUM, SCALAR, UNION, INPUT_OBJECT, INTERFACE)"`
	Headers  map[string]string `json:"headers,omitempty" jsonschema:"HTTP headers to include"`
	NoCache  bool              `json:"noCache,omitempty" jsonschema:"Skip cache and force fresh schema introspection"`
}

// ListTypesOutput defines the output schema for the list_types tool
type ListTypesOutput struct {
	TypeNames []string `json:"type_names" jsonschema:"List of matching type names"`
	Count     int      `json:"count" jsonschema:"Total number of matching types"`
}

// VersionInput defines the input schema for the version tool
type VersionInput struct {
	// No input parameters required
}

// VersionOutput defines the output schema for the version tool
type VersionOutput struct {
	Version string `json:"version" jsonschema:"The current version of gqlt"`
}

func (s *SDKServer) handleExecuteQuery(ctx context.Context, req *mcp.CallToolRequest, input ExecuteQueryInput) (
	*mcp.CallToolResult,
	ExecuteQueryOutput,
	error,
) {
	// Create a new client for this specific endpoint
	client := NewClient(input.Endpoint, nil)

	// Set headers if provided
	if len(input.Headers) > 0 {
		client.SetHeaders(input.Headers)
	}

	// Execute the query with or without files
	start := time.Now()
	var result *Response
	var err error

	if len(input.Files) > 0 {
		// Use ExecuteWithFiles for file uploads
		result, err = client.ExecuteWithFiles(input.Query, input.Variables, input.OperationName, input.Files)
	} else {
		// Use regular Execute for queries without files
		result, err = client.Execute(input.Query, input.Variables, input.OperationName)
	}

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

func (s *SDKServer) handleDescribeType(ctx context.Context, req *mcp.CallToolRequest, input DescribeTypeInput) (
	*mcp.CallToolResult,
	DescribeTypeOutput,
	error,
) {
	var schemaData interface{}
	var exists bool

	// Check cache first, unless NoCache is true
	if !input.NoCache {
		s.cacheMutex.RLock()
		schemaData, exists = s.schemaCache[input.Endpoint]
		s.cacheMutex.RUnlock()
	}

	// If not in cache or NoCache is true, introspect and cache it
	if !exists || input.NoCache {
		client := NewClient(input.Endpoint, nil)

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
			}, DescribeTypeOutput{}, nil
		}

		// Check if introspection returned data
		if result.Data == nil {
			errorMsg := "Schema introspection returned no data"
			if len(result.Errors) > 0 {
				if errMap, ok := result.Errors[0].(map[string]interface{}); ok {
					if msg, ok := errMap["message"].(string); ok {
						errorMsg = msg
					}
				}
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Schema introspection failed: %s", errorMsg),
					},
				},
				IsError: true,
			}, DescribeTypeOutput{}, nil
		}

		// Cache the schema
		s.cacheMutex.Lock()
		s.schemaCache[input.Endpoint] = result.Data
		schemaData = result.Data
		s.cacheMutex.Unlock()
	}

	// Parse the schema to find the specific type
	typeInfo, err := s.extractTypeInfo(schemaData, input.TypeName, input.Endpoint)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to extract type info: %v", err),
				},
			},
			IsError: true,
		}, DescribeTypeOutput{}, nil
	}

	return nil, DescribeTypeOutput{
		TypeInfo: typeInfo,
	}, nil
}

func (s *SDKServer) handleListTypes(ctx context.Context, req *mcp.CallToolRequest, input ListTypesInput) (
	*mcp.CallToolResult,
	ListTypesOutput,
	error,
) {
	var schemaData interface{}
	var exists bool

	// Check cache first, unless NoCache is true
	if !input.NoCache {
		s.cacheMutex.RLock()
		schemaData, exists = s.schemaCache[input.Endpoint]
		s.cacheMutex.RUnlock()
	}

	// If not in cache or NoCache is true, introspect and cache it
	if !exists || input.NoCache {
		client := NewClient(input.Endpoint, nil)

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
			}, ListTypesOutput{TypeNames: []string{}, Count: 0}, nil
		}

		// Check if introspection returned data
		if result.Data == nil {
			errorMsg := "Schema introspection returned no data"
			if len(result.Errors) > 0 {
				if errMap, ok := result.Errors[0].(map[string]interface{}); ok {
					if msg, ok := errMap["message"].(string); ok {
						errorMsg = msg
					}
				}
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Schema introspection failed: %s", errorMsg),
					},
				},
				IsError: true,
			}, ListTypesOutput{TypeNames: []string{}, Count: 0}, nil
		}

		// Cache the schema
		s.cacheMutex.Lock()
		s.schemaCache[input.Endpoint] = result.Data
		schemaData = result.Data
		s.cacheMutex.Unlock()
	}

	// Parse the schema to find matching types
	typeNames, err := s.listMatchingTypes(schemaData, input.Filter, input.Kind)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("Failed to list types: %v", err),
				},
			},
			IsError: true,
		}, ListTypesOutput{TypeNames: []string{}, Count: 0}, nil
	}

	return nil, ListTypesOutput{
		TypeNames: typeNames,
		Count:     len(typeNames),
	}, nil
}

// extractTypeInfo parses the schema data to extract information about a specific type
func (s *SDKServer) extractTypeInfo(schemaData interface{}, typeName, endpoint string) (string, error) {
	// Parse the schema structure
	schemaMap, ok := schemaData.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid schema data format")
	}

	// Navigate to the schema types
	__schema, ok := schemaMap["__schema"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("schema missing __schema field")
	}

	types, ok := __schema["types"].([]interface{})
	if !ok {
		return "", fmt.Errorf("schema missing types array")
	}

	// Find the specific type
	var targetType map[string]interface{}
	for _, typeItem := range types {
		if typeDef, ok := typeItem.(map[string]interface{}); ok {
			if name, ok := typeDef["name"].(string); ok && name == typeName {
				targetType = typeDef
				break
			}
		}
	}

	if targetType == nil {
		return "", fmt.Errorf("type '%s' not found in schema", typeName)
	}

	// Format the type information
	return s.formatTypeDefinition(targetType, typeName, endpoint), nil
}

// formatTypeDefinition formats a type definition into a readable string
func (s *SDKServer) formatTypeDefinition(typeDef map[string]interface{}, typeName, endpoint string) string {
	result := fmt.Sprintf("Type: %s (from %s)\n\n", typeName, endpoint)

	// Get type kind
	if kind, ok := typeDef["kind"].(string); ok {
		result += fmt.Sprintf("Kind: %s\n", kind)
	}

	// Get description
	if description, ok := typeDef["description"].(string); ok && description != "" {
		result += fmt.Sprintf("Description: %s\n", description)
	}

	// Handle different type kinds
	switch typeDef["kind"] {
	case "OBJECT", "INTERFACE":
		s.formatObjectType(typeDef, &result)
	case "ENUM":
		s.formatEnumType(typeDef, &result)
	case "SCALAR":
		s.formatScalarType(typeDef, &result)
	case "UNION":
		s.formatUnionType(typeDef, &result)
	case "INPUT_OBJECT":
		s.formatInputObjectType(typeDef, &result)
	}

	return result
}

// formatObjectType formats object/interface types with their fields
func (s *SDKServer) formatObjectType(typeDef map[string]interface{}, result *string) {
	if fields, ok := typeDef["fields"].([]interface{}); ok && len(fields) > 0 {
		*result += "\nFields:\n"
		for _, field := range fields {
			if fieldDef, ok := field.(map[string]interface{}); ok {
				s.formatField(fieldDef, result)
			}
		}
	}
}

// formatEnumType formats enum types with their values
func (s *SDKServer) formatEnumType(typeDef map[string]interface{}, result *string) {
	if enumValues, ok := typeDef["enumValues"].([]interface{}); ok && len(enumValues) > 0 {
		*result += "\nValues:\n"
		for _, value := range enumValues {
			if valueDef, ok := value.(map[string]interface{}); ok {
				if name, ok := valueDef["name"].(string); ok {
					*result += fmt.Sprintf("  - %s", name)
					if description, ok := valueDef["description"].(string); ok && description != "" {
						*result += fmt.Sprintf(" (%s)", description)
					}
					*result += "\n"
				}
			}
		}
	}
}

// formatScalarType formats scalar types
func (s *SDKServer) formatScalarType(typeDef map[string]interface{}, result *string) {
	*result += "\nThis is a scalar type.\n"
}

// formatUnionType formats union types with their possible types
func (s *SDKServer) formatUnionType(typeDef map[string]interface{}, result *string) {
	if possibleTypes, ok := typeDef["possibleTypes"].([]interface{}); ok && len(possibleTypes) > 0 {
		*result += "\nPossible Types:\n"
		for _, possibleType := range possibleTypes {
			if typeRef, ok := possibleType.(map[string]interface{}); ok {
				if name, ok := typeRef["name"].(string); ok {
					*result += fmt.Sprintf("  - %s\n", name)
				}
			}
		}
	}
}

// formatInputObjectType formats input object types with their input fields
func (s *SDKServer) formatInputObjectType(typeDef map[string]interface{}, result *string) {
	if inputFields, ok := typeDef["inputFields"].([]interface{}); ok && len(inputFields) > 0 {
		*result += "\nInput Fields:\n"
		for _, field := range inputFields {
			if fieldDef, ok := field.(map[string]interface{}); ok {
				s.formatField(fieldDef, result)
			}
		}
	}
}

// formatField formats a field definition
func (s *SDKServer) formatField(fieldDef map[string]interface{}, result *string) {
	if name, ok := fieldDef["name"].(string); ok {
		*result += fmt.Sprintf("  - %s", name)

		// Add type information
		if typeInfo, ok := fieldDef["type"].(map[string]interface{}); ok {
			*result += fmt.Sprintf(": %s", s.formatType(typeInfo))
		}

		// Add description
		if description, ok := fieldDef["description"].(string); ok && description != "" {
			*result += fmt.Sprintf(" - %s", description)
		}

		// Add arguments if present
		if args, ok := fieldDef["args"].([]interface{}); ok && len(args) > 0 {
			*result += "\n    Arguments:"
			for _, arg := range args {
				if argDef, ok := arg.(map[string]interface{}); ok {
					if argName, ok := argDef["name"].(string); ok {
						*result += fmt.Sprintf("\n      - %s", argName)
						if argType, ok := argDef["type"].(map[string]interface{}); ok {
							*result += fmt.Sprintf(": %s", s.formatType(argType))
						}
					}
				}
			}
		}

		*result += "\n"
	}
}

// formatType formats a type reference
func (s *SDKServer) formatType(typeInfo map[string]interface{}) string {
	if kind, ok := typeInfo["kind"].(string); ok {
		switch kind {
		case "NON_NULL":
			if ofType, ok := typeInfo["ofType"].(map[string]interface{}); ok {
				return s.formatType(ofType) + "!"
			}
		case "LIST":
			if ofType, ok := typeInfo["ofType"].(map[string]interface{}); ok {
				return "[" + s.formatType(ofType) + "]"
			}
		}
	}

	if name, ok := typeInfo["name"].(string); ok {
		return name
	}
	return "Unknown"
}

// listMatchingTypes finds types matching the given filter and kind
func (s *SDKServer) listMatchingTypes(schemaData interface{}, filter, kind string) ([]string, error) {
	// Parse the schema structure
	schemaMap, ok := schemaData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid schema data format")
	}

	// Navigate to the schema types
	__schema, ok := schemaMap["__schema"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("schema missing __schema field")
	}

	types, ok := __schema["types"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("schema missing types array")
	}

	matchingTypes := []string{} // Initialize as empty slice, not nil

	// Iterate through all types
	for _, typeItem := range types {
		if typeDef, ok := typeItem.(map[string]interface{}); ok {
			if name, ok := typeDef["name"].(string); ok {
				// Skip introspection types
				if name == "" || name[0] == '_' {
					continue
				}

				// Check kind filter
				if kind != "" {
					if typeKind, ok := typeDef["kind"].(string); ok {
						if typeKind != kind {
							continue
						}
					} else {
						continue
					}
				}

				// Check name filter (regex matching)
				if filter != "" {
					if !s.matchesRegex(name, filter) {
						continue
					}
				}

				matchingTypes = append(matchingTypes, name)
			}
		}
	}

	return matchingTypes, nil
}

// matchesRegex performs regex pattern matching
func (s *SDKServer) matchesRegex(name, pattern string) bool {
	// Compile the regex pattern
	regex, err := regexp.Compile(pattern)
	if err != nil {
		// If regex compilation fails, fall back to exact match
		return name == pattern
	}

	// Check if the name matches the pattern
	return regex.MatchString(name)
}

func (s *SDKServer) handleVersion(ctx context.Context, req *mcp.CallToolRequest, input VersionInput) (
	*mcp.CallToolResult,
	VersionOutput,
	error,
) {
	// Return the current version
	return nil, VersionOutput{
		Version: Version(),
	}, nil
}

// TODO: Add resource and prompt handlers later
