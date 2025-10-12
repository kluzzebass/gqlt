================================================================
IMPLEMENTATION PLAN — Mock GraphQL Server Subcommand
Language: Go
Goal: Add a simple mock GraphQL server for testing and development
================================================================

**Date:** 2025-10-12

**Execution Mode:** TBD

## CONTEXT

A simple mock GraphQL server would be valuable for:
- Testing GraphQL clients
- Learning GraphQL features
- Quick local development without backend
- Demonstrating all GraphQL features (queries, mutations, subscriptions, types, etc.)

## OBJECTIVE

Create a `gqlt serve` subcommand that runs a simple mock GraphQL server with:
- Fixed schema with example types (scalars, objects, enums, unions, interfaces)
- Simple query handlers
- Simple mutation handlers
- Subscription that emits data periodically
- Full introspection support
- Optional SDL endpoint
- Optional WebSocket support for subscriptions

## SCHEMA DESIGN

### Simple but comprehensive schema:

```graphql
scalar DateTime
scalar URL

enum Status {
  ACTIVE
  INACTIVE
  PENDING
}

enum UserRole {
  ADMIN
  USER
  GUEST
}

type User {
  id: ID!
  name: String!
  email: String!
  status: Status!
  role: UserRole!
  createdAt: DateTime!
  website: URL
}

type Post {
  id: ID!
  title: String!
  content: String!
  author: User!
  published: Boolean!
}

interface Node {
  id: ID!
}

type Product implements Node {
  id: ID!
  name: String!
  price: Float!
}

type Service implements Node {
  id: ID!
  name: String!
  hourlyRate: Float!
}

union SearchResult = User | Post | Product | Service

type Query {
  # Simple queries
  hello: String!
  echo(message: String!): String!
  
  # Object queries
  user(id: ID!): User
  users: [User!]!
  
  # Search with union
  search(term: String!): [SearchResult!]!
  
  # Testing different types
  currentTime: DateTime!
}

type Mutation {
  # Simple mutation
  createUser(name: String!, email: String!): User!
  
  # Mutation with enum
  updateUserStatus(id: ID!, status: Status!): User!
  
  # Mutation with file upload
  uploadFile(file: Upload!): String!
}

type Subscription {
  # Emit a counter every second
  counter: Int!
  
  # Emit user events (created, updated, deleted)
  userEvents: User!
  
  # Emit timestamp every N seconds
  tick(interval: Int): DateTime!
}
```

## IMPLEMENTATION STEPS

### Phase 1: Command Structure
- [ ] Create `cmd/serve.go` with Cobra command
- [ ] Add flags:
  - [ ] `--port` (default: 4000)
  - [ ] `--host` (default: localhost)
  - [ ] `--quiet` (suppress startup messages)
  - [ ] `--cors` (enable CORS for web clients)
- [ ] Register command in root

### Phase 2: HTTP Server Setup
- [ ] Create actual HTTP server (not httptest)
- [ ] Serve GraphQL at `/graphql`
- [ ] Serve SDL at `/graphql/schema.graphql`
- [ ] Add graceful shutdown on signals (SIGINT, SIGTERM)
- [ ] Add startup message with server URL

### Phase 3: Schema & Introspection
- [ ] Define the fixed schema (as shown above)
- [ ] Generate introspection JSON from schema
- [ ] Handle introspection queries
- [ ] Serve SDL at GET /graphql/schema.graphql

### Phase 4: Query Handlers
- [ ] Implement `hello` - returns "Hello, GraphQL!"
- [ ] Implement `echo` - returns the input message
- [ ] Implement `user(id)` - returns mock user by ID
- [ ] Implement `users` - returns list of mock users
- [ ] Implement `search` - returns union results
- [ ] Implement `currentTime` - returns current timestamp

### Phase 5: Mutation Handlers
- [ ] Implement `createUser` - creates mock user (in-memory)
- [ ] Implement `updateUserStatus` - updates user status
- [ ] Implement `uploadFile` - accepts file upload, returns filename

### Phase 6: WebSocket & Subscription Handlers
- [ ] Add WebSocket upgrade handler at `/graphql`
- [ ] Implement graphql-transport-ws protocol
- [ ] Implement `counter` subscription - emit 1, 2, 3, ... every second
- [ ] Implement `userEvents` subscription - emit mock user events
- [ ] Implement `tick(interval)` subscription - emit timestamp every N seconds
- [ ] Handle subscription cancellation
- [ ] Clean up subscriptions on disconnect

