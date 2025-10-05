package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kluzzebass/gqlt/internal/graphql"
)

// Analyzer handles GraphQL schema analysis and description
type Analyzer struct {
	schemaData map[string]interface{}
}

// NewAnalyzer creates a new schema analyzer
func NewAnalyzer(schema *graphql.Response) (*Analyzer, error) {
	// Extract schema data
	schemaData, ok := schema.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid schema data format")
	}

	schemaObj, ok := schemaData["__schema"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid schema format")
	}

	return &Analyzer{
		schemaData: schemaObj,
	}, nil
}

// LoadAnalyzerFromFile creates an analyzer from a schema file
func LoadAnalyzerFromFile(filePath string) (*Analyzer, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file: %w", err)
	}

	var result graphql.Response
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse schema file: %w", err)
	}

	return NewAnalyzer(&result)
}

// GetSummary returns a summary of the schema
func (a *Analyzer) GetSummary() (*Summary, error) {
	types, ok := a.schemaData["types"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid types format")
	}

	queryType, _ := a.schemaData["queryType"].(map[string]interface{})
	mutationType, _ := a.schemaData["mutationType"].(map[string]interface{})
	subscriptionType, _ := a.schemaData["subscriptionType"].(map[string]interface{})

	summary := &Summary{
		TotalTypes: len(types),
	}

	if queryType != nil {
		if name, ok := queryType["name"].(string); ok {
			summary.QueryType = name
		}
	}

	if mutationType != nil {
		if name, ok := mutationType["name"].(string); ok {
			summary.MutationType = name
		}
	}

	if subscriptionType != nil {
		if name, ok := subscriptionType["name"].(string); ok {
			summary.SubscriptionType = name
		}
	}

	return summary, nil
}

// FindType finds a type by name
func (a *Analyzer) FindType(typeName string) (map[string]interface{}, error) {
	types, ok := a.schemaData["types"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid types format")
	}

	for _, t := range types {
		typeObj, ok := t.(map[string]interface{})
		if !ok {
			continue
		}
		if name, ok := typeObj["name"].(string); ok && name == typeName {
			return typeObj, nil
		}
	}

	return nil, fmt.Errorf("type '%s' not found in schema", typeName)
}

// FindField finds a field in a root type
func (a *Analyzer) FindField(rootType, fieldName string) (map[string]interface{}, error) {
	// Find the root type
	rootTypeObj, err := a.FindType(rootType)
	if err != nil {
		return nil, err
	}

	// Find the field
	fields, ok := rootTypeObj["fields"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("type '%s' has no fields", rootType)
	}

	for _, f := range fields {
		fieldObj, ok := f.(map[string]interface{})
		if !ok {
			continue
		}
		if name, ok := fieldObj["name"].(string); ok && name == fieldName {
			return fieldObj, nil
		}
	}

	return nil, fmt.Errorf("field '%s' not found in type '%s'", fieldName, rootType)
}

// FormatTypeDescription formats a type description
func (a *Analyzer) FormatTypeDescription(typeObj map[string]interface{}) (*TypeDescription, error) {
	name, _ := typeObj["name"].(string)
	kind, _ := typeObj["kind"].(string)
	description, _ := typeObj["description"].(string)

	desc := &TypeDescription{
		Name:        name,
		Kind:        kind,
		Description: description,
	}

	// Format fields if available
	if fields, ok := typeObj["fields"].([]interface{}); ok && len(fields) > 0 {
		desc.Fields = make([]FieldSummary, 0, len(fields))
		for _, f := range fields {
			fieldObj, ok := f.(map[string]interface{})
			if !ok {
				continue
			}
			fieldSummary := a.formatFieldSummary(fieldObj)
			desc.Fields = append(desc.Fields, fieldSummary)
		}
	}

	// Format input fields if available
	if inputFields, ok := typeObj["inputFields"].([]interface{}); ok && len(inputFields) > 0 {
		desc.InputFields = make([]FieldSummary, 0, len(inputFields))
		for _, f := range inputFields {
			fieldObj, ok := f.(map[string]interface{})
			if !ok {
				continue
			}
			fieldSummary := a.formatFieldSummary(fieldObj)
			desc.InputFields = append(desc.InputFields, fieldSummary)
		}
	}

	// Format enum values if available
	if enumValues, ok := typeObj["enumValues"].([]interface{}); ok && len(enumValues) > 0 {
		desc.EnumValues = make([]EnumValue, 0, len(enumValues))
		for _, e := range enumValues {
			enumObj, ok := e.(map[string]interface{})
			if !ok {
				continue
			}
			enumName, _ := enumObj["name"].(string)
			enumDesc, _ := enumObj["description"].(string)
			desc.EnumValues = append(desc.EnumValues, EnumValue{
				Name:        enumName,
				Description: enumDesc,
			})
		}
	}

	return desc, nil
}

