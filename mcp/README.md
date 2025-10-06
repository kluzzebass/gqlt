# gqlt MCP Server

The gqlt MCP (Model Context Protocol) server provides AI agents with access to GraphQL functionality through a standardized JSON-RPC 2.0 interface.

## Overview

The MCP server exposes gqlt's GraphQL capabilities to AI agents, allowing them to:
- Execute GraphQL queries, mutations, and subscriptions
- Introspect GraphQL schemas
- Analyze and describe GraphQL types
- Validate GraphQL queries
- Handle file uploads
- Manage GraphQL configurations
- Configure authentication

## Starting the Server

```bash
# Start with default settings (localhost:8080)
gqlt mcp

# Start on specific address
gqlt mcp --address localhost:9000

# Start with specific configuration
gqlt mcp --use-config production
```

## Integration with AI Clients

### Cursor Integration

1. **Start the gqlt MCP server:**
   ```bash
   gqlt mcp --address localhost:8080
   ```

2. **Add to Cursor MCP settings:**
   - Go to Cursor menu → "Settings" → "Cursor Settings" → "Tools & MCP"
   - Choose "New MCP Server"
   - Edit the mcp.json file that pops up:
   ```json
   {
     "mcpServers": {
       "gqlt": {
         "url": "http://localhost:8080/mcp"
       }
     }
   }
   ```

3. **Restart Cursor** to load the MCP server

### Claude Desktop Integration

1. **Start the gqlt MCP server:**
   ```bash
   gqlt mcp --address localhost:8080
   ```

2. **Add to Claude Desktop config** (`~/.config/claude-desktop/config.json`):
   ```json
   {
     "mcpServers": {
       "gqlt": {
         "command": "gqlt",
         "args": ["mcp", "--address", "localhost:8080"],
         "env": {}
       }
     }
   }
   ```

3. **Restart Claude Desktop**

### Other MCP Clients

- **Server URL:** `http://localhost:8080/mcp`
- **Protocol:** JSON-RPC 2.0 over HTTP POST
- **Endpoint:** All requests should be sent to the `/mcp` endpoint

## Usage Examples

Once integrated with an AI client, you can use natural language to interact with GraphQL:

**Example prompts:**
- "Show me the GraphQL schema for this API"
- "Execute a query to get all users with their email addresses"
- "Validate this GraphQL query before running it"
- "Set up authentication for the production environment"
- "Upload a file using this GraphQL mutation"

The AI agent will use the MCP tools to:
1. **Introspect schemas** to understand available types and fields
2. **Execute queries** with proper variables and authentication
3. **Validate queries** before execution
4. **Manage configurations** for different environments
5. **Handle file uploads** in mutations

## MCP Tools

### execute_query
Execute a GraphQL query, mutation, or subscription.

**Parameters:**
- `query` (string, required): The GraphQL query string
- `variables` (object, optional): Variables to pass to the query
- `operationName` (string, optional): The operation name to execute
- `endpoint` (string, optional): GraphQL endpoint URL (uses config if not provided)
- `headers` (object, optional): HTTP headers to include

**Example:**
```json
{
  "name": "execute_query",
  "arguments": {
    "query": "query GetUser($id: ID!) { user(id: $id) { name email } }",
    "variables": {"id": "123"},
    "operationName": "GetUser"
  }
}
```

### introspect_schema
Get complete GraphQL schema information via introspection.

**Parameters:**
- `endpoint` (string, optional): GraphQL endpoint URL (uses config if not provided)
- `headers` (object, optional): HTTP headers to include

### describe_type
Analyze specific GraphQL types and fields.

**Parameters:**
- `typeName` (string, required): The GraphQL type name to describe
- `endpoint` (string, optional): GraphQL endpoint URL (uses config if not provided)
- `headers` (object, optional): HTTP headers to include

### validate_query
Check GraphQL query validity against schema.

**Parameters:**
- `query` (string, required): The GraphQL query to validate
- `endpoint` (string, optional): GraphQL endpoint URL (uses config if not provided)
- `headers` (object, optional): HTTP headers to include

### upload_files
Handle file uploads in GraphQL mutations.

**Parameters:**
- `query` (string, required): The GraphQL mutation with file uploads
- `variables` (object, optional): Variables to pass to the mutation
- `files` (object, required): File mappings (field name to file path)
- `endpoint` (string, optional): GraphQL endpoint URL (uses config if not provided)
- `headers` (object, optional): HTTP headers to include

### get_config
Retrieve and manage GraphQL endpoint configurations.

**Parameters:**
- `action` (string, optional): Configuration action ("list", "show", "current")
- `name` (string, optional): Configuration name (for "show" action)

### set_auth
Configure authentication (Bearer, Basic, API Key).

**Parameters:**
- `type` (string, required): Authentication type ("bearer", "basic", "api_key")
- `token` (string, optional): Bearer token or API key
- `username` (string, optional): Username for basic auth
- `password` (string, optional): Password for basic auth
- `configName` (string, optional): Configuration name to update

## MCP Resources

### config://current
The currently active GraphQL configuration.

### config://list
List of all available GraphQL configurations.

### schema://introspection
Complete GraphQL schema via introspection.

## MCP Prompts

### graphql_query
Generate a GraphQL query for a specific use case.

**Arguments:**
- `use_case` (string, required): The use case or requirement for the query
- `schema_context` (string, optional): Optional schema context to help generate the query

### graphql_mutation
Generate a GraphQL mutation for a specific operation.

**Arguments:**
- `operation` (string, required): The operation to perform (create, update, delete, etc.)
- `entity_type` (string, required): The type of entity to operate on
- `schema_context` (string, optional): Optional schema context to help generate the mutation

### schema_analysis
Analyze a GraphQL schema and provide insights.

**Arguments:**
- `analysis_type` (string, required): Type of analysis (complexity, patterns, best_practices)
- `schema_data` (string, required): The schema data to analyze

## Protocol Details

The server implements the Model Context Protocol specification with:
- JSON-RPC 2.0 for request/response handling
- Session management for concurrent AI agent requests
- Proper error handling with standardized error codes
- Support for tools, resources, and prompts

## Error Codes

- `-32700`: Parse error
- `-32600`: Invalid Request
- `-32601`: Method not found
- `-32602`: Invalid params
- `-32603`: Internal error

## Testing

Run the test suite:

```bash
go test ./mcp/... -v
```

The test suite includes:
- Unit tests for all handlers
- Integration tests with HTTP client
- Parameter validation tests
- Error handling tests

## Configuration

The MCP server uses the same configuration system as the gqlt CLI. Configuration files are automatically loaded from the standard locations:

- Linux: `~/.config/gqlt/config.json`
- macOS: `~/Library/Application Support/gqlt/config.json`
- Windows: `%APPDATA%/gqlt/config.json`

## Security Considerations

- The server runs on localhost by default
- Authentication credentials are handled securely through the configuration system
- File uploads are validated before processing
- All network requests respect configured headers and authentication
