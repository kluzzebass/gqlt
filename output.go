package gqlt

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
)

// Formatter defines the interface for output formatting
type Formatter interface {
	FormatStructured(data interface{}, quiet bool) error
	FormatStructuredError(err error, code string, quiet bool) error
	FormatStructuredErrorWithContext(err error, code string, errorType string, context map[string]interface{}, quiet bool) error
	FormatResponse(response *Response, mode string) error
}

// FormatterFactory creates a new formatter instance
type FormatterFactory func() Formatter

// FormatterRegistry manages available formatters
type FormatterRegistry struct {
	formatters map[string]FormatterFactory
}

// NewFormatterRegistry creates a new formatter registry with default formatters
func NewFormatterRegistry() *FormatterRegistry {
	registry := &FormatterRegistry{
		formatters: make(map[string]FormatterFactory),
	}

	// Register default formatters
	registry.Register("json", func() Formatter { return &JSONFormatter{} })
	registry.Register("table", func() Formatter { return &TableFormatter{} })
	registry.Register("yaml", func() Formatter { return &YAMLFormatter{} })

	return registry
}

// Register adds a new formatter to the registry
func (r *FormatterRegistry) Register(name string, factory FormatterFactory) {
	r.formatters[name] = factory
}

// Get creates a new formatter instance for the specified format
func (r *FormatterRegistry) Get(format string) (Formatter, error) {
	factory, exists := r.formatters[format]
	if !exists {
		return nil, fmt.Errorf("unknown formatter: %s", format)
	}
	return factory(), nil
}

// List returns all registered formatter names
func (r *FormatterRegistry) List() []string {
	names := make([]string, 0, len(r.formatters))
	for name := range r.formatters {
		names = append(names, name)
	}
	return names
}

// Global formatter registry
var defaultRegistry = NewFormatterRegistry()

// NewFormatter creates a new formatter using the default registry
// Returns the default JSON formatter if the requested format is not found
func NewFormatter(format string) Formatter {
	formatter, err := defaultRegistry.Get(format)
	if err != nil {
		// Fall back to JSON formatter for unknown formats
		formatter, _ = defaultRegistry.Get("json")
	}
	return formatter
}

// RegisterFormatter registers a custom formatter with the default registry
func RegisterFormatter(name string, factory FormatterFactory) {
	defaultRegistry.Register(name, factory)
}

// GetAvailableFormatters returns all available formatter names
func GetAvailableFormatters() []string {
	return defaultRegistry.List()
}

// JSONFormatter implements Formatter for JSON output
type JSONFormatter struct{}

// TableFormatter implements Formatter for table output
type TableFormatter struct{}

// YAMLFormatter implements Formatter for YAML output
type YAMLFormatter struct{}

// StructuredOutput represents a structured response for AI agents
type StructuredOutput struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *MetaInfo   `json:"meta,omitempty"`
}

