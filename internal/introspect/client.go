package introspect

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kluzzebass/gqlt/internal/graphql"
)

// Client handles GraphQL schema introspection
type Client struct {
	graphqlClient *graphql.Client
}

// NewClient creates a new introspection client
func NewClient(graphqlClient *graphql.Client) *Client {
	return &Client{
		graphqlClient: graphqlClient,
	}
}

// IntrospectQuery represents the introspection query
const IntrospectQuery = `query IntrospectionQuery {
	__schema {
		types {
			name
			kind
			description
			fields {
				name
				description
				type {
					name
					kind
					ofType {
						name
						kind
					}
				}
				args {
					name
					description
					type {
						name
						kind
						ofType {
							name
							kind
						}
					}
					defaultValue
				}
			}
			inputFields {
				name
				description
				type {
					name
					kind
					ofType {
						name
						kind
					}
				}
				defaultValue
			}
			enumValues {
				name
				description
			}
		}
		queryType {
			name
		}
		mutationType {
			name
		}
		subscriptionType {
			name
		}
	}
}`

// FetchSchema executes the introspection query and returns the schema
func (c *Client) FetchSchema() (*graphql.Response, error) {
	result, err := c.graphqlClient.Execute(IntrospectQuery, nil, "IntrospectionQuery")
	if err != nil {
		return nil, fmt.Errorf("failed to execute introspection query: %w", err)
	}

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("introspection query failed: %v", result.Errors)
	}

	return result, nil
}

// SaveSchema saves the schema to a file
func SaveSchema(schema *graphql.Response, filePath string) error {
	// Create output directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Save schema to file
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	return nil
}

// LoadSchema loads a schema from a file
func LoadSchema(filePath string) (*graphql.Response, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	var result graphql.Response
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse schema file: %w", err)
	}

	return &result, nil
}

// SchemaExists checks if a schema file exists
func SchemaExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
