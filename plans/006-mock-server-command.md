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

### Simple but comprehensive schema:

```graphql
# ============================================================================
# CUSTOM SCALARS
# ============================================================================

"""
DateTime is a custom scalar representing a date and time in ISO 8601 format.
Example: "2025-10-12T14:30:00Z"
"""
scalar DateTime

"""
URL is a custom scalar representing a valid HTTP/HTTPS URL.
Example: "https://example.com"
"""
scalar URL

# ============================================================================
# DIRECTIVES
# ============================================================================

"""
Marks a field or enum value as deprecated, providing a reason for deprecation.
This directive is used to inform clients that a field should no longer be used.
"""
directive @deprecated(
  reason: String = "No longer supported"
) on FIELD_DEFINITION | ENUM_VALUE

# ============================================================================
# ENUMS
# ============================================================================

"""
Status represents the current state of a user account.
"""
enum Status {
  "User account is active and can perform operations"
  ACTIVE
  
  "User account is temporarily disabled"
  INACTIVE
  
  "User account is awaiting activation or approval"
  PENDING
  
  "User account has been marked as deleted (deprecated, use INACTIVE instead)"
  DELETED @deprecated(reason: "Use INACTIVE instead")
}

"""
UserRole defines the permission level of a user in the system.
"""
enum UserRole {
  "Administrator with full access"
  ADMIN
  
  "Regular user with standard permissions"
  USER
  
  "Guest user with limited read-only access"
  GUEST
}

# ============================================================================
# TYPES
# ============================================================================

"""
User represents a person who can access the system.
This type demonstrates object types, custom scalars, enums, deprecated fields,
and field-level arguments with default values.
"""
type User {
  "Unique identifier for the user"
  id: ID!
  
  "Full name of the user"
  name: String!
  
  "Email address (must be unique)"
  email: String!
  
  "Current status of the user account"
  status: Status!
  
  "Permission level assigned to the user"
  role: UserRole!
  
  "Timestamp when the user account was created"
  createdAt: DateTime!
  
  "Optional personal or professional website URL"
  website: URL
  
  "Biographical information (deprecated in favor of structured profile)"
  bio: String @deprecated(reason: "Use profile.bio instead")
  
  """
  Posts created by this user.
  Demonstrates field arguments with default values and pagination.
  """
  posts(
    "Maximum number of posts to return (default: 10)"
    limit: Int = 10
    
    "Number of posts to skip for pagination (default: 0)"
    offset: Int = 0
  ): [Post!]!
}

"""
Post represents a blog post or article written by a user.
Demonstrates relationships between types.
"""
type Post {
  "Unique identifier for the post"
  id: ID!
  
  "Title of the post"
  title: String!
  
  "Full content of the post (can be markdown)"
  content: String!
  
  "The user who authored this post"
  author: User!
  
  "Whether the post is publicly visible"
  published: Boolean!
}

# ============================================================================
# INTERFACES
# ============================================================================

"""
Node is a common interface for types that have a unique identifier.
Demonstrates GraphQL interfaces.
"""
interface Node {
  "Unique identifier"
  id: ID!
}

# ============================================================================
# INTERFACE IMPLEMENTATIONS
# ============================================================================

"""
Product represents a physical or digital product for sale.
Implements the Node interface.
"""
type Product implements Node {
  "Unique identifier for the product"
  id: ID!
  
  "Name of the product"
  name: String!
  
  "Price in USD"
  price: Float!
}

"""
Service represents a billable service offering.
Implements the Node interface.
"""
type Service implements Node {
  "Unique identifier for the service"
  id: ID!
  
  "Name of the service"
  name: String!
  
  "Cost per hour in USD"
  hourlyRate: Float!
}

# ============================================================================
# UNIONS
# ============================================================================

"""
SearchResult is a union type that can be one of several different types.
Demonstrates union types for polymorphic results.
"""
union SearchResult = User | Post | Product | Service

# ============================================================================
# INPUT TYPES
# ============================================================================

"""
CreateUserInput contains all fields needed to create a new user.
Demonstrates input types for structured mutation arguments.
"""
input CreateUserInput {
  "Full name of the user (required)"
  name: String!
  
  "Email address (required, must be unique)"
  email: String!
  
  "Optional role assignment (defaults to USER if not provided)"
  role: UserRole
  
  "Optional website URL"
  website: URL
}

"""
UpdateUserInput contains optional fields for updating an existing user.
All fields are optional to support partial updates.
"""
input UpdateUserInput {
  "Updated name (optional)"
  name: String
  
  "Updated email address (optional)"
  email: String
  
  "Updated status (optional)"
  status: Status
  
  "Updated role (optional)"
  role: UserRole
  
  "Updated website URL (optional)"
  website: URL
}

"""
SearchFilters provides optional criteria for filtering user queries.
Demonstrates input types for filtering and search operations.
"""
input SearchFilters {
  "Filter users by status"
  status: Status
  
  "Filter users by role"
  role: UserRole
}

# ============================================================================
# QUERY ROOT TYPE
# ============================================================================

"""
Query defines all read operations available in the API.
"""
type Query {
  # ------------------------------------------------------------------
  # Simple Queries
  # ------------------------------------------------------------------
  
  "Returns a greeting message - useful for health checks"
  hello: String!
  
  """
  Echoes back the provided message.
  Demonstrates simple string arguments and returns.
  """
  echo(message: String!): String!
  
  # ------------------------------------------------------------------
  # Object Queries
  # ------------------------------------------------------------------
  
  """
  Retrieves a single user by ID.
  Returns null if user is not found.
  """
  user(id: ID!): User
  
  """
  Retrieves a list of users with optional filtering and pagination.
  Demonstrates input types, default values, and filtering.
  """
  users(
    "Optional filters to apply"
    filters: SearchFilters
    
    "Maximum number of users to return (default: 100)"
    limit: Int = 100
    
    "Number of users to skip for pagination (default: 0)"
    offset: Int = 0
  ): [User!]!
  
  # ------------------------------------------------------------------
  # Union Type Queries
  # ------------------------------------------------------------------
  
  """
  Searches across multiple types (User, Post, Product, Service).
  Demonstrates union types and polymorphic results.
  """
  search(
    "Search term to match against"
    term: String!
    
    "Maximum number of results to return (default: 10)"
    limit: Int = 10
  ): [SearchResult!]!
  
  # ------------------------------------------------------------------
  # Custom Scalar Queries
  # ------------------------------------------------------------------
  
  """
  Returns the current server time.
  Demonstrates custom scalar types.
  """
  currentTime: DateTime!
  
  # ------------------------------------------------------------------
  # Deprecated Fields
  # ------------------------------------------------------------------
  
  """
  Returns the server version string.
  Deprecated - use serverInfo.version instead.
  """
  version: String! @deprecated(reason: "Use serverInfo.version instead")
}

# ============================================================================
# MUTATION ROOT TYPE
# ============================================================================

"""
Mutation defines all write operations available in the API.
"""
type Mutation {
  # ------------------------------------------------------------------
  # Simple Mutations
  # ------------------------------------------------------------------
  
  """
  Creates a new user with simple arguments.
  Demonstrates basic mutation with direct arguments (backward compatible).
  """
  createUser(
    name: String!
    email: String!
  ): User!
  
  # ------------------------------------------------------------------
  # Input Type Mutations
  # ------------------------------------------------------------------
  
  """
  Creates a new user using an input type.
  Demonstrates structured mutation inputs for complex operations.
  """
  createUserWithInput(
    "User creation data"
    input: CreateUserInput!
  ): User!
  
  """
  Updates an existing user with partial data.
  Demonstrates input types for updates with optional fields.
  """
  updateUser(
    "ID of the user to update"
    id: ID!
    
    "Fields to update (all optional)"
    input: UpdateUserInput!
  ): User!
  
  # ------------------------------------------------------------------
  # Enum Mutations
  # ------------------------------------------------------------------
  
  """
  Updates only the status of a user.
  Demonstrates enum types in mutations.
  """
  updateUserStatus(
    "ID of the user to update"
    id: ID!
    
    "New status value"
    status: Status!
  ): User!
  
  # ------------------------------------------------------------------
  # File Upload Mutations
  # ------------------------------------------------------------------
  
  """
  Uploads a file and returns the filename.
  Demonstrates the Upload scalar type for file handling.
  """
  uploadFile(
    "File to upload"
    file: Upload!
  ): String!
}

# ============================================================================
# SUBSCRIPTION ROOT TYPE
# ============================================================================

"""
Subscription defines all real-time streaming operations.
Subscriptions use WebSocket or SSE for continuous data flow.
"""
type Subscription {
  """
  Emits an incrementing counter every second: 1, 2, 3, ...
  Useful for testing basic subscription functionality.
  """
  counter: Int!
  
  """
  Emits user objects whenever users are created, updated, or deleted.
  Demonstrates real-time notifications for entity changes.
  """
  userEvents: User!
  
  """
  Emits the current timestamp at a configurable interval.
  Demonstrates subscriptions with arguments and custom scalars.
  """
  tick(
    "Interval in seconds between emissions (default: 1)"
    interval: Int
  ): DateTime!
}
```

