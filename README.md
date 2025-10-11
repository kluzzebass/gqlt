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
```

**MCP Server Usage:**
```bash
# Start MCP server for AI agents
gqlt mcp

# Add to Cursor's mcp.json or Claude Desktop config:
# {
#   "mcpServers": {
#     "gqlt": {
#       "command": "gqlt",
#       "args": ["mcp"]
#     }
#   }
# }
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

## Documentation


## Completion


Generate the autocompletion script for the specified shell

### Synopsis

Generate the autocompletion script for gqlt for the specified shell.
See each sub-command's help for details on how to use the generated script.
### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Completion Bash


Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(gqlt completion bash)

To load completions for every new session, execute once:

#### Linux:

	gqlt completion bash > /etc/bash_completion.d/gqlt

#### macOS:

	gqlt completion bash > $(brew --prefix)/etc/bash_completion.d/gqlt

You will need to start a new shell for this setup to take effect.
```
gqlt completion bash
```

### Options

```
  -h, --help              help for bash
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Completion Fish


Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	gqlt completion fish | source

To load completions for every new session, execute once:

	gqlt completion fish > ~/.config/fish/completions/gqlt.fish

You will need to start a new shell for this setup to take effect.
```
gqlt completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Completion Powershell


Generate the autocompletion script for powershell

### Synopsis

Generate the autocompletion script for powershell.

To load completions in your current shell session:

	gqlt completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.
```
gqlt completion powershell [flags]
```

### Options

```
  -h, --help              help for powershell
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Completion Zsh


Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(gqlt completion zsh)

To load completions for every new session, execute once:

#### Linux:

	gqlt completion zsh > "${fpath[1]}/_gqlt"

#### macOS:

	gqlt completion zsh > $(brew --prefix)/share/zsh/site-functions/_gqlt

You will need to start a new shell for this setup to take effect.
```
gqlt completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Config


Manage gqlt configuration files

### Synopsis

Manage gqlt configuration files with support for multiple named configurations.
This allows you to store different settings for different environments (production, staging, local, etc.).

AI-FRIENDLY FEATURES:
- Structured output with --format json|table|yaml
- Machine-readable error codes
- Quiet mode for automation

### Examples

```
# Initialize and setup
gqlt config init
gqlt config create production
gqlt config set production endpoint https://api.prod.com/graphql
gqlt config set production headers '{"Authorization": "Bearer token"}'
gqlt config use production

# Environment management
gqlt config create staging
gqlt config set staging endpoint https://api.staging.com/graphql
gqlt config create local
gqlt config set local endpoint http://localhost:4000/graphql

# Configuration inspection
gqlt config list --format table
gqlt config show production --format json
gqlt config validate

# With authentication
gqlt config set myapi auth.token "your-bearer-token"
gqlt config set myapi auth.username "username"
gqlt config set myapi auth.password "password"
gqlt config set myapi auth.api_key "api-key"

# Clone configuration
gqlt config clone production staging

# Structured output for AI agents
gqlt config list --format json --quiet
gqlt config show --format yaml
```

### Options

```
  -h, --help   help for config
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Config Clone


Clone an existing configuration

### Synopsis

Create a new configuration by copying an existing one.

```
gqlt config clone <source> <target> [flags]
```

### Options

```
  -h, --help   help for clone
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Config Create


Create a new configuration

### Synopsis

Create a new named configuration with default values.

```
gqlt config create <name> [flags]
```

### Examples

```
gqlt config create production
gqlt config create staging
gqlt config create local
```

### Options

```
  -h, --help   help for create
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Config Delete


Delete a configuration

### Synopsis

Delete a named configuration (cannot delete default).

```
gqlt config delete <name> [flags]
```

### Options

```
  -h, --help   help for delete
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Config Init


Initialize configuration file

### Synopsis

Create a new configuration file with default settings.

```
gqlt config init [flags]
```

### Examples

```
gqlt config init
```

### Options

```
  -h, --help   help for init
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Config List


List all configurations

### Synopsis

List all available configurations with their current status.

```
gqlt config list [flags]
```

### Options

```
  -h, --help   help for list
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Config Set


Set a configuration value

### Synopsis

Set a configuration value for a named configuration.

Available keys:
  endpoint                    - GraphQL endpoint URL (required)
  headers.<name>              - Custom HTTP header (e.g., headers.X-Custom "value")
  headers.Authorization       - Authorization header (e.g., "Bearer token")
  headers.X-API-Key           - API key header
  auth.token                  - Bearer token for authentication
  auth.username               - Username for basic authentication
  auth.password               - Password for basic authentication
  auth.api_key                - API key for authentication
  defaults.out                - Default output mode (json|pretty|raw)

Authentication precedence:
  1. Basic auth (auth.username + auth.password)
  2. Bearer token (auth.token)
  3. API key (auth.api_key)
  4. Custom headers (headers.Authorization, headers.X-API-Key)

