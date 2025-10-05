package schema

import (
	"testing"

	"github.com/kluzzebass/gqlt/internal/graphql"
)

func TestNewAnalyzer(t *testing.T) {
	// Create mock schema response
	schemaData := map[string]interface{}{
		"__schema": map[string]interface{}{
			"types": []interface{}{
				map[string]interface{}{
					"name": "User",
					"kind": "OBJECT",
					"fields": []interface{}{
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
	}

	response := &graphql.Response{
		Data: schemaData,
	}

	analyzer, err := NewAnalyzer(response)
	if err != nil {
		t.Errorf("NewAnalyzer failed: %v", err)
	}

	if analyzer == nil {
		t.Error("Expected analyzer to be created")
		return
	}

	if analyzer.schemaData == nil {
		t.Error("Expected schemaData to be set")
	}
}

func TestNewAnalyzerInvalidData(t *testing.T) {
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
}

func TestGetSummary(t *testing.T) {
	// Create mock schema with all root types
	schemaData := map[string]interface{}{
		"__schema": map[string]interface{}{
			"types": []interface{}{
				map[string]interface{}{"name": "User", "kind": "OBJECT"},
				map[string]interface{}{"name": "Product", "kind": "OBJECT"},
				map[string]interface{}{"name": "String", "kind": "SCALAR"},
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
	analyzer, _ := NewAnalyzer(response)

	summary, err := analyzer.GetSummary()
	if err != nil {
		t.Errorf("GetSummary failed: %v", err)
	}

	if summary.TotalTypes != 3 {
		t.Errorf("Expected 3 types, got %d", summary.TotalTypes)
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
}

func TestFindType(t *testing.T) {
	schemaData := map[string]interface{}{
		"__schema": map[string]interface{}{
			"types": []interface{}{
				map[string]interface{}{
					"name":        "User",
					"kind":        "OBJECT",
					"description": "A user in the system",
				},
				map[string]interface{}{
					"name":        "Product",
					"kind":        "OBJECT",
					"description": "A product in the catalog",
				},
			},
		},
	}

	response := &graphql.Response{Data: schemaData}
	analyzer, _ := NewAnalyzer(response)

	// Test finding existing type
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

	// Test finding non-existent type
	_, err = analyzer.FindType("NonExistent")
	if err == nil {
		t.Error("Expected error for non-existent type")
	}
}

func TestFindField(t *testing.T) {
	schemaData := map[string]interface{}{
		"__schema": map[string]interface{}{
			"types": []interface{}{
				map[string]interface{}{
					"name": "Query",
					"kind": "OBJECT",
					"fields": []interface{}{
						map[string]interface{}{
							"name": "user",
							"type": map[string]interface{}{
								"name": "User",
								"kind": "OBJECT",
							},
						},
						map[string]interface{}{
							"name": "products",
							"type": map[string]interface{}{
								"name": "Product",
								"kind": "OBJECT",
							},
						},
					},
				},
			},
		},
	}

	response := &graphql.Response{Data: schemaData}
	analyzer, _ := NewAnalyzer(response)

	// Test finding existing field
	userField, err := analyzer.FindField("Query", "user")
	if err != nil {
		t.Errorf("FindField failed: %v", err)
	}

	if userField["name"] != "user" {
		t.Errorf("Expected name 'user', got %v", userField["name"])
	}

	// Test finding non-existent field
	_, err = analyzer.FindField("Query", "nonExistent")
	if err == nil {
		t.Error("Expected error for non-existent field")
	}

	// Test finding field in non-existent type
	_, err = analyzer.FindField("NonExistent", "user")
	if err == nil {
		t.Error("Expected error for non-existent type")
	}
}

func TestFormatTypeString(t *testing.T) {
	analyzer := &Analyzer{}

	tests := []struct {
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
		{
			name: "NonNull List",
			typeObj: map[string]interface{}{
				"kind": "NON_NULL",
				"ofType": map[string]interface{}{
					"kind": "LIST",
					"ofType": map[string]interface{}{
						"name": "String",
						"kind": "SCALAR",
					},
				},
			},
			expected: "[String]!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.formatTypeString(tt.typeObj)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFormatFieldSummary(t *testing.T) {
	analyzer := &Analyzer{}

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

	summary := analyzer.formatFieldSummary(fieldObj)

	if summary.Name != "user" {
		t.Errorf("Expected name 'user', got %s", summary.Name)
	}

	if summary.Description != "Get a user by ID" {
		t.Errorf("Expected description 'Get a user by ID', got %s", summary.Description)
	}

	if summary.Type != "User" {
		t.Errorf("Expected type 'User', got %s", summary.Type)
	}

	// Check signature format
	expectedSignature := "user(id: ID): User"
	if summary.Signature != expectedSignature {
		t.Errorf("Expected signature '%s', got %s", expectedSignature, summary.Signature)
	}
}
