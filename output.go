package gqlt

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
)

// Formatter handles different output modes
type Formatter struct{}

// NewFormatter creates a new formatter
func NewFormatter() *Formatter {
	return &Formatter{}
}

// FormatResponse formats and prints the GraphQL response
func (f *Formatter) FormatResponse(response *Response, mode string) error {
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

// formatJSON outputs formatted JSON
func (f *Formatter) formatJSON(response *Response) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}

// formatPretty outputs colorized formatted JSON
func (f *Formatter) formatPretty(response *Response) error {
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
func (f *Formatter) formatRaw(response *Response) error {
	encoder := json.NewEncoder(os.Stdout)
	return encoder.Encode(response)
}

// printValue recursively prints a value with colors
func (f *Formatter) printValue(v interface{}, indent string, dataColor, errorColor, keyColor, stringColor, numberColor, boolColor *color.Color) {
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
func (f *Formatter) FormatJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// FormatPretty formats data as colorized JSON
func (f *Formatter) FormatPretty(data interface{}) error {
	// For now, just use regular JSON formatting
	// TODO: Implement colorized formatting for arbitrary data
	return f.FormatJSON(data)
}

// WriteToFile writes data to a file
func (f *Formatter) WriteToFile(data interface{}, filename string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return os.WriteFile(filename, jsonData, 0644)
}