```
gqlt config set <name> <key> <value> [flags]
```

### Examples

```
# Basic configuration
gqlt config set production endpoint https://api.example.com/graphql
gqlt config set production defaults.out pretty

# Authentication methods
gqlt config set production auth.token "your-bearer-token"
gqlt config set production auth.username "admin"
gqlt config set production auth.password "secret"
gqlt config set production auth.api_key "api-key-123"

# Custom headers
gqlt config set production headers.X-Custom "custom-value"
gqlt config set production headers.Authorization "Bearer manual-token"
gqlt config set production headers.X-API-Key "manual-api-key"
```

### Options

```
  -h, --help   help for set
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Config Show


Show current or named configuration

### Synopsis

Show the current configuration or a specific named configuration.

```
gqlt config show [name] [flags]
```

### Options

```
  -h, --help   help for show
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Config Use


Switch to a configuration

### Synopsis

Switch the current active configuration to the specified name.

```
gqlt config use <name> [flags]
```

### Options

```
  -h, --help   help for use
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Config Validate


Validate configuration

### Synopsis

Check the configuration file for errors and provide suggestions.

```
gqlt config validate [flags]
```

### Options

```
  -h, --help   help for validate
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Describe


Describe GraphQL types and fields from cached schema

### Synopsis

Describe GraphQL types and fields from a cached schema file.
Use this to explore the GraphQL schema structure.

```
gqlt describe [type|field] [flags]
```

### Examples

```
# Describe a type
gqlt describe User

# Describe a field
gqlt describe Query.users

# Describe with JSON output
gqlt describe User --json

# Show summary only
gqlt describe User --summary
```

### Options

```
  -h, --help            help for describe
      --json            output exact node JSON
      --schema string   schema file path (default is OS-specific)
      --summary         output plain text summary
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Docs


Generate documentation

### Synopsis

Generate documentation in various formats from the command structure

```
gqlt docs [flags]
```

### Examples

```
# Generate README.md
gqlt docs --format md --output README.md

# Generate man pages
gqlt docs --format man --output man/

# Generate multiple markdown files
gqlt docs --format md --tree --output docs/

# Output to stdout
gqlt docs --format md --output -
```

### Options

```
  -f, --format string   Output format: md or man (default "md")
  -h, --help            help for docs
  -o, --output string   Output destination (file for md, directory for man, '-' for stdout) (default "-")
      --tree            Generate multiple files (one per command) instead of single file
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Introspect


Fetch and cache GraphQL schema via introspection

### Synopsis

Fetch the GraphQL schema from an endpoint using introspection
and save it to a local cache file for use with other commands.

```
gqlt introspect [flags]
```

### Examples

```
# Fetch schema from URL
gqlt introspect --url https://api.example.com/graphql

# Fetch schema with authentication
gqlt introspect --url https://api.example.com/graphql --token "bearer-token"

# Force refresh cached schema
gqlt introspect --refresh

# Show schema summary
gqlt introspect --summary

# Save to specific file
gqlt introspect --output schema.json
```

### Options

```
  -h, --help         help for introspect
      --out string   output file path (default is OS-specific)
      --refresh      ignore cache and fetch fresh schema
      --summary      show summary instead of saving to file
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Mcp


Start MCP (Model Context Protocol) server for AI agent integration

### Synopsis

Start an MCP server that provides GraphQL query execution and schema exploration to AI agents via JSON-RPC 2.0.
This allows AI agents to execute GraphQL queries, introspect schemas, and explore types
through a standardized protocol using stdin/stdout communication.

The MCP server provides tools for:
- execute_query: Run GraphQL queries, mutations, and subscriptions (supports file uploads via local paths)
- describe_type: Analyze specific GraphQL types and fields with detailed information
- list_types: List GraphQL type names with optional regex filtering
- version: Get the current version of gqlt

```
gqlt mcp [flags]
```

### Examples

```
# Start MCP server (stdin/stdout mode)
gqlt mcp

# For Cursor integration, add to mcp.json:
{
  "mcpServers": {
    "gqlt": {
      "command": "gqlt",
      "args": ["mcp"],
      "env": {}
    }
  }
}

# For Claude Desktop, add to ~/.config/claude-desktop/config.json:
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

### Options

```
  -h, --help   help for mcp
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Run


Execute a GraphQL operation against an endpoint

### Synopsis

Execute a GraphQL operation (query or mutation) against a GraphQL endpoint.
You can provide the query inline, from a file, or via stdin.

```
gqlt run [flags]
```

### Examples

```
# Basic query
gqlt run --url https://api.example.com/graphql --query "{ users { id name } }"

# Query with variables
gqlt run --url https://api.example.com/graphql --query "query($id: ID!) { user(id: $id) { name } }" --vars '{"id": "123"}'

# Query from stdin
echo "{ users { id name } }" | gqlt run --url https://api.example.com/graphql