## IMPLEMENTATION STEPS

### [ ] 0) Planning document and alignment

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

### [ ] 1) Add gqlgen dependency and initialize structure

Add the gqlgen library and create the directory structure for the mock server.

- Add `github.com/99designs/gqlgen` to go.mod
- Create `internal/mockserver/` directory
- Create subdirectories for organization (if needed)

How to test:
- Run `go mod tidy` and verify no errors
- Verify `internal/mockserver/` directory exists

### [ ] 2) Create GraphQL schema file

Create the complete GraphQL schema with comprehensive documentation.

- Create `internal/mockserver/schema.graphqls`
- Copy the fully documented schema from this plan (lines 55-488)
- Ensure all types, fields, and arguments are documented

How to test:
- File exists at `internal/mockserver/schema.graphqls`
- File contains ~435 lines of schema with documentation

### [ ] 3) Configure gqlgen code generation

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
- CRITICAL: All files must be in `internal/mockserver/`, not scattered elsewhere
- `ls internal/mockserver/` shows: `gqlgen.yml`, `schema.graphqls`, `generated.go`, `models_gen.go`, `resolver.go`
- Code compiles: `cd internal/mockserver && go build`
- No errors about missing types or packages
- Resolver methods exist with correct signatures (check `resolver.go`)

### [ ] 4) Implement in-memory data store

