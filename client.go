package gqlt

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// Client represents a GraphQL client
type Client struct {
	endpoint   string
	headers    map[string]string
	httpClient *http.Client
}

// Response represents a GraphQL response
type Response struct {
	Data       interface{}            `json:"data"`
	Errors     []interface{}          `json:"errors,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// NewClient creates a new GraphQL client
func NewClient(endpoint string, headers map[string]string) *Client {
	return &Client{
		endpoint:   endpoint,
		headers:    headers,
		httpClient: &http.Client{},
	}
}

// SetAuth sets basic authentication for the client
func (c *Client) SetAuth(username, password string) {
	c.httpClient = &http.Client{
		Transport: &basicAuthTransport{
			username: username,
			password: password,
		},
	}
}

// SetHeaders sets additional headers for the client
func (c *Client) SetHeaders(headers map[string]string) {
	if c.headers == nil {
		c.headers = make(map[string]string)
	}
	for k, v := range headers {
		c.headers[k] = v
	}
}

// Execute executes a GraphQL operation
func (c *Client) Execute(query string, variables map[string]interface{}, operationName string) (*Response, error) {
	// Build GraphQL request payload
	payload := map[string]interface{}{
		"query": query,
	}

	if operationName != "" {
		payload["operationName"] = operationName
	}

	if len(variables) > 0 {
		payload["variables"] = variables
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal GraphQL request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "POST", c.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute GraphQL request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var result Response
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	return &result, nil
}

// ExecuteWithFiles executes a GraphQL operation with file uploads using multipart/form-data
func (c *Client) ExecuteWithFiles(query string, variables map[string]interface{}, operationName string, files map[string]string) (*Response, error) {
	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add GraphQL operation fields
	operations := map[string]interface{}{
		"query": query,
	}
	if operationName != "" {
		operations["operationName"] = operationName
	}
	if len(variables) > 0 {
		operations["variables"] = variables
	}

	operationsJSON, err := json.Marshal(operations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal operations: %w", err)
	}

	// Add operations field
	operationsField, err := writer.CreateFormField("operations")
	if err != nil {
		return nil, fmt.Errorf("failed to create operations field: %w", err)
	}
	_, err = operationsField.Write(operationsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to write operations: %w", err)
	}

	// Add map field for file mappings
	if len(files) > 0 {
		mapField, err := writer.CreateFormField("map")
		if err != nil {
			return nil, fmt.Errorf("failed to create map field: %w", err)
		}

		// Create file mapping JSON
		fileMap := make(map[string][]string)
		for name := range files {
			fileMap[name] = []string{"variables." + name}
		}

		mapJSON, err := json.Marshal(fileMap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal file map: %w", err)
		}

		_, err = mapField.Write(mapJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to write file map: %w", err)
		}

		// Add files
		for name, path := range files {
			file, err := os.Open(path)
			if err != nil {
				return nil, fmt.Errorf("failed to open file %s: %w", path, err)
			}
			defer file.Close()

			part, err := writer.CreateFormFile(name, filepath.Base(path))
			if err != nil {
				return nil, fmt.Errorf("failed to create form file for %s: %w", name, err)
			}

			_, err = io.Copy(part, file)
			if err != nil {
				return nil, fmt.Errorf("failed to copy file %s: %w", path, err)
			}
		}
	}

	// Close the writer
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(context.Background(), "POST", c.endpoint, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute GraphQL request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response
	var result Response
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse GraphQL response: %w", err)
	}

	return &result, nil
}

// Introspect performs GraphQL introspection to get the schema
func (c *Client) Introspect() (*Response, error) {
	introspectionQuery := `
		query IntrospectionQuery {
			__schema {
				queryType { name }
				mutationType { name }
				subscriptionType { name }
				types {
					...FullType
				}
				directives {
					name
					description
					locations
					args {
						...InputValue
					}
				}
			}
		}

		fragment FullType on __Type {
			kind
			name
			description
			fields(includeDeprecated: true) {
				name
				description
				args {
					...InputValue
				}
				type {
					...TypeRef
				}
				isDeprecated
				deprecationReason
			}
			inputFields {
				...InputValue
			}
			interfaces {
				...TypeRef
			}
			possibleTypes {
				...TypeRef
			}
			enumValues(includeDeprecated: true) {
				name
				description
				isDeprecated
				deprecationReason
			}
		}

		fragment InputValue on __InputValue {
			name
			description
			type { ...TypeRef }
			defaultValue
		}

		fragment TypeRef on __Type {
			kind
			name
			ofType {
				kind
				name
				ofType {
					kind
					name
					ofType {
						kind
						name
						ofType {
							kind
							name
							ofType {
								kind
								name
								ofType {
									kind
									name
									ofType {
										kind
										name
									}
								}
							}
						}
					}
				}
			}
		}
	`

	return c.Execute(introspectionQuery, nil, "IntrospectionQuery")
}

// basicAuthTransport implements HTTP transport with basic authentication
type basicAuthTransport struct {
	username string
	password string
}

func (t *basicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	auth := t.username + ":" + t.password
	encoded := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", "Basic "+encoded)
	return http.DefaultTransport.RoundTrip(req)
}
