# gqlt

A powerful GraphQL CLI tool and Go library for querying GraphQL APIs, schema introspection, and testing. Perfect for automation, seeding, testing, and AI agents.

## Features

- **GraphQL Operations**: Execute queries, mutations, and subscriptions
- **Schema Introspection**: Analyze and explore GraphQL schemas
- **File Uploads**: Support for multipart/form-data file uploads
- **Authentication**: Bearer tokens, Basic auth, and API keys
- **Configuration Management**: Multiple named configurations
- **AI-Friendly**: Structured output, machine-readable errors, quiet mode
- **Testing Utilities**: Mock servers, test helpers, and assertions
- **Library Support**: Use as a Go library in your applications

## Installation

```bash
go install github.com/kluzzebass/gqlt@latest
```

## CLI Usage

### Basic Query

```bash
# Simple query
gqlt run --query '{ users { id name } }' --url https://api.example.com/graphql

# Query with variables
gqlt run --query 'query GetUser($id: ID!) { user(id: $id) { name email } }' \
  --variables '{"id": "123"}' \
  --url https://api.example.com/graphql

# Query from file
gqlt run --query-file query.graphql --url https://api.example.com/graphql
```

### Configuration Management

```bash
# Initialize configuration
gqlt config init

# Create a new configuration
gqlt config create production --endpoint https://api.production.com/graphql

# Switch to a configuration
gqlt config use production

# List configurations
gqlt config list
```

### Schema Introspection

```bash
# Introspect and save schema
gqlt introspect --url https://api.example.com/graphql

# Describe schema types
gqlt describe --url https://api.example.com/graphql
```

### Validation

```bash
# Validate query against schema
gqlt validate query --query '{ users { id } }' --url https://api.example.com/graphql

# Validate configuration
gqlt validate config

# Validate schema
gqlt validate schema --url https://api.example.com/graphql
```

## Library Usage

### Basic Client

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/kluzzebass/gqlt"
)

func main() {
    // Create client
    client := gqlt.NewClient("https://api.example.com/graphql", nil)
    
    // Execute query
    response, err := client.Execute(
        `query GetUsers { users { id name email } }`,
        nil,
        "GetUsers",
    )
    if err != nil {
        log.Fatal(err)
    }
    
    // Check for errors
    if len(response.Errors) > 0 {
        fmt.Printf("GraphQL errors: %v\n", response.Errors)
    }
    
    // Use response data
    fmt.Printf("Response: %+v\n", response.Data)
}
```

### Authentication

```go
// Bearer token authentication
client := gqlt.NewClient("https://api.example.com/graphql", nil)
client.SetHeaders(map[string]string{
    "Authorization": "Bearer your-token",
})

// Basic authentication
client.SetAuth("username", "password")

// API key authentication
client.SetHeaders(map[string]string{
    "X-API-Key": "your-api-key",
})
```

### Schema Introspection

```go
// Create introspection client
client := gqlt.NewClient("https://api.example.com/graphql", nil)
introspect := gqlt.NewIntrospect(client)

// Introspect schema
schema, err := introspect.IntrospectSchema()
if err != nil {
    log.Fatal(err)
}

// Analyze schema
analyzer, err := gqlt.NewAnalyzer(schema)
if err != nil {
    log.Fatal(err)
}

summary, err := analyzer.GetSummary()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Schema has %d types\n", summary.TotalTypes)
fmt.Printf("Query type: %s\n", summary.QueryType)
```

### Configuration Management

```go
// Load configuration
config, err := gqlt.Load("/path/to/config")
if err != nil {
    log.Fatal(err)
}

// Get current configuration
current := config.GetCurrent()
fmt.Printf("Current endpoint: %s\n", current.Endpoint)

// Create new configuration
config.Configs["production"] = gqlt.ConfigEntry{
    Endpoint: "https://api.production.com/graphql",
    Headers: map[string]string{
        "Authorization": "Bearer prod-token",
    },
}

// Save configuration
err = config.Save("/path/to/config")
if err != nil {
    log.Fatal(err)
}
```

### File Uploads

```go
// Execute mutation with file upload
response, err := client.ExecuteWithFiles(
    `mutation UploadFile($file: Upload!) { uploadFile(file: $file) { id } }`,
    map[string]interface{}{"file": nil},
    "UploadFile",
    map[string]string{"file": "/path/to/file.jpg"},
)
```

### Testing with Mock Server

```go
package main

import (
    "testing"
    "github.com/kluzzebass/gqlt"
)

func TestGraphQL(t *testing.T) {
    // Create mock server
    mock := gqlt.NewMockGraphQLServer()
    defer mock.Close()
    
    // Add handler
    mock.AddHandler("GetUsers", func(response *gqlt.Response) {
        response.Data = map[string]interface{}{
            "users": []map[string]interface{}{
                {"id": "1", "name": "John Doe"},
            },
        }
    })
    
    // Test with client
    client := gqlt.NewClient(mock.URL(), nil)
    response, err := client.Execute(
        `query GetUsers { users { id name } }`,
        nil,
        "GetUsers",
    )
    
    if err != nil {
        t.Fatalf("Query failed: %v", err)
    }
    
    // Verify response
    if len(response.Errors) > 0 {
        t.Errorf("Unexpected errors: %v", response.Errors)
    }
}
```

### Test Utilities

```go
func TestWithHelper(t *testing.T) {
    helper := gqlt.NewGraphQLTestHelper(t, "https://api.example.com/graphql")
    
    // Set authentication
    helper.SetAuth("bearer", map[string]string{
        "token": "your-token",
    })
    
    // Execute query
    response := helper.ExecuteQuery("{ users { id name } }", nil, "")
    
    // Use assertions
    helper.AssertNoErrors(response)
    helper.AssertFieldExists(response, "users")
    helper.AssertArrayLength(response, "users", 2)
}
```

## Output Formats

gqlt supports multiple output formats for different use cases:

- **json**: Formatted JSON output (default)
- **pretty**: Colorized JSON output
- **raw**: Compact JSON output
- **table**: Human-readable table format
- **yaml**: YAML format

```bash
# Use different output formats
gqlt run --query '{ users { id } }' --format table
gqlt run --query '{ users { id } }' --format yaml
gqlt run --query '{ users { id } }' --format raw
```

## AI-Friendly Features

gqlt is designed to work well with AI agents and automation:

- **Structured Output**: All commands support `--format json` for machine-readable output
- **Error Codes**: Standardized error codes for programmatic handling
- **Quiet Mode**: `--quiet` flag for automation scenarios
- **Validation Commands**: Structured validation results
- **Configuration**: Self-documenting configuration system

```bash
# AI-friendly usage
gqlt run --query '{ users { id } }' --format json --quiet
gqlt validate query --query '{ users { id } }' --format json
```

## Examples

See the [examples/](examples/) directory for comprehensive examples including:

- Basic GraphQL operations
- Authentication patterns
- Schema introspection
- Configuration management
- Testing utilities
- Mock servers
- Integration tests

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.
