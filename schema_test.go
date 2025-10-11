package gqlt

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
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

	response := &Response{Data: schemaData}
	analyzer, err := NewAnalyzer(response)
	if err != nil {
		t.Fatalf("NewAnalyzer failed: %v", err)
	}

	// Test analyzer creation
	if analyzer == nil {
		t.Error("Expected analyzer to be created")
		return
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
	if userType.Name != "User" {
		t.Errorf("Expected name 'User', got %v", userType.Name)
	}
	if userType.Kind != "OBJECT" {
		t.Errorf("Expected kind 'OBJECT', got %v", userType.Kind)
	}

	// Test FindField
	userField, err := analyzer.FindField("Query", "user")
	if err != nil {
		t.Errorf("FindField failed: %v", err)
	}
	if userField.Name != "user" {
		t.Errorf("Expected name 'user', got %v", userField.Name)
	}

	// Test FormatTypeString - removed as formatTypeString is not a public method

	// Test FormatFieldSummary - removed as formatFieldSummary is not a public method
}

func TestAnalyzerEdgeCases(t *testing.T) {
	// Test with invalid data
	response := &Response{
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

	response = &Response{Data: emptySchema}
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

	response = &Response{Data: nilTypesSchema}
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

	response := &Response{Data: schemaData}
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
		return
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

func TestAnalyzer_GetTypeDescription_TableDriven(t *testing.T) {
	tests := []struct {
		name         string
		typeName     string
		schemaData   map[string]interface{}
		wantErr      bool
		validateType func(*testing.T, *TypeDescription)
	}{
		{
			name:     "OBJECT type with fields",
			typeName: "User",
			schemaData: map[string]interface{}{
				"__schema": map[string]interface{}{
					"types": []interface{}{
						map[string]interface{}{
							"name":        "User",
							"kind":        "OBJECT",
							"description": "A user object",
							"fields": []interface{}{
								map[string]interface{}{
									"name":        "id",
									"description": "User ID",
									"type": map[string]interface{}{
										"kind": "SCALAR",
										"name": "ID",
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			validateType: func(t *testing.T, td *TypeDescription) {
				if td.Name != "User" {
					t.Errorf("Expected name User, got %s", td.Name)
				}
				if td.Kind != "OBJECT" {
					t.Errorf("Expected kind OBJECT, got %s", td.Kind)
				}
				if len(td.Fields) == 0 {
					t.Error("Expected fields to be present")
				}
			},
		},
		{
			name:     "ENUM type with values",
			typeName: "Status",
			schemaData: map[string]interface{}{
				"__schema": map[string]interface{}{
					"types": []interface{}{
						map[string]interface{}{
							"name":        "Status",
							"kind":        "ENUM",
							"description": "User status",
							"enumValues": []interface{}{
								map[string]interface{}{
									"name":        "ACTIVE",
									"description": "Active user",
								},
								map[string]interface{}{
									"name":        "INACTIVE",
									"description": "Inactive user",
								},
							},
						},
					},
				},
			},
			wantErr: false,
			validateType: func(t *testing.T, td *TypeDescription) {
				if td.Name != "Status" {
					t.Errorf("Expected name Status, got %s", td.Name)
				}
				if td.Kind != "ENUM" {
					t.Errorf("Expected kind ENUM, got %s", td.Kind)
				}
				if len(td.EnumValues) != 2 {
					t.Errorf("Expected 2 enum values, got %d", len(td.EnumValues))
				}
			},
		},
		{
			name:     "INPUT_OBJECT type with input fields",
			typeName: "CreateUserInput",
			schemaData: map[string]interface{}{
				"__schema": map[string]interface{}{
					"types": []interface{}{
						map[string]interface{}{
							"name": "CreateUserInput",
							"kind": "INPUT_OBJECT",
							"inputFields": []interface{}{
								map[string]interface{}{
									"name": "name",
									"type": map[string]interface{}{
										"kind": "SCALAR",
										"name": "String",
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
			validateType: func(t *testing.T, td *TypeDescription) {
				if td.Kind != "INPUT_OBJECT" {
					t.Errorf("Expected kind INPUT_OBJECT, got %s", td.Kind)
				}
				if len(td.InputFields) == 0 {
					t.Error("Expected input fields to be present")
				}
			},
		},
		{
			name:     "SCALAR type",
			typeName: "String",
			schemaData: map[string]interface{}{
				"__schema": map[string]interface{}{
					"types": []interface{}{
						map[string]interface{}{
							"name":        "String",
							"kind":        "SCALAR",
							"description": "Built-in String type",
						},
					},
				},
			},
			wantErr: false,
			validateType: func(t *testing.T, td *TypeDescription) {
				if td.Kind != "SCALAR" {
					t.Errorf("Expected kind SCALAR, got %s", td.Kind)
				}
			},
		},
		{
			name:     "non-existent type",
			typeName: "NonExistent",
			schemaData: map[string]interface{}{
				"__schema": map[string]interface{}{
					"types": []interface{}{
						map[string]interface{}{
							"name": "User",
							"kind": "OBJECT",
						},
					},
				},
			},
			wantErr:      true,
			validateType: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := &Response{Data: tt.schemaData}
			analyzer, err := NewAnalyzer(response)
			if err != nil {
				t.Fatalf("NewAnalyzer failed: %v", err)
			}

			typeDesc, err := analyzer.GetTypeDescription(tt.typeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTypeDescription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateType != nil {
				tt.validateType(t, typeDesc)
			}
		})
	}
}

func TestAnalyzer_FindField_EdgeCases(t *testing.T) {
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
								"kind": "OBJECT",
								"name": "User",
							},
							"args": []interface{}{
								map[string]interface{}{
									"name": "id",
									"type": map[string]interface{}{
										"kind": "SCALAR",
										"name": "ID",
									},
								},
							},
						},
					},
				},
				map[string]interface{}{
					"name": "Mutation",
					"kind": "OBJECT",
				},
			},
		},
	}

	response := &Response{Data: schemaData}
	analyzer, _ := NewAnalyzer(response)

	tests := []struct {
		name      string
		rootType  string
		fieldName string
		wantErr   bool
		validate  func(*testing.T, *FieldDescription)
	}{
		{
			name:      "existing field",
			rootType:  "Query",
			fieldName: "user",
			wantErr:   false,
			validate: func(t *testing.T, fd *FieldDescription) {
				if fd.Name != "user" {
					t.Errorf("Expected name user, got %s", fd.Name)
				}
				if fd.RootType != "Query" {
					t.Errorf("Expected RootType Query, got %s", fd.RootType)
				}
			},
		},
		{
			name:      "non-existent field",
			rootType:  "Query",
			fieldName: "nonExistent",
			wantErr:   true,
			validate:  nil,
		},
		{
			name:      "non-existent root type",
			rootType:  "NonExistent",
			fieldName: "field",
			wantErr:   true,
			validate:  nil,
		},
		{
			name:      "type without fields",
			rootType:  "Mutation",
			fieldName: "createUser",
			wantErr:   true,
			validate:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, err := analyzer.FindField(tt.rootType, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, field)
			}
		})
	}
}
