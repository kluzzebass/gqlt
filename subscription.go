package gqlt

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// SubscriptionClient handles GraphQL subscriptions over WebSocket
type SubscriptionClient struct {
	url     string
	headers map[string]string
	conn    *websocket.Conn
	mu      sync.Mutex
}

// GraphQL WebSocket Protocol Messages (graphql-transport-ws)
const (
	MessageTypeConnectionInit = "connection_init"
	MessageTypeConnectionAck  = "connection_ack"
	MessageTypePing           = "ping"
	MessageTypePong           = "pong"
	MessageTypeSubscribe      = "subscribe"
	MessageTypeNext           = "next"
	MessageTypeError          = "error"
	MessageTypeComplete       = "complete"
)

// WebSocket message structure
type wsMessage struct {
	ID      string                 `json:"id,omitempty"`
	Type    string                 `json:"type"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// SubscriptionMessage represents a message received from a subscription
type SubscriptionMessage struct {
	Data   interface{}   `json:"data,omitempty"`
	Errors []interface{} `json:"errors,omitempty"`
}

// NewSubscriptionClient creates a new WebSocket subscription client
func NewSubscriptionClient(url string, headers map[string]string) *SubscriptionClient {
	return &SubscriptionClient{
		url:     url,
		headers: headers,
	}
}

// Connect establishes a WebSocket connection and performs the connection handshake
func (c *SubscriptionClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Set up WebSocket dial options with headers
	opts := &websocket.DialOptions{
		HTTPHeader:   http.Header{},
		Subprotocols: []string{"graphql-transport-ws", "graphql-ws", "apollo-ws"},
	}

	// Add custom headers
	for k, v := range c.headers {
		opts.HTTPHeader.Set(k, v)
	}

	// Establish WebSocket connection
	conn, _, err := websocket.Dial(ctx, c.url, opts)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	c.conn = conn

	// Send connection_init message
	initMsg := wsMessage{
		Type: MessageTypeConnectionInit,
		Payload: map[string]interface{}{
			// Include headers as connection params for auth
			"headers": c.headers,
		},
	}

	if err := wsjson.Write(ctx, c.conn, initMsg); err != nil {
		c.conn.Close(websocket.StatusInternalError, "Failed to send connection_init")
		return fmt.Errorf("failed to send connection_init: %w", err)
	}

	// Wait for connection_ack
	var ackMsg wsMessage
	if err := wsjson.Read(ctx, c.conn, &ackMsg); err != nil {
		c.conn.Close(websocket.StatusInternalError, "Failed to receive connection_ack")
		return fmt.Errorf("failed to receive connection_ack: %w", err)
	}

	if ackMsg.Type != MessageTypeConnectionAck {
		c.conn.Close(websocket.StatusPolicyViolation, "Expected connection_ack")
		return fmt.Errorf("expected connection_ack, got %s", ackMsg.Type)
	}

	return nil
}

// Subscribe sends a subscription request and returns a channel of messages
func (c *SubscriptionClient) Subscribe(ctx context.Context, query string, variables map[string]interface{}, operationName string) (<-chan *SubscriptionMessage, <-chan error, error) {
	c.mu.Lock()
	if c.conn == nil {
		c.mu.Unlock()
		return nil, nil, fmt.Errorf("not connected - call Connect() first")
	}
	c.mu.Unlock()

	// Generate subscription ID
	subscriptionID := fmt.Sprintf("sub_%d", time.Now().UnixNano())

	// Send subscribe message
	subscribeMsg := wsMessage{
		ID:   subscriptionID,
		Type: MessageTypeSubscribe,
		Payload: map[string]interface{}{
			"query": query,
		},
	}

	if operationName != "" {
		subscribeMsg.Payload["operationName"] = operationName
	}

	if variables != nil {
		subscribeMsg.Payload["variables"] = variables
	}

	if err := wsjson.Write(ctx, c.conn, subscribeMsg); err != nil {
		return nil, nil, fmt.Errorf("failed to send subscribe message: %w", err)
	}

	// Create channels for messages and errors
	messages := make(chan *SubscriptionMessage, 10)
	errors := make(chan error, 1)

	// Start goroutine to receive messages
	go c.receiveMessages(ctx, subscriptionID, messages, errors)

	return messages, errors, nil
}

// receiveMessages listens for WebSocket messages and sends them to the appropriate channel
func (c *SubscriptionClient) receiveMessages(ctx context.Context, subscriptionID string, messages chan<- *SubscriptionMessage, errors chan<- error) {
	defer close(messages)
	defer close(errors)

	for {
		select {
		case <-ctx.Done():
			// Context cancelled - clean up and return
			c.Unsubscribe(context.Background(), subscriptionID)
			return
		default:
			// Try to read a message with a timeout to check context periodically
			readCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			var msg wsMessage
			err := wsjson.Read(readCtx, c.conn, &msg)
			cancel()

			if err != nil {
				// Check if context was cancelled
				if ctx.Err() != nil {
					return
				}
				// Check if it's just a timeout (continue reading)
				if readCtx.Err() == context.DeadlineExceeded {
					continue
				}
				// Real error
				errors <- fmt.Errorf("failed to read message: %w", err)
				return
			}

			// Handle message based on type
			switch msg.Type {
			case MessageTypeNext:
				// Subscription data message
				if msg.ID == subscriptionID {
					if payload, ok := msg.Payload["data"].(map[string]interface{}); ok {
						subMsg := &SubscriptionMessage{
							Data: payload,
						}
						if errs, ok := msg.Payload["errors"].([]interface{}); ok {
							subMsg.Errors = errs
						}
						messages <- subMsg
					}
				}

			case MessageTypeError:
				// Subscription error
				if msg.ID == subscriptionID {
					errors <- fmt.Errorf("subscription error: %v", msg.Payload)
					return
				}

			case MessageTypeComplete:
				// Subscription completed
				if msg.ID == subscriptionID {
					return
				}

			case MessageTypePing:
				// Respond to ping with pong
				pongMsg := wsMessage{Type: MessageTypePong}
				wsjson.Write(ctx, c.conn, pongMsg)
			}
		}
	}
}

// Unsubscribe sends a complete message to stop the subscription
func (c *SubscriptionClient) Unsubscribe(ctx context.Context, subscriptionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	completeMsg := wsMessage{
		ID:   subscriptionID,
		Type: MessageTypeComplete,
	}

	return wsjson.Write(ctx, c.conn, completeMsg)
}

// Close closes the WebSocket connection
func (c *SubscriptionClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	err := c.conn.Close(websocket.StatusNormalClosure, "")
	c.conn = nil
	return err
}
