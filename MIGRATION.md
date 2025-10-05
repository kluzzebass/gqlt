# Migration Guide

This guide helps you migrate from other GraphQL libraries to gqlt.

## From graphql-request

### Before (graphql-request)

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/jensneuse/graphql-go-tools/pkg/astparser"
    "github.com/machinebox/graphql"
)

func main() {
    // Create client
    client := graphql.NewClient("https://api.example.com/graphql")
    
    // Set headers
    client.Header.Set("Authorization", "Bearer token")
    
    // Define query
    query := `
        query GetUser($id: ID!) {
            user(id: $id) {
                id
                name
                email
            }
        }
    `
    
    // Set variables
    variables := map[string]interface{}{
        "id": "123",
    }
    
    // Execute request
    var response struct {
        User struct {
            ID    string `json:"id"`
            Name  string `json:"name"`
            Email string `json:"email"`
        } `json:"user"`
    }
    
    err := client.Query(context.Background(), &response, variables, graphql.OperationName("GetUser"))
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("User: %+v\n", response.User)
}
```

### After (gqlt)

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/kluzzebass/gqlt"
)

func main() {
    // Create client
    client := gqlt.NewClient("https://api.example.com/graphql", nil)
    
    // Set headers
    client.SetHeaders(map[string]string{
        "Authorization": "Bearer token",
    })
    
    // Define query
    query := `
        query GetUser($id: ID!) {
            user(id: $id) {
                id
                name
                email
            }
        }
    `
    
    // Set variables
    variables := map[string]interface{}{
        "id": "123",
    }
    
    // Execute request
    response, err := client.Execute(query, variables, "GetUser")
    if err != nil {
        log.Fatal(err)
    }
    
    // Check for GraphQL errors
    if len(response.Errors) > 0 {
        log.Printf("GraphQL errors: %v", response.Errors)
    }
    
    // Access response data
    data, ok := response.Data.(map[string]interface{})
    if !ok {
        log.Fatal("Expected data to be a map")
    }
    
    user, exists := data["user"]
    if !exists {
        log.Fatal("Expected 'user' field")
    }
    
    fmt.Printf("User: %+v\n", user)
}
```

### Key Differences

1. **No struct binding**: gqlt returns `interface{}` instead of binding to structs
2. **Explicit error handling**: Check `response.Errors` for GraphQL errors
3. **Type assertions**: Use type assertions to access response data
4. **Simpler API**: Fewer dependencies and simpler method signatures

## From graphql-go

### Before (graphql-go)

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/graphql-go/graphql"
    "github.com/graphql-go/handler"
    "net/http"
)

func main() {
    // Define schema
    userType := graphql.NewObject(graphql.ObjectConfig{
        Name: "User",
        Fields: graphql.Fields{
            "id": &graphql.Field{
                Type: graphql.String,
            },
            "name": &graphql.Field{
                Type: graphql.String,
            },
        },
    })
    
    queryType := graphql.NewObject(graphql.ObjectConfig{
        Name: "Query",
        Fields: graphql.Fields{
            "user": &graphql.Field{
                Type: userType,
                Args: graphql.FieldConfigArgument{
                    "id": &graphql.ArgumentConfig{
                        Type: graphql.String,
                    },
                },
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    // Resolver logic
                    return map[string]interface{}{
                        "id":   "123",
                        "name": "John Doe",
                    }, nil
                },
            },
        },
    })
    
    schema, err := graphql.NewSchema(graphql.SchemaConfig{
        Query: queryType,
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Create handler
    h := handler.New(&handler.Config{
        Schema: &schema,
    })
    
    http.Handle("/graphql", h)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### After (gqlt)

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/kluzzebass/gqlt"
)

func main() {
    // gqlt is a client library, not a server library
    // For GraphQL servers, consider using graphql-go or other server libraries
    
    // Use gqlt to query existing GraphQL APIs
    client := gqlt.NewClient("https://api.example.com/graphql", nil)
    
    response, err := client.Execute(
        `query GetUser($id: ID!) { user(id: $id) { id name } }`,
        map[string]interface{}{"id": "123"},
        "GetUser",
    )
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Response: %+v\n", response.Data)
}
```

### Key Differences

1. **Client vs Server**: gqlt is a client library, not a server library
2. **Different purpose**: Use gqlt to query GraphQL APIs, not to build GraphQL servers
3. **Simpler usage**: No schema definition required for client usage

## From graphql-js (Node.js)

### Before (graphql-js)

```javascript
const { request } = require('graphql-request');

async function main() {
    const endpoint = 'https://api.example.com/graphql';
    
    const query = `
        query GetUser($id: ID!) {
            user(id: $id) {
                id
                name
                email
            }
        }
    `;
    
    const variables = { id: '123' };
    
    try {
        const data = await request(endpoint, query, variables);
        console.log('User:', data.user);
    } catch (error) {
        console.error('Error:', error);
    }
}
```

### After (gqlt)

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/kluzzebass/gqlt"
)

func main() {
    client := gqlt.NewClient("https://api.example.com/graphql", nil)
    
    query := `
        query GetUser($id: ID!) {
            user(id: $id) {
                id
                name
                email
            }
        }
    `
    
    variables := map[string]interface{}{
        "id": "123",
    }
    
    response, err := client.Execute(query, variables, "GetUser")
    if err != nil {
        log.Fatal(err)
    }
    
    if len(response.Errors) > 0 {
        log.Printf("GraphQL errors: %v", response.Errors)
    }
    
    data, ok := response.Data.(map[string]interface{})
    if !ok {
        log.Fatal("Expected data to be a map")
    }
    
    user, exists := data["user"]
    if !exists {
        log.Fatal("Expected 'user' field")
    }
    
    fmt.Printf("User: %+v\n", user)
}
```

### Key Differences

1. **Language**: JavaScript vs Go
2. **Error handling**: Explicit error checking in Go
3. **Type safety**: Go's type system requires type assertions
4. **Async**: JavaScript uses async/await, Go uses synchronous calls

## From Apollo Client

### Before (Apollo Client)

```javascript
import { ApolloClient, InMemoryCache, gql } from '@apollo/client';

const client = new ApolloClient({
    uri: 'https://api.example.com/graphql',
    cache: new InMemoryCache(),
    headers: {
        'Authorization': 'Bearer token'
    }
});

const GET_USER = gql`
    query GetUser($id: ID!) {
        user(id: $id) {
            id
            name
            email
        }
    }
`;

async function getUser(id) {
    try {
        const { data, error } = await client.query({
            query: GET_USER,
            variables: { id }
        });
        
        if (error) {
            console.error('GraphQL error:', error);
            return;
        }
        
        console.log('User:', data.user);
    } catch (error) {
        console.error('Network error:', error);
    }
}
```

### After (gqlt)

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/kluzzebass/gqlt"
)

