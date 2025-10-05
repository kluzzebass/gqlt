package gqlt

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Input handles input operations for loading queries, variables, headers, and files.
// It provides utilities for parsing and loading various types of input data.
type Input struct{}

// NewInput creates a new input handler instance.
//
// Example:
//   input := gqlt.NewInput()
//   query, err := input.LoadQuery("", "query.graphql")
func NewInput() *Input {
	return &Input{}
}

// LoadQuery loads a GraphQL query from a string or file.
// If query is provided, it returns the query string directly.
// If queryFile is provided, it reads and returns the file contents.
// If both are provided, query takes precedence.
//
// Example:
//   query, err := input.LoadQuery("", "query.graphql")
//   if err != nil {
//       log.Fatal(err)
//   }
func (i *Input) LoadQuery(query, queryFile string) (string, error) {
	if query != "" {
		return query, nil
	}

	if queryFile != "" {
		data, err := os.ReadFile(queryFile)
		if err != nil {
			return "", fmt.Errorf("failed to read query file: %w", err)
		}
		return string(data), nil
	}

	return "", fmt.Errorf("either query or queryFile must be provided")
}

// LoadVariables loads GraphQL variables from a JSON string or file.
// If vars is provided, it parses the JSON string directly.
// If varsFile is provided, it reads and parses the file contents.
// If both are provided, vars takes precedence.
//
// Example:
//   variables, err := input.LoadVariables(`{"id": "123"}`, "")
//   if err != nil {
//       log.Fatal(err)
//   }
func (i *Input) LoadVariables(vars, varsFile string) (map[string]interface{}, error) {
	var varsStr string

	if vars != "" {
		varsStr = vars
	} else if varsFile != "" {
		data, err := os.ReadFile(varsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read variables file: %w", err)
		}
		varsStr = string(data)
	} else {
		// No variables provided
		return make(map[string]interface{}), nil
	}

	// Parse JSON variables
	var varsMap map[string]interface{}
	err := json.Unmarshal([]byte(varsStr), &varsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse variables JSON: %w", err)
	}

	return varsMap, nil
}

// LoadHeaders parses header strings into a map.
// Each header string should be in the format "Key: Value".
//
// Example:
//   headers := input.LoadHeaders([]string{
//       "Authorization: Bearer token",
//       "Content-Type: application/json",
//   })
func (i *Input) LoadHeaders(headers []string) map[string]string {
	headersMap := make(map[string]string)

	for _, header := range headers {
		parts := strings.SplitN(header, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			headersMap[key] = value
		}
	}

	return headersMap
}

// ParseFiles parses file upload specifications
func (i *Input) ParseFiles(files []string) (map[string]string, error) {
	filesMap := make(map[string]string)

	for _, file := range files {
		// Parse "name=path" format
		parts := strings.SplitN(file, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid file format '%s', expected 'name=path'", file)
		}

		name := strings.TrimSpace(parts[0])
		path := strings.TrimSpace(parts[1])

		if name == "" {
			return nil, fmt.Errorf("file name cannot be empty in '%s'", file)
		}

		if path == "" {
			return nil, fmt.Errorf("file path cannot be empty in '%s'", file)
		}

		// Validate file exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, fmt.Errorf("file does not exist: %s", path)
		}

		filesMap[name] = path
	}

	return filesMap, nil
}

// ParseFilesFromList parses file upload specifications from a file
func (i *Input) ParseFilesFromList(filesListPath string) ([]string, error) {
	if filesListPath == "" {
		return []string{}, nil
	}

	data, err := os.ReadFile(filesListPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read files list: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var files []string

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		// Validate format - just check for = and use everything after it as the filename
		if !strings.Contains(line, "=") {
			return nil, fmt.Errorf("invalid file format at line %d: '%s', expected 'name=path'", lineNum+1, line)
		}

		// Resolve the path to handle relative paths, ~, etc.
		resolvedLine, err := i.resolveFilePath(line)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve path at line %d: %v", lineNum+1, err)
		}

		files = append(files, resolvedLine)
	}

	return files, nil
}

// resolveFilePath resolves a file path, handling ~ expansion and relative paths
func (i *Input) resolveFilePath(line string) (string, error) {
	// Split the line into name and path
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return line, nil // Return as-is if no = found
	}

	name := strings.TrimSpace(parts[0])
	path := strings.TrimSpace(parts[1])

	// Handle ~ expansion
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %v", err)
		}
		path = filepath.Join(homeDir, path[2:])
	} else if path == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %v", err)
		}
		path = homeDir
	}

	// Resolve relative paths to absolute paths
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve absolute path for '%s': %v", path, err)
	}

	return name + "=" + absPath, nil
}