Create thread-safe in-memory storage for users and other entities.

- Create `internal/mockserver/store.go`
- Define `Store` struct with `sync.RWMutex`
- Implement methods: `GetUser`, `GetUsers`, `CreateUser`, `UpdateUser`
- Pre-seed with 3 sample users with different roles and statuses

How to test:
- Unit test: Create store, verify pre-seeded users exist
- Unit test: Create user, retrieve it, verify fields match
- Unit test: Concurrent access (multiple goroutines)

### [ ] 5) Implement query resolvers

Implement all query field resolvers.

- Implement `Query.hello` -> "Hello, GraphQL!"
- Implement `Query.echo` -> return input message
- Implement `Query.user` -> fetch from store by ID
- Implement `Query.users` -> fetch all, apply filters if provided
- Implement `Query.search` -> return mock union results
- Implement `Query.currentTime` -> return `time.Now()`
- Implement `Query.version` -> return "1.0.0" (deprecated field)

How to test:
- Test each resolver independently
- Verify filtering works for `users(filters: {...})`
- Verify union types are correctly returned for `search`

### [ ] 6) Implement mutation resolvers

Implement all mutation field resolvers.

- Implement `Mutation.createUser` -> simple args version
- Implement `Mutation.createUserWithInput` -> input type version
- Implement `Mutation.updateUser` -> partial updates via input type
- Implement `Mutation.updateUserStatus` -> status-only update
- Implement `Mutation.uploadFile` -> save file info, return filename

How to test:
- Test creating user, verify it appears in store
- Test updating user, verify changes persist
- Test both createUser variants produce same result
- Test file upload with mock file

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

- Create `cmd/serve.go`
- Add flags: `--port` (default: 4000), `--host` (default: localhost)
- Add flags: `--quiet`, `--cors`
- Register command in `cmd/root.go`
- Add Examples section to command

How to test:
- Run `gqlt serve --help`, verify all flags are listed
- Verify command is registered: `gqlt --help` shows serve

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

