package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kluzzebass/gqlt/mcp"
	"github.com/spf13/cobra"
)

var (
	mcpAddress string
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP (Model Context Protocol) server for AI agent integration",
	Long: `Start an MCP server that exposes gqlt functionality to AI agents via JSON-RPC 2.0.
This allows AI agents to execute GraphQL queries, introspect schemas, and manage configurations
through a standardized protocol.

The MCP server provides tools for:
- execute_query: Run GraphQL queries, mutations, and subscriptions
- introspect_schema: Get complete GraphQL schema information
- describe_type: Analyze specific GraphQL types and fields
- validate_query: Check GraphQL query validity against schema
- upload_files: Handle file uploads in GraphQL mutations
- get_config: Retrieve and manage GraphQL endpoint configurations
- set_auth: Configure authentication (Bearer, Basic, API Key)

And resources for:
- config://current: Current GraphQL configuration
- config://list: List of all configurations
- schema://introspection: Complete GraphQL schema

The server also provides prompts for common GraphQL operations and patterns.

INTEGRATION WITH AI CLIENTS:

Cursor Integration:
1. Start the gqlt MCP server:
   gqlt mcp --address localhost:8080

2. Add to Cursor MCP settings:
   - Go to Cursor menu → "Settings" → "Cursor Settings" → "Tools & MCP"
   - Choose "New MCP Server"
   - Edit the mcp.json file that pops up:
   {
     "mcpServers": {
       "gqlt": {
         "url": "http://localhost:8080/mcp"
       }
     }
   }

3. Restart Cursor to load the MCP server

Claude Desktop Integration:
1. Start the gqlt MCP server:
   gqlt mcp --address localhost:8080

2. Add to Claude Desktop config (~/.config/claude-desktop/config.json):
   {
     "mcpServers": {
       "gqlt": {
         "command": "gqlt",
         "args": ["mcp", "--address", "localhost:8080"],
         "env": {}
       }
     }
   }

3. Restart Claude Desktop

Other MCP Clients:
- Use the server URL: http://localhost:8080/mcp
- The server implements JSON-RPC 2.0 over HTTP POST
- All requests should be sent to the /mcp endpoint`,
	Example: `# Start MCP server on default address
gqlt mcp

# Start MCP server on specific address
gqlt mcp --address localhost:8080

# Start MCP server with custom config directory
gqlt mcp --config-dir /custom/config/dir

# Start MCP server with specific configuration
gqlt mcp --use-config production`,
	RunE: runMCPServer,
}

func init() {
	rootCmd.AddCommand(mcpCmd)

	mcpCmd.Flags().StringVarP(&mcpAddress, "address", "a", "localhost:8080", "Address to listen on")
}

func runMCPServer(cmd *cobra.Command, args []string) error {
	// Load configuration using the same system as other commands
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// If a specific config is requested, switch to it
	if configName != "" {
		if err := cfg.SetCurrent(configName); err != nil {
			return fmt.Errorf("failed to switch to config '%s': %w", configName, err)
		}
	}

	// Create MCP server with the loaded config using official SDK
	server, err := mcp.NewSDKServer(cfg)
	if err != nil {
		return fmt.Errorf("failed to create MCP server: %w", err)
	}

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down MCP server...")
		cancel()
	}()

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		fmt.Printf("Starting MCP server on %s\n", mcpAddress)
		fmt.Println("Press Ctrl+C to stop the server")
		serverErr <- server.Start(context.Background(), mcpAddress)
	}()

	// Wait for server to start or error
	select {
	case err := <-serverErr:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("MCP server error: %w", err)
		}
	case <-ctx.Done():
		// Graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := server.Stop(shutdownCtx); err != nil {
			return fmt.Errorf("failed to stop MCP server: %w", err)
		}
	}

	return nil
}
