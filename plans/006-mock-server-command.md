================================================================
IMPLEMENTATION PLAN — Mock GraphQL Server Subcommand
Language: Go
Goal: Add a simple mock GraphQL server for testing and development
================================================================

**Date:** 2025-10-12

**Execution Mode:** supervised

Progress legend: [x] Completed, [ ] Pending

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

### Todo List Application Schema

**Concept:** A practical todo list application that demonstrates all GraphQL features.
Each todo can have attachments (files or links via interface), assigned users, priority,
and status. This makes the schema cohesive while covering all GraphQL capabilities.

**Key Features:**
- **Users** with roles (ADMIN, USER, GUEST)
- **Todos** with status (PENDING, IN_PROGRESS, COMPLETED), priority, assignees, due dates, and tags
- **Attachments** (interface) with two implementations:
  - **FileAttachment** - uploaded files (demonstrates Upload scalar)
  - **LinkAttachment** - URL links (demonstrates interface)
- **Search** across users and todos (demonstrates unions)
- **Subscriptions** for real-time todo and user events

**Complete Schema:** See `plans/006-todo-schema.graphqls` for the full documented schema.

### Schema Highlights:

The complete schema is in `plans/006-todo-schema.graphqls`. Key type relationships:

```
User (implements Node)
  ├── role: UserRole (enum)
  └── todos: [Todo!]! (with filtering)

Todo (implements Node)
  ├── status: TodoStatus (enum)
  ├── priority: TodoPriority (enum)
  ├── createdBy: User
  ├── assignedTo: User
  ├── attachments: [Attachment!]! (interface)
  │   ├── FileAttachment (implements Attachment) - for uploads
  │   └── LinkAttachment (implements Attachment) - for URLs
  ├── dueDate: DateTime (custom scalar)
  └── tags: [String!]!

SearchResult (union) = User | Todo

Mutations:
  - User CRUD (createUser)
  - Todo CRUD (createTodo, updateTodo, deleteTodo, completeTodo)
  - Attachments (addFileAttachment, addLinkAttachment, removeAttachment)

Subscriptions:
  - counter (testing)
  - todoEvents (real-time todo changes)
  - userEvents (real-time user changes)
  - tick(interval) (configurable timer)
```

**GraphQL Features Demonstrated:**
- Custom Scalars: DateTime, URL, Upload
- Enums: TodoStatus, TodoPriority, UserRole
- Interfaces: Attachment (FileAttachment, LinkAttachment), Node
- Unions: SearchResult
- Input Types: CreateTodoInput, UpdateTodoInput, CreateUserInput, TodoFilters
- Directives: @deprecated
- Default Values: On all pagination and interval arguments
- Field Arguments: todos(filters, limit, offset), User.todos(status, limit, offset)

**Schema Location:** `plans/006-todo-schema.graphqls` (626 lines, fully documented)

The schema has been extracted to a separate file for:
- Easier reference during implementation
- Direct copying to `internal/mockserver/graph/schema.graphqls`
- Reduced plan file size
- Better git diffs when schema changes

