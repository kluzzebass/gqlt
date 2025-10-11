package gqlt

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Formatter defines the interface for output formatting.
// Implementations can format data as JSON, table, YAML, or other formats.
type Formatter interface {
	FormatStructured(data interface{}, quiet bool) error
	FormatStructuredError(err error, code string, quiet bool) error
	FormatStructuredErrorWithContext(err error, code string, errorType string, context map[string]interface{}, quiet bool) error
	FormatResponse(response *Response, mode string) error
	SetOutput(writer io.Writer)
	SetErrorOutput(writer io.Writer)
}

// FormatterFactory creates a new formatter instance.
// This function type is used to register formatters in the registry.
type FormatterFactory func() Formatter

// FormatterRegistry manages available formatters and provides a way to register
// and retrieve formatters by name.
type FormatterRegistry struct {
	formatters map[string]FormatterFactory
}

// NewFormatterRegistry creates a new formatter registry with default formatters
// (JSON, Table, YAML) already registered.
//
// Example:
//
//	registry := gqlt.NewFormatterRegistry()
//	formatter, err := registry.Get("json")
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
		// Return nil for unknown formats
		return nil
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
type JSONFormatter struct {
	output      io.Writer
	errorOutput io.Writer
}

// SetOutput sets the output writer for the formatter
func (f *JSONFormatter) SetOutput(writer io.Writer) {
	f.output = writer
}

// SetErrorOutput sets the error output writer for the formatter
func (f *JSONFormatter) SetErrorOutput(writer io.Writer) {
	f.errorOutput = writer
}

// getOutput returns the output writer, defaulting to os.Stdout if not set
func (f *JSONFormatter) getOutput() io.Writer {
	if f.output != nil {
		return f.output
	}
	return os.Stdout
}

// getErrorOutput returns the error output writer, defaulting to os.Stderr if not set
func (f *JSONFormatter) getErrorOutput() io.Writer {
	if f.errorOutput != nil {
		return f.errorOutput
	}
	return os.Stderr
}

// TableFormatter implements Formatter for table output
type TableFormatter struct {
	output      io.Writer
	errorOutput io.Writer
}

// SetOutput sets the output writer for the formatter
func (f *TableFormatter) SetOutput(writer io.Writer) {
	f.output = writer
}

// SetErrorOutput sets the error output writer for the formatter
func (f *TableFormatter) SetErrorOutput(writer io.Writer) {
	f.errorOutput = writer
}

// getOutput returns the output writer, defaulting to os.Stdout if not set
func (f *TableFormatter) getOutput() io.Writer {
	if f.output != nil {
		return f.output
	}
	return os.Stdout
}

// getErrorOutput returns the error output writer, defaulting to os.Stderr if not set
func (f *TableFormatter) getErrorOutput() io.Writer {
	if f.errorOutput != nil {
		return f.errorOutput
	}
	return os.Stderr
}

// YAMLFormatter implements Formatter for YAML output
type YAMLFormatter struct {
	output      io.Writer
	errorOutput io.Writer
}

// SetOutput sets the output writer for the formatter
func (f *YAMLFormatter) SetOutput(writer io.Writer) {
	f.output = writer
}

// SetErrorOutput sets the error output writer for the formatter
func (f *YAMLFormatter) SetErrorOutput(writer io.Writer) {
	f.errorOutput = writer
}

// getOutput returns the output writer, defaulting to os.Stdout if not set
func (f *YAMLFormatter) getOutput() io.Writer {
	if f.output != nil {
		return f.output
	}
	return os.Stdout
}

// getErrorOutput returns the error output writer, defaulting to os.Stderr if not set
func (f *YAMLFormatter) getErrorOutput() io.Writer {
	if f.errorOutput != nil {
		return f.errorOutput
	}
	return os.Stderr
}

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
	message := ""
	if err != nil {
		message = err.Error()
	}

	output := &StructuredOutput{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
	return f.formatStructuredJSONToError(output)
}

