package gqlt

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

// FetchSDL attempts to fetch the GraphQL schema in SDL format from common endpoint paths
func (c *Client) FetchSDL() (string, error) {
	// Parse the base URL
	baseURL, err := url.Parse(c.endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint URL: %w", err)
	}

	// Common SDL endpoint paths to try
	sdlPaths := []string{
		"/schema.graphql", // Relative to graphql endpoint
		strings.TrimSuffix(baseURL.Path, "/") + "/schema.graphql", // Relative to current path
		"/graphql/schema.graphql",                                 // Common pattern
		"/sdl",                                                    // Alternative path
	}

	// Try each path
	for _, path := range sdlPaths {
		sdlURL := *baseURL
		sdlURL.Path = path

		req, err := http.NewRequest("GET", sdlURL.String(), nil)
		if err != nil {
			continue
		}

		// Add headers
		for k, v := range c.headers {
			req.Header.Set(k, v)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				continue
			}

			sdl := string(body)
			// Basic validation - check if it looks like SDL
			if strings.Contains(sdl, "type") || strings.Contains(sdl, "schema") {
				return sdl, nil
			}
		}
	}

	return "", fmt.Errorf("could not fetch SDL from any common paths")
}

// SDLToIntrospection converts SDL schema text to introspection JSON format
func SDLToIntrospection(sdl string) (interface{}, error) {
	// Parse the SDL
	schema, err := gqlparser.LoadSchema(&ast.Source{
		Name:  "schema.graphql",
		Input: sdl,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse SDL: %w", err)
	}

	// Convert to introspection format
	introspection := map[string]interface{}{
		"__schema": schemaToIntrospection(schema),
	}

	return introspection, nil
}

// schemaToIntrospection converts a gqlparser schema to introspection format
func schemaToIntrospection(schema *ast.Schema) map[string]interface{} {
	result := map[string]interface{}{}

	// Query type
	if schema.Query != nil {
		result["queryType"] = map[string]interface{}{
			"name": schema.Query.Name,
		}
	}

	// Mutation type
	if schema.Mutation != nil {
		result["mutationType"] = map[string]interface{}{
			"name": schema.Mutation.Name,
		}
	}

	// Subscription type
	if schema.Subscription != nil {
		result["subscriptionType"] = map[string]interface{}{
			"name": schema.Subscription.Name,
		}
	}

	// Types
	types := []interface{}{}
	for _, typeDef := range schema.Types {
		types = append(types, typeToIntrospection(typeDef))
	}
	result["types"] = types

	// Directives
	directives := []interface{}{}
	for _, directive := range schema.Directives {
		directives = append(directives, directiveToIntrospection(directive))
	}
	result["directives"] = directives

	return result
}

// typeToIntrospection converts a gqlparser type definition to introspection format
func typeToIntrospection(typeDef *ast.Definition) map[string]interface{} {
	result := map[string]interface{}{
		"kind":        string(typeDef.Kind),
		"name":        typeDef.Name,
		"description": typeDef.Description,
	}

	// Fields (for OBJECT and INTERFACE types)
	if typeDef.Kind == ast.Object || typeDef.Kind == ast.Interface {
		fields := []interface{}{}
		for _, field := range typeDef.Fields {
			fields = append(fields, fieldToIntrospection(field))
		}
		result["fields"] = fields

		// Interfaces (for OBJECT types)
		if typeDef.Kind == ast.Object {
			interfaces := []interface{}{}
			for _, iface := range typeDef.Interfaces {
				interfaces = append(interfaces, typeRefToIntrospection(ast.NamedType(iface, nil)))
			}
			result["interfaces"] = interfaces
		}
	}

	// Input fields (for INPUT_OBJECT types)
	if typeDef.Kind == ast.InputObject {
		inputFields := []interface{}{}
		for _, field := range typeDef.Fields {
			inputFields = append(inputFields, inputValueToIntrospection(field))
		}
		result["inputFields"] = inputFields
	}

	// Enum values (for ENUM types)
	if typeDef.Kind == ast.Enum {
		enumValues := []interface{}{}
		for _, val := range typeDef.EnumValues {
			enumValues = append(enumValues, map[string]interface{}{
				"name":              val.Name,
				"description":       val.Description,
				"isDeprecated":      val.Directives.ForName("deprecated") != nil,
				"deprecationReason": getDeprecationReason(val.Directives),
			})
		}
		result["enumValues"] = enumValues
	}

	// Possible types (for UNION types)
	if typeDef.Kind == ast.Union {
		possibleTypes := []interface{}{}
		for _, unionType := range typeDef.Types {
			possibleTypes = append(possibleTypes, map[string]interface{}{
				"kind": "OBJECT",
				"name": unionType,
			})
		}
		result["possibleTypes"] = possibleTypes
	}

	return result
}

// fieldToIntrospection converts a field definition to introspection format
func fieldToIntrospection(field *ast.FieldDefinition) map[string]interface{} {
	args := []interface{}{}
	for _, arg := range field.Arguments {
		args = append(args, argumentToIntrospection(arg))
	}

	return map[string]interface{}{
		"name":              field.Name,
		"description":       field.Description,
		"args":              args,
		"type":              typeRefToIntrospection(field.Type),
		"isDeprecated":      field.Directives.ForName("deprecated") != nil,
		"deprecationReason": getDeprecationReason(field.Directives),
	}
}

// inputValueToIntrospection converts an input field to introspection format
func inputValueToIntrospection(field *ast.FieldDefinition) map[string]interface{} {
	result := map[string]interface{}{
		"name":        field.Name,
		"description": field.Description,
		"type":        typeRefToIntrospection(field.Type),
	}

	if field.DefaultValue != nil {
		result["defaultValue"] = field.DefaultValue.String()
	}

	return result
}

// argumentToIntrospection converts an argument to introspection format
func argumentToIntrospection(arg *ast.ArgumentDefinition) map[string]interface{} {
	result := map[string]interface{}{
		"name":        arg.Name,
		"description": arg.Description,
		"type":        typeRefToIntrospection(arg.Type),
	}

	if arg.DefaultValue != nil {
		result["defaultValue"] = arg.DefaultValue.String()
	}

	return result
}

// typeRefToIntrospection converts a type reference to introspection format
func typeRefToIntrospection(typeRef *ast.Type) map[string]interface{} {
	if typeRef.NonNull {
		return map[string]interface{}{
			"kind":   "NON_NULL",
			"name":   nil,
			"ofType": typeRefToIntrospection(&ast.Type{NamedType: typeRef.NamedType, Elem: typeRef.Elem}),
		}
	}

	if typeRef.Elem != nil {
		return map[string]interface{}{
			"kind":   "LIST",
			"name":   nil,
			"ofType": typeRefToIntrospection(typeRef.Elem),
		}
	}

	return map[string]interface{}{
		"kind":   getTypeKind(typeRef.NamedType),
		"name":   typeRef.NamedType,
		"ofType": nil,
	}
}

// directiveToIntrospection converts a directive to introspection format
func directiveToIntrospection(directive *ast.DirectiveDefinition) map[string]interface{} {
	args := []interface{}{}
	for _, arg := range directive.Arguments {
		args = append(args, argumentToIntrospection(arg))
	}

	locations := []string{}
	for _, loc := range directive.Locations {
		locations = append(locations, string(loc))
	}

	return map[string]interface{}{
		"name":        directive.Name,
		"description": directive.Description,
		"locations":   locations,
		"args":        args,
	}
}

// getTypeKind returns the introspection kind for a named type
func getTypeKind(typeName string) string {
	// Built-in scalars
	switch typeName {
	case "Int", "Float", "String", "Boolean", "ID":
		return "SCALAR"
	}
	// Default to OBJECT - will be corrected by actual type definition
	return "OBJECT"
}

// getDeprecationReason extracts the deprecation reason from directives
func getDeprecationReason(directives ast.DirectiveList) string {
	deprecated := directives.ForName("deprecated")
	if deprecated == nil {
		return ""
	}

	reason := deprecated.Arguments.ForName("reason")
	if reason == nil || reason.Value == nil {
		return "No longer supported"
	}

	return reason.Value.String()
}
