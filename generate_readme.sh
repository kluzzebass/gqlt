#!/bin/bash

# Generate comprehensive README.md from markdown documentation tree
# This script creates a single README.md file by combining all command documentation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Generating comprehensive README.md...${NC}"

# Build the binary first
echo -e "${YELLOW}Building gqlt...${NC}"
mkdir -p dist
CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/gqlt ./cmd

# Generate markdown tree to temp directory
TEMP_DIR=$(mktemp -d)
echo -e "${YELLOW}Generating markdown tree to $TEMP_DIR...${NC}"
./dist/gqlt docs --format md --tree --output "$TEMP_DIR"

# Create comprehensive README.md
echo -e "${YELLOW}Creating comprehensive README.md...${NC}"
cat > README.md << 'EOF'
# gqlt

A triple-threat GraphQL tool: CLI client, MCP server, and Go library.

## Overview

**gqlt** operates in three distinct modes:

### 1. **CLI Mode**: Composable Unix Tool
A minimal command-line client for running GraphQL operations that follows Unix philosophy:
- Accepts input from stdin, files, or arguments
- Outputs structured data (JSON/YAML/table) to stdout
- Composes with other Unix tools via pipes (`jq`, `grep`, `awk`, etc.)
- Supports queries, mutations, subscriptions, and introspection

### 2. **MCP Mode**: AI Agent Server
A Model Context Protocol (MCP) server that provides GraphQL capabilities to AI agents:
- Execute queries against any GraphQL endpoint
- Introspect and explore GraphQL schemas
- List and describe types with intelligent filtering
- Check version information
- Integrates with Cursor, Claude Desktop, and other MCP-compatible tools

### 3. **Library Mode**: Go Package
A clean, testable Go library for GraphQL operations in your own applications:
- Pure functions with no side effects (`import "github.com/kluzzebass/gqlt"`)
- Embed GraphQL client in your Go applications
- Perfect for testing GraphQL integrations
- Comprehensive API with full type safety
- Mock server infrastructure included (`gqlt/internal/testutil`)

### Quick Start

**CLI Usage:**
```bash
# Basic query execution
gqlt run --url https://api.example.com/graphql --query "{ users { id name } }"

# Compose with jq to extract data
gqlt run --query "{ users { id name email } }" --format json --quiet | \
  jq '.data.users[] | select(.email | contains("@example.com"))'

# Using configuration
gqlt config create production
gqlt config set production endpoint https://api.example.com/graphql
gqlt run --query "{ users { id name } }"

# Check version
gqlt version

# GraphQL subscriptions (WebSocket)
gqlt run --url wss://api.example.com/graphql --query 'subscription { events { id type data } }'

# Limit subscription messages
gqlt run --url wss://api.example.com/graphql \
  --query 'subscription { updates }' \
  --max-messages 10 \
  --timeout 30s
```

**Mock Server Usage:**
```bash
# Start comprehensive mock GraphQL server for testing
gqlt serve

# Custom address
gqlt serve --listen :3000

# Test against the mock server
gqlt serve &
gqlt run --url http://localhost:8090/graphql --query '{ users { id name email role } }'
gqlt run --url http://localhost:8090/graphql --query 'subscription { counter }' --timeout 5s
```

**MCP Server Usage:**
```bash
# Start MCP server for AI agents
gqlt mcp
```

Add to Cursor's `mcp.json` or Claude Desktop config:
```json
{
  "mcpServers": {
    "gqlt": {
      "command": "gqlt",
      "args": ["mcp"]
    }
  }
}
```

**Library Usage:**
```go
import "github.com/kluzzebass/gqlt"

// Create client and execute query
client := gqlt.NewClient("https://api.example.com/graphql", nil)
response, err := client.Execute(`query { users { id name } }`, nil, "")

// Use mock server for testing
import "github.com/kluzzebass/gqlt/internal/testutil"

mockServer := testutil.NewMockGraphQLServer()
defer mockServer.Close()
mockServer.AddHandler("GetUsers", func(req testutil.Request) *gqlt.Response {
    return testutil.SuccessResponse(map[string]interface{}{
        "users": []map[string]interface{}{{"id": "1", "name": "Test User"}},
    })
})
```

