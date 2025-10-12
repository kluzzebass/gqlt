================================================================
IMPLEMENTATION PLAN — GraphQL Subscription Support
Language: Go
Goal: Add WebSocket-based GraphQL subscription support for CLI, MCP, and Library modes
================================================================

**Date:** 2025-10-11

**Execution Mode:** TBD

## CONTEXT

GraphQL subscriptions enable real-time data streaming over WebSocket connections. While initially considered incompatible with gqlt's design, subscriptions are valuable for:

1. **Library Mode:** Testing real-time GraphQL features
2. **CLI Mode:** Monitoring live data streams (with proper termination)
3. **MCP Mode:** Potentially useful for streaming results (needs exploration)

## OBJECTIVE

Add full subscription support to gqlt across all three modes (CLI, MCP, Library) while maintaining Unix composability and focused design.

## REQUIREMENTS

### Functional Requirements
- [ ] Establish WebSocket connections to GraphQL subscription endpoints
- [ ] Send subscription operations conforming to graphql-ws protocol
- [ ] Receive and parse subscription messages
- [ ] Handle connection lifecycle (connect, subscribe, unsubscribe, close)
- [ ] Support authentication headers in WebSocket handshake
- [ ] Handle reconnection on connection drops
- [ ] Graceful shutdown on signals (SIGINT, SIGTERM)

### CLI Requirements
- [ ] `gqlt run` command should detect subscription operations
- [ ] Stream messages to stdout as JSON (one per line)
- [ ] Run until user interrupts (Ctrl+C) or connection closes
- [ ] Option to limit messages: `--max-messages N`
- [ ] Option to timeout: `--timeout 30s`
- [ ] Composable output: each message on separate line for piping

### Library Requirements
- [ ] `Client.Subscribe()` method that returns a channel or callback
- [ ] Clean API for testing subscriptions
- [ ] Support in mock server for testing subscription handlers
- [ ] Proper resource cleanup

### MCP Requirements (Exploration Needed)
- [ ] Determine if subscriptions make sense in MCP context
- [ ] If yes: streaming response format?
- [ ] If no: document why in Limitations

## ARCHITECTURE

### WebSocket Protocol
- Use `graphql-transport-ws` protocol (modern standard)
- Fallback to legacy `graphql-ws` protocol if needed
- Connection init with authentication headers

### CLI Output Format
```bash
# Each message on separate line for composability
{"data":{"userCreated":{"id":"123","name":"Alice"}}}
{"data":{"userCreated":{"id":"124","name":"Bob"}}}

# Pipe to jq for filtering
gqlt run --query 'subscription { userCreated { id name } }' | jq 'select(.data.userCreated.name == "Alice")'

# Count messages
gqlt run --query 'subscription { events }' --max-messages 10 | wc -l
```

### Library API
```go
// Option 1: Channel-based
messages := make(chan *Response)
err := client.Subscribe(query, variables, messages)

for msg := range messages {
    // Process message
}

// Option 2: Callback-based
err := client.Subscribe(query, variables, func(msg *Response) error {
    // Process message
    return nil // or error to stop
})
```

## IMPLEMENTATION STEPS

### Phase 0: MCP Cancellation & Timeout Strategy (CRITICAL)
- [x] Research MCP cancellation support
  - [x] Confirmed: MCP SDK supports `notifications/cancelled` automatically
  - [x] When user clicks Cancel in Cursor, our handler's `ctx` gets cancelled
  - [x] We just need to respect `ctx.Done()` in subscription loop
- [ ] Test Cursor's timeout behavior empirically
  - [ ] Create test MCP tool that sleeps for varying durations (10s, 30s, 60s, 120s)
  - [ ] Determine if Cursor has hard timeout or relies on user Cancel button
  - [ ] Document actual timeout behavior
- [ ] Implement multi-stop-condition subscriptions
  - [ ] Stop on `ctx.Done()` (Cursor Cancel button)
  - [ ] Stop on user timeout (e.g., `"timeout": "10s"`)
  - [ ] Stop on max messages (e.g., `"maxMessages": 10`)
  - [ ] Default timeout: 30s (conservative, user can override)
  - [ ] Return partial results when stopped early

### Phase 1: Research & Dependencies
- [ ] Research Go WebSocket libraries (gorilla/websocket, nhooyr/websocket)
- [ ] Research GraphQL subscription protocols (graphql-transport-ws vs graphql-ws)
- [ ] Choose library and protocol
- [ ] Add dependencies

