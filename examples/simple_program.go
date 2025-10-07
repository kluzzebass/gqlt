package main

import (
	"fmt"
	"log"

	"github.com/kluzzebass/gqlt"
)

// This is a simple example program showing how to use gqlt as a library
func main() {
	// Example 1: Basic query execution
	fmt.Println("=== Basic Query Example ===")

	// Create a GraphQL client
	client := gqlt.NewClient("https://api.github.com/graphql", nil)

	// Define a simple query
	query := `
		query {
			viewer {
				login
				name
			}
		}
	`

	// Execute the query
	response, err := client.Execute(query, nil, "")
	if err != nil {
		log.Printf("Query execution failed: %v", err)
	} else {
		fmt.Printf("Response: %+v\n", response)
	}

	// Example 2: Query with variables
	fmt.Println("\n=== Query with Variables Example ===")

	queryWithVars := `
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

	response, err = client.Execute(queryWithVars, variables, "GetRepository")
	if err != nil {
		log.Printf("Query execution failed: %v", err)
	} else {
		fmt.Printf("Repository data: %+v\n", response.Data)
		if len(response.Errors) > 0 {
			fmt.Printf("Errors: %+v\n", response.Errors)
		}
	}

	// Example 3: Schema introspection
	fmt.Println("\n=== Schema Introspection Example ===")

	introspectClient := gqlt.NewIntrospect(client)
	schema, err := introspectClient.IntrospectSchema()
	if err != nil {
		log.Printf("Schema introspection failed: %v", err)
	} else {
		fmt.Printf("Schema introspection successful\n")
		if len(schema.Errors) > 0 {
			fmt.Printf("Schema errors: %+v\n", schema.Errors)
		}

		// Analyze the schema
		analyzer, err := gqlt.NewAnalyzer(schema)
		if err != nil {
			log.Printf("Failed to create analyzer: %v", err)
		} else {
			summary, err := analyzer.GetSummary()
			if err != nil {
				log.Printf("Failed to get schema summary: %v", err)
			} else {
				fmt.Printf("Schema summary:\n")
				fmt.Printf("  Total types: %d\n", summary.TotalTypes)
				fmt.Printf("  Query type: %s\n", summary.QueryType)
				fmt.Printf("  Mutation type: %s\n", summary.MutationType)
				fmt.Printf("  Subscription type: %s\n", summary.SubscriptionType)
			}
		}
	}

	// Example 4: Configuration management
	fmt.Println("\n=== Configuration Management Example ===")

	// Load configuration
	cfg, err := gqlt.Load(".")
	if err != nil {
		log.Printf("Failed to load config: %v", err)
	} else {
		fmt.Printf("Current config: %s\n", cfg.Current)
		configNames := make([]string, 0, len(cfg.Configs))
		for name := range cfg.Configs {
			configNames = append(configNames, name)
		}
		fmt.Printf("Available configs: %v\n", configNames)

		// Create a new configuration
		cfg.Configs["example"] = gqlt.ConfigEntry{
			Endpoint: "https://api.example.com/graphql",
			Headers: map[string]string{
				"Authorization": "Bearer example-token",
			},
			Defaults: struct {
				Out string `json:"out"`
			}{Out: "json"},
		}

		// Save configuration
		err = cfg.Save(".")
		if err != nil {
			log.Printf("Failed to save config: %v", err)
		} else {
			fmt.Println("Configuration saved successfully")
		}
	}

	// Example 5: Input handling
	fmt.Println("\n=== Input Handling Example ===")

	inputHandler := gqlt.NewInput()

	// Test query loading
	queryStr, err := inputHandler.LoadQuery("{ users { id name } }", "")
	if err != nil {
		log.Printf("Failed to load query: %v", err)
	} else {
		fmt.Printf("Loaded query: %s\n", queryStr)
	}

	// Test variables loading
	variablesJSON := `{"id": "123", "name": "test"}`
	variablesMap, err := inputHandler.LoadVariables(variablesJSON, "")
	if err != nil {
		log.Printf("Failed to load variables: %v", err)
	} else {
		fmt.Printf("Loaded variables: %+v\n", variablesMap)
	}

	// Test headers loading
	headersList := []string{
		"Authorization: Bearer token",
		"Content-Type: application/json",
	}
	headersMap := inputHandler.LoadHeaders(headersList)
	fmt.Printf("Loaded headers: %+v\n", headersMap)
}