# Mutation with file upload
gqlt run --url https://api.example.com/graphql --query "mutation($file: Upload!) { uploadFile(file: $file) }" --file avatar=./photo.jpg

# Using configuration
gqlt run --query "{ users { id name } }"  # Uses configured endpoint

# Authentication (precedence: Basic Auth > Bearer Token > API Key)
gqlt run --username user --password pass --query "{ me { id } }"  # Basic auth (highest precedence)
gqlt run --token "bearer-token" --query "{ me { id } }"          # Bearer token
gqlt run --api-key "api-key" --query "{ me { id } }"             # API key (lowest precedence)

# Structured output for AI agents
gqlt run --format json --quiet --query "{ users { id } }"

# Multiple file uploads
gqlt run --query "mutation($files: [Upload!]!) { uploadFiles(files: $files) }" --files-list files.txt
```

### Options

```
  -k, --api-key string       API key for authentication (sets X-API-Key header)
  -f, --file stringArray     File upload (name=path, repeatable, e.g. avatar=./photo.jpg)
  -F, --files-list string    File containing list of files to upload (one per line, format: name=path, supports # comments, ~ expansion, and relative paths)
  -H, --header stringArray   HTTP header (key=value, repeatable)
  -h, --help                 help for run
  -o, --operation string     Operation name
  -O, --out string           Output mode: json|pretty|raw (default "json")
  -p, --password string      Password for basic authentication
  -q, --query string         Inline GraphQL document
  -Q, --query-file string    Path to .graphql file
  -t, --token string         Bearer token for authentication
  -u, --url string           GraphQL endpoint URL (required if not in config)
  -U, --username string      Username for basic authentication
  -v, --vars string          JSON object with variables
  -V, --vars-file string     Path to JSON file with variables
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Validate


Validate GraphQL queries, schemas, and configurations

### Synopsis

Validate GraphQL queries, schemas, and configurations.
This command provides structured validation results for AI agents and automation.

AI-FRIENDLY FEATURES:
- Structured JSON output with validation results
- Machine-readable error codes
- Detailed validation information
- Quiet mode for automation

### Examples

```
# Validate a query against a schema
gqlt validate query --query "{ users { id name } }" --url https://api.example.com/graphql

# Validate query from file
gqlt validate query --query-file query.graphql --url https://api.example.com/graphql

# Validate configuration
gqlt validate config

# Validate schema
gqlt validate schema --url https://api.example.com/graphql

# Structured output for AI agents
gqlt validate query --query "{ users { id } }" --format json --quiet
```

### Options

```
  -h, --help   help for validate
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Validate Config


Validate configuration files

### Synopsis

Validate configuration files for syntax and completeness.
Returns structured validation results with detailed error information.

```
gqlt validate config [flags]
```

### Examples

```
gqlt validate config
gqlt validate config --format json --quiet
```

### Options

```
  -h, --help   help for config
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Validate Query


Validate a GraphQL query against a schema

### Synopsis

Validate a GraphQL query against a schema.
Returns structured validation results including syntax errors, type errors, and field availability.

```
gqlt validate query [flags]
```

### Examples

```
gqlt validate query --query "{ users { id name } }" --url https://api.example.com/graphql
gqlt validate query --query-file query.graphql --url https://api.example.com/graphql
gqlt validate query --query "{ users { id } }" --format json --quiet
```

### Options

```
  -h, --help                help for query
  -q, --query string        GraphQL query to validate
  -Q, --query-file string   Path to GraphQL query file
  -u, --url string          GraphQL endpoint URL
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Validate Schema


Validate a GraphQL schema

### Synopsis

Validate a GraphQL schema for correctness and completeness.
Returns structured validation results with schema analysis.

```
gqlt validate schema [flags]
```

### Examples

```
gqlt validate schema --url https://api.example.com/graphql
gqlt validate schema --url https://api.example.com/graphql --format json --quiet
```

### Options

```
  -h, --help         help for schema
  -u, --url string   GraphQL endpoint URL
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Version


Display the current version of gqlt

### Synopsis

Display the current version of gqlt.

```
gqlt version [flags]
```

### Examples

```
# Show version
gqlt version

# Use in scripts
VERSION=$(gqlt version)
echo "Using gqlt version: $VERSION"
```

### Options

```
  -h, --help   help for version
```

### Options inherited from parent commands

```
      --config-dir string   config directory (default is OS-specific)
      --format string       Output format: json|table|yaml (default: json) (default "json")
      --quiet               Quiet mode - suppress non-essential output for automation
      --use-config string   use specific configuration by name (overrides current selection)
```

## Limitations

gqlt is designed to be a focused, composable tool. The following are intentional limitations:

**GraphQL Subscriptions:**
- gqlt does not support GraphQL subscriptions over WebSockets
- Subscriptions require persistent connections incompatible with:
  - CLI's request/response model
  - MCP's synchronous tool call pattern
  - Unix philosophy of composable, discrete operations
- For subscription support, use a full-featured GraphQL client library

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
