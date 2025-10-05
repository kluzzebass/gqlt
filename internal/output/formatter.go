package output

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/kluzzebass/gqlt/internal/graphql"
)

// Formatter handles different output modes
type Formatter struct {
	mode string
}

// NewFormatter creates a new formatter
func NewFormatter(mode string) *Formatter {
	return &Formatter{mode: mode}
}

// Format formats and prints the GraphQL response
func (f *Formatter) Format(response *graphql.Response) error {
	switch f.mode {
	case "json":
		return f.formatJSON(response)
	case "pretty":
		return f.formatPretty(response)
	case "raw":
		return f.formatRaw(response)
	default:
		return fmt.Errorf("unknown output mode: %s", f.mode)
	}
}

// formatJSON outputs formatted JSON
func (f *Formatter) formatJSON(response *graphql.Response) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}

// formatPretty outputs colorized formatted JSON
func (f *Formatter) formatPretty(response *graphql.Response) error {
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
func (f *Formatter) formatRaw(response *graphql.Response) error {
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
