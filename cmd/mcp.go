package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kluzzebass/gqlt"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP (Model Context Protocol) server for AI agent integration",
	Long: `Start an MCP server that provides GraphQL query execution and schema exploration to AI agents via JSON-RPC 2.0.
This allows AI agents to execute GraphQL queries, introspect schemas, and explore types
through a standardized protocol using stdin/stdout communication.

The MCP server provides tools for:
- execute_query: Run GraphQL queries, mutations, and subscriptions (supports file uploads via local paths)
- describe_type: Analyze specific GraphQL types and fields with detailed information
- list_types: List GraphQL type names with optional regex filtering
- version: Get the current version of gqlt`,
	Example: `# Start MCP server (stdin/stdout mode)
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
}`,
	RunE: runMCPServer,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

func runMCPServer(cmd *cobra.Command, args []string) error {
	// Create MCP server using official SDK
	server, err := gqlt.NewSDKServer()
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
		fmt.Println("Starting MCP server (stdin/stdout mode)")
		fmt.Println("Press Ctrl+C to stop the server")
		serverErr <- server.Start(context.Background(), "")
	}()

	// Wait for server to start or error
	select {
	case err := <-serverErr:
		if err != nil {
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
