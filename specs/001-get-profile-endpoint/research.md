# Research: Get Profile Endpoint

**Created**: 2026-06-14  
**Feature**: Get Profile Endpoint  
**Spec**: [spec.md](spec.md)  
**Plan**: [plan.md](plan.md)

## Technical Decisions

### Decision 1: Endpoint Location and Routing

**What was chosen**: Implement handler in `cmd/handlers.go` following existing pattern, register route in `internal/router/BlogAPIRouter.go`

**Rationale**: 
- Existing handlers (users, articles, comments) are in `cmd/handlers.go`
- Router pattern already established in `internal/router/`
- Maintains consistency with current codebase structure
- Minimizes refactoring overhead

**Alternatives considered**:
- Create separate `cmd/profiles.go` file: Rejected - would fragment handler logic unnecessarily for a single endpoint
- Create `internal/handler/profile.go`: Rejected - current project uses `cmd/` for all handlers

---

### Decision 2: Authentication Handling

**What was chosen**: Extract JWT token if present in `Authorization: Token <token>` header, validate silently, return `following: false` if no token or invalid token

**Rationale**:
- Authentication is OPTIONAL per RealWorld spec
- Invalid/missing token should not cause 401 error (unlike protected endpoints)
- Following status calculation requires knowing current authenticated user
- Silent validation provides better UX - public profiles are always accessible

**Implementation pattern**:
```go
// Pseudocode - actual implementation in tasks phase
token := extractTokenFromHeader(request)
if token != "" && validateToken(token) {
    currentUser := getUserFromToken(token)
    following := checkFollowingStatus(currentUser, profileUser)
    return Profile{..., Following: following}
}
return Profile{..., Following: false}
```

**Alternatives considered**:
- Require authentication and return 401 if missing: Rejected - violates RealWorld spec requirement for optional auth
- Always return `following: false`: Rejected - authenticated users should see accurate following status

---

### Decision 3: Database Query Strategy

**What was chosen**: Single SQL query with LEFT JOIN to fetch user profile and following status in one call

**Rationale**:
- Minimizes database round trips
- LEFT JOIN handles both authenticated and unauthenticated cases
- Existing `users` and `followers` tables support this pattern
- Better performance than separate queries

**Query pattern**:
```sql
-- For authenticated users
SELECT u.username, u.bio, u.image, 
       CASE WHEN f.following_id IS NOT NULL THEN true ELSE false END as following
FROM users u
LEFT JOIN followers f ON f.followed_id = u.id AND f.following_id = $1
WHERE u.username = $2

-- For unauthenticated (simpler, no following check)
SELECT username, bio, image, false as following
FROM users
WHERE username = $1
```

**Alternatives considered**:
- Separate queries for profile and following status: Rejected - N+1 query problem, slower
- Load user first, then check following in application code: Rejected - more network round trips

---

### Decision 4: Error Response Format

**What was chosen**: Standard RealWorld format: `{"errors": {"body": ["User not found"]}}` with 404 status

**Rationale**:
- Constitution Principle II (API-First) mandates RealWorld API compliance
- Consistent with other error responses in the application
- Clear error message for debugging
- Proper HTTP semantics (404 for resource not found)

**Alternatives considered**:
- Return `{"error": "User not found"}`: Rejected - inconsistent with RealWorld spec
- Return `{"errors": {"username": ["User not found"]}}`: Rejected - spec uses "body" key for general errors

---

### Decision 5: Profile Image Null Handling

**What was chosen**: Return `null` (JSON null) when user has no profile image

**Rationale**:
- User clarified this choice during specification phase
- Explicitly indicates absence of value
- Frontend can handle null gracefully with placeholder image
- Standard JSON practice for optional fields

**Alternatives considered**:
- Return empty string `""`: Rejected - can cause broken image links
- Return default image URL: Rejected - frontend should control placeholder logic
- Omit field entirely: Rejected - inconsistent response structure

---

## RealWorld API Contract Reference

**Endpoint**: `GET /api/profiles/:username`

**Authentication**: Optional (JWT token in `Authorization: Token <token>` header)

**Success Response (200)**:
```json
{
  "profile": {
    "username": "jake",
    "bio": "I work at statefarm",
    "image": "https://api.realworld.io/images/smiley-cyrus.jpg",
    "following": false
  }
}
```

**Error Response (404)**:
```json
{
  "errors": {
    "body": ["User not found"]
  }
}
```

---

## Existing Codebase Analysis

### Current Handler Pattern (from `cmd/handlers.go`)

- Handlers receive `*gin.Context`
- Extract authentication via middleware
- Use `core` package for business logic
- Return JSON responses with appropriate status codes
- Error handling with `c.JSON()` and proper status codes

### Current Router Pattern (from `internal/router/`)

- `BlogAPIRouter` sets up API routes under `/api` prefix
- Uses Gin routing with path parameters (`:username`)
- Middleware chain includes logging, authentication (where needed)

### Database Schema (existing)

- `users` table: `id`, `username`, `email`, `password`, `bio`, `image`, `created_at`, `updated_at`
- `followers` table: `follower_id`, `followed_id` (many-to-many relationship)

---

## Best Practices Applied

### Go HTTP Handler Patterns

1. **Early returns**: Handle errors immediately, keep happy path left-aligned
2. **Context usage**: Pass `context.Context` for database operations
3. **Error wrapping**: Use `fmt.Errorf` with `%w` for error context
4. **Structured logging**: Log request details and errors with context

### Database Access

1. **Parameterized queries**: Prevent SQL injection
2. **Connection pooling**: Use existing `*sql.DB` pool
3. **Proper cleanup**: `defer rows.Close()` after queries
4. **Transaction not needed**: Read-only operation

### Testing Approach

1. **Integration tests**: Full stack testing with test database
2. **Table-driven tests**: Multiple test cases in single test function
3. **Test isolation**: Truncate tables between tests
4. **Authentication scenarios**: Test both authenticated and unauthenticated paths

---

## Clarifications Resolved

All `NEEDS CLARIFICATION` items from the plan have been resolved through research and user input:

✅ Authentication behavior: Optional, returns `following: false` when unauthenticated, calculates actual status when authenticated  
✅ Profile image handling: Return `null` if not uploaded  
✅ Bio field format: Plain text only, max 500 characters  
✅ Error response format: Standard RealWorld format with 404 status

---

## Next Steps

Proceed to Phase 1: Design & Contracts
- Create `data-model.md` documenting Profile entity and relationships
- Create `contracts/` directory with API contract specification
- Create `quickstart.md` with validation commands
- Update agent context in `AGENTS.md`