```

## IMPLEMENTATION STEPS

### [x] 0) Planning document and alignment

This plan outlines the implementation of a comprehensive mock GraphQL server using `gqlgen`.
The server will demonstrate all GraphQL features and support both WebSocket and SSE transports
for subscriptions, making it perfect for testing gqlt itself and other GraphQL clients.

**Architecture Decision: Use gqlgen**
- Chosen: gqlgen for code generation and type safety
- Rejected alternatives:
  - Building from scratch: Too time-consuming (days vs hours)
  - graphql-go/graphql: Runtime-based, less type-safe, more manual work
  - Thunder: Less mature, smaller community
- Rationale: gqlgen provides automatic introspection, WebSocket subscriptions, and generates
  type-safe Go code from schema, reducing implementation time by ~80%

Status: Complete - Plan finalized with comprehensive schema, proper step structure, and gqlgen gotchas documented.

### [x] 1) Add gqlgen dependency and initialize structure

Add the gqlgen library and create the directory structure for the mock server.

- Add `github.com/99designs/gqlgen` to go.mod
- Create `internal/mockserver/` directory
- Create subdirectories for organization (if needed)

How to test:
- Run `go mod tidy` and verify no errors
- Verify `internal/mockserver/` directory exists

Status: Complete - Added gqlgen v0.17.81 dependency and created internal/mockserver/ directory.

### [x] 2) Create GraphQL schema file

Create the complete GraphQL schema with comprehensive documentation.

- Create `internal/mockserver/graph/schema.graphqls`
- Copy the fully documented schema from `plans/006-todo-schema.graphqls`
- Ensure all types, fields, and arguments are documented

How to test:
- File exists at `internal/mockserver/graph/schema.graphqls`
- File contains ~626 lines of schema with documentation
- Schema includes all GraphQL features: scalars, enums, interfaces, unions, input types, directives

Status: Complete - Copied comprehensive todo-list schema from plans/006-todo-schema.graphqls.

### [x] 3) Configure gqlgen code generation

Create the gqlgen configuration file and run code generation.

CRITICAL: gqlgen is very particular about directory structure and configuration.
Incorrect paths will cause silent failures or generate code in wrong locations.

- Create `internal/mockserver/gqlgen.yml` with exact configuration:
  ```yaml
  # Schema location (must match actual file path)
  schema:
    - schema.graphqls
  
  # Where to generate the GraphQL server code
  exec:
    filename: generated.go
    package: mockserver
  
  # Where to generate the resolver stubs
  resolver:
    filename: resolver.go
    package: mockserver
    type: Resolver
  
  # Where to generate the models
  model:
    filename: models_gen.go
    package: mockserver
  
  # Autobind tells gqlgen to use Go types that match GraphQL type names
  autobind: []
  
  # Custom scalar mappings
  models:
    DateTime:
      model: time.Time
    URL:
      model: string
    Upload:
      model: github.com/99designs/gqlgen/graphql.Upload
  ```

- Run `cd internal/mockserver && gqlgen generate`
- Verify generated files are in `internal/mockserver/` (NOT in root or wrong directory)
- Review `generated.go` - contains GraphQL execution engine
- Review `models_gen.go` - contains Go structs for GraphQL types
- Review `resolver.go` - contains resolver stub methods

How to test:
- CRITICAL: All files must be in `internal/mockserver/graph/`, not scattered elsewhere
- `ls internal/mockserver/graph/` shows: `generated.go`, `models_gen.go`, `resolver.go`, `schema.resolvers.go`, `schema.graphqls`
- Code compiles: `cd internal/mockserver && go build`
- No errors about missing types or packages
- Resolver methods exist with correct signatures (check `schema.resolvers.go`)

Status: Complete - Generated all code successfully. 22 resolver stubs created for comprehensive schema.
Added custom scalar mappings for DateTime, URL, and Upload to gqlgen.yml.

### [x] 4) Implement in-memory data store

Create thread-safe in-memory storage for users and other entities.

- Create `internal/mockserver/graph/store.go`
- Define `Store` struct with `sync.RWMutex`
- Implement maps for: users, todos, fileAttachments, linkAttachments
- Implement methods: GetUser, GetUsers, CreateUser, GetTodo, GetTodos, CreateTodo, UpdateTodo, DeleteTodo
- Implement attachment methods: GetFileAttachment, GetLinkAttachment, CreateFileAttachment, CreateLinkAttachment
- Pre-seed with 3 sample users (Admin, User, Guest roles)
- Use global ID format "TypeName:localId" for Relay Node pattern
- Update Resolver struct to use Store instead of direct todos slice

How to test:
- Unit test: Create store, verify pre-seeded users exist
- Unit test: Create user, retrieve it, verify fields match
- Unit test: Concurrent access (multiple goroutines)
- Unit test: Create, update, delete todo
- All tests pass

Status: Complete - Created comprehensive thread-safe store with all CRUD operations.
6 tests pass including concurrent access test. Global IDs properly formatted.

### [x] 5) Implement query resolvers

Implement all query field resolvers.

- Implement `Query.node` -> parse global ID, route to correct store (Relay pattern)
- Implement `Query.hello` -> "Hello, GraphQL!"
- Implement `Query.echo` -> return input message
- Implement `Query.user` -> fetch from store by ID
- Implement `Query.users` -> fetch all with pagination (limit, offset)
- Implement `Query.todo` -> fetch from store by ID
- Implement `Query.todos` -> fetch all with filtering and pagination
- Implement `Query.search` -> search users and todos, return union results
- Implement `Query.currentTime` -> return `time.Now()`
- Implement `Query.version` -> return "1.0.0" (deprecated field)

How to test:
- Test each resolver independently
- Verify pagination works for `users(limit, offset)`
- Verify filtering works for `todos(filters: {...})`
- Verify union types are correctly returned for `search`
- Verify Relay node() query works with global IDs

Status: Complete - All 10 query resolvers implemented with pagination, filtering, and Relay Node support.

### [ ] 6) Implement mutation resolvers

Implement all mutation field resolvers.

- Implement `Mutation.createUser` -> create user with input type
- Implement `Mutation.createTodo` -> create todo with input type
- Implement `Mutation.updateTodo` -> update todo via input type (with ID in input)
- Implement `Mutation.deleteTodo` -> delete todo by ID
- Implement `Mutation.completeTodo` -> mark todo as completed
- Implement `Mutation.addFileAttachment` -> upload file and attach to todo
- Implement `Mutation.addLinkAttachment` -> add link attachment to todo
- Implement `Mutation.removeAttachment` -> remove attachment from todo

How to test:
- Test creating user, verify it appears in store with global ID
- Test creating todo, verify createdBy and timestamps set
- Test updating todo, verify changes persist and updatedAt changes
- Test deleting todo, verify it's removed from store
- Test completing todo, verify status changes to COMPLETED
- Test file upload attachment
- Test link attachment
- Test removing attachment

### [ ] 7) Implement field resolvers

Implement nested field resolvers for complex types.

- Implement `User.posts` -> return mock posts for user (with pagination)
- Implement union type resolvers for `SearchResult`
- Ensure all custom scalars (DateTime, URL) are handled

How to test:
- Query user with posts, verify pagination works
- Query search, verify union type resolution
- Verify DateTime returns ISO 8601 format
- Verify URL validates format

### [ ] 8) Implement WebSocket subscription resolvers

Implement real-time subscriptions using Go channels.

- Implement `Subscription.counter` -> emit 1, 2, 3... every second
- Implement `Subscription.userEvents` -> emit on user changes
- Implement `Subscription.tick` -> emit timestamp every N seconds
- Handle context cancellation for all subscriptions
- Use Go channels to stream data

How to test:
- Subscribe to counter, verify sequential numbers
- Subscribe to tick with custom interval, verify timing
- Cancel subscription mid-stream, verify cleanup

### [ ] 9) Create serve command

Create the Cobra CLI command for starting the server.

NOTE: The server's listening address MUST be configurable via flags since common ports
(8080, 4000, etc.) are often already in use during development.

- Create `cmd/serve.go`
- Add flags: `--port` (default: 4000), `--host` (default: localhost)
- Add flags: `--quiet`, `--cors`
- Register command in `cmd/root.go`
- Add Examples section to command
- Pass host and port to the server initialization

How to test:
- Run `gqlt serve --help`, verify all flags are listed
- Verify command is registered: `gqlt --help` shows serve
- Test with custom port: `gqlt serve --port 8090`

### [ ] 10) Implement HTTP server with gqlgen handler

Create the HTTP server using gqlgen's generated handler.

- Create server initialization function
- Mount gqlgen handler at POST `/graphql`
- Mount GET `/graphql` for GraphQL Playground (optional)
- Add graceful shutdown on SIGINT/SIGTERM
- Add startup message with server URL
- Implement CORS middleware (if `--cors` flag set)

How to test:
- Start server, verify it listens on configured port
- Send POST request to `/graphql`, verify response
- Send SIGINT, verify graceful shutdown
- Test CORS headers if flag is enabled

### [ ] 11) Implement SSE transport for subscriptions

Add Server-Sent Events support as alternative to WebSocket.

- Create `internal/mockserver/sse.go`
- Implement SSE handler for POST `/graphql` with `Accept: text/event-stream`
- Wrap gqlgen subscriptions to work with SSE
- Send `event: next` with GraphQL data
- Send `event: complete` on subscription end
- Support `graphql-preflight: 1` header
- Handle client disconnection

How to test:
- Subscribe via SSE using `curl` with `Accept: text/event-stream`
- Verify `event: next` messages arrive
- Verify `event: complete` is sent on completion
- Test client disconnect, verify cleanup

### [ ] 12) Add comprehensive integration tests

Test the complete server with real GraphQL queries.

- Test introspection queries (`__schema`, `__type`)
- Test SDL endpoint at GET `/graphql/schema.graphql`
- Test all queries return expected data
- Test mutations modify state correctly
- Test WebSocket subscriptions (`graphql-transport-ws`)
- Test SSE subscriptions (`graphql-sse`)
- Test subscription cancellation and cleanup
- Integration test: Use gqlt CLI to connect to mock server
- Integration test: Use gqlt introspect/describe commands

How to test:
- Run integration test suite
- All tests pass
- No memory leaks or goroutine leaks

### [ ] 13) Update documentation

Add mock server documentation to README and examples.

- Add `gqlt serve` section to README
- Document all queries, mutations, and subscriptions
- Add examples using gqlt client against mock server
- Document use cases (testing, learning, demos)
- Add section about self-testing (gqlt testing gqlt)

How to test:
- README includes serve command documentation
- Examples are runnable and produce expected output

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

## SUCCESS CRITERIA

The mock server implementation is complete when:

1. `gqlt serve` command starts a working GraphQL server on http://localhost:4000/graphql
2. All GraphQL features are demonstrated: scalars, objects, enums, unions, interfaces, input types, directives
3. Full introspection support - `gqlt introspect` works against the mock server
4. SDL endpoint at GET /graphql/schema.graphql serves the complete schema
5. All queries, mutations, and subscriptions work as documented
6. Both WebSocket (graphql-transport-ws) and SSE (graphql-sse) transports work for subscriptions
7. File upload support works via Upload scalar
8. In-memory state persists across operations within a session
9. Comprehensive integration tests pass using gqlt client
10. Documentation includes complete examples of using gqlt to test gqlt

## NOTES

This is intentionally simple - just enough to test all GraphQL features. Not meant to be a full-featured mock server like graphql-faker or similar tools. The goal is a batteries-included testing server that just works out of the box.

The implementation uses gqlgen for code generation, which provides automatic introspection, type safety, and WebSocket subscriptions. We only need to implement resolvers and add SSE transport support.

**IMPORTANT gqlgen Gotchas:**
1. Directory structure MUST match gqlgen.yml configuration exactly
2. gqlgen may silently fail or generate in wrong locations if paths are incorrect
3. Always run `gqlgen generate` from the directory containing gqlgen.yml
4. Schema file must be relative to gqlgen.yml location
5. Generated files will overwrite existing files - never edit generated.go or models_gen.go
6. Only edit resolver.go and add new files (store.go, etc.)
7. Re-run `gqlgen generate` after schema changes to update generated code
8. Package names in gqlgen.yml must match actual Go package declarations

Status: Planning - Ready for implementation

