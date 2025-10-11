# Plan: Add SDL Schema Endpoint Support

**Status:** In Progress  
**Created:** 2025-10-11  
**Goal:** Support fetching GraphQL schemas from SDL endpoints when introspection is disabled

## Problem

Some GraphQL servers disable introspection queries but provide the complete schema in SDL (Schema Definition Language) format at a GET endpoint, typically at `/graphql/schema.graphql` or similar.

Currently, gqlt only supports introspection queries, which fails on these servers.

## Solution

Add fallback mechanism to fetch schema from SDL endpoints when introspection fails.

## Implementation Steps

### Phase 1: SDL Fetching
- [ ] Add method to fetch SDL from GET endpoint (try multiple common paths)
- [ ] Common paths to try:
  - `/schema.graphql` (relative to graphql endpoint)
  - `/graphql/schema.graphql`
  - `/sdl`
  - Configurable custom path
- [ ] Add HTTP client logic for GET requests

### Phase 2: SDL Parsing
- [ ] Research Go GraphQL SDL parsing libraries (graphql-go, gqlgen, etc.)
- [ ] Add dependency for SDL parsing
- [ ] Convert SDL to introspection-compatible format
- [ ] Handle schema types, queries, mutations, subscriptions, etc.

### Phase 3: Integration
- [ ] Update `Client.Introspect()` to try SDL fallback when introspection fails
- [ ] Update MCP handlers to use new logic
- [ ] Update CLI introspect command to support SDL endpoints
- [ ] Add configuration option for SDL endpoint path

### Phase 4: Testing
- [ ] Add tests for SDL fetching
- [ ] Add tests for SDL parsing
- [ ] Add tests for fallback mechanism
- [ ] Test with real servers that use SDL endpoints

### Phase 5: Documentation
- [ ] Update README with SDL support information
- [ ] Add examples for SDL endpoints
- [ ] Document configuration options

## Open Questions
1. Which SDL parsing library to use?
2. Should we try SDL first or introspection first?
3. How to handle custom SDL endpoint paths?
4. Should this be automatic or opt-in?

## Notes
- SDL format is text-based schema definition
- Need to convert SDL → introspection JSON format for compatibility
- Some servers may provide both introspection and SDL endpoints

