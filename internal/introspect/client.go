package introspect

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kluzzebass/gqlt/internal/graphql"
	"github.com/kluzzebass/gqlt/internal/paths"
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

// SaveSchemaDual saves both JSON and GraphQL schema files
func SaveSchemaDual(schema *graphql.Response, configName, configDir string) error {
	// Save JSON schema
	jsonPath := paths.GetJSONSchemaPathForConfigInDir(configName, configDir)
	if err := SaveSchema(schema, jsonPath); err != nil {
		return fmt.Errorf("failed to save JSON schema: %w", err)
	}

	// Convert to GraphQL SDL and save
	graphqlPath := paths.GetGraphQLSchemaPathForConfigInDir(configName, configDir)
	if err := SaveGraphQLSchema(schema, graphqlPath); err != nil {
		return fmt.Errorf("failed to save GraphQL schema: %w", err)
	}

	return nil
}

// SaveGraphQLSchema saves the schema as GraphQL SDL
func SaveGraphQLSchema(schema *graphql.Response, filePath string) error {
	// Create output directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Convert JSON introspection to GraphQL SDL
	sdl, err := convertIntrospectionToSDL(schema)
	if err != nil {
		return fmt.Errorf("failed to convert introspection to SDL: %w", err)
	}

	// Save GraphQL schema to file
	if err := os.WriteFile(filePath, []byte(sdl), 0644); err != nil {
		return fmt.Errorf("failed to write GraphQL schema file: %w", err)
	}

	return nil
}

// convertIntrospectionToSDL converts introspection JSON to GraphQL SDL
func convertIntrospectionToSDL(schema *graphql.Response) (string, error) {
	// Extract schema data
	schemaData, ok := schema.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid schema data format")
	}

	schemaObj, ok := schemaData["__schema"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid schema format")
	}

	// Build SDL from introspection data
	var sdl strings.Builder

	// Add types
	types, ok := schemaObj["types"].([]interface{})
	if !ok {
		return "", fmt.Errorf("invalid types format")
	}

	for _, typeObj := range types {
		typeMap, ok := typeObj.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := typeMap["name"].(string)
		kind, _ := typeMap["kind"].(string)
		description, _ := typeMap["description"].(string)

		// Skip introspection types
		if strings.HasPrefix(name, "__") {
			continue
		}

		// Add description if present
		if description != "" {
			sdl.WriteString(fmt.Sprintf("\"\"\"%s\"\"\"\n", description))
		}

		// Add type definition based on kind
		switch kind {
		case "OBJECT":
			sdl.WriteString(fmt.Sprintf("type %s {\n", name))
			// Add fields
			if fields, ok := typeMap["fields"].([]interface{}); ok {
				for _, field := range fields {
					if fieldMap, ok := field.(map[string]interface{}); ok {
						fieldName, _ := fieldMap["name"].(string)
						fieldType := formatType(fieldMap["type"])
						fieldDesc, _ := fieldMap["description"].(string)

						if fieldDesc != "" {
							sdl.WriteString(fmt.Sprintf("  \"\"\"%s\"\"\"\n", fieldDesc))
						}
						sdl.WriteString(fmt.Sprintf("  %s: %s\n", fieldName, fieldType))
					}
				}
			}
			sdl.WriteString("}\n\n")
		case "INPUT_OBJECT":
			sdl.WriteString(fmt.Sprintf("input %s {\n", name))
			// Add input fields
			if inputFields, ok := typeMap["inputFields"].([]interface{}); ok {
				for _, field := range inputFields {
					if fieldMap, ok := field.(map[string]interface{}); ok {
						fieldName, _ := fieldMap["name"].(string)
						fieldType := formatType(fieldMap["type"])
						fieldDesc, _ := fieldMap["description"].(string)

						if fieldDesc != "" {
							sdl.WriteString(fmt.Sprintf("  \"\"\"%s\"\"\"\n", fieldDesc))
						}
						sdl.WriteString(fmt.Sprintf("  %s: %s\n", fieldName, fieldType))
					}
				}
			}
			sdl.WriteString("}\n\n")
		case "ENUM":
			sdl.WriteString(fmt.Sprintf("enum %s {\n", name))
			// Add enum values
			if enumValues, ok := typeMap["enumValues"].([]interface{}); ok {
				for _, value := range enumValues {
					if valueMap, ok := value.(map[string]interface{}); ok {
						valueName, _ := valueMap["name"].(string)
						valueDesc, _ := valueMap["description"].(string)

						if valueDesc != "" {
							sdl.WriteString(fmt.Sprintf("  \"\"\"%s\"\"\"\n", valueDesc))
						}
						sdl.WriteString(fmt.Sprintf("  %s\n", valueName))
					}
				}
			}
			sdl.WriteString("}\n\n")
		case "SCALAR":
			sdl.WriteString(fmt.Sprintf("scalar %s\n\n", name))
		}
	}

	return sdl.String(), nil
}

// formatType formats a GraphQL type from introspection data
func formatType(typeObj interface{}) string {
	if typeMap, ok := typeObj.(map[string]interface{}); ok {
		kind, _ := typeMap["kind"].(string)
		name, _ := typeMap["name"].(string)
		ofType := typeMap["ofType"]

		switch kind {
		case "NON_NULL":
			return formatType(ofType) + "!"
		case "LIST":
			return "[" + formatType(ofType) + "]"
		case "SCALAR", "OBJECT", "ENUM", "INPUT_OBJECT":
			if name != "" {
				return name
			}
		}
	}
	return "String" // fallback
}
