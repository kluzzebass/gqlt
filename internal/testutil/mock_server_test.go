package testutil

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kluzzebass/gqlt"
)

func TestNewMockGraphQLServer(t *testing.T) {
	server := NewMockGraphQLServer()
	defer server.Close()

	if server == nil {
		t.Fatal("Expected non-nil server")
	}

	if server.URL() == "" {
		t.Error("Expected non-empty URL")
	}

	if !server.introspectionOn {
		t.Error("Expected introspection to be enabled by default")
	}

	if len(server.handlers) != 0 {
		t.Error("Expected empty handlers map initially")
	}
}

func TestMockGraphQLServer_BasicQuery(t *testing.T) {
	server := NewMockGraphQLServer()
	defer server.Close()

	// Add handler for test operation
	server.AddHandler("GetUser", func(req Request) *gqlt.Response {
		return SuccessResponse(map[string]interface{}{
			"user": map[string]interface{}{
				"id":   "123",
				"name": "Test User",
			},
		})
	})

	// Create client and execute query
	client := gqlt.NewClient(server.URL(), nil)
	response, err := client.Execute(
		`query GetUser { user { id name } }`,
		nil,
		"GetUser",
	)

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if response.Data == nil {
		t.Fatal("Expected data in response")
	}

	if len(response.Errors) > 0 {
		t.Errorf("Unexpected errors: %v", response.Errors)
	}
}

func TestMockGraphQLServer_QueryWithVariables(t *testing.T) {
	server := NewMockGraphQLServer()
	defer server.Close()

	// Add handler that uses variables
	server.AddHandler("GetUserById", func(req Request) *gqlt.Response {
		id, ok := req.Variables["id"].(string)
		if !ok {
			return ErrorResponse("Invalid id variable")
		}

		return SuccessResponse(map[string]interface{}{
			"user": map[string]interface{}{
				"id":   id,
				"name": "User " + id,
			},
		})
	})

	client := gqlt.NewClient(server.URL(), nil)
	response, err := client.Execute(
		`query GetUserById($id: ID!) { user(id: $id) { id name } }`,
		map[string]interface{}{"id": "456"},
		"GetUserById",
	)

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	user, ok := data["user"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected user to be a map")
	}

	if user["id"] != "456" {
		t.Errorf("Expected user id 456, got %v", user["id"])
	}
}

func TestMockGraphQLServer_DefaultHandler(t *testing.T) {
	server := NewMockGraphQLServer()
	defer server.Close()

	// Set default handler
	server.SetDefaultHandler(func(req Request) *gqlt.Response {
		return SuccessResponse(map[string]interface{}{
			"message": "Default handler called",
		})
	})

	client := gqlt.NewClient(server.URL(), nil)
	response, err := client.Execute(
		`query UnknownOperation { test }`,
		nil,
		"UnknownOperation",
	)

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	if data["message"] != "Default handler called" {
		t.Error("Expected default handler to be called")
	}
}

func TestMockGraphQLServer_NoHandler(t *testing.T) {
	server := NewMockGraphQLServer()
	defer server.Close()

	client := gqlt.NewClient(server.URL(), nil)
	response, err := client.Execute(
		`query UnhandledOperation { test }`,
		nil,
		"UnhandledOperation",
	)

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(response.Errors) == 0 {
		t.Error("Expected error for unhandled operation")
	}
}

