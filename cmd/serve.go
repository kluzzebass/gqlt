package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
	"github.com/kluzzebass/gqlt/internal/mockserver/graph"
	"github.com/spf13/cobra"
	"github.com/vektah/gqlparser/v2/ast"
)

var (
	servePort       string
	servePlayground bool
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a mock GraphQL server for testing",
	Long: `Start a mock GraphQL server with a comprehensive todo-list schema.

The server includes:
- Queries: users, todos, search, and more
- Mutations: create/update/delete users and todos
- Subscriptions: real-time events for todos and users
- File uploads via multipart/form-data
- WebSocket and SSE transport for subscriptions
- GraphQL Playground for interactive testing
- Relay Node pattern for global object identification

The server is pre-seeded with sample data and ready to use immediately.`,
	Example: `  # Start server on default port 8090
  gqlt serve

  # Start on custom port
  gqlt serve --port 3000

  # Start without playground
  gqlt serve --no-playground

  # Test with queries
  gqlt serve &
  gqlt run --url http://localhost:8090/graphql --query '{ users { id name email } }'
  
  # Test subscriptions
  gqlt serve &
  gqlt run --url http://localhost:8090/graphql --query 'subscription { counter }' --timeout 10s`,
	RunE: serve,
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(&servePort, "port", "p", "8090", "Port to listen on")
	serveCmd.Flags().BoolVar(&servePlayground, "playground", true, "Enable GraphQL Playground")
}

func serve(cmd *cobra.Command, args []string) error {
	// Create GraphQL server
	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: graph.NewResolver()}))

	// Add transports - order matters for subscription support
	// Per gqlgen docs: SSE first, then WebSocket, then HTTP transports
	srv.AddTransport(transport.SSE{})
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for testing
				return true
			},
		},
	})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	// Configure caching and extensions
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	// Setup HTTP handlers
	http.Handle("/graphql", srv)

	if servePlayground {
		http.Handle("/", playground.Handler("GraphQL Playground", "/graphql"))
		log.Printf("GraphQL Playground available at http://localhost:%s/", servePort)
	}

	log.Printf("GraphQL endpoint: http://localhost:%s/graphql", servePort)
	log.Printf("Starting mock GraphQL server on port %s...", servePort)

	// Start server
	if err := http.ListenAndServe(":"+servePort, nil); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