### Phase 7: In-Memory State
- [ ] Simple in-memory store for users
- [ ] Start with 3 pre-seeded users
- [ ] Mutations modify the in-memory state
- [ ] Subscriptions can emit when state changes

### Phase 8: Testing
- [ ] Test starting server and making HTTP requests
- [ ] Test queries return expected data
- [ ] Test mutations modify state
- [ ] Test subscriptions emit periodic messages
- [ ] Test graceful shutdown
- [ ] Test CORS headers if enabled

### Phase 9: Documentation
- [ ] Add serve command to README
- [ ] Document all available queries/mutations/subscriptions
- [ ] Add examples for testing with gqlt itself
- [ ] Document use cases (testing, learning, demos)

## USAGE EXAMPLES

### Starting the Server
```bash
# Start with defaults
gqlt serve

# Custom port
gqlt serve --port 5000

# With CORS for browser clients
gqlt serve --cors

# Quiet mode
gqlt serve --quiet
```

### Testing the Server
```bash
# In another terminal - test queries
gqlt run --url http://localhost:4000/graphql --query '{ hello }'
gqlt run --url http://localhost:4000/graphql --query '{ users { id name } }'

# Test mutations
gqlt run --url http://localhost:4000/graphql \
  --query 'mutation { createUser(name: "Alice", email: "alice@example.com") { id name } }'

# Test subscriptions
gqlt run --url http://localhost:4000/graphql \
  --query 'subscription { counter }'

# With jq filtering
gqlt run --url http://localhost:4000/graphql \
  --query 'subscription { tick(interval: 2) }' | jq -r '.data.tick'
```

### Use Cases
- **Testing gqlt itself:** Perfect for integration tests
- **Testing other GraphQL clients:** Simple endpoint for client development
- **Learning GraphQL:** All features demonstrated
- **Demos:** Quick GraphQL server for presentations

## ARCHITECTURE

### Server Structure
```go
type MockServer struct {
    port       int
    host       string
    httpServer *http.Server
    users      []User      // In-memory data
    userID     int         // Counter for new users
    mu         sync.RWMutex
}

func (s *MockServer) Start() error {
    // Set up HTTP server
    // Register /graphql handler
    // Register /graphql/schema.graphql handler
    // Start listening
}

func (s *MockServer) handleGraphQL(w http.ResponseWriter, r *http.Request) {
    if r.Header.Get("Upgrade") == "websocket" {
        s.handleWebSocket(w, r)
        return
    }
    s.handleHTTP(w, r)
}
```

### Subscription Implementation
```go
func (s *MockServer) handleCounterSubscription(ctx context.Context) <-chan int {
    ch := make(chan int)
    go func() {
        defer close(ch)
        counter := 0
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                counter++
                ch <- counter
            }
        }
    }()
    return ch
}
```

## MOCK DATA

### Pre-seeded Users
```go
{ID: "1", Name: "Alice", Email: "alice@example.com", Status: ACTIVE, Role: ADMIN},
{ID: "2", Name: "Bob", Email: "bob@example.com", Status: ACTIVE, Role: USER},
{ID: "3", Name: "Charlie", Email: "charlie@example.com", Status: INACTIVE, Role: USER},
```

### Generated Data
- User IDs auto-increment from 4
- Timestamps use actual current time
- Posts reference existing users as authors

## SUCCESS CRITERIA

- [ ] `gqlt serve` starts a server on specified port
- [ ] All queries return sensible mock data
- [ ] Mutations modify in-memory state
- [ ] Subscriptions emit data periodically
- [ ] Introspection works
- [ ] SDL endpoint serves schema
- [ ] WebSocket upgrades work for subscriptions
- [ ] Ctrl+C shuts down gracefully
- [ ] Can test the server using gqlt itself
- [ ] Well-documented with examples

## DEFERRED FEATURES

These can be added later if needed:
- Custom schema via config file
- Custom response data via config file
- Multiple schemas/endpoints
- Request delay simulation
- Error simulation
- Rate limiting

## NOTES

This is intentionally simple - just enough to test all GraphQL features. Not meant to be a full-featured mock server like graphql-faker or similar tools. The goal is a batteries-included testing server that just works out of the box.

Status: Planning - Ready for implementation

