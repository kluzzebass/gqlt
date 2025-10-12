package gqlt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// SSESubscriptionClient handles GraphQL subscriptions over Server-Sent Events
type SSESubscriptionClient struct {
	url     string
	headers map[string]string
	client  *http.Client
	mu      sync.Mutex
}

// SSE message types for graphql-sse protocol
const (
	SSEMessageTypeNext     = "next"
	SSEMessageTypeError    = "error"
	SSEMessageTypeComplete = "complete"
)

// SSEMessage represents a message received from SSE subscription
type SSEMessage struct {
	Type    string                 `json:"type"`
	ID      string                 `json:"id,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// NewSSESubscriptionClient creates a new SSE subscription client
func NewSSESubscriptionClient(url string, headers map[string]string) *SSESubscriptionClient {
	return &SSESubscriptionClient{
		url:     url,
		headers: headers,
		client:  &http.Client{Timeout: 0}, // No timeout for SSE
	}
}

// Subscribe starts a subscription and returns channels for messages and errors
func (c *SSESubscriptionClient) Subscribe(ctx context.Context, query string, variables map[string]interface{}, operationName string) (<-chan *SubscriptionMessage, <-chan error, error) {
	// Prepare the subscription payload
	payload := map[string]interface{}{
		"query":         query,
		"variables":     variables,
		"operationName": operationName,
	}

	// Convert to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal subscription payload: %w", err)
	}

	// Create HTTP request with POST body
	req, err := http.NewRequestWithContext(ctx, "POST", c.url, strings.NewReader(string(payloadJSON)))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// Add custom headers
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// Add graphql-sse specific headers
	req.Header.Set("graphql-preflight", "1")

	// Make the request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start SSE subscription: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, nil, fmt.Errorf("SSE subscription failed with status: %d", resp.StatusCode)
	}

	// Create channels
	messages := make(chan *SubscriptionMessage, 10)
	errors := make(chan error, 10)

	// Start reading SSE stream in goroutine
	go func() {
		defer close(messages)
		defer close(errors)
		defer resp.Body.Close()

		reader := NewSSEReader(resp.Body)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				event, err := reader.ReadEvent()
				if err != nil {
					if err == io.EOF {
						return
					}
					errors <- fmt.Errorf("failed to read SSE event: %w", err)
					return
				}

				if event.Type == "next" && event.Data != "" {
					// Parse the GraphQL response directly from the SSE data
					var graphqlResponse map[string]interface{}
					if err := json.Unmarshal([]byte(event.Data), &graphqlResponse); err != nil {
						errors <- fmt.Errorf("failed to parse GraphQL response: %w", err)
						continue
					}

					// Extract GraphQL response
					subMsg := &SubscriptionMessage{
						Data:   graphqlResponse["data"],
						Errors: []interface{}{},
					}
					if errs, ok := graphqlResponse["errors"].([]interface{}); ok {
						subMsg.Errors = errs
					}
					messages <- subMsg
				} else if event.Type == "complete" {
					return
				}
			}
		}
	}()

	return messages, errors, nil
}

// SSEReader reads Server-Sent Events from an io.Reader
type SSEReader struct {
	reader io.Reader
	buffer []byte
}

// SSEEvent represents a parsed SSE event
type SSEEvent struct {
	Type string
	Data string
	ID   string
}

// NewSSEReader creates a new SSE reader
func NewSSEReader(reader io.Reader) *SSEReader {
	return &SSEReader{
		reader: reader,
		buffer: make([]byte, 4096),
	}
}

// ReadEvent reads the next SSE event
func (r *SSEReader) ReadEvent() (*SSEEvent, error) {
	var event SSEEvent

	for {
		n, err := r.reader.Read(r.buffer)
		if err != nil {
			return nil, err
		}

		lines := strings.Split(string(r.buffer[:n]), "\n")
		
		for _, line := range lines {
			line = strings.TrimSpace(line)
			
			if line == "" {
				// Empty line indicates end of event
				if event.Type != "" || event.Data != "" {
					return &event, nil
				}
				continue
			}

			if strings.HasPrefix(line, ":") {
				// Comment line, ignore
				continue
			}

			colonIndex := strings.Index(line, ":")
			if colonIndex == -1 {
				continue
			}

			field := line[:colonIndex]
			value := strings.TrimSpace(line[colonIndex+1:])

			switch field {
			case "event":
				event.Type = value
			case "data":
				if event.Data != "" {
					event.Data += "\n" + value
				} else {
					event.Data = value
				}
			case "id":
				event.ID = value
			}
		}

		// If we have a complete event, return it
		if event.Type != "" || event.Data != "" {
			return &event, nil
		}
	}
}
