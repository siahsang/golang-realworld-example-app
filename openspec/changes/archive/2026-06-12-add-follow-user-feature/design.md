## Context

The blog application currently has a profile system with optional authentication for viewing profiles. The core follow/unfollow functionality exists in the Core layer (`internal/core/profiles.go`) with `FollowUser()` and `UnfollowUser()` methods, and the database schema includes a `followers` table. However, there are no API endpoints to expose this functionality to clients.

The existing architecture uses Gin for HTTP routing with:
- Optional auth middleware for public endpoints (e.g., viewing profiles)
- Controller pattern (UserController) for handling requests
- Core layer for business logic
- Existing error handling with simple format: `{"errors": {"field": "message"}}`

## Goals / Non-Goals

**Goals:**
- Expose follow/unfollow functionality through REST API endpoints
- Require authentication for follow operations (users must be logged in)
- Follow existing architectural patterns (Controller → Core → Database)
- Return profile with updated following status
- Handle errors with appropriate HTTP status codes (401, 400, 404)

**Non-Goals:**
- Modifying the existing followers table schema
- Fixing the bug in core/profiles.go where wrong profile is returned
- Adding follower count or following count to profiles
- Implementing notifications for new followers
- Adding pagination for follower/following lists

## Decisions

### 1. Route Structure: `/api/profiles/:username/follow`

**Decision:** Use nested route under profiles resource

**Rationale:**
- Follows RealWorld API specification convention
- Clear resource hierarchy (following is an action on a profile)
- Consistent with existing `/api/profiles/:username` endpoint
- Alternative considered: `/api/users/:username/follow` - rejected because profile is the public-facing resource

### 2. HTTP Methods: POST for follow, DELETE for unfollow

**Decision:** Use POST to follow, DELETE to unfollow

**Rationale:**
- POST creates a new follower relationship (idempotent in practice due to DB constraint)
- DELETE removes the follower relationship
- Alternative considered: PUT/PATCH with boolean flag - rejected because separate endpoints are more RESTful and clearer
- Alternative considered: Using PUT for both - rejected because POST/DELETE better expresses intent

### 3. Required Authentication Middleware

**Decision:** Create new `RequiredAuthMiddleware()` separate from existing `OptionalAuthMiddleware()`

**Rationale:**
- Follow/unfollow requires user to be authenticated (unlike viewing profiles)
- Existing optional auth silently ignores invalid tokens - not suitable for protected routes
- New middleware will return 401 for missing/invalid tokens
- Keeps authentication concerns separated and reusable for future protected routes

### 4. Error Response Format

**Decision:** Use simple error format: `{"errors": {"field": "message"}}`

**Rationale:**
- Matches existing application error handling pattern
- Simpler than array format for single errors
- Consistent with RealWorld specification examples
- Already implemented in handler.ErrorResponse()

### 5. Response Payload

**Decision:** Return full profile object with updated following status

**Rationale:**
- Consistent with existing GET /api/profiles/:username response
- Client receives immediate feedback on following status
- No need for separate profile fetch after follow operation
- Matches RealWorld API specification

## Risks / Trade-offs

**Risk:** Duplicate follow requests could cause race conditions
**Mitigation:** Database unique constraint on (user_id, follower_id) prevents duplicates, returns error on second insert

**Risk:** New middleware adds complexity to server setup
**Mitigation:** Middleware follows same pattern as existing OptionalAuthMiddleware, minimal code duplication

**Risk:** Protected route group registration could be forgotten for future endpoints
**Mitigation:** Clear code structure with separate `RegisterAuthAPIRouter()` method makes intent explicit

**Trade-off:** Not fixing the profile return bug in core/profiles.go
**Reason:** Out of scope for this change - will be addressed in separate bug fix to keep changes atomic

**Trade-off:** Using POST instead of PUT for follow
**Reason:** POST better expresses "create relationship" semantics, though PUT would also be valid for idempotent operations