func main() {
    client := gqlt.NewClient("https://api.example.com/graphql", nil)
    client.SetHeaders(map[string]string{
        "Authorization": "Bearer token",
    })
    
    query := `
        query GetUser($id: ID!) {
            user(id: $id) {
                id
                name
                email
            }
        }
    `
    
    response, err := client.Execute(query, map[string]interface{}{"id": "123"}, "GetUser")
    if err != nil {
        log.Fatal(err)
    }
    
    if len(response.Errors) > 0 {
        log.Printf("GraphQL errors: %v", response.Errors)
    }
    
    data, ok := response.Data.(map[string]interface{})
    if !ok {
        log.Fatal("Expected data to be a map")
    }
    
    user, exists := data["user"]
    if !exists {
        log.Fatal("Expected 'user' field")
    }
    
    fmt.Printf("User: %+v\n", user)
}
```

### Key Differences

1. **No caching**: gqlt doesn't include caching (use external caching if needed)
2. **No reactive updates**: gqlt is stateless, no automatic re-fetching
3. **Simpler API**: Fewer features but simpler to use
4. **Explicit error handling**: Check for errors manually

## Migration Checklist

### ✅ What to Migrate

- [ ] **Query strings**: Copy GraphQL queries directly
- [ ] **Variables**: Convert to `map[string]interface{}`
- [ ] **Headers**: Use `SetHeaders()` method
- [ ] **Authentication**: Use `SetAuth()` or `SetHeaders()`
- [ ] **Error handling**: Check `response.Errors` array
- [ ] **Response data**: Use type assertions to access data

### ❌ What's Different

- [ ] **No automatic struct binding**: Use type assertions instead
- [ ] **No caching**: Implement external caching if needed
- [ ] **No reactive updates**: Manual re-fetching required
- [ ] **No subscriptions**: Use WebSocket libraries for real-time updates
- [ ] **No optimistic updates**: Manual state management required

### 🔄 What to Consider

- [ ] **Type safety**: Use Go's type system for better safety
- [ ] **Error handling**: Implement comprehensive error handling
- [ ] **Testing**: Use gqlt's testing utilities for better test coverage
- [ ] **Configuration**: Use gqlt's configuration system for multiple environments

## Best Practices

### 1. Use Type Assertions Safely

```go
// Good: Safe type assertion
if data, ok := response.Data.(map[string]interface{}); ok {
    if user, exists := data["user"]; exists {
        if userMap, ok := user.(map[string]interface{}); ok {
            fmt.Printf("User ID: %v\n", userMap["id"])
        }
    }
}

// Better: Use helper functions
func getStringField(data map[string]interface{}, field string) string {
    if value, exists := data[field]; exists {
        if str, ok := value.(string); ok {
            return str
        }
    }
    return ""
}
```

### 2. Handle Errors Properly

```go
response, err := client.Execute(query, variables, operationName)
if err != nil {
    // Network or HTTP error
    log.Printf("Request failed: %v", err)
    return
}

if len(response.Errors) > 0 {
    // GraphQL errors
    for _, gqlErr := range response.Errors {
        log.Printf("GraphQL error: %v", gqlErr)
    }
    return
}

// Success - use response.Data
```

### 3. Use Configuration Management

```go
// Load configuration
config, err := gqlt.Load("")
if err != nil {
    log.Fatal(err)
}

// Get current configuration
current := config.GetCurrent()
client := gqlt.NewClient(current.Endpoint, current.Headers)
```

### 4. Use Testing Utilities

```go
func TestGraphQL(t *testing.T) {
    helper := gqlt.NewGraphQLTestHelper(t, "https://api.example.com/graphql")
    
    response := helper.ExecuteQuery("{ users { id name } }", nil, "")
    helper.AssertNoErrors(response)
    helper.AssertFieldExists(response, "users")
}
```

## Need Help?

- Check the [examples/](examples/) directory for comprehensive examples
- Read the [README.md](README.md) for detailed documentation
- Open an issue on GitHub for questions or bug reports