func TestMockGraphQLServer_Introspection(t *testing.T) {
	server := NewMockGraphQLServer()
	defer server.Close()

	client := gqlt.NewClient(server.URL(), nil)
	response, err := client.Introspect()

	if err != nil {
		t.Fatalf("Introspection failed: %v", err)
	}

	if response.Data == nil {
		t.Fatal("Expected data in introspection response")
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	schema, ok := data["__schema"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected __schema field")
	}

	if schema["queryType"] == nil {
		t.Error("Expected queryType in schema")
	}

	types, ok := schema["types"].([]interface{})
	if !ok || len(types) == 0 {
		t.Error("Expected non-empty types array")
	}
}

func TestMockGraphQLServer_IntrospectionDisabled(t *testing.T) {
	server := NewMockGraphQLServer()
	defer server.Close()

	// Disable introspection
	server.EnableIntrospection(false)

	// Add default handler to catch introspection query
	server.SetDefaultHandler(func(req Request) *gqlt.Response {
		return ErrorResponse("Introspection disabled")
	})

	client := gqlt.NewClient(server.URL(), nil)
	response, err := client.Introspect()

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if len(response.Errors) == 0 {
		t.Error("Expected error when introspection is disabled")
	}
}

func TestMockGraphQLServer_CustomSchema(t *testing.T) {
	server := NewMockGraphQLServer()
	defer server.Close()

	// Set custom schema
	customSchema := map[string]interface{}{
		"__schema": map[string]interface{}{
			"queryType": map[string]interface{}{
				"name": "CustomQuery",
			},
			"types": []interface{}{
				map[string]interface{}{
					"kind": "OBJECT",
					"name": "CustomType",
				},
			},
		},
	}
	server.SetSchema(customSchema)

	client := gqlt.NewClient(server.URL(), nil)
	response, err := client.Introspect()

	if err != nil {
		t.Fatalf("Introspection failed: %v", err)
	}

	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	schema, ok := data["__schema"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected __schema field")
	}

	queryType, ok := schema["queryType"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected queryType field")
	}

	if queryType["name"] != "CustomQuery" {
		t.Errorf("Expected CustomQuery, got %v", queryType["name"])
	}
}

func TestMockGraphQLServer_Delay(t *testing.T) {
	server := NewMockGraphQLServer()
	defer server.Close()

	// Set 100ms delay
	server.SetDelay(100 * time.Millisecond)

	server.AddHandler("Test", func(req Request) *gqlt.Response {
		return SuccessResponse(map[string]interface{}{"result": "ok"})
	})

	client := gqlt.NewClient(server.URL(), nil)

	start := time.Now()
	_, err := client.Execute(`query Test { test }`, nil, "Test")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if elapsed < 100*time.Millisecond {
		t.Errorf("Expected delay of at least 100ms, got %v", elapsed)
	}
}

func TestMockGraphQLServer_RequestLog(t *testing.T) {
	server := NewMockGraphQLServer()
	defer server.Close()

	server.AddHandler("Op1", func(req Request) *gqlt.Response {
		return SuccessResponse(nil)
	})

	server.AddHandler("Op2", func(req Request) *gqlt.Response {
		return SuccessResponse(nil)
	})

	client := gqlt.NewClient(server.URL(), nil)

	// Execute multiple queries
	client.Execute(`query Op1 { test }`, nil, "Op1")
	client.Execute(`query Op2 { test }`, nil, "Op2")

	log := server.GetRequestLog()
	if len(log) != 2 {
		t.Errorf("Expected 2 logged requests, got %d", len(log))
	}

	if log[0].OperationName != "Op1" {
		t.Errorf("Expected first operation to be Op1, got %s", log[0].OperationName)
	}

	if log[1].OperationName != "Op2" {
		t.Errorf("Expected second operation to be Op2, got %s", log[1].OperationName)
	}

	// Clear log
	server.ClearRequestLog()
	log = server.GetRequestLog()
	if len(log) != 0 {
		t.Errorf("Expected empty log after clear, got %d entries", len(log))
	}
}

func TestMockGraphQLServer_ErrorResponse(t *testing.T) {
	server := NewMockGraphQLServer()
	defer server.Close()

	server.AddHandler("ErrorOp", func(req Request) *gqlt.Response {
		return ErrorResponse("Test error message")
	})

	client := gqlt.NewClient(server.URL(), nil)
	response, err := client.Execute(`query ErrorOp { test }`, nil, "ErrorOp")

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if len(response.Errors) == 0 {
		t.Fatal("Expected errors in response")
	}

	errorMap, ok := response.Errors[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected error to be a map")
	}

	if errorMap["message"] != "Test error message" {
		t.Errorf("Expected test error message, got %v", errorMap["message"])
	}
}

func TestMockGraphQLServer_DataWithErrors(t *testing.T) {
	server := NewMockGraphQLServer()
	defer server.Close()

	server.AddHandler("PartialSuccess", func(req Request) *gqlt.Response {
		return DataWithErrors(
			map[string]interface{}{"partialData": true},
			[]string{"Error 1", "Error 2"},
		)
	})

	client := gqlt.NewClient(server.URL(), nil)
	response, err := client.Execute(`query PartialSuccess { test }`, nil, "PartialSuccess")

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if response.Data == nil {
		t.Error("Expected data in response")
	}

	if len(response.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(response.Errors))
	}
}