## Installation

### CLI Binary

**From Source:**
```bash
git clone https://github.com/kluzzebass/gqlt.git
cd gqlt
go build -o gqlt ./cmd
```

**From Releases:**
Download the latest release for your platform from the [releases page](https://github.com/kluzzebass/gqlt/releases).

### Go Library

```bash
go get github.com/kluzzebass/gqlt
```

## The Trinity: Three Tools in One

### Mode 1: Composable CLI

The CLI mode outputs clean, structured data that composes naturally with Unix tools:

```bash
# Extract and filter user emails
gqlt run --query "{ users { id name email } }" --format json --quiet | \
  jq -r '.data.users[] | select(.email | contains("@example.com")) | .email'

# Count total types in a schema
gqlt introspect --format json --quiet | jq '.data.__schema.types | length'

# Find types matching a pattern
gqlt introspect --format json --quiet | \
  jq -r '.data.__schema.types[].name' | grep -i "user"

# Chain multiple queries
user_id=$(gqlt run --query "{ users { id } }" --format json --quiet | jq -r '.data.users[0].id')
gqlt run --query "query(\$id: ID!) { user(id: \$id) { name email } }" \
  --vars "{\"id\": \"$user_id\"}" --format json --quiet

# GraphQL subscriptions (streamed output)
gqlt run --url wss://api.example.com/graphql \
  --query 'subscription { events { id type data } }' | jq .

# Limit subscription duration and message count
gqlt run --url wss://api.example.com/graphql \
  --query 'subscription { updates }' \
  --max-messages 10 --timeout 1m | jq -c '.data.updates'
```

### Mode 2: MCP Server for AI Agents

Run as a server to provide GraphQL capabilities to AI agents:

```bash
# Start the MCP server (stdin/stdout mode)
gqlt mcp
```

**Integration with Cursor:**

Add to your `~/.cursor/mcp.json` or workspace `.cursor/mcp.json`:
```json
{
  "mcpServers": {
    "gqlt": {
      "command": "gqlt",
      "args": ["mcp"],
      "env": {}
    }
  }
}
```

**Integration with Claude Desktop:**

Add to `~/.config/claude-desktop/config.json` (macOS/Linux) or `%APPDATA%\Claude\config.json` (Windows):
```json
{
  "mcpServers": {
    "gqlt": {
      "command": "/usr/local/bin/gqlt",
      "args": ["mcp"],
      "env": {}
    }
  }
}
```

**Available MCP Tools:**
- `execute_query`: Run GraphQL queries, mutations, and subscriptions (supports file uploads)
- `describe_type`: Analyze specific GraphQL types with detailed field information
- `list_types`: List and filter GraphQL type names (supports regex patterns and kind filtering)
- `version`: Get the current version of gqlt

**File Upload Support:**
The `execute_query` tool supports file uploads for mutations with Upload scalar types. Provide local file paths:

```json
{
  "query": "mutation($file: Upload!) { uploadFile(file: $file) { id url } }",
  "variables": {"file": null},
  "endpoint": "https://api.example.com/graphql",
  "files": {
    "file": "/Users/you/photos/avatar.jpg"
  }
}
```

**Subscription Support:**
The `execute_query` tool supports GraphQL subscriptions via WebSocket. Use `timeout` and `maxMessages` to control execution:

```json
{
  "query": "subscription { events { id type data } }",
  "endpoint": "wss://api.example.com/graphql",
  "timeout": "30s",
  "maxMessages": 10
}
```

Subscriptions will:
- Automatically convert `https://` URLs to `wss://` and `http://` to `ws://`
- Stream messages until timeout, maxMessages limit, or Cursor cancellation
- Return each message as a complete GraphQL response
- Clean up WebSocket connection on completion or error

**Tool Parameters:**
- Schema-related tools (`describe_type` and `list_types`) support `noCache` parameter to force fresh schema introspection
- `execute_query` supports `files` parameter for file uploads via local filesystem paths

### Mode 3: Go Library

Use gqlt as a library in your own Go applications:

```go
package main

import (
    "fmt"
    "github.com/kluzzebass/gqlt"
)

func main() {
    // Create a GraphQL client
    client := gqlt.NewClient("https://api.example.com/graphql", nil)
    
    // Set authentication
    client.SetHeaders(map[string]string{
        "Authorization": "Bearer your-token",
    })
    
    // Execute a query
    response, err := client.Execute(
        `query GetUsers { users { id name email } }`,
        nil,
        "GetUsers",
    )
    
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Response: %+v\n", response.Data)
}
```

**Testing with Mock Server:**

```go
package yourpackage_test

import (
    "testing"
    "github.com/kluzzebass/gqlt"
    "github.com/kluzzebass/gqlt/internal/testutil"
)

func TestYourGraphQLIntegration(t *testing.T) {
    // Create mock GraphQL server
    mockServer := testutil.NewMockGraphQLServer()
    defer mockServer.Close()
    
    // Configure mock responses
    mockServer.AddHandler("GetUser", func(req testutil.Request) *gqlt.Response {
        userID := req.Variables["id"].(string)
        return testutil.SuccessResponse(map[string]interface{}{
            "user": map[string]interface{}{
                "id":   userID,
                "name": "Test User",
                "email": "test@example.com",
            },
        })
    })
    
    // Use the mock server in your tests
    client := gqlt.NewClient(mockServer.URL(), nil)
    response, err := client.Execute(
        `query GetUser($id: ID!) { user(id: $id) { id name email } }`,
        map[string]interface{}{"id": "123"},
        "GetUser",
    )
    
    if err != nil {
        t.Fatalf("Query failed: %v", err)
    }
    
    // Validate response
    data := response.Data.(map[string]interface{})
    user := data["user"].(map[string]interface{})
    
    if user["id"] != "123" {
        t.Errorf("Expected user id 123, got %v", user["id"])
    }
}
```

## Mock GraphQL Server

**gqlt** includes a comprehensive mock GraphQL server powered by `gqlgen` for testing and development:

### Features
- **Complete Schema**: Todo-list application with users, todos, attachments, and search
- **All GraphQL Features**: Queries, mutations, subscriptions, unions, interfaces, custom scalars, directives
- **Relay Node Pattern**: Global object identification with `node(id: "Type:ID")` queries
- **Real-Time Subscriptions**: WebSocket and SSE support with event broadcasting
- **File Uploads**: Multipart form-data for testing file upload mutations
- **Pre-Seeded Data**: 3 sample users ready to use immediately
- **Introspection Enabled**: Full schema introspection for tooling

### Quick Start

```bash
# Start mock server (default: localhost:8090)
gqlt serve

# Custom address
gqlt serve --listen :3000
gqlt serve --listen 0.0.0.0:8080

# Without GraphQL Playground
gqlt serve --no-playground
```

### Example Usage

```bash
# Start server in background
gqlt serve &

# Query pre-seeded users
gqlt run --url http://localhost:8090/graphql \
  --query '{ users { id name email role } }'

# Create a todo
gqlt run --url http://localhost:8090/graphql \
  --query 'mutation { createTodo(input: { title: "Test Todo" }) { id title status } }'

# Subscribe to real-time todo events
gqlt run --url http://localhost:8090/graphql \
  --query 'subscription { todoEvents { id title status } }' \
  --timeout 30s &

# Trigger events by creating todos
gqlt run --url http://localhost:8090/graphql \
  --query 'mutation { createTodo(input: { title: "Event Test" }) { id } }'

# Search across users and todos (union types)
gqlt run --url http://localhost:8090/graphql \
  --query '{ search(term: "admin") { ... on User { id name } ... on Todo { id title } } }'

# Test Relay Node pattern
gqlt run --url http://localhost:8090/graphql \
  --query '{ node(id: "User:1") { id ... on User { name email } } }'
```

### Schema Highlights

- **Queries**: `hello`, `echo`, `user`, `users`, `todo`, `todos`, `search`, `currentTime`, `version`, `node`
- **Mutations**: `createUser`, `createTodo`, `updateTodo`, `deleteTodo`, `completeTodo`, `addFileAttachment`, `addLinkAttachment`, `removeAttachment`
- **Subscriptions**: `counter`, `todoEvents`, `userEvents`, `tick`
- **Types**: User, Todo, FileAttachment, LinkAttachment (with Node interface)
- **Unions**: SearchResult (User | Todo | Post | Product | Service)
- **Enums**: UserRole, TodoStatus, TodoPriority
- **Custom Scalars**: DateTime, URL, Upload

Access the GraphQL Playground at `http://localhost:8090/` to explore the complete schema interactively.

## Documentation

EOF

# Add each command's documentation to README.md
for md_file in "$TEMP_DIR"/*.md; do
    if [ -f "$md_file" ]; then
        filename=$(basename "$md_file")
        command_name="${filename%.md}"
        
        # Skip the root command file (gqlt.md) as we already have overview
        if [ "$command_name" = "gqlt" ]; then
            continue
        fi
        
        echo -e "${YELLOW}Adding $command_name documentation...${NC}"
        
        # Convert command name to readable format
        # Remove gqlt_ prefix and convert to Title Case
        readable_name=$(echo "$command_name" | sed -E 's/^gqlt_([a-z]+)(_[a-z]+)*/\1\2/' | sed 's/_/ /g' | awk '{for(i=1;i<=NF;i++){$i=toupper(substr($i,1,1)) substr($i,2)};print}')
        
        # Add command section header
        echo "" >> README.md
        echo "## $readable_name" >> README.md
        echo "" >> README.md
        
        # Extract content from the markdown file, cleaning up artifacts
        tail -n +2 "$md_file" | \
        sed '/^### SEE ALSO$/,$d' | \
        sed '/^###### Auto generated by spf13\/cobra/d' | \
        sed '/^$/N;/^\n$/d' >> README.md
    fi
done

# Add footer
cat >> README.md << 'EOF'

## Limitations

gqlt is designed to be a focused, composable tool. The following are intentional limitations:

**GraphQL Subscriptions:**
- Full WebSocket subscription support via graphql-transport-ws protocol
- CLI: Streams messages to stdout (one JSON per line) until Ctrl+C or completion
- MCP: Collects messages with timeout/max-messages limits, respects Cancel button
- Library: Channel-based API for testing subscriptions
- Auto-detects subscription operations and routes to WebSocket
- Works with mixed operation documents (queries + mutations + subscriptions)

**Response Filtering:**
- gqlt outputs raw GraphQL responses without built-in filtering
- This is intentional - filtering should be done with specialized tools like `jq`
- Example: `gqlt run ... | jq '.data.users[] | select(.active)'`
- Follows Unix philosophy: do one thing well, compose with other tools

**Schema Features:**
- SDL fallback supports standard GraphQL schemas
- Some server-specific features may not be fully represented in introspection format
- Custom scalars are preserved but not validated beyond GraphQL spec

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

*This documentation is auto-generated from the command structure. Last updated: $(date)*
EOF

# Clean up temp directory
rm -rf "$TEMP_DIR"

echo -e "${GREEN}✅ Comprehensive README.md generated successfully!${NC}"
echo -e "${GREEN}📄 README.md now contains all command documentation${NC}"
