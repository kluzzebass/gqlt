package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kluzzebass/gqlt"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// DocsGenerator defines the interface for documentation generators
type DocsGenerator interface {
	GenerateTo(rootCmd *cobra.Command, output string) error
}

// DocsGeneratorRegistry manages available documentation generators
type DocsGeneratorRegistry struct {
	generators map[string]DocsGenerator
}

// NewDocsGeneratorRegistry creates a new registry with all available generators
func NewDocsGeneratorRegistry() *DocsGeneratorRegistry {
	registry := &DocsGeneratorRegistry{
		generators: make(map[string]DocsGenerator),
	}

	// Register built-in generators
	registry.Register("md", &MarkdownGenerator{})
	registry.Register("man", &ManPageGenerator{})

	return registry
}

// Register adds a generator to the registry
func (r *DocsGeneratorRegistry) Register(name string, generator DocsGenerator) {
	r.generators[name] = generator
}

var (
	docsFormat string
	docsOutput string
	docsTree   bool
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate documentation",
	Long:  "Generate documentation in various formats from the command structure",
	Example: `# Generate README.md
gqlt docs --format md --output README.md

# Generate man pages
gqlt docs --format man --output man/

# Generate multiple markdown files
gqlt docs --format md --tree --output docs/

# Output to stdout
gqlt docs --format md --output -`,
	Args: cobra.NoArgs,
	RunE: docs,
}

func init() {
	rootCmd.AddCommand(docsCmd)
	docsCmd.Flags().StringVarP(&docsFormat, "format", "f", "md", "Output format: md or man")
	docsCmd.Flags().StringVarP(&docsOutput, "output", "o", "-", "Output destination (file for md, directory for man, '-' for stdout)")
	docsCmd.Flags().BoolVar(&docsTree, "tree", false, "Generate multiple files (one per command) instead of single file")
}

func docs(cmd *cobra.Command, args []string) error {
	registry := NewDocsGeneratorRegistry()

	format := docsFormat
	output := docsOutput

	// Generate specific format
	generator, exists := registry.generators[format]
	if !exists {
		available := make([]string, 0, len(registry.generators))
		for name := range registry.generators {
			available = append(available, name)
		}
		return fmt.Errorf("unsupported format: %s. Available: %s", format, strings.Join(available, ", "))
	}

	// Generate documentation
	if err := generator.GenerateTo(rootCmd, output); err != nil {
		return fmt.Errorf("error generating %s: %w", format, err)
	}

	if output != "-" {
		fmt.Printf("Documentation generated successfully: %s -> %s\n", format, output)
	}

	return nil
}

// MarkdownGenerator generates README.md
type MarkdownGenerator struct{}

func (g *MarkdownGenerator) GenerateTo(rootCmd *cobra.Command, output string) error {
	// Handle tree mode - generate separate files for each command
	if docsTree {
		// For tree mode, output must be a directory
		if output == "-" {
			return fmt.Errorf("tree mode requires a directory output, not stdout")
		}

		// If output is empty or ".", default to "docs" directory
		if output == "" || output == "." {
			output = "docs"
		}

		// Create output directory if it doesn't exist
		if err := os.MkdirAll(output, 0755); err != nil {
			return err
		}

		// Use Cobra's tree markdown generator
		return doc.GenMarkdownTree(rootCmd, output)
	}

	// Single file mode - generate markdown for root command only
	if output == "-" {
		// Output to stdout using Cobra's built-in markdown generator
		return doc.GenMarkdown(rootCmd, os.Stdout)
	} else if output == "." || output == "" {
		output = "README.md"
	} else if stat, err := os.Stat(output); err == nil && stat.IsDir() {
		// If output is a directory, create README.md in that directory
		output = filepath.Join(output, "README.md")
	}
	// Otherwise, treat output as the exact file path

	// Create the output file
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	// Use Cobra's built-in markdown generator for single file (root command only)
	return doc.GenMarkdown(rootCmd, file)
}

// ManPageGenerator generates man pages
type ManPageGenerator struct{}

func (g *ManPageGenerator) GenerateTo(rootCmd *cobra.Command, output string) error {
	// Generate man page header
	header := &doc.GenManHeader{
		Title:   "GQLT",
		Section: "1",
		Source:  fmt.Sprintf("gqlt %s", gqlt.Version()),
		Manual:  "User Commands",
	}

	// Handle tree mode - generate separate files for each command
	if docsTree {
		// For tree mode, output must be a directory
		if output == "-" {
			return fmt.Errorf("tree mode requires a directory output, not stdout")
		}

		// If output is empty or ".", default to "man" directory
		if output == "" || output == "." {
			output = "man"
		}

		// Create output directory if it doesn't exist
		if err := os.MkdirAll(output, 0755); err != nil {
			return err
		}

		// Generate man pages for all commands recursively (separate files)
		return doc.GenManTree(rootCmd, header, output)
	}

	// Single file mode - generate man page for root command only
	if output == "-" {
		// Output to stdout using Cobra's built-in man page generator
		return doc.GenMan(rootCmd, header, os.Stdout)
	} else if output == "." || output == "" {
		output = "gqlt.1"
	} else if stat, err := os.Stat(output); err == nil && stat.IsDir() {
		// If output is a directory, create gqlt.1 in that directory
		output = filepath.Join(output, "gqlt.1")
	}
	// Otherwise, treat output as the exact file path

	// Create the output file
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	// Use Cobra's built-in man page generator for single file (root command only)
	return doc.GenMan(rootCmd, header, file)
}
