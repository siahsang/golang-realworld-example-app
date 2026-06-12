## MODIFIED Requirements

### Requirement: Follow User Endpoint

The system SHALL allow authenticated users to follow other users by username via POST /api/profiles/:username/follow. Users SHALL NOT be allowed to follow themselves.

#### Scenario: Successfully follow a user
- **WHEN** authenticated user sends POST request to /api/profiles/:username/follow with valid Authorization header for a different user
- **THEN** system returns 200 OK with profile object containing following: true

#### Scenario: Follow without authentication
- **WHEN** user sends POST request without Authorization header
- **THEN** system returns 401 Unauthorized with error message

#### Scenario: Follow with invalid token
- **WHEN** user sends POST request with expired or invalid token
- **THEN** system returns 401 Unauthorized with error message

#### Scenario: Follow non-existent user
- **WHEN** authenticated user tries to follow a username that doesn't exist
- **THEN** system returns 404 Not Found with error message

#### Scenario: Follow already followed user
- **WHEN** authenticated user tries to follow a user they already follow
- **THEN** system returns 400 Bad Request with error message

#### Scenario: Follow self
- **WHEN** authenticated user tries to follow their own username
- **THEN** system returns 400 Bad Request with error message "Cannot follow yourself"