// FormatFieldDescription formats a field description
func (a *Analyzer) FormatFieldDescription(fieldObj map[string]interface{}, rootType string) (*FieldDescription, error) {
	name, _ := fieldObj["name"].(string)
	description, _ := fieldObj["description"].(string)
	fieldType, _ := fieldObj["type"].(map[string]interface{})

	desc := &FieldDescription{
		RootType:    rootType,
		Name:        name,
		Description: description,
		Type:        a.formatTypeString(fieldType),
	}

	// Format arguments if available
	if args, ok := fieldObj["args"].([]interface{}); ok && len(args) > 0 {
		desc.Arguments = make([]FieldSummary, 0, len(args))
		for _, arg := range args {
			argObj, ok := arg.(map[string]interface{})
			if !ok {
				continue
			}
			argSummary := a.formatFieldSummary(argObj)
			desc.Arguments = append(desc.Arguments, argSummary)
		}
	}

	return desc, nil
}

func (a *Analyzer) formatFieldSummary(fieldObj map[string]interface{}) FieldSummary {
	name, _ := fieldObj["name"].(string)
	description, _ := fieldObj["description"].(string)
	fieldType, _ := fieldObj["type"].(map[string]interface{})
	defaultValue, _ := fieldObj["defaultValue"].(string)

	// Format type
	typeStr := a.formatTypeString(fieldType)

	// Format arguments if available
	argsStr := ""
	if args, ok := fieldObj["args"].([]interface{}); ok && len(args) > 0 {
		var argStrs []string
		for _, arg := range args {
			argObj, ok := arg.(map[string]interface{})
			if !ok {
				continue
			}
			argName, _ := argObj["name"].(string)
			argType, _ := argObj["type"].(map[string]interface{})
			argTypeStr := a.formatTypeString(argType)
			argStrs = append(argStrs, argName+": "+argTypeStr)
		}
		if len(argStrs) > 0 {
			argsStr = "(" + strings.Join(argStrs, ", ") + ")"
		}
	}

	// Build the field signature
	signature := name + argsStr + ": " + typeStr
	if defaultValue != "" {
		signature += " = " + defaultValue
	}

	return FieldSummary{
		Name:         name,
		Description:  description,
		Type:         typeStr,
		Signature:    signature,
		DefaultValue: defaultValue,
	}
}

func (a *Analyzer) formatTypeString(typeObj map[string]interface{}) string {
	if typeObj == nil {
		return "Unknown"
	}

	name, _ := typeObj["name"].(string)
	kind, _ := typeObj["kind"].(string)
	ofType, _ := typeObj["ofType"].(map[string]interface{})

	if ofType != nil {
		// This is a wrapper type (List, NonNull, etc.)
		innerType := a.formatTypeString(ofType)
		switch kind {
		case "LIST":
			return "[" + innerType + "]"
		case "NON_NULL":
			return innerType + "!"
		default:
			return kind + "(" + innerType + ")"
		}
	}

	// Base type
	if name != "" {
		return name
	}
	return kind
}

// Summary represents a schema summary
type Summary struct {
	TotalTypes       int    `json:"totalTypes"`
	QueryType        string `json:"queryType,omitempty"`
	MutationType     string `json:"mutationType,omitempty"`
	SubscriptionType string `json:"subscriptionType,omitempty"`
}

// TypeDescription represents a type description
type TypeDescription struct {
	Name        string         `json:"name"`
	Kind        string         `json:"kind"`
	Description string         `json:"description,omitempty"`
	Fields      []FieldSummary `json:"fields,omitempty"`
	InputFields []FieldSummary `json:"inputFields,omitempty"`
	EnumValues  []EnumValue    `json:"enumValues,omitempty"`
}

// FieldDescription represents a field description
type FieldDescription struct {
	RootType    string         `json:"rootType"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Type        string         `json:"type"`
	Arguments   []FieldSummary `json:"arguments,omitempty"`
}

// FieldSummary represents a field summary
type FieldSummary struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	Type         string `json:"type"`
	Signature    string `json:"signature"`
	DefaultValue string `json:"defaultValue,omitempty"`
}

// EnumValue represents an enum value
type EnumValue struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
