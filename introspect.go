package gqlt

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Introspect handles GraphQL schema introspection operations.
// It provides utilities for introspecting GraphQL schemas and saving them to files.
type Introspect struct {
	client *Client
}

// NewIntrospect creates a new introspection handler for the specified client.
//
// Example:
//   client := gqlt.NewClient("https://api.example.com/graphql", nil)
//   introspect := gqlt.NewIntrospect(client)
func NewIntrospect(client *Client) *Introspect {
	return &Introspect{
		client: client,
	}
}

// IntrospectSchema performs GraphQL introspection to get the schema from the endpoint.
// Returns a Response containing the complete GraphQL schema information.
//
// Example:
//   schema, err := introspect.IntrospectSchema()
//   if err != nil {
//       log.Fatal(err)
//   }
func (i *Introspect) IntrospectSchema() (*Response, error) {
	return i.client.Introspect()
}

// SaveSchema saves a schema response to a JSON file.
// The file will contain the complete introspection response in formatted JSON.
//
// Example:
//   err := introspect.SaveSchema(schema, "schema.json")
//   if err != nil {
//       log.Fatal(err)
//   }
func (i *Introspect) SaveSchema(schema *Response, filePath string) error {
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}
	return os.WriteFile(filePath, data, 0644)
}

// SaveSchemaDual saves schema in both JSON and GraphQL formats
func SaveSchemaDual(result *Response, configName, configDir string) error {
	// Save JSON schema
	jsonPath := getJSONSchemaPathForConfigInDir(configName, configDir)
	if err := SaveSchema(result, jsonPath); err != nil {
		return fmt.Errorf("failed to save JSON schema: %w", err)
	}

	// Convert to GraphQL SDL and save
	graphqlPath := getGraphQLSchemaPathForConfigInDir(configName, configDir)
	if err := SaveGraphQLSchema(result, graphqlPath); err != nil {
		return fmt.Errorf("failed to save GraphQL schema: %w", err)
	}

	return nil
}

// SaveSchema saves schema to a single file
func SaveSchema(result *Response, path string) error {
	// Create output directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Save schema to file
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	return nil
}

// SaveGraphQLSchema saves the schema as GraphQL SDL
func SaveGraphQLSchema(schema *Response, filePath string) error {
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

// SchemaExists checks if a schema file exists
func SchemaExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// convertIntrospectionToSDL converts introspection JSON to GraphQL SDL
func convertIntrospectionToSDL(schema *Response) (string, error) {
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