### Phase 2: Core WebSocket Client
- [ ] Create `SubscriptionClient` struct
- [ ] Implement WebSocket connection establishment
- [ ] Implement graphql-transport-ws protocol messages:
  - [ ] `connection_init`
  - [ ] `connection_ack`
  - [ ] `subscribe`
  - [ ] `next` (receive data)
  - [ ] `error`
  - [ ] `complete`
- [ ] Handle authentication in connection_init
- [ ] Implement connection keep-alive (ping/pong)

### Phase 3: Client API
- [ ] Add `Client.Subscribe()` method
- [ ] Decide on API (channel vs callback)
- [ ] Implement message streaming
- [ ] Handle errors and connection drops
- [ ] Resource cleanup (defer, context)
- [ ] Reconnection logic (optional)

### Phase 4: Operation Detection
- [ ] Parse GraphQL document to extract operation definitions
- [ ] Identify operation type (query/mutation/subscription) by name
- [ ] Handle documents with multiple operations:
  - [ ] Use `--operation` flag to select specific operation
  - [ ] Detect the selected operation's type
  - [ ] Error if subscription without operation name in multi-op document
- [ ] Route to appropriate handler:
  - [ ] HTTP client for queries/mutations
  - [ ] WebSocket client for subscriptions

### Phase 5: CLI Integration
- [ ] Integrate operation detection into `run` command
- [ ] Route to subscription handler when operation type is subscription
- [ ] Stream messages to stdout (one JSON per line)
- [ ] Add `--max-messages` flag
- [ ] Add `--timeout` flag
- [ ] Handle SIGINT for graceful shutdown
- [ ] Test with real subscription endpoints
- [ ] Test with mixed operation documents

### Phase 6: Testing Infrastructure
- [ ] Add WebSocket support to mock server
- [ ] Mock subscription message streaming
- [ ] Test subscription lifecycle
- [ ] Test authentication
- [ ] Test error handling
- [ ] Test timeout/max-messages
- [ ] Test signal handling
- [ ] Test operation detection with mixed documents

### Phase 7: MCP Implementation
- [ ] Implement time-bounded subscription for MCP
- [ ] Add `maxMessages` and `timeout` parameters to execute_query
- [ ] Respect `context.Context` deadline from MCP client
- [ ] Collect messages until limit/timeout/context cancellation
- [ ] Return array of messages with metadata
- [ ] Test with various timeout values
- [ ] Ensure subscription completes before MCP timeout
- [ ] Handle context cancellation gracefully

### Phase 8: Documentation
- [ ] Update README with subscription examples
- [ ] Document WebSocket protocol support
- [ ] Add CLI subscription examples with mixed operations
- [ ] Add library subscription examples
- [ ] Document operation detection behavior
- [ ] Document limitations (if any)
- [ ] Update Limitations section if subscriptions don't work in MCP

## OPEN QUESTIONS

1. **MCP Timeout Constraints:** What is Cursor's timeout for MCP tool calls?
   - Need to research/test actual timeout limits (may be none/very long)
   - Cursor provides Cancel button for user control
   - Default timeout should be reasonable (e.g., 30 seconds)
   - Make timeout fully configurable - no artificial caps
   - Respect ctx.Done() when user cancels in Cursor
2. **Library API:** Channel-based or callback-based Subscribe method?
3. **MCP Support:** Can subscriptions work in MCP given timeout constraints?
   - If MCP timeout is 30s, max subscription timeout should be ~25s
   - Need to respect `context.Context` deadline from MCP handler
4. **Protocol:** Support both graphql-transport-ws and legacy graphql-ws?
5. **Reconnection:** Auto-reconnect on connection drop, or let user handle it?
6. **Backpressure:** How to handle slow consumers of subscription messages?

## TESTING STRATEGY

### Unit Tests
- WebSocket connection lifecycle
- Protocol message parsing
- Error handling
- Timeout behavior

### Integration Tests  
- Real WebSocket server (using mock server)
- Authentication flow
- Message streaming
- Graceful shutdown

### Real-World Tests
- Public GraphQL subscription endpoint
- Verify compatibility with real servers

## SUCCESS CRITERIA

