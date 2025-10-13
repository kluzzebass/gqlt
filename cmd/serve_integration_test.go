package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// TestServeCommand_Integration tests the mock server end-to-end
func TestServeCommand_Integration(t *testing.T) {
	// Start server in background
	go func() {
		cmd := rootCmd
		cmd.SetArgs([]string{"serve", "--listen", "localhost:18090", "--playground=false"})
		_ = cmd.Execute()
	}()

	// Wait for server to start
	time.Sleep(2 * time.Second)
	defer func() {
		// Server will be cleaned up when test ends
	}()

	baseURL := "http://localhost:18090/graphql"

	t.Run("Introspection", func(t *testing.T) {
		query := `{ __schema { queryType { name } mutationType { name } subscriptionType { name } } }`
		resp := executeHTTPQuery(t, baseURL, query, nil)

		schemaData, ok := resp["__schema"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected __schema in response")
		}

		if queryType, ok := schemaData["queryType"].(map[string]interface{}); ok {
			if name := queryType["name"]; name != "Query" {
				t.Errorf("Expected Query type, got %v", name)
			}
		}
		if mutationType, ok := schemaData["mutationType"].(map[string]interface{}); ok {
			if name := mutationType["name"]; name != "Mutation" {
				t.Errorf("Expected Mutation type, got %v", name)
			}
		}
		if subscriptionType, ok := schemaData["subscriptionType"].(map[string]interface{}); ok {
			if name := subscriptionType["name"]; name != "Subscription" {
				t.Errorf("Expected Subscription type, got %v", name)
			}
		}
	})

	t.Run("Query_Hello", func(t *testing.T) {
		resp := executeHTTPQuery(t, baseURL, `{ hello }`, nil)
		if hello := resp["hello"]; hello != "Hello, GraphQL!" {
			t.Errorf("Expected 'Hello, GraphQL!', got %v", hello)
		}
	})

	t.Run("Query_Echo", func(t *testing.T) {
		resp := executeHTTPQuery(t, baseURL, `{ echo(message: "test") }`, nil)
		if echo := resp["echo"]; echo != "test" {
			t.Errorf("Expected 'test', got %v", echo)
		}
	})

	t.Run("Query_Users", func(t *testing.T) {
		resp := executeHTTPQuery(t, baseURL, `{ users { id name email role } }`, nil)
		users, ok := resp["users"].([]interface{})
		if !ok {
			t.Fatalf("Expected users array")
		}
		if len(users) != 3 {
			t.Errorf("Expected 3 pre-seeded users, got %d", len(users))
		}

		// Check first user is Alice Admin
		if user, ok := users[0].(map[string]interface{}); ok {
			if user["name"] != "Alice Admin" || user["role"] != "ADMIN" {
				t.Errorf("Expected Alice Admin with ADMIN role, got %v", user)
			}
		}
	})

	t.Run("Query_User_ById", func(t *testing.T) {
		resp := executeHTTPQuery(t, baseURL, `{ user(id: "User:1") { id name email } }`, nil)
		user, ok := resp["user"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected user object")
		}
		if user["id"] != "User:1" || user["name"] != "Alice Admin" {
			t.Errorf("Expected User:1 (Alice Admin), got %v", user)
		}
	})

	t.Run("Query_RelayNode", func(t *testing.T) {
		resp := executeHTTPQuery(t, baseURL, `{ node(id: "User:2") { id ... on User { name email } } }`, nil)
		node, ok := resp["node"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected node object")
		}
		if node["id"] != "User:2" || node["name"] != "Bob User" {
			t.Errorf("Expected User:2 (Bob User), got %v", node)
		}
	})

	t.Run("Query_Search_Union", func(t *testing.T) {
		resp := executeHTTPQuery(t, baseURL, `{ search(term: "alice") { ... on User { id name } ... on Todo { id title } } }`, nil)
		results, ok := resp["search"].([]interface{})
		if !ok {
			t.Fatalf("Expected search results array")
		}
		if len(results) < 1 {
			t.Errorf("Expected at least 1 search result for 'alice'")
		}
	})

	t.Run("Mutation_CreateUser", func(t *testing.T) {
		mutation := `mutation { createUser(input: { name: "Test User", email: "test@example.com" }) { id name email role } }`
		resp := executeHTTPQuery(t, baseURL, mutation, nil)
		user, ok := resp["createUser"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected createUser object")
		}
		if user["name"] != "Test User" || user["email"] != "test@example.com" {
			t.Errorf("Expected created user data, got %v", user)
		}
		if !strings.HasPrefix(user["id"].(string), "User:") {
			t.Errorf("Expected User: ID prefix, got %v", user["id"])
		}
	})

	t.Run("Mutation_CreateTodo", func(t *testing.T) {
		mutation := `mutation { createTodo(input: { title: "Integration Test Todo" }) { id title status priority } }`
		resp := executeHTTPQuery(t, baseURL, mutation, nil)
		todo, ok := resp["createTodo"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected createTodo object")
		}
		if todo["title"] != "Integration Test Todo" || todo["status"] != "PENDING" {
			t.Errorf("Expected created todo data, got %v", todo)
		}
	})

	t.Run("Mutation_UpdateTodo", func(t *testing.T) {
		// First create a todo
		createMutation := `mutation { createTodo(input: { title: "To Update" }) { id } }`
		createResp := executeHTTPQuery(t, baseURL, createMutation, nil)
		todo := createResp["createTodo"].(map[string]interface{})
		todoID := todo["id"].(string)

		// Then update it
		updateMutation := fmt.Sprintf(`mutation { updateTodo(input: { id: "%s", title: "Updated Title" }) { id title } }`, todoID)
		updateResp := executeHTTPQuery(t, baseURL, updateMutation, nil)
		updated := updateResp["updateTodo"].(map[string]interface{})
		if updated["title"] != "Updated Title" {
			t.Errorf("Expected 'Updated Title', got %v", updated["title"])
		}
	})

	t.Run("Mutation_DeleteTodo", func(t *testing.T) {
		// First create a todo
		createMutation := `mutation { createTodo(input: { title: "To Delete" }) { id } }`
		createResp := executeHTTPQuery(t, baseURL, createMutation, nil)
		todo := createResp["createTodo"].(map[string]interface{})
		todoID := todo["id"].(string)

		// Then delete it
		deleteMutation := fmt.Sprintf(`mutation { deleteTodo(id: "%s") }`, todoID)
		deleteResp := executeHTTPQuery(t, baseURL, deleteMutation, nil)
		if deleted := deleteResp["deleteTodo"]; deleted != true {
			t.Errorf("Expected true for deleteTodo, got %v", deleted)
		}

		// Verify it's gone
		query := fmt.Sprintf(`{ todo(id: "%s") { id } }`, todoID)
		queryResp := executeHTTPQuery(t, baseURL, query, nil)
		if queryResp["todo"] != nil {
			t.Errorf("Expected todo to be deleted, but it still exists")
		}
	})

	t.Run("Subscription_Counter_WebSocket", func(t *testing.T) {
		wsURL := "ws://localhost:18090/graphql"
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Connect
		conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
			Subprotocols: []string{"graphql-ws"},
		})
		if err != nil {
			t.Fatalf("WebSocket dial failed: %v", err)
		}
		defer conn.Close(websocket.StatusNormalClosure, "")

		// Send connection_init
		if err := wsjson.Write(ctx, conn, map[string]interface{}{"type": "connection_init"}); err != nil {
			t.Fatalf("Failed to send connection_init: %v", err)
		}

		// Read connection_ack
		var ackMsg map[string]interface{}
		if err := wsjson.Read(ctx, conn, &ackMsg); err != nil {
			t.Fatalf("Failed to read connection_ack: %v", err)
		}
		if ackMsg["type"] != "connection_ack" {
			t.Fatalf("Expected connection_ack, got %v", ackMsg["type"])
		}

		// Subscribe to counter
		subMsg := map[string]interface{}{
			"id":      "1",
			"type":    "start",
			"payload": map[string]interface{}{"query": "subscription { counter }"},
		}
		if err := wsjson.Write(ctx, conn, subMsg); err != nil {
			t.Fatalf("Failed to send subscription: %v", err)
		}

		// Read messages and verify counter increments
		receivedCount := 0
		lastCounter := 0
		for i := 0; i < 3; i++ {
			var msg map[string]interface{}
			if err := wsjson.Read(ctx, conn, &msg); err != nil {
				t.Fatalf("Failed to read message %d: %v", i, err)
			}

			// Skip keep-alive messages
			if msg["type"] == "ka" {
				i--
				continue
			}

			if msg["type"] != "data" {
				t.Errorf("Expected 'data' message, got %v", msg["type"])
				continue
			}

			payload, ok := msg["payload"].(map[string]interface{})
			if !ok {
				t.Errorf("Expected payload in message")
				continue
			}

			data, ok := payload["data"].(map[string]interface{})
			if !ok {
				t.Errorf("Expected data in payload")
				continue
			}

			counterValue := int(data["counter"].(float64))
			if counterValue <= lastCounter {
				t.Errorf("Counter should increment, got %d after %d", counterValue, lastCounter)
			}
			lastCounter = counterValue
			receivedCount++
		}

		if receivedCount != 3 {
			t.Errorf("Expected 3 counter messages, got %d", receivedCount)
		}
	})

	t.Run("Subscription_TodoEvents_SSE", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Start SSE subscription in background
		events := make(chan map[string]interface{}, 10)
		errors := make(chan error, 1)

		go func() {
			defer close(events)
			defer close(errors)

			req, _ := http.NewRequestWithContext(ctx, "POST", baseURL, strings.NewReader(`{"query":"subscription { todoEvents { id title status } }"}`))
			req.Header.Set("Accept", "text/event-stream")
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				errors <- err
				return
			}
			defer resp.Body.Close()

			// Read SSE events
			reader := resp.Body
			buf := make([]byte, 1024)
			var eventData string

			for {
				n, err := reader.Read(buf)
				if err != nil {
					if err != io.EOF {
						errors <- err
					}
					return
				}

				chunk := string(buf[:n])
				lines := strings.Split(chunk, "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "data: ") {
						eventData = strings.TrimPrefix(line, "data: ")
						var eventMsg map[string]interface{}
						if err := json.Unmarshal([]byte(eventData), &eventMsg); err == nil {
							events <- eventMsg
						}
					}
				}
			}
		}()

		// Wait a bit for subscription to be ready
		time.Sleep(500 * time.Millisecond)

		// Trigger a todo creation to generate an event
		mutation := `mutation { createTodo(input: { title: "SSE Test Todo" }) { id title } }`
		_ = executeHTTPQuery(t, baseURL, mutation, nil)

		// Wait for event
		select {
		case event := <-events:
			data, ok := event["data"].(map[string]interface{})
			if !ok {
				t.Fatalf("Expected data in event")
			}
			todoEvents, ok := data["todoEvents"].(map[string]interface{})
			if !ok {
				t.Fatalf("Expected todoEvents in data")
			}
			if todoEvents["title"] != "SSE Test Todo" {
				t.Errorf("Expected 'SSE Test Todo' in event, got %v", todoEvents)
			}
		case err := <-errors:
			t.Fatalf("SSE subscription error: %v", err)
		case <-time.After(3 * time.Second):
			t.Fatal("Timeout waiting for todo event")
		}
	})

	t.Run("Query_Pagination", func(t *testing.T) {
		// Create 5 users
		for i := 1; i <= 5; i++ {
			mutation := fmt.Sprintf(`mutation { createUser(input: { name: "User %d", email: "user%d@test.com" }) { id } }`, i, i)
			_ = executeHTTPQuery(t, baseURL, mutation, nil)
		}

		// Query with limit
		resp := executeHTTPQuery(t, baseURL, `{ users(limit: 3) { id name } }`, nil)
		users, ok := resp["users"].([]interface{})
		if !ok {
			t.Fatalf("Expected users array")
		}
		if len(users) != 3 {
			t.Errorf("Expected 3 users with limit, got %d", len(users))
		}
	})

	t.Run("Query_Filtering", func(t *testing.T) {
		// Create todos with different statuses
		executeHTTPQuery(t, baseURL, `mutation { createTodo(input: { title: "Pending Todo" }) { id } }`, nil)

		resp := executeHTTPQuery(t, baseURL, `{ todos(filters: { status: PENDING }) { title status } }`, nil)
		todos, ok := resp["todos"].([]interface{})
		if !ok {
			t.Fatalf("Expected todos array")
		}
		
		// Verify all returned todos are PENDING
		for _, todoInterface := range todos {
			todo := todoInterface.(map[string]interface{})
			if todo["status"] != "PENDING" {
				t.Errorf("Expected PENDING status, got %v", todo["status"])
			}
		}
	})

	t.Run("Mutation_CompleteTodo", func(t *testing.T) {
		// Create a todo
		createResp := executeHTTPQuery(t, baseURL, `mutation { createTodo(input: { title: "To Complete" }) { id status } }`, nil)
		todo := createResp["createTodo"].(map[string]interface{})
		todoID := todo["id"].(string)

		if todo["status"] != "PENDING" {
			t.Errorf("New todo should be PENDING, got %v", todo["status"])
		}

		// Complete it
		completeMutation := fmt.Sprintf(`mutation { completeTodo(id: "%s") { id status } }`, todoID)
		completeResp := executeHTTPQuery(t, baseURL, completeMutation, nil)
		completed := completeResp["completeTodo"].(map[string]interface{})

		if completed["status"] != "COMPLETED" {
			t.Errorf("Expected COMPLETED status, got %v", completed["status"])
		}
	})

	t.Run("Query_Search_Union", func(t *testing.T) {
		// Search for something that exists
		resp := executeHTTPQuery(t, baseURL, `{ search(term: "User") { ... on User { id name } ... on Todo { id title } } }`, nil)
		results, ok := resp["search"].([]interface{})
		if !ok {
			t.Fatalf("Expected search results")
		}
		if len(results) == 0 {
			t.Error("Expected search results for 'User'")
		}

		// Verify results have either name (User) or title (Todo)
		for _, result := range results {
			r := result.(map[string]interface{})
			hasName := r["name"] != nil
			hasTitle := r["title"] != nil
			if !hasName && !hasTitle {
				t.Errorf("Search result should have name or title, got %v", r)
			}
		}
	})
}

// executeHTTPQuery is a helper to execute GraphQL queries via HTTP POST
func executeHTTPQuery(t *testing.T, url, query string, variables map[string]interface{}) map[string]interface{} {
	t.Helper()

	payload := map[string]interface{}{
		"query": query,
	}
	if variables != nil {
		payload["variables"] = variables
	}

	payloadBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, strings.NewReader(string(payloadBytes)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if errors, ok := result["errors"]; ok && errors != nil {
		t.Fatalf("GraphQL errors: %v", errors)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected data in response, got %v", result)
	}

	return data
}

