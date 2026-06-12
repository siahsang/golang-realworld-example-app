## Context

The follow user feature was recently implemented with POST/DELETE `/api/profiles/:username/follow` endpoints. The core logic in `internal/core/profiles.go` and controller in `internal/controller/user_controller.go` handle follow/unfollow operations, but neither layer validates whether a user is attempting to follow themselves.

Current flow:
```
UserController.Follow() 
  → Core.FollowUser() 
    → INSERT INTO followers (user_id, follower_id)
```

The database schema allows `(user_id, follower_id)` where both values are the same user ID, as there's no CHECK constraint preventing self-referential relationships.

Existing error handling pattern in the codebase:
- Controller validates input format and returns HTTP-appropriate status codes
- Core layer enforces business rules with custom error types
- Tests cover edge cases comprehensively (12 test cases for follow/unfollow)

## Goals / Non-Goals

**Goals:**
- Prevent users from following themselves at the API level
- Return 400 Bad Request with clear error message for self-follow attempts
- Add validation in both controller (early fail) and core (business rule enforcement)
- Add comprehensive test coverage for self-follow scenario
- Update spec to document this constraint

**Non-Goals:**
- Modifying the followers table schema or adding database constraints
- Changing existing follow/unfollow behavior for valid scenarios
- Adding validation for other edge cases (e.g., following bots, inactive users)
- Modifying the response format for successful follow operations

## Decisions

### 1. Validation Location: Core + Controller (Defense in Depth)

**Decision:** Add validation in both layers

**Rationale:**
- **Controller layer**: Early validation provides fast feedback with clear HTTP 400 semantics
- **Core layer**: Business rule enforcement ensures validation works regardless of call source (controller, CLI, background jobs)
- Follows existing pattern in codebase where controller handles HTTP concerns and core handles domain logic

**Alternatives considered:**
- Controller only: Rejected because business logic belongs in core, and FollowUser() could be called from other contexts
- Core only: Rejected because controller should provide early validation for clear HTTP error responses

### 2. Error Type: New Custom Error in Core

**Decision:** Add `CannotFollowSelf` error constant in `internal/core/profiles.go`

**Rationale:**
- Consistent with existing error handling pattern (`NoRecordFound`, `UserIsAlreadyFollowed`, `UserIsNotFollowed`)
- Allows controller to use `errors.Is()` for type-safe error handling
- Clear semantic meaning for future maintainers

**Error constant name:** `CannotFollowSelf`
**Error message:** "Cannot follow yourself"

### 3. HTTP Status Code: 400 Bad Request

**Decision:** Return 400 for self-follow attempts

**Rationale:**
- Self-follow is a client error (invalid request), not an authentication (401) or authorization (403) issue
- Consistent with existing validation errors in the follow endpoint
- Matches RealWorld API specification pattern for validation failures

### 4. Error Response Format: Simple Field-Based Format

**Decision:** Use existing simple format: `{"errors": {"body": "Cannot follow yourself"}}`

**Rationale:**
- Matches existing error response pattern in follow_test.go
- Consistent with other follow endpoint errors
- Simpler than array format for single validation errors

### 5. Validation Logic: Username Comparison

**Decision:** Compare authenticated user's username with target username

**Rationale:**
- Username is the API identifier used in the route parameter
- Case-sensitive comparison (consistent with existing username handling)
- No need to fetch user by ID first (username already available from route param and auth context)

**Implementation:**
```go
if authenticatedUser.Username == username {
    return 400 error
}
```

## Risks / Trade-offs

**Risk:** Duplicate validation logic between controller and core
**Mitigation:** Clear separation of concerns - controller handles HTTP semantics, core handles business rule. Comments in code will clarify intent.

**Risk:** Existing clients might rely on self-follow behavior (if any)
**Mitigation:** Self-follow is unlikely to be used intentionally; if discovered, can be addressed as a breaking change with versioning

**Risk:** Validation might not catch all edge cases (e.g., case-insensitive usernames)
**Mitigation:** Current implementation follows existing username comparison patterns in the codebase; can be enhanced if needed

**Trade-off:** Not adding database-level CHECK constraint
**Reason:** Application-level validation is sufficient for this use case; database constraint would add migration complexity and provide minimal additional value

**Trade-off:** Not refactoring existing error handling pattern
**Reason:** Current pattern works well; this change follows established conventions rather than introducing new patterns