- [ ] Can execute subscriptions from CLI
- [ ] Messages stream to stdout as newline-delimited JSON
- [ ] Ctrl+C cleanly terminates subscription
- [ ] Library API works in tests
- [ ] Mock server supports subscription testing
- [ ] All existing tests still pass
- [ ] Documentation is complete
- [ ] MCP subscription support determined (yes/no with rationale)

## NOTES

### MCP Cancellation & Timeout Handling

**MCP SDK Cancellation Support (Confirmed):**
- When user clicks Cancel in Cursor, the `context.Context` gets cancelled
- SDK automatically sends `notifications/cancelled` to server
- Our handler observes cancellation via `ctx.Done()`
- We return partial results collected so far

**Implementation with 3 stop conditions:**

```go
func handleExecuteQuery(ctx context.Context, input ExecuteQueryInput) (*CallToolResult, ExecuteQueryOutput, error) {
    if isSubscription {
        // Parse user timeout (default: 30s)
        timeout := 30 * time.Second
        if input.Timeout != "" {
            timeout, _ = time.ParseDuration(input.Timeout)
        }
        
        // Create timeout context (inherits from parent ctx for cancellation)
        timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
        defer cancel()
        
        messages := []Response{}
        messageCount := 0
        stoppedReason := ""
        
        // Establish WebSocket and subscribe...
        
        // Collect messages until one of three conditions:
        for {
            select {
            case <-timeoutCtx.Done():
                // Could be: user timeout, Cursor cancel, or context deadline
                if ctx.Err() == context.Canceled {
                    stoppedReason = "cancelled_by_user"  // Cursor Cancel button
                } else {
                    stoppedReason = "timeout_reached"
                }
                return &CallToolResult{}, ExecuteQueryOutput{
                    Data: map[string]interface{}{
                        "messages": messages,
                        "count": messageCount,
                        "stoppedReason": stoppedReason,
                    },
                }, nil
                
            case msg := <-subscriptionChannel:
                messages = append(messages, msg)
                messageCount++
                
                // Stop condition 3: Max messages
                if input.MaxMessages > 0 && messageCount >= input.MaxMessages {
                    stoppedReason = "max_messages_reached"
                    return &CallToolResult{}, ExecuteQueryOutput{
                        Data: map[string]interface{}{
                            "messages": messages,
                            "count": messageCount,
                            "stoppedReason": stoppedReason,
                        },
                    }, nil
                }
            }
        }
    }
}
```

**Stop Conditions (in priority order):**
1. **User Cancel** - Cursor Cancel button (highest priority, immediate stop)
2. **Max Messages** - User-specified message limit
3. **Timeout** - User-specified or default 30s timeout

**Benefits:**
- User always in control (Cancel button, timeout, maxMessages)
- No arbitrary caps - user can set long timeouts if needed
- Partial results always returned (even on cancel/timeout)
- Stopped reason included for debugging

### Protocol References
- graphql-transport-ws: https://github.com/enisdenjo/graphql-ws/blob/master/PROTOCOL.md
- Legacy graphql-ws: https://github.com/apollographql/subscriptions-transport-ws/blob/master/PROTOCOL.md

### Composability Examples
```bash
# Filter subscription messages
gqlt run --query 'subscription { events }' | jq 'select(.data.events.type == "ERROR")'

# Count events over 1 minute
timeout 60 gqlt run --query 'subscription { events }' | wc -l

# Save first 100 messages
gqlt run --query 'subscription { events }' --max-messages 100 > events.jsonl
```

### Library Testing Value
```go
func TestSubscription(t *testing.T) {
    mockServer := testutil.NewMockGraphQLServer()
    defer mockServer.Close()
    
    mockServer.SetupSubscription("Events", func() chan *Response {
        // Return channel that streams test messages
    })
    
    // Test your subscription handler
}
```

## RISKS

- **Complexity:** WebSocket handling adds significant complexity
- **State Management:** Long-lived connections require careful resource management
- **Testing:** Subscription testing is inherently harder than request/response
- **MCP Compatibility:** May not fit MCP's synchronous model

## ALTERNATIVES CONSIDERED

1. **No subscriptions** - Keep gqlt focused on queries/mutations only
2. **Polling mode** - Simulate subscriptions with repeated queries
3. **External tool** - Point users to dedicated subscription tools

Status: Planning - Awaiting approval to proceed

