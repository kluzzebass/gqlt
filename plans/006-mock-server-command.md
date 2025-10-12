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

**Implementation Approach:** Use `gqlgen` (https://gqlgen.com/) - a Go library that generates
type-safe GraphQL server code from schema definitions. This eliminates the need to build a
GraphQL execution engine from scratch and provides automatic introspection, WebSocket subscriptions,
and SDL support out of the box.

**Complexity Estimate:**
- **~500-800 lines of code** (resolvers, store, command, SSE wrapper)
- **~300-400 lines** (tests)
- **Time: 4-8 hours** of focused work (instead of multiple days)

**What gqlgen Provides for Free:**
- ✅ Schema parsing and validation
- ✅ Query/mutation execution engine
- ✅ Full introspection support
- ✅ WebSocket subscriptions (`graphql-transport-ws`)
- ✅ Type-safe Go code generation
- ✅ Automatic resolver scaffolding

## OBJECTIVE

Create a `gqlt serve` subcommand that runs a simple mock GraphQL server with:
- Fixed schema with all GraphQL features:
  - Types: scalars, objects, enums, unions, interfaces, input types
  - Directives: `@deprecated` with custom reasons
  - Default values on field arguments
- Simple query handlers (with optional input type filters and pagination)
- Simple mutation handlers (with both simple args and input types)
- Subscription that emits data periodically
- Full introspection support (standard `__schema` and `__type` queries)
- SDL endpoint at GET /graphql/schema.graphql
- Multi-transport subscription support (WebSocket and SSE)

## SCHEMA DESIGN

### Simple but comprehensive schema:

```graphql
scalar DateTime
scalar URL

directive @deprecated(
  reason: String = "No longer supported"
) on FIELD_DEFINITION | ENUM_VALUE

enum Status {
  ACTIVE
  INACTIVE
  PENDING
  DELETED @deprecated(reason: "Use INACTIVE instead")
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
  bio: String @deprecated(reason: "Use profile.bio instead")
  posts(limit: Int = 10, offset: Int = 0): [Post!]!
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

input CreateUserInput {
  name: String!
  email: String!
  role: UserRole
  website: URL
}

input UpdateUserInput {
  name: String
  email: String
  status: Status
  role: UserRole
  website: URL
}

input SearchFilters {
  status: Status
  role: UserRole
}

type Query {
  # Simple queries
  hello: String!
  echo(message: String!): String!
  
  # Object queries
  user(id: ID!): User
  users(filters: SearchFilters, limit: Int = 100, offset: Int = 0): [User!]!
  
  # Search with union
  search(term: String!, limit: Int = 10): [SearchResult!]!
  
  # Testing different types
  currentTime: DateTime!
  
  # Deprecated field for testing
  version: String! @deprecated(reason: "Use serverInfo.version instead")
}

type Mutation {
  # Simple mutation (backward compatible)
  createUser(name: String!, email: String!): User!
  
  # Mutation with input type
  createUserWithInput(input: CreateUserInput!): User!
  
  # Mutation with input type for updates
  updateUser(id: ID!, input: UpdateUserInput!): User!
  
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

### Phase 1: gqlgen Setup
- [ ] Add `github.com/99designs/gqlgen` dependency to go.mod
- [ ] Create `internal/mockserver/` directory structure
- [ ] Create `internal/mockserver/schema.graphqls` with the complete schema
- [ ] Create `internal/mockserver/gqlgen.yml` configuration
- [ ] Run `gqlgen generate` to create resolvers and models
- [ ] Review generated code (`generated.go`, `models_gen.go`, `resolver.go`)

### Phase 2: Command Structure
- [ ] Create `cmd/serve.go` with Cobra command
- [ ] Add flags:
  - [ ] `--port` (default: 4000)
  - [ ] `--host` (default: localhost)
  - [ ] `--quiet` (suppress startup messages)
  - [ ] `--cors` (enable CORS for web clients)
- [ ] Register command in root

### Phase 3: HTTP Server Setup
- [ ] Create HTTP server using gqlgen's handler
- [ ] Serve GraphQL at `/graphql` (queries, mutations, subscriptions)
- [ ] Enable GraphQL Playground at `/` (optional, for debugging)
- [ ] Add graceful shutdown on signals (SIGINT, SIGTERM)
- [ ] Add startup message with server URL
- [ ] Configure CORS middleware if `--cors` flag is set

### Phase 4: In-Memory Data Store
- [ ] Create `internal/mockserver/store.go` with simple in-memory storage
- [ ] Define User struct matching generated model
- [ ] Pre-seed with 3 sample users
- [ ] Add methods: GetUser, GetUsers, CreateUser, UpdateUser
- [ ] Use sync.RWMutex for thread-safe access

### Phase 5: Resolver Implementation - Queries
- [ ] Implement `hello` resolver - returns "Hello, GraphQL!"
- [ ] Implement `echo` resolver - returns the input message
- [ ] Implement `user(id)` resolver - fetches from store
- [ ] Implement `users(filters)` resolver - filters by status/role if provided
- [ ] Implement `search` resolver - returns union of different types
- [ ] Implement `currentTime` resolver - returns current timestamp
- [ ] Implement `version` resolver - returns mock version string

### Phase 6: Resolver Implementation - Mutations
- [ ] Implement `createUser` resolver - simple args version
- [ ] Implement `createUserWithInput` resolver - input type version
- [ ] Implement `updateUser` resolver - updates via input type
- [ ] Implement `updateUserStatus` resolver - status update only
- [ ] Implement `uploadFile` resolver - saves file info, returns filename

### Phase 7: Resolver Implementation - Subscriptions
gqlgen provides WebSocket subscriptions via `graphql-transport-ws` out of the box.

- [ ] Implement `counter` subscription - channel that emits 1, 2, 3...
- [ ] Implement `userEvents` subscription - channel that emits on user changes
- [ ] Implement `tick(interval)` subscription - configurable interval timer
- [ ] Use Go channels for all subscriptions
- [ ] Handle context cancellation properly

### Phase 8: SSE Support (Additional Transport)
gqlgen natively supports WebSocket. For SSE support:

- [ ] Add custom SSE handler for POST `/graphql` with `Accept: text/event-stream`
- [ ] Implement `graphql-sse` protocol wrapper around gqlgen subscriptions
- [ ] Reuse existing subscription resolvers
- [ ] Send `event: next`, `event: complete` messages
- [ ] Support `graphql-preflight: 1` header

### Phase 9: Field Resolvers
- [ ] Implement `User.posts` resolver - returns mock posts for user
- [ ] Implement union type resolvers for SearchResult
- [ ] Ensure all fields return appropriate mock data

### Phase 10: Testing
- [ ] Test starting server and making HTTP requests
- [ ] Test introspection queries (`__schema`, `__type`)
- [ ] Test SDL endpoint at GET /graphql/schema.graphql
- [ ] Test queries return expected data
- [ ] Test mutations modify state
- [ ] Test subscriptions via WebSocket (graphql-transport-ws)
- [ ] Test subscriptions via SSE (graphql-sse)
- [ ] Test subscription cancellation and cleanup
- [ ] Test graceful shutdown
- [ ] Test CORS headers if enabled
- [ ] Integration test: use gqlt client to connect to mock server
- [ ] Integration test: use gqlt introspect/describe against mock server

### Phase 11: Documentation
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
# In another terminal - test introspection
gqlt introspect --url http://localhost:4000/graphql
gqlt describe User --url http://localhost:4000/graphql
gqlt list-types --url http://localhost:4000/graphql

# Test queries
gqlt run --url http://localhost:4000/graphql --query '{ hello }'
gqlt run --url http://localhost:4000/graphql --query '{ users { id name role } }'

# Test queries with input type filters
gqlt run --url http://localhost:4000/graphql \
  --query '{ users(filters: {status: ACTIVE, role: ADMIN}) { id name status role } }'

# Test mutations (simple args)
gqlt run --url http://localhost:4000/graphql \
  --query 'mutation { createUser(name: "Alice", email: "alice@example.com") { id name } }'

# Test mutations with input types
gqlt run --url http://localhost:4000/graphql \
  --query 'mutation { createUserWithInput(input: {name: "Bob", email: "bob@example.com", role: ADMIN}) { id name role } }'

gqlt run --url http://localhost:4000/graphql \
  --query 'mutation { updateUser(id: "1", input: {status: INACTIVE, website: "https://example.com"}) { id status website } }'

# Test subscriptions (will use WebSocket by default, with SSE fallback)
gqlt run --url http://localhost:4000/graphql \
  --query 'subscription { counter }' \
  --timeout 10s

# Test with message limit
gqlt run --url http://localhost:4000/graphql \
  --query 'subscription { tick(interval: 2) }' \
  --max-messages 5 | jq -r '.data.tick'

# Test subscription with custom interval
gqlt run --url http://localhost:4000/graphql \
  --query 'subscription { tick(interval: 1) }' \
  --timeout 5s --max-messages 3
```

### Use Cases
- **Testing gqlt itself:** Perfect for integration tests with both WebSocket and SSE transports
- **Testing other GraphQL clients:** Simple endpoint for client development
- **Learning GraphQL:** All features demonstrated including real-time subscriptions
- **Demos:** Quick GraphQL server for presentations
- **Protocol Testing:** Verify client behavior with both WebSocket and SSE protocols

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

