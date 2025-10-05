package schema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kluzzebass/gqlt/internal/graphql"
)

func TestAnalyzer(t *testing.T) {
	// Create comprehensive test schema
	schemaData := map[string]interface{}{
		"__schema": map[string]interface{}{
			"types": []interface{}{
				map[string]interface{}{
					"name":        "User",
					"kind":        "OBJECT",
					"description": "A user in the system",
					"fields": []interface{}{
						map[string]interface{}{
							"name":        "id",
							"description": "User ID",
							"type": map[string]interface{}{
								"name": "ID",
								"kind": "SCALAR",
							},
						},
						map[string]interface{}{
							"name":        "name",
							"description": "User name",
							"type": map[string]interface{}{
								"name": "String",
								"kind": "SCALAR",
							},
						},
					},
				},
				map[string]interface{}{
					"name":        "Query",
					"kind":        "OBJECT",
					"description": "Root query type",
					"fields": []interface{}{
						map[string]interface{}{
							"name":        "user",
							"description": "Get a user by ID",
							"type": map[string]interface{}{
								"name": "User",
								"kind": "OBJECT",
							},
							"args": []interface{}{
								map[string]interface{}{
									"name": "id",
									"type": map[string]interface{}{
										"name": "ID",
										"kind": "SCALAR",
									},
								},
							},
						},
					},
				},
				map[string]interface{}{
					"name": "String",
					"kind": "SCALAR",
				},
				map[string]interface{}{
					"name": "ID",
					"kind": "SCALAR",
				},
			},
			"queryType": map[string]interface{}{
				"name": "Query",
			},
			"mutationType": map[string]interface{}{
				"name": "Mutation",
			},
			"subscriptionType": map[string]interface{}{
				"name": "Subscription",
			},
		},
	}

	response := &graphql.Response{Data: schemaData}
	analyzer, err := NewAnalyzer(response)
	if err != nil {
		t.Fatalf("NewAnalyzer failed: %v", err)
	}

	// Test analyzer creation
	if analyzer == nil {
		t.Error("Expected analyzer to be created")
	}
	if analyzer.schemaData == nil {
		t.Error("Expected schemaData to be set")
	}

	// Test GetSummary
	summary, err := analyzer.GetSummary()
	if err != nil {
		t.Errorf("GetSummary failed: %v", err)
	}
	if summary.TotalTypes != 4 {
		t.Errorf("Expected 4 types, got %d", summary.TotalTypes)
	}
	if summary.QueryType != "Query" {
		t.Errorf("Expected QueryType 'Query', got %s", summary.QueryType)
	}
	if summary.MutationType != "Mutation" {
		t.Errorf("Expected MutationType 'Mutation', got %s", summary.MutationType)
	}
	if summary.SubscriptionType != "Subscription" {
		t.Errorf("Expected SubscriptionType 'Subscription', got %s", summary.SubscriptionType)
	}

	// Test FindType
	userType, err := analyzer.FindType("User")
	if err != nil {
		t.Errorf("FindType failed: %v", err)
	}
	if userType["name"] != "User" {
		t.Errorf("Expected name 'User', got %v", userType["name"])
	}
	if userType["kind"] != "OBJECT" {
		t.Errorf("Expected kind 'OBJECT', got %v", userType["kind"])
	}

	// Test FindField
	userField, err := analyzer.FindField("Query", "user")
	if err != nil {
		t.Errorf("FindField failed: %v", err)
	}
	if userField["name"] != "user" {
		t.Errorf("Expected name 'user', got %v", userField["name"])
	}

	// Test FormatTypeString
	typeTests := []struct {
		name     string
		typeObj  map[string]interface{}
		expected string
	}{
		{
			name: "Simple scalar",
			typeObj: map[string]interface{}{
				"name": "String",
				"kind": "SCALAR",
			},
			expected: "String",
		},
		{
			name: "List type",
			typeObj: map[string]interface{}{
				"kind": "LIST",
				"ofType": map[string]interface{}{
					"name": "String",
					"kind": "SCALAR",
				},
			},
			expected: "[String]",
		},
		{
			name: "NonNull type",
			typeObj: map[string]interface{}{
				"kind": "NON_NULL",
				"ofType": map[string]interface{}{
					"name": "String",
					"kind": "SCALAR",
				},
			},
			expected: "String!",
		},
	}

	for _, tt := range typeTests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.formatTypeString(tt.typeObj)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}

	// Test FormatFieldSummary
	fieldObj := map[string]interface{}{
		"name":        "user",
		"description": "Get a user by ID",
		"type": map[string]interface{}{
			"name": "User",
			"kind": "OBJECT",
		},
		"args": []interface{}{
			map[string]interface{}{
				"name": "id",
				"type": map[string]interface{}{
					"name": "ID",
					"kind": "SCALAR",
				},
			},
		},
	}

	summary_field := analyzer.formatFieldSummary(fieldObj)
	if summary_field.Name != "user" {
		t.Errorf("Expected name 'user', got %s", summary_field.Name)
	}
	if summary_field.Description != "Get a user by ID" {
		t.Errorf("Expected description 'Get a user by ID', got %s", summary_field.Description)
	}
	if summary_field.Type != "User" {
		t.Errorf("Expected type 'User', got %s", summary_field.Type)
	}
}