func TestMockGraphQLServer_InvalidJSON(t *testing.T) {
	server := NewMockGraphQLServer()
	defer server.Close()

	// Send invalid JSON
	resp, err := http.Post(server.URL(), "application/json", strings.NewReader("{invalid json"))
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}
}

func TestMockGraphQLServer_FileUpload(t *testing.T) {
	server := NewMockGraphQLServer()
	defer server.Close()

	server.AddHandler("UploadFile", func(req Request) *gqlt.Response {
		if req.Files == nil || len(req.Files) == 0 {
			return ErrorResponse("No files uploaded")
		}

		fileData, ok := req.Files["file"]
		if !ok {
			return ErrorResponse("File not found")
		}

		return SuccessResponse(map[string]interface{}{
			"upload": map[string]interface{}{
				"filename": "test.txt",
				"size":     len(fileData),
			},
		})
	})

	// Create multipart request manually
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add operations
	operations := `{"query":"mutation UploadFile { upload }","operationName":"UploadFile"}`
	writer.WriteField("operations", operations)

	// Add map
	writer.WriteField("map", `{"file":["variables.file"]}`)

	// Add file
	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	io.WriteString(part, "test file content")

	writer.Close()

	// Send request
	resp, err := http.Post(server.URL(), writer.FormDataContentType(), &buf)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check request log
	log := server.GetRequestLog()
	if len(log) != 1 {
		t.Fatalf("Expected 1 logged request, got %d", len(log))
	}

	if len(log[0].Files) == 0 {
		t.Error("Expected files in logged request")
	}

	if string(log[0].Files["file"]) != "test file content" {
		t.Errorf("Expected file content 'test file content', got %s", log[0].Files["file"])
	}
}

func TestMockGraphQLServer_HeadersInRequest(t *testing.T) {
	server := NewMockGraphQLServer()
	defer server.Close()

	server.AddHandler("CheckHeaders", func(req Request) *gqlt.Response {
		auth := req.Headers.Get("Authorization")
		if auth == "" {
			return ErrorResponse("Missing authorization header")
		}

		return SuccessResponse(map[string]interface{}{
			"authenticated": true,
		})
	})

	// Test without header
	client1 := gqlt.NewClient(server.URL(), nil)
	response1, _ := client1.Execute(`query CheckHeaders { test }`, nil, "CheckHeaders")

	if len(response1.Errors) == 0 {
		t.Error("Expected error without authorization header")
	}

	// Test with header
	client2 := gqlt.NewClient(server.URL(), map[string]string{
		"Authorization": "Bearer token123",
	})
	response2, err := client2.Execute(`query CheckHeaders { test }`, nil, "CheckHeaders")

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(response2.Errors) > 0 {
		t.Errorf("Unexpected errors: %v", response2.Errors)
	}

	data, ok := response2.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	if data["authenticated"] != true {
		t.Error("Expected authenticated to be true")
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test SuccessResponse
	success := SuccessResponse(map[string]interface{}{"test": "data"})
	if success.Data == nil {
		t.Error("Expected data in success response")
	}
	if len(success.Errors) > 0 {
		t.Error("Expected no errors in success response")
	}

	// Test ErrorResponse
	errResp := ErrorResponse("test error")
	if len(errResp.Errors) == 0 {
		t.Error("Expected errors in error response")
	}

	// Test DataWithErrors
	partial := DataWithErrors(
		map[string]interface{}{"data": true},
		[]string{"error1", "error2"},
	)
	if partial.Data == nil {
		t.Error("Expected data in partial response")
	}
	if len(partial.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(partial.Errors))
	}
}