// ErrorInfo provides structured error information
type ErrorInfo struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details string                 `json:"details,omitempty"`
	Type    string                 `json:"type,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// Common error codes for AI agents
const (
	// Configuration errors
	ErrorCodeConfigLoad     = "CONFIG_LOAD_ERROR"
	ErrorCodeConfigNotFound = "CONFIG_NOT_FOUND"
	ErrorCodeConfigCreate   = "CONFIG_CREATE_ERROR"
	ErrorCodeConfigSave     = "CONFIG_SAVE_ERROR"
	ErrorCodeConfigDelete   = "CONFIG_DELETE_ERROR"
	ErrorCodeConfigValidate = "CONFIG_VALIDATE_ERROR"

	// Input validation errors
	ErrorCodeInputValidation = "INPUT_VALIDATION_ERROR"
	ErrorCodeQueryLoad       = "QUERY_LOAD_ERROR"
	ErrorCodeVariablesLoad   = "VARIABLES_LOAD_ERROR"
	ErrorCodeFilesParse      = "FILES_PARSE_ERROR"
	ErrorCodeFilesListParse  = "FILES_LIST_PARSE_ERROR"

	// GraphQL execution errors
	ErrorCodeGraphQLExecution = "GRAPHQL_EXECUTION_ERROR"
	ErrorCodeGraphQLErrors    = "GRAPHQL_ERRORS"
	ErrorCodeNetworkError     = "NETWORK_ERROR"
	ErrorCodeAuthError        = "AUTH_ERROR"

	// Schema errors
	ErrorCodeSchemaLoad       = "SCHEMA_LOAD_ERROR"
	ErrorCodeSchemaIntrospect = "SCHEMA_INTROSPECT_ERROR"
	ErrorCodeSchemaSave       = "SCHEMA_SAVE_ERROR"

	// System errors
	ErrorCodeSystemError      = "SYSTEM_ERROR"
	ErrorCodeFileNotFound     = "FILE_NOT_FOUND"
	ErrorCodePermissionDenied = "PERMISSION_DENIED"
)

// MetaInfo provides metadata about the operation
type MetaInfo struct {
	Command   string                 `json:"command,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
	Duration  string                 `json:"duration,omitempty"`
	Config    string                 `json:"config,omitempty"`
	Endpoint  string                 `json:"endpoint,omitempty"`
	Operation string                 `json:"operation,omitempty"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// JSONFormatter implementation

// FormatStructured formats data as structured JSON output
func (f *JSONFormatter) FormatStructured(data interface{}, quiet bool) error {
	output := &StructuredOutput{
		Success: true,
		Data:    data,
	}
	return f.formatStructuredJSON(output)
}

// FormatStructuredError formats an error as structured output
func (f *JSONFormatter) FormatStructuredError(err error, code string, quiet bool) error {
	output := &StructuredOutput{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: err.Error(),
		},
	}
	return f.formatStructuredJSON(output)
}

// FormatStructuredErrorWithContext formats an error with additional context
func (f *JSONFormatter) FormatStructuredErrorWithContext(err error, code string, errorType string, context map[string]interface{}, quiet bool) error {
	output := &StructuredOutput{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: err.Error(),
			Type:    errorType,
			Context: context,
		},
	}
	return f.formatStructuredJSON(output)
}

// FormatResponse formats a GraphQL response
func (f *JSONFormatter) FormatResponse(response *Response, mode string) error {
	switch mode {
	case "json":
		return f.formatJSON(response)
	case "pretty":
		return f.formatPretty(response)
	case "raw":
		return f.formatRaw(response)
	default:
		return fmt.Errorf("unknown output mode: %s", mode)
	}
}

func (f *JSONFormatter) formatStructuredJSON(output *StructuredOutput) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// TableFormatter implementation

// FormatStructured formats data as structured table output
func (f *TableFormatter) FormatStructured(data interface{}, quiet bool) error {
	output := &StructuredOutput{
		Success: true,
		Data:    data,
	}
	return f.formatStructuredTable(output, quiet)
}

// FormatStructuredError formats an error as structured table output
func (f *TableFormatter) FormatStructuredError(err error, code string, quiet bool) error {
	output := &StructuredOutput{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: err.Error(),
		},
	}
	return f.formatStructuredTable(output, quiet)
}

// FormatStructuredErrorWithContext formats an error with additional context
func (f *TableFormatter) FormatStructuredErrorWithContext(err error, code string, errorType string, context map[string]interface{}, quiet bool) error {
	output := &StructuredOutput{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: err.Error(),
			Type:    errorType,
			Context: context,
		},
	}
	return f.formatStructuredTable(output, quiet)
}

// FormatResponse formats a GraphQL response
func (f *TableFormatter) FormatResponse(response *Response, mode string) error {
	// Table formatter doesn't support GraphQL response modes, fall back to JSON
	jsonFormatter := &JSONFormatter{}
	return jsonFormatter.FormatResponse(response, mode)
}

func (f *TableFormatter) formatStructuredTable(output *StructuredOutput, quiet bool) error {
	if quiet {
		// In quiet mode, just show the data or error message
		if output.Success {
			if output.Data != nil {
				fmt.Println(output.Data)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\n", output.Error.Message)
		}
		return nil
	}

	// Full table output
	if output.Success {
		fmt.Println("✓ Success")
		if output.Data != nil {
			fmt.Printf("Data: %v\n", output.Data)
		}
	} else {
		fmt.Println("✗ Error")
		fmt.Printf("Code: %s\n", output.Error.Code)
		if output.Error.Type != "" {
			fmt.Printf("Type: %s\n", output.Error.Type)
		}
		fmt.Printf("Message: %s\n", output.Error.Message)
		if output.Error.Details != "" {
			fmt.Printf("Details: %s\n", output.Error.Details)
		}
		if len(output.Error.Context) > 0 {
			fmt.Println("Context:")
			for key, value := range output.Error.Context {
				fmt.Printf("  %s: %v\n", key, value)
			}
		}
	}

	if output.Meta != nil {
		fmt.Println("\nMetadata:")
		if output.Meta.Command != "" {
			fmt.Printf("  Command: %s\n", output.Meta.Command)
		}
		if output.Meta.Config != "" {
			fmt.Printf("  Config: %s\n", output.Meta.Config)
		}
		if output.Meta.Endpoint != "" {
			fmt.Printf("  Endpoint: %s\n", output.Meta.Endpoint)
		}
		if output.Meta.Operation != "" {
			fmt.Printf("  Operation: %s\n", output.Meta.Operation)
		}
	}

	return nil
}

// YAMLFormatter implementation

// FormatStructured formats data as structured YAML output
func (f *YAMLFormatter) FormatStructured(data interface{}, quiet bool) error {
	output := &StructuredOutput{
		Success: true,
		Data:    data,
	}
	return f.formatStructuredYAML(output)
}

// FormatStructuredError formats an error as structured YAML output
func (f *YAMLFormatter) FormatStructuredError(err error, code string, quiet bool) error {
	output := &StructuredOutput{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: err.Error(),
		},
	}
	return f.formatStructuredYAML(output)
}

// FormatStructuredErrorWithContext formats an error with additional context
func (f *YAMLFormatter) FormatStructuredErrorWithContext(err error, code string, errorType string, context map[string]interface{}, quiet bool) error {
	output := &StructuredOutput{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: err.Error(),
			Type:    errorType,
			Context: context,
		},
	}
	return f.formatStructuredYAML(output)
}

// FormatResponse formats a GraphQL response
func (f *YAMLFormatter) FormatResponse(response *Response, mode string) error {
	// YAML formatter doesn't support GraphQL response modes, fall back to JSON
	jsonFormatter := &JSONFormatter{}
	return jsonFormatter.FormatResponse(response, mode)
}

func (f *YAMLFormatter) formatStructuredYAML(output *StructuredOutput) error {
	// Simple YAML-like output
	if output.Success {
		fmt.Println("success: true")
		if output.Data != nil {
			fmt.Printf("data: %v\n", output.Data)
		}
	} else {
		fmt.Println("success: false")
		fmt.Printf("error:\n")
		fmt.Printf("  code: %s\n", output.Error.Code)
		if output.Error.Type != "" {
			fmt.Printf("  type: %s\n", output.Error.Type)
		}
		fmt.Printf("  message: %s\n", output.Error.Message)
		if output.Error.Details != "" {
			fmt.Printf("  details: %s\n", output.Error.Details)
		}
		if len(output.Error.Context) > 0 {
			fmt.Printf("  context:\n")
			for key, value := range output.Error.Context {
				fmt.Printf("    %s: %v\n", key, value)
			}
		}
	}

	if output.Meta != nil {
		fmt.Println("meta:")
		if output.Meta.Command != "" {
			fmt.Printf("  command: %s\n", output.Meta.Command)
		}
		if output.Meta.Config != "" {
			fmt.Printf("  config: %s\n", output.Meta.Config)
		}
		if output.Meta.Endpoint != "" {
			fmt.Printf("  endpoint: %s\n", output.Meta.Endpoint)
		}
		if output.Meta.Operation != "" {
			fmt.Printf("  operation: %s\n", output.Meta.Operation)
		}
	}

	return nil
}

// FormatResponse formats and prints the GraphQL response
// This old FormatResponse method is no longer needed - it's been moved to JSONFormatter

// formatJSON outputs formatted JSON
func (f *JSONFormatter) formatJSON(response *Response) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}

// formatPretty outputs colorized formatted JSON
func (f *JSONFormatter) formatPretty(response *Response) error {
	// Color definitions
	dataColor := color.New(color.FgGreen, color.Bold)
	errorColor := color.New(color.FgRed, color.Bold)
	keyColor := color.New(color.FgBlue, color.Bold)
	stringColor := color.New(color.FgYellow)
	numberColor := color.New(color.FgCyan)
	boolColor := color.New(color.FgMagenta)

	fmt.Print("{\n")

	// Data field
	if response.Data != nil {
		dataColor.Print("  \"data\": ")
		f.printValue(response.Data, "  ", dataColor, errorColor, keyColor, stringColor, numberColor, boolColor)
		fmt.Print(",\n")
	}

	// Errors field
	if len(response.Errors) > 0 {
		errorColor.Print("  \"errors\": ")
		f.printValue(response.Errors, "  ", dataColor, errorColor, keyColor, stringColor, numberColor, boolColor)
		fmt.Print(",\n")
	}

	// Extensions field
	if len(response.Extensions) > 0 {
		keyColor.Print("  \"extensions\": ")
		f.printValue(response.Extensions, "  ", dataColor, errorColor, keyColor, stringColor, numberColor, boolColor)
		fmt.Print(",\n")
	}

	fmt.Print("}\n")
	return nil
}

// formatRaw outputs unformatted JSON
func (f *JSONFormatter) formatRaw(response *Response) error {
	encoder := json.NewEncoder(os.Stdout)
	return encoder.Encode(response)
}

// printValue recursively prints a value with colors
func (f *JSONFormatter) printValue(v interface{}, indent string, dataColor, errorColor, keyColor, stringColor, numberColor, boolColor *color.Color) {
	switch val := v.(type) {
	case map[string]interface{}:
		fmt.Print("{\n")
		first := true
		for k, v := range val {
			if !first {
				fmt.Print(",\n")
			}
			first = false
			keyColor.Printf("%s  \"%s\": ", indent, k)
			f.printValue(v, indent+"  ", dataColor, errorColor, keyColor, stringColor, numberColor, boolColor)
		}
		fmt.Printf("\n%s}", indent)
	case []interface{}:
		fmt.Print("[\n")
		for i, item := range val {
			if i > 0 {
				fmt.Print(",\n")
			}
			fmt.Printf("%s  ", indent)
			f.printValue(item, indent+"  ", dataColor, errorColor, keyColor, stringColor, numberColor, boolColor)
		}
		fmt.Printf("\n%s]", indent)
	case string:
		stringColor.Printf("\"%s\"", val)
	case float64:
		numberColor.Print(val)
	case bool:
		boolColor.Print(val)
	case nil:
		fmt.Print("null")
	default:
		fmt.Printf("%v", val)
	}
}

// FormatJSON formats data as JSON with indentation
func (f *JSONFormatter) FormatJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// FormatPretty formats data as colorized JSON
func (f *JSONFormatter) FormatPretty(data interface{}) error {
	// For now, just use regular JSON formatting
	// TODO: Implement colorized formatting for arbitrary data
	return f.FormatJSON(data)
}

// WriteToFile writes data to a file
func (f *JSONFormatter) WriteToFile(data interface{}, filename string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return os.WriteFile(filename, jsonData, 0644)
}
