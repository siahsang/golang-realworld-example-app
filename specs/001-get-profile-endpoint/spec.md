# Feature Specification: Get Profile Endpoint

**Feature Branch**: `001-get-profile-endpoint`

**Created**: 2026-06-14

**Status**: Draft

**Input**: User description: "Get Profile - GET /api/profiles/:username - Authentication optional, returns a Profile"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View User Profile (Priority: P1)

As a blog reader, I want to view another user's profile so I can learn about their background and decide whether to follow them.

**Why this priority**: This is the core functionality of the endpoint - retrieving profile information. Without this, users cannot discover other authors on the platform.

**Independent Test**: Can be fully tested by making a GET request to `/api/profiles/:username` and verifying the response contains correct profile data.

**Acceptance Scenarios**:

1. **Given** a user "jake" exists with bio "I work at statefarm" and an image, **When** I request GET `/api/profiles/jake` without authentication, **Then** I receive status 200 with profile containing username "jake", bio "I work at statefarm", image URL, and following: false

2. **Given** a user "jake" exists, **When** I request GET `/api/profiles/jake` with a valid JWT token for a different user who is NOT following jake, **Then** I receive status 200 with following: false

3. **Given** a user "jake" exists, **When** I request GET `/api/profiles/jake` with a valid JWT token for a user who IS following jake, **Then** I receive status 200 with following: true

4. **Given** any user exists, **When** I request their profile WITHOUT providing authentication, **Then** the request succeeds (authentication is OPTIONAL, not required)

---

### User Story 2 - Handle Non-existent User (Priority: P2)

As a blog reader, I want to be informed when I try to view a profile that doesn't exist so I know the user is not on the platform.

**Why this priority**: Error handling is critical for user experience and API consistency. Users need clear feedback when requesting invalid resources.

**Independent Test**: Can be tested by requesting a username that doesn't exist and verifying the 404 response format.

**Acceptance Scenarios**:

1. **Given** no user with username "unknown" exists, **When** I request GET `/api/profiles/unknown`, **Then** I receive status 404 with error message "User not found" in the standard RealWorld error format

---

### User Story 3 - View Profile with Null Image (Priority: P3)

As a blog reader, I want to view profiles of users who haven't uploaded an image so I can still access their bio and other information.

**Why this priority**: Not all users upload profile images; the system must gracefully handle missing images without breaking the profile display.

**Independent Test**: Can be tested by requesting a profile for a user who has no image uploaded and verifying the response contains null for the image field.

**Acceptance Scenarios**:

1. **Given** a user "noimage" exists with bio "Hello" but no profile image, **When** I request GET `/api/profiles/noimage`, **Then** I receive status 200 with image: null in the response

---

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept GET requests to `/api/profiles/:username` endpoint
- **FR-002**: System MUST return HTTP status 200 with profile data when the requested user exists
- **FR-003**: System MUST return HTTP status 404 with error message "User not found" when the requested user does not exist
- **FR-004**: System MUST return profile with username matching the requested :username parameter
- **FR-005**: System MUST return profile with bio field (plain text, max 500 characters)
- **FR-006**: System MUST return profile with image field (URL string or null if not set)
- **FR-007**: System MUST return profile with following field as boolean
- **FR-008**: System MUST return following: false when request is made without authentication
- **FR-009**: System MUST calculate and return actual following status when request includes valid JWT authentication
- **FR-010**: System MUST return error response in format `{"errors": {"body": ["error message"]}}` for validation errors
- **FR-011**: System MUST NOT require authentication - the endpoint MUST work with or without JWT token

### Key Entities

- **Profile**: Represents a user's public profile information containing username, bio, image URL, and following status
- **User**: Represents a registered user account in the system (source of profile data)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can retrieve any existing profile with a single API call
- **SC-002**: 100% of profile requests return correct username matching the requested :username parameter
- **SC-003**: 100% of requests for non-existent users return 404 status code
- **SC-004**: Following status is accurate for authenticated requests (verified by integration tests)
- **SC-005**: Response time for profile retrieval is under 200ms for 95th percentile requests

## Assumptions

- Users have already registered accounts in the system before profiles can be retrieved
- The username in the URL path is case-sensitive (exact match required)
- Profile images are hosted externally (URLs stored, not files)
- Bio content is stored as plain text without markdown or HTML formatting
- JWT authentication uses the `Authorization: Token <token>` header format as per RealWorld spec
- The following relationship is unidirectional (user A can follow user B without reciprocity)
- **Authentication is OPTIONAL**: Clients can call this endpoint with or without providing a JWT token
