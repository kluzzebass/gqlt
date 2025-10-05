package graphql

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client represents a GraphQL client
type Client struct {
	endpoint  string
	headers   map[string]string
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
