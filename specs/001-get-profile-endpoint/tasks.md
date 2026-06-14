# Tasks: Get Profile Endpoint

**Feature**: `GET /api/profiles/:username`  
**Spec**: `specs/001-get-profile-endpoint/spec.md`  
**Plan**: `specs/001-get-profile-endpoint/plan.md`  
**Constitution**: `.specify/memory/constitution.md` (v1.0.3)

---

## Phase 1: Setup

- [x] **1.1** Project structure exists (cmd/, internal/, tests/)
- [x] **1.2** Go modules initialized (go.mod present)
- [x] **1.3** Gin framework available (v1.11.0 in go.mod)

---

## Phase 2: Foundational

- [x] **2.1** Database connection exists (check cmd/main.go or similar)
- [x] **2.2** PostgreSQL driver available
- [x] **2.3** Existing handlers pattern in `cmd/handlers.go`

---

## Phase 3: User Story 1 - View User Profile

### Backend Implementation

- [ ] **3.1** Create handler function `GetProfile` in `cmd/handlers.go`
  - Accept `gin.Context` parameter
  - Extract `username` from URL path parameter
  - Extract JWT token from `Authorization: Token <token>` header (optional)
  - Call service/repository layer to fetch profile data
  - Return JSON response with profile object

- [ ] **3.2** Create database query in `internal/repository/profile_repository.go` (or similar)
  - SQL query to fetch user by username: `id, username, bio, image`
  - LEFT JOIN with `followers` table to check following status
  - Return `following: false` if no JWT token provided
  - Return `following: true/false` based on follower relationship if JWT valid

- [ ] **3.3** Create profile model/struct in `internal/model/profile.go` (or similar)
  - Fields: `Username`, `Bio`, `Image`, `Following`
  - JSON tags matching RealWorld API spec

- [ ] **3.4** Register route in `internal/router/BlogAPIRouter.go`
  - Add `GET /api/profiles/:username` route
  - Map to `GetProfile` handler

### Integration Tests

- [ ] **3.5** Create test file `tests/api_tests/profile_test.go`
  - Set up test database connection
  - Create test helper functions for API requests

- [ ] **3.6** Test: Get existing user profile (unauthenticated)
  - Seed test user in database
  - Make GET request without Authorization header
  - Assert 200 OK response
  - Assert profile data matches seeded user
  - Assert `following: false`

- [ ] **3.7** Test: Get existing user profile (authenticated, not following)
  - Seed test user and authenticated user
  - Make GET request with valid JWT token
  - Assert 200 OK response
  - Assert `following: false` (no follower relationship)

- [ ] **3.8** Test: Get existing user profile (authenticated, following)
  - Seed test user, authenticated user, and follower relationship
  - Make GET request with valid JWT token
  - Assert 200 OK response
  - Assert `following: true`

---

## Phase 4: User Story 2 - Handle Non-existent User

### Backend Implementation

- [ ] **4.1** Add error handling in `GetProfile` handler
  - Check if user exists in database
  - Return 404 with RealWorld error format if not found: `{"errors": {"body": ["profile not found"]}}`

### Integration Tests

- [ ] **4.2** Test: Get non-existent user profile
  - Make GET request with username that doesn't exist
  - Assert 404 Not Found response
  - Assert error body matches RealWorld format

---

## Phase 5: User Story 3 - View Profile with Null Image

### Backend Implementation

- [ ] **5.1** Handle null image in response
  - Ensure `image` field returns `null` (not empty string) when user has no image
  - Use pointer type or omitempty with proper JSON marshaling

### Integration Tests

- [ ] **5.2** Test: Get profile with null image
  - Seed user without image (NULL in database)
  - Make GET request
  - Assert 200 OK response
  - Assert `image` field is `null` in JSON response

---

## Phase 6: Polish & Cross-Cutting Concerns

- [ ] **6.1** Add input validation
  - Username format validation (alphanumeric, underscores, etc.)
  - Return 422 if username is empty or invalid format

- [ ] **6.2** Add logging
  - Log profile access attempts
  - Log errors with appropriate context

- [ ] **6.3** Run all integration tests
  - Execute `go test ./tests/api_tests/...` or appropriate test command
  - Verify all tests pass

- [ ] **6.4** Manual validation
  - Test with curl commands from `quickstart.md`
  - Verify response format matches RealWorld spec

- [ ] **6.5** Code review
  - Check adherence to constitution principles
  - Verify idiomatic Go patterns
  - Ensure error handling follows project conventions

---

## Task Execution Order

1. **Start with Phase 3** (main functionality)
2. **Then Phase 4** (error handling)
3. **Then Phase 5** (null handling)
4. **Finally Phase 6** (polish & validation)

**Note**: Phases 1 & 2 are marked complete as infrastructure already exists. Verify during implementation.

---

## Test Commands

```bash
# Run integration tests
go test ./tests/api_tests/... -v

# Run specific profile tests
go test ./tests/api_tests/profile_test.go -v

# Manual curl test (see quickstart.md for full examples)
curl http://localhost:8080/api/profiles/testuser
```

---

## Definition of Done

- [ ] All user stories implemented
- [ ] All integration tests passing
- [ ] Manual validation completed
- [ ] Code follows constitution principles
- [ ] No linting errors
- [ ] Ready for merge to `gin` branch
