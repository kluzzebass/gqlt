package main

import (
	"testing"

	"github.com/kluzzebass/gqlt"
)

// ExampleIntegrationTest demonstrates a complete integration test
func ExampleIntegrationTest(t *testing.T) {
	// This would typically use a real GraphQL endpoint
	// For this example, we'll use a public GraphQL API
	endpoint := "https://api.github.com/graphql"

	// Create client with authentication
	client := gqlt.NewClient(endpoint, nil)

	// Set authentication (you would need a real GitHub token)
	// client.SetAuth("bearer", "your-github-token", "")

	// Test 1: Simple query
	t.Run("SimpleQuery", func(t *testing.T) {
		query := `
			query {
				viewer {
					login
					name
				}
			}
		`

		response, err := client.Execute(query, nil, "")
		if err != nil {
			t.Fatalf("Query execution failed: %v", err)
		}

		// Check for authentication errors (expected without token)
		if len(response.Errors) > 0 {
			// This is expected without authentication
			t.Logf("Expected authentication error: %v", response.Errors)
		}
	})

	// Test 2: Query with variables
	t.Run("QueryWithVariables", func(t *testing.T) {
		query := `
			query GetRepository($owner: String!, $name: String!) {
				repository(owner: $owner, name: $name) {
					name
					description
					stargazerCount
				}
			}
		`

		variables := map[string]interface{}{
			"owner": "facebook",
			"name":  "react",
		}

		response, err := client.Execute(query, variables, "GetRepository")
		if err != nil {
			t.Fatalf("Query execution failed: %v", err)
		}

		// This should work even without authentication for public repos
		if len(response.Errors) > 0 {
			t.Logf("GraphQL errors: %v", response.Errors)
		}
	})
}

// ExampleTestWithSchemaIntrospection demonstrates testing with schema introspection
func ExampleTestWithSchemaIntrospection(t *testing.T) {
	endpoint := "https://api.github.com/graphql"
	client := gqlt.NewClient(endpoint, nil)

	// Create introspection client
	introspectClient := gqlt.NewIntrospect(client)

	// Test schema introspection
	t.Run("IntrospectSchema", func(t *testing.T) {
		schema, err := introspectClient.IntrospectSchema()
		if err != nil {
			t.Fatalf("Schema introspection failed: %v", err)
		}

		// Verify schema structure
		if schema == nil || schema.Data == nil {
			t.Fatal("Expected schema data")
		}

		// Check for GraphQL errors
		if len(schema.Errors) > 0 {
			t.Errorf("Schema introspection errors: %v", schema.Errors)
		}

		// Verify schema contains expected fields
		data, ok := schema.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Expected schema data to be a map")
		}

		if _, exists := data["__schema"]; !exists {
			t.Error("Expected '__schema' field in introspection response")
		}
	})

	// Test schema analysis
	t.Run("AnalyzeSchema", func(t *testing.T) {
		schema, err := introspectClient.IntrospectSchema()
		if err != nil {
			t.Fatalf("Schema introspection failed: %v", err)
		}

		// Create analyzer
		analyzer, err := gqlt.NewAnalyzer(schema)
		if err != nil {
			t.Fatalf("Failed to create analyzer: %v", err)
		}

		// Get schema summary
		summary, err := analyzer.GetSummary()
		if err != nil {
			t.Fatalf("Failed to get schema summary: %v", err)
		}

		// Verify summary contains expected information
		if summary.TotalTypes == 0 {
			t.Error("Expected schema to have types")
		}

		t.Logf("Schema has %d types", summary.TotalTypes)
		t.Logf("Query type: %s", summary.QueryType)
		t.Logf("Mutation type: %s", summary.MutationType)
		t.Logf("Subscription type: %s", summary.SubscriptionType)
	})
}

// ExampleTestWithConfiguration demonstrates testing with configuration
func ExampleTestWithConfiguration(t *testing.T) {
	// Create temporary directory for config
	tempDir := t.TempDir()

	// Test configuration management
	t.Run("ConfigManagement", func(t *testing.T) {
		// Load configuration
		cfg, err := gqlt.Load(tempDir)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		// Create a test configuration
		cfg.Configs["test"] = gqlt.ConfigEntry{
			Endpoint: "https://api.example.com/graphql",
			Headers: map[string]string{
				"Authorization": "Bearer test-token",
			},
		}

		// Save configuration
		err = cfg.Save(tempDir)
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Reload and verify
		cfg2, err := gqlt.Load(tempDir)
		if err != nil {
			t.Fatalf("Failed to reload config: %v", err)
		}

		testConfig, exists := cfg2.Configs["test"]
		if !exists {
			t.Error("Expected test configuration to exist")
		}

		if testConfig.Endpoint != "https://api.example.com/graphql" {
			t.Errorf("Expected endpoint to be 'https://api.example.com/graphql', got '%s'", testConfig.Endpoint)
		}
	})
}

// ExampleTestWithValidation demonstrates testing with validation
func ExampleTestWithValidation(t *testing.T) {
	// This would typically use a real GraphQL endpoint
	endpoint := "https://api.github.com/graphql"
	client := gqlt.NewClient(endpoint, nil)

	// Test query validation
	t.Run("ValidateQuery", func(t *testing.T) {
		// This would use the validate command functionality
		// For now, we'll demonstrate the concept

		query := `
			query GetUser($id: ID!) {
				user(id: $id) {
					id
					name
					email
				}
			}
		`

		// Test with valid variables
		variables := map[string]interface{}{
			"id": "123",
		}

		response, err := client.Execute(query, variables, "GetUser")
		if err != nil {
			t.Fatalf("Query execution failed: %v", err)
		}

		// Check for validation errors
		if len(response.Errors) > 0 {
			// Some errors might be expected (e.g., authentication)
			t.Logf("Query validation errors: %v", response.Errors)
		}
	})
}

// ExampleTestWithFileUploads demonstrates testing file uploads
func ExampleTestWithFileUploads(t *testing.T) {
	// This would use a GraphQL endpoint that supports file uploads
	// For this example, we'll show the structure

	t.Run("FileUpload", func(t *testing.T) {
		// Create test files
		files := map[string]string{
			"file1.txt": "Hello, World!",
			"file2.txt": "Test content",
		}

		// Parse files using gqlt input handler
		inputHandler := gqlt.NewInput()
		fileList := make([]string, 0, len(files))
		for filename := range files {
			fileList = append(fileList, filename)
		}
		parsedFiles, err := inputHandler.ParseFiles(fileList)
		if err != nil {
			t.Fatalf("Failed to parse files: %v", err)
		}

		// Verify files were parsed correctly
		if len(parsedFiles) != len(files) {
			t.Errorf("Expected %d files, got %d", len(files), len(parsedFiles))
		}

		// Test file upload mutation (conceptual)
		mutation := `
			mutation UploadFiles($files: [Upload!]!) {
				uploadFiles(files: $files) {
					id
					filename
					size
				}
			}
		`

		// This would be used with a real GraphQL endpoint
		t.Logf("Files ready for upload: %v", parsedFiles)
		t.Logf("Mutation: %s", mutation)
	})
}
