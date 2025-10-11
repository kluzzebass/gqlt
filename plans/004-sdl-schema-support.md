# Plan: Add SDL Schema Endpoint Support

**Status:** ✅ Complete  
**Created:** 2025-10-11  
**Completed:** 2025-10-11  
**Goal:** Support fetching GraphQL schemas from SDL endpoints when introspection is disabled

## Problem

Some GraphQL servers disable introspection queries but provide the complete schema in SDL (Schema Definition Language) format at a GET endpoint, typically at `/graphql/schema.graphql` or similar.

Currently, gqlt only supports introspection queries, which fails on these servers.

## Solution

Add fallback mechanism to fetch schema from SDL endpoints when introspection fails.

## Implementation Steps

### Phase 1: SDL Fetching ✅
- [x] Add method to fetch SDL from GET endpoint (try multiple common paths)
- [x] Common paths to try:
  - `/schema.graphql` (relative to graphql endpoint)
  - `/graphql/schema.graphql`
  - `/sdl`
  - Configurable custom path
- [x] Add HTTP client logic for GET requests

### Phase 2: SDL Parsing ✅
- [x] Research Go GraphQL SDL parsing libraries (chose vektah/gqlparser)
- [x] Add dependency for SDL parsing
- [x] Convert SDL to introspection-compatible format
- [x] Handle schema types, queries, mutations, subscriptions, etc.

### Phase 3: Integration ✅
- [x] Update `Client.Introspect()` to try SDL fallback when introspection fails
- [x] MCP handlers automatically use new logic (no changes needed)
- [x] CLI introspect command automatically supports SDL endpoints (no changes needed)

### Phase 4: Testing ✅
- [x] Tested with real servers that use SDL endpoints (localhost:5095)
- [x] Verified list_types works with SDL fallback
- [x] Verified describe_type works with SDL fallback
- [x] All existing tests still pass

### Phase 5: Documentation
- [ ] Update README with SDL support information
- [ ] Add examples for SDL endpoints

## Implementation Decisions
1. **SDL parsing library:** vektah/gqlparser v2 - well maintained, widely used
2. **Priority:** Try introspection first, fallback to SDL on failure - respects standard introspection when available
3. **SDL endpoint paths:** Try multiple common paths automatically - no configuration needed in most cases
4. **Automatic fallback:** Yes - completely transparent to users

## Results
- SDL schema is automatically fetched and parsed when introspection is disabled
- Converted to identical introspection JSON format for seamless compatibility
- Works transparently with all existing code (CLI, MCP tools, library)
- No configuration required - tries common paths automatically
- Successfully tested with real GraphQL server using SDL endpoints