// FormatStructuredErrorWithContext formats an error with additional context
func (f *JSONFormatter) FormatStructuredErrorWithContext(err error, code string, errorType string, context map[string]interface{}, quiet bool) error {
	message := ""
	if err != nil {
		message = err.Error()
	}

	output := &StructuredOutput{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Type:    errorType,
			Context: context,
		},
	}
	return f.formatStructuredJSONToError(output)
}

// FormatResponse formats a GraphQL response
func (f *JSONFormatter) FormatResponse(response *Response, mode string) error {
	switch mode {
	case "json":
		return f.formatJSON(response)
	case "raw":
		return f.formatRaw(response)
	default:
		return fmt.Errorf("unknown output mode: %s (valid modes: json, raw)", mode)
	}
}

func (f *JSONFormatter) formatStructuredJSON(output *StructuredOutput) error {
	encoder := json.NewEncoder(f.getOutput())
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

func (f *JSONFormatter) formatStructuredJSONToError(output *StructuredOutput) error {
	encoder := json.NewEncoder(f.getErrorOutput())
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
	message := ""
	if err != nil {
		message = err.Error()
	}

	output := &StructuredOutput{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
	return f.formatStructuredTableToError(output, quiet)
}

// FormatStructuredErrorWithContext formats an error with additional context
func (f *TableFormatter) FormatStructuredErrorWithContext(err error, code string, errorType string, context map[string]interface{}, quiet bool) error {
	message := ""
	if err != nil {
		message = err.Error()
	}

	output := &StructuredOutput{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Type:    errorType,
			Context: context,
		},
	}
	return f.formatStructuredTableToError(output, quiet)
}

// FormatResponse formats a GraphQL response
func (f *TableFormatter) FormatResponse(response *Response, mode string) error {
	// Table formatter doesn't support GraphQL response modes, fall back to JSON
	jsonFormatter := &JSONFormatter{}
	jsonFormatter.SetOutput(f.getOutput())
	jsonFormatter.SetErrorOutput(f.getErrorOutput())
	return jsonFormatter.FormatResponse(response, mode)
}

func (f *TableFormatter) formatStructuredTable(output *StructuredOutput, quiet bool) error {
	if quiet {
		// In quiet mode, just show the data or error message
		if output.Success {
			if output.Data != nil {
				fmt.Fprintln(f.getOutput(), output.Data)
			}
		} else {
			fmt.Fprintf(f.getErrorOutput(), "Error: %s\n", output.Error.Message)
		}
		return nil
	}

	// Full table output
	if output.Success {
		fmt.Fprintln(f.getOutput(), "✓ Success")
		if output.Data != nil {
			fmt.Fprintf(f.getOutput(), "Data: %v\n", output.Data)
		}
	} else {
		fmt.Fprintln(f.getOutput(), "✗ Error")
		fmt.Fprintf(f.getOutput(), "Code: %s\n", output.Error.Code)
		if output.Error.Type != "" {
			fmt.Fprintf(f.getOutput(), "Type: %s\n", output.Error.Type)
		}
		fmt.Fprintf(f.getOutput(), "Message: %s\n", output.Error.Message)
		if output.Error.Details != "" {
			fmt.Fprintf(f.getOutput(), "Details: %s\n", output.Error.Details)
		}
		if len(output.Error.Context) > 0 {
			fmt.Fprintln(f.getOutput(), "Context:")
			for key, value := range output.Error.Context {
				fmt.Fprintf(f.getOutput(), "  %s: %v\n", key, value)
			}
		}
	}

	if output.Meta != nil {
		fmt.Fprintln(f.getOutput(), "\nMetadata:")
		if output.Meta.Command != "" {
			fmt.Fprintf(f.getOutput(), "  Command: %s\n", output.Meta.Command)
		}
		if output.Meta.Config != "" {
			fmt.Fprintf(f.getOutput(), "  Config: %s\n", output.Meta.Config)
		}
		if output.Meta.Endpoint != "" {
			fmt.Fprintf(f.getOutput(), "  Endpoint: %s\n", output.Meta.Endpoint)
		}
		if output.Meta.Operation != "" {
			fmt.Fprintf(f.getOutput(), "  Operation: %s\n", output.Meta.Operation)
		}
	}

	return nil
}

func (f *TableFormatter) formatStructuredTableToError(output *StructuredOutput, quiet bool) error {
	if quiet {
		// In quiet mode, just show the error message to error output
		if !output.Success {
			fmt.Fprintf(f.getErrorOutput(), "Error: %s\n", output.Error.Message)
		}
		return nil
	}

	// Full table output to error stream
	if !output.Success {
		fmt.Fprintln(f.getErrorOutput(), "✗ Error")
		fmt.Fprintf(f.getErrorOutput(), "Code: %s\n", output.Error.Code)
		if output.Error.Type != "" {
			fmt.Fprintf(f.getErrorOutput(), "Type: %s\n", output.Error.Type)
		}
		fmt.Fprintf(f.getErrorOutput(), "Message: %s\n", output.Error.Message)
		if output.Error.Details != "" {
			fmt.Fprintf(f.getErrorOutput(), "Details: %s\n", output.Error.Details)
		}
		if len(output.Error.Context) > 0 {
			fmt.Fprintf(f.getErrorOutput(), "Context:\n")
			for key, value := range output.Error.Context {
				fmt.Fprintf(f.getErrorOutput(), "  %s: %v\n", key, value)
			}
		}
	}

	if output.Meta != nil {
		fmt.Fprintln(f.getErrorOutput(), "\nMetadata:")
		if output.Meta.Command != "" {
			fmt.Fprintf(f.getErrorOutput(), "  Command: %s\n", output.Meta.Command)
		}
		if output.Meta.Config != "" {
			fmt.Fprintf(f.getErrorOutput(), "  Config: %s\n", output.Meta.Config)
		}
		if output.Meta.Endpoint != "" {
			fmt.Fprintf(f.getErrorOutput(), "  Endpoint: %s\n", output.Meta.Endpoint)
		}
		if output.Meta.Operation != "" {
			fmt.Fprintf(f.getErrorOutput(), "  Operation: %s\n", output.Meta.Operation)
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
	message := ""
	if err != nil {
		message = err.Error()
	}

	output := &StructuredOutput{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}
	return f.formatStructuredYAMLToError(output)
}

// FormatStructuredErrorWithContext formats an error with additional context
func (f *YAMLFormatter) FormatStructuredErrorWithContext(err error, code string, errorType string, context map[string]interface{}, quiet bool) error {
	message := ""
	if err != nil {
		message = err.Error()
	}

	output := &StructuredOutput{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Type:    errorType,
			Context: context,
		},
	}
	return f.formatStructuredYAMLToError(output)
}

// FormatResponse formats a GraphQL response
func (f *YAMLFormatter) FormatResponse(response *Response, mode string) error {
	// YAML formatter doesn't support GraphQL response modes, fall back to JSON
	jsonFormatter := &JSONFormatter{}
	jsonFormatter.SetOutput(f.getOutput())
	jsonFormatter.SetErrorOutput(f.getErrorOutput())
	return jsonFormatter.FormatResponse(response, mode)
}

func (f *YAMLFormatter) formatStructuredYAML(output *StructuredOutput) error {
	// Simple YAML-like output
	if output.Success {
		fmt.Fprintln(f.getOutput(), "success: true")
		if output.Data != nil {
			fmt.Fprintf(f.getOutput(), "data: %v\n", output.Data)
		}
	} else {
		fmt.Fprintln(f.getOutput(), "success: false")
		fmt.Fprintf(f.getOutput(), "error:\n")
		fmt.Fprintf(f.getOutput(), "  code: %s\n", output.Error.Code)
		if output.Error.Type != "" {
			fmt.Fprintf(f.getOutput(), "  type: %s\n", output.Error.Type)
		}
		fmt.Fprintf(f.getOutput(), "  message: %s\n", output.Error.Message)
		if output.Error.Details != "" {
			fmt.Fprintf(f.getOutput(), "  details: %s\n", output.Error.Details)
		}
		if len(output.Error.Context) > 0 {
			fmt.Fprintf(f.getOutput(), "  context:\n")
			for key, value := range output.Error.Context {
				fmt.Fprintf(f.getOutput(), "    %s: %v\n", key, value)
			}
		}
	}

	if output.Meta != nil {
		fmt.Fprintln(f.getOutput(), "meta:")
		if output.Meta.Command != "" {
			fmt.Fprintf(f.getOutput(), "  command: %s\n", output.Meta.Command)
		}
		if output.Meta.Config != "" {
			fmt.Fprintf(f.getOutput(), "  config: %s\n", output.Meta.Config)
		}
		if output.Meta.Endpoint != "" {
			fmt.Fprintf(f.getOutput(), "  endpoint: %s\n", output.Meta.Endpoint)
		}
		if output.Meta.Operation != "" {
			fmt.Fprintf(f.getOutput(), "  operation: %s\n", output.Meta.Operation)
		}
	}

	return nil
}

func (f *YAMLFormatter) formatStructuredYAMLToError(output *StructuredOutput) error {
	// Simple YAML-like output to error stream
	if !output.Success {
		fmt.Fprintln(f.getErrorOutput(), "success: false")
		fmt.Fprintf(f.getErrorOutput(), "error:\n")
		fmt.Fprintf(f.getErrorOutput(), "  code: %s\n", output.Error.Code)
		if output.Error.Type != "" {
			fmt.Fprintf(f.getErrorOutput(), "  type: %s\n", output.Error.Type)
		}
		fmt.Fprintf(f.getErrorOutput(), "  message: %s\n", output.Error.Message)
		if output.Error.Details != "" {
			fmt.Fprintf(f.getErrorOutput(), "  details: %s\n", output.Error.Details)
		}
		if len(output.Error.Context) > 0 {
			fmt.Fprintf(f.getErrorOutput(), "  context:\n")
			for key, value := range output.Error.Context {
				fmt.Fprintf(f.getErrorOutput(), "    %s: %v\n", key, value)
			}
		}
	}

	if output.Meta != nil {
		fmt.Fprintln(f.getErrorOutput(), "meta:")
		if output.Meta.Command != "" {
			fmt.Fprintf(f.getErrorOutput(), "  command: %s\n", output.Meta.Command)
		}
		if output.Meta.Config != "" {
			fmt.Fprintf(f.getErrorOutput(), "  config: %s\n", output.Meta.Config)
		}
		if output.Meta.Endpoint != "" {
			fmt.Fprintf(f.getErrorOutput(), "  endpoint: %s\n", output.Meta.Endpoint)
		}
		if output.Meta.Operation != "" {
			fmt.Fprintf(f.getErrorOutput(), "  operation: %s\n", output.Meta.Operation)
		}
	}

	return nil
}

// FormatResponse formats and prints the GraphQL response
// This old FormatResponse method is no longer needed - it's been moved to JSONFormatter

// formatJSON outputs formatted JSON
func (f *JSONFormatter) formatJSON(response *Response) error {
	encoder := json.NewEncoder(f.getOutput())
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}


// formatRaw outputs unformatted JSON
func (f *JSONFormatter) formatRaw(response *Response) error {
	encoder := json.NewEncoder(f.getOutput())
	return encoder.Encode(response)
}

// FormatJSON formats data as JSON with indentation
func (f *JSONFormatter) FormatJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Fprintln(f.getOutput(), string(jsonData))
	return nil
}

// WriteToFile writes data to a file
func (f *JSONFormatter) WriteToFile(data interface{}, filename string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return os.WriteFile(filename, jsonData, 0644)
}
