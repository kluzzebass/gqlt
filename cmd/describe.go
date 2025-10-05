package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kluzzebass/gqlt"
	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:   "describe [type|field]",
	Short: "Describe GraphQL types and fields from cached schema",
	Long: `Describe GraphQL types and fields from a cached schema file.
Use this to explore the GraphQL schema structure.`,
	Args: cobra.ExactArgs(1),
	RunE: describe,
}

var (
	describeJSON    bool
	describeSummary bool
	describeSchema  string
)

func init() {
	rootCmd.AddCommand(describeCmd)

	// Define flags
	describeCmd.Flags().BoolVar(&describeJSON, "json", false, "output exact node JSON")
	describeCmd.Flags().BoolVar(&describeSummary, "summary", false, "output plain text summary")
	describeCmd.Flags().StringVar(&describeSchema, "schema", "", "schema file path (default is OS-specific)")
}

func describe(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := gqlt.Load(configDir)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Merge config with flags
	mergeConfigWithFlags(cfg)

	// Determine schema path
	schemaPath := describeSchema
	if schemaPath == "" {
		// Use config-specific schema path
		if configDir != "" {
			schemaPath = gqlt.GetSchemaPathForConfigInDir(cfg.Current, configDir)
		} else {
			schemaPath = gqlt.GetSchemaPathForConfig(cfg.Current)
		}
	}

	// Load schema analyzer
	analyzer, err := gqlt.LoadAnalyzerFromFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to load schema: %w", err)
	}

	// Parse the target
	target := args[0]

	// Handle different target formats
	if strings.HasPrefix(target, "Query.") || strings.HasPrefix(target, "Mutation.") || strings.HasPrefix(target, "Subscription.") {
		// Field reference: Query.product, Mutation.createUser, etc.
		parts := strings.SplitN(target, ".", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid field reference format: %s", target)
		}
		return describeField(analyzer, parts[0], parts[1])
	} else if strings.HasPrefix(target, "Type.") {
		// Type reference: Type.Product, Type.User, etc.
		typeName := strings.TrimPrefix(target, "Type.")
		return describeType(analyzer, typeName)
	} else {
		// Direct type name: Product, User, etc.
		return describeType(analyzer, target)
	}
}

func describeType(analyzer *gqlt.Analyzer, typeName string) error {
	// Find the type
	typeObj, err := analyzer.FindType(typeName)
	if err != nil {
		return err
	}

	if describeJSON {
		// Output raw JSON
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(typeObj)
	}

	// Output formatted description
	desc, err := analyzer.GetTypeDescription(typeName)
	if err != nil {
		return fmt.Errorf("failed to format type description: %w", err)
	}

	return printTypeDescription(desc)
}

func describeField(analyzer *gqlt.Analyzer, rootType, fieldName string) error {
	// Find the field
	fieldObj, err := analyzer.FindField(rootType, fieldName)
	if err != nil {
		return err
	}

	if describeJSON {
		// Output raw JSON
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(fieldObj)
	}

	// Output formatted description
	desc, err := analyzer.FindField(rootType, fieldName)
	if err != nil {
		return fmt.Errorf("failed to format field description: %w", err)
	}

	return printFieldDescription(desc)
}

func printTypeDescription(desc *gqlt.TypeDescription) error {
	fmt.Printf("TYPE %s (%s)\n", desc.Name, desc.Kind)
	if desc.Description != "" {
		fmt.Printf("  %s\n", desc.Description)
	}

	// Show fields if available
	if len(desc.Fields) > 0 {
		fmt.Printf("\nFields:\n")
		for _, field := range desc.Fields {
			fmt.Printf("  %s\n", field.Signature)
			if field.Description != "" {
				fmt.Printf("    %s\n", field.Description)
			}
		}
	}

	// Show input fields if available
	if len(desc.InputFields) > 0 {
		fmt.Printf("\nInput Fields:\n")
		for _, field := range desc.InputFields {
			fmt.Printf("  %s\n", field.Signature)
			if field.Description != "" {
				fmt.Printf("    %s\n", field.Description)
			}
		}
	}

	// Show enum values if available
	if len(desc.EnumValues) > 0 {
		fmt.Printf("\nEnum Values:\n")
		for _, enum := range desc.EnumValues {
			if enum.Description != "" {
				fmt.Printf("  %s - %s\n", enum.Name, enum.Description)
			} else {
				fmt.Printf("  %s\n", enum.Name)
			}
		}
	}

	return nil
}

func printFieldDescription(desc *gqlt.FieldDescription) error {
	fmt.Printf("FIELD %s.%s\n", desc.RootType, desc.Name)
	if desc.Description != "" {
		fmt.Printf("  %s\n", desc.Description)
	}

	// Show type information
	fmt.Printf("  Type: %s\n", desc.Type)

	// Show arguments if available
	if len(desc.Arguments) > 0 {
		fmt.Printf("\nArguments:\n")
		for _, arg := range desc.Arguments {
			fmt.Printf("  %s\n", arg.Signature)
			if arg.Description != "" {
				fmt.Printf("    %s\n", arg.Description)
			}
		}
	}

	return nil
}
