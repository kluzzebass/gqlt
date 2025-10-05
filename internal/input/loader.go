package input

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// LoadQuery loads a GraphQL query from string or file
func LoadQuery(query, queryFile string) (string, error) {
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

// LoadVariables loads variables from string or file
func LoadVariables(vars, varsFile string) (map[string]interface{}, error) {
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

// LoadHeaders parses header strings into a map
func LoadHeaders(headers []string) map[string]string {
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
// Currently returns empty map as file uploads are not yet implemented
func ParseFiles(files []string) (map[string]string, error) {
	// TODO: Implement file upload parsing
	// This will handle multipart/form-data for file uploads
	return make(map[string]string), nil
}