func TestAnalyzerEdgeCases(t *testing.T) {
	// Test with invalid data
	response := &graphql.Response{
		Data: map[string]interface{}{
			"invalid": "data",
		},
	}

	_, err := NewAnalyzer(response)
	if err == nil {
		t.Error("Expected error for invalid schema data")
	}

	// Test with empty schema
	emptySchema := map[string]interface{}{
		"__schema": map[string]interface{}{
			"types": []interface{}{},
		},
	}

	response = &graphql.Response{Data: emptySchema}
	analyzer, _ := NewAnalyzer(response)
	summary, err := analyzer.GetSummary()
	if err != nil {
		t.Errorf("GetSummary failed: %v", err)
	}
	if summary.TotalTypes != 0 {
		t.Errorf("Expected 0 types, got %d", summary.TotalTypes)
	}

	// Test with nil types
	nilTypesSchema := map[string]interface{}{
		"__schema": map[string]interface{}{
			"types": nil,
		},
	}

	response = &graphql.Response{Data: nilTypesSchema}
	analyzer, _ = NewAnalyzer(response)
	summary, err = analyzer.GetSummary()
	if err != nil {
		// It's okay if this fails with nil types
		return
	}
	if summary.TotalTypes != 0 {
		t.Errorf("Expected 0 types, got %d", summary.TotalTypes)
	}

	// Test finding non-existent type
	_, err = analyzer.FindType("NonExistent")
	if err == nil {
		t.Error("Expected error for non-existent type")
	}

	// Test finding non-existent field
	_, err = analyzer.FindField("Query", "nonExistent")
	if err == nil {
		t.Error("Expected error for non-existent field")
	}
}

func TestAnalyzerFileOperations(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	schemaFile := filepath.Join(tempDir, "schema.json")

	// Create mock schema data
	schemaData := map[string]interface{}{
		"__schema": map[string]interface{}{
			"types": []interface{}{
				map[string]interface{}{
					"name": "User",
					"kind": "OBJECT",
				},
			},
		},
	}

	response := &graphql.Response{Data: schemaData}
	data, err := json.Marshal(response)
	if err != nil {
		t.Errorf("Failed to marshal schema: %v", err)
	}

	err = os.WriteFile(schemaFile, data, 0644)
	if err != nil {
		t.Errorf("Failed to write schema file: %v", err)
	}

	// Test LoadAnalyzerFromFile
	analyzer, err := LoadAnalyzerFromFile(schemaFile)
	if err != nil {
		t.Errorf("LoadAnalyzerFromFile failed: %v", err)
	}
	if analyzer == nil {
		t.Error("Expected analyzer to be created")
	}
	if analyzer.schemaData == nil {
		t.Error("Expected schemaData to be set")
	}

	// Test loading non-existent file
	_, err = LoadAnalyzerFromFile("/non/existent/schema.json")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test loading invalid JSON
	invalidFile := filepath.Join(tempDir, "invalid.json")
	err = os.WriteFile(invalidFile, []byte("invalid json"), 0644)
	if err != nil {
		t.Errorf("Failed to create invalid JSON file: %v", err)
	}

	_, err = LoadAnalyzerFromFile(invalidFile)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}
